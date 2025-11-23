package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
)

// TaskUpdateCallback 任务更新回调函数类型
type TaskUpdateCallback func(task *Task)

// WorkerPool 协程池
type WorkerPool struct {
	taskID      string
	threadCount int
	workers     []*Worker
	taskChan    chan *WorkItem
	resultChan  chan *WorkResult
	ctx         context.Context
	cancel      context.CancelFunc
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	isPaused    bool
	mu          sync.RWMutex
	db          database.Database
	engine      *generator.Engine
	config      *generator.TableConfig
	task        *Task
	onUpdate    TaskUpdateCallback // 任务更新回调
	onComplete  TaskUpdateCallback // 任务完成回调
	stopWorkers []chan struct{}    // 用于停止特定 worker 的通道
}

// WorkItem 工作项
type WorkItem struct {
	BatchSize  int
	StartIndex int64
}

// WorkResult 工作结果
type WorkResult struct {
	SuccessCount int64
	FailedCount  int64
	Error        error
}

// Worker 工作协程
type Worker struct {
	id       int
	pool     *WorkerPool
	stopChan chan struct{} // 用于接收停止信号
}

// NewWorkerPool 创建协程池
func NewWorkerPool(taskID string, threadCount int, db database.Database, engine *generator.Engine, config *generator.TableConfig, task *Task) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		taskID:      taskID,
		threadCount: threadCount,
		taskChan:    make(chan *WorkItem, DefaultChannelBufferSize),
		resultChan:  make(chan *WorkResult, DefaultChannelBufferSize),
		ctx:         ctx,
		cancel:      cancel,
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
		db:          db,
		engine:      engine,
		config:      config,
		task:        task,
		onUpdate:    nil,
		onComplete:  nil,
		stopWorkers: make([]chan struct{}, 0),
	}
}

// SetUpdateCallback 设置任务更新回调
func (p *WorkerPool) SetUpdateCallback(callback TaskUpdateCallback) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onUpdate = callback
}

// SetCompleteCallback 设置任务完成回调
func (p *WorkerPool) SetCompleteCallback(callback TaskUpdateCallback) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onComplete = callback
}

// Start 启动协程池
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 开始任务级别的事务
	if err := p.db.BeginTransaction(); err != nil {
		// 如果开始事务失败，记录错误但不阻止任务启动
		// 任务会继续执行，但使用批次级别的事务（向后兼容）
		_ = err
	}

	p.workers = make([]*Worker, p.threadCount)
	p.stopWorkers = make([]chan struct{}, p.threadCount)
	for i := 0; i < p.threadCount; i++ {
		stopChan := make(chan struct{})
		worker := &Worker{
			id:       i,
			pool:     p,
			stopChan: stopChan,
		}
		p.workers[i] = worker
		p.stopWorkers[i] = stopChan
		go worker.run()
	}

	// 启动结果收集协程
	go p.collectResults()
}

// Stop 停止协程池（带超时）
func (p *WorkerPool) Stop() {
	p.StopWithTimeout(5 * time.Second)
}

// StopWithTimeout 停止协程池（带超时）
func (p *WorkerPool) StopWithTimeout(timeout time.Duration) {
	// 1. 取消 context，通知所有 worker 停止
	p.cancel()

	// 2. 回滚事务（任务被停止，不应该提交）
	_ = p.db.RollbackTransaction()

	// 3. 关闭通道，防止新的工作项被添加
	p.mu.Lock()
	close(p.taskChan)
	close(p.resultChan)

	// 3. 向所有 worker 发送停止信号
	for _, stopChan := range p.stopWorkers {
		select {
		case stopChan <- struct{}{}:
		default:
			// 如果通道已满或已关闭，直接关闭通道
			close(stopChan)
		}
	}
	p.mu.Unlock()

	// 4. 等待所有 worker 退出（带超时）
	// 由于无法直接等待 goroutine 退出，我们使用超时机制
	// 如果超时，强制关闭所有停止通道
	select {
	case <-time.After(timeout):
		// 超时，强制清理
		p.mu.Lock()
		// 关闭所有 worker 的停止通道
		for _, stopChan := range p.stopWorkers {
			select {
			case <-stopChan:
				// 已经关闭
			default:
				close(stopChan)
			}
		}
		p.mu.Unlock()
	}
}

// Pause 暂停
func (p *WorkerPool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isPaused {
		p.isPaused = true
		close(p.pauseChan)
		p.pauseChan = make(chan struct{})
	}
}

// Resume 恢复
func (p *WorkerPool) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isPaused {
		p.isPaused = false
		close(p.resumeChan)
		p.resumeChan = make(chan struct{})
	}
}

// SetThreadCount 设置线程数
func (p *WorkerPool) SetThreadCount(count int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if count > p.threadCount {
		// 增加工作协程
		for i := p.threadCount; i < count; i++ {
			stopChan := make(chan struct{})
			worker := &Worker{
				id:       i,
				pool:     p,
				stopChan: stopChan,
			}
			p.workers = append(p.workers, worker)
			p.stopWorkers = append(p.stopWorkers, stopChan)
			go worker.run()
		}
	} else if count < p.threadCount {
		// 减少工作协程：优雅停止多余的 worker
		excessCount := p.threadCount - count
		if excessCount > 0 && excessCount <= len(p.workers) {
			// 从末尾开始停止多余的 worker
			startIdx := len(p.workers) - excessCount
			for i := startIdx; i < len(p.workers); i++ {
				// 发送停止信号
				select {
				case p.stopWorkers[i] <- struct{}{}:
				default:
					// 如果通道已关闭或阻塞，直接关闭通道
					close(p.stopWorkers[i])
				}
			}
			// 从列表中移除
			p.workers = p.workers[:startIdx]
			p.stopWorkers = p.stopWorkers[:startIdx]
		}
	}
	p.threadCount = count
}

// AddWork 添加工作项
func (p *WorkerPool) AddWork(item *WorkItem) {
	select {
	case p.taskChan <- item:
	case <-p.ctx.Done():
		return
	}
}

// run 工作协程运行
func (w *Worker) run() {
	for {
		select {
		case <-w.pool.ctx.Done():
			return
		case <-w.stopChan:
			// 收到停止信号，优雅退出
			return
		case item, ok := <-w.pool.taskChan:
			if !ok {
				return
			}

			// 检查是否暂停
			w.pool.mu.RLock()
			if w.pool.isPaused {
				w.pool.mu.RUnlock()
				select {
				case <-w.pool.resumeChan:
					w.pool.mu.RLock()
					w.pool.mu.RUnlock()
				case <-w.pool.ctx.Done():
					return
				case <-w.stopChan:
					return
				}
			} else {
				w.pool.mu.RUnlock()
			}

			// 执行工作
			result := w.execute(item)
			select {
			case w.pool.resultChan <- result:
			case <-w.pool.ctx.Done():
				return
			case <-w.stopChan:
				return
			}
		}
	}
}

// execute 执行工作项
func (w *Worker) execute(item *WorkItem) *WorkResult {
	// 检查是否已取消
	select {
	case <-w.pool.ctx.Done():
		return &WorkResult{
			FailedCount: int64(item.BatchSize),
			Error:       w.pool.ctx.Err(),
		}
	case <-w.stopChan:
		return &WorkResult{
			FailedCount: int64(item.BatchSize),
			Error:       fmt.Errorf("任务已停止"),
		}
	default:
	}

	// 生成数据
	rows, err := w.pool.engine.GenerateBatch(w.pool.ctx, w.pool.config, item.StartIndex, item.BatchSize)
	if err != nil {
		return &WorkResult{
			FailedCount: int64(item.BatchSize),
			Error:       err,
		}
	}

	// 再次检查是否已取消（生成数据可能耗时）
	select {
	case <-w.pool.ctx.Done():
		return &WorkResult{
			FailedCount: int64(len(rows)),
			Error:       w.pool.ctx.Err(),
		}
	case <-w.stopChan:
		return &WorkResult{
			FailedCount: int64(len(rows)),
			Error:       fmt.Errorf("任务已停止"),
		}
	default:
	}

	// 插入数据库（这里可能阻塞）
	// 使用 goroutine 执行数据库操作，以便能够响应停止信号
	insertDone := make(chan error, 1)
	go func() {
		// 使用事务中的批量插入（任务级别事务）
		insertDone <- w.pool.db.BatchInsertInTransaction(w.pool.config.Database, w.pool.config.TableName, rows)
	}()

	// 等待数据库操作完成或收到停止信号
	select {
	case <-w.pool.ctx.Done():
		// Context 已取消，返回失败结果
		return &WorkResult{
			FailedCount: int64(len(rows)),
			Error:       w.pool.ctx.Err(),
		}
	case <-w.stopChan:
		// 收到停止信号，返回失败结果
		return &WorkResult{
			FailedCount: int64(len(rows)),
			Error:       fmt.Errorf("任务已停止"),
		}
	case err := <-insertDone:
		// 数据库操作完成
		if err != nil {
			// 再次检查是否是因为取消导致的错误
			select {
			case <-w.pool.ctx.Done():
				return &WorkResult{
					FailedCount: int64(len(rows)),
					Error:       w.pool.ctx.Err(),
				}
			case <-w.stopChan:
				return &WorkResult{
					FailedCount: int64(len(rows)),
					Error:       fmt.Errorf("任务已停止"),
				}
			default:
				return &WorkResult{
					SuccessCount: 0,
					FailedCount:  int64(len(rows)),
					Error:        err,
				}
			}
		}
	}

	return &WorkResult{
		SuccessCount: int64(len(rows)),
		FailedCount:  0,
	}
}

// collectResults 收集结果
func (p *WorkerPool) collectResults() {
	var totalGenerated, totalSuccess, totalFailed int64
	startTime := time.Now()

	for {
		select {
		case <-p.ctx.Done():
			return
		case result, ok := <-p.resultChan:
			if !ok {
				return
			}

			totalGenerated += result.SuccessCount + result.FailedCount
			totalSuccess += result.SuccessCount
			totalFailed += result.FailedCount

			// 如果结果有错误，记录错误信息
			if result.Error != nil {
				p.task.mu.Lock()
				p.task.Error = result.Error.Error()
				p.task.mu.Unlock()
			}

			// 更新任务进度
			elapsed := time.Since(startTime)
			speed := float64(totalGenerated) / elapsed.Seconds()
			remaining := p.config.TotalRows - totalGenerated
			eta := time.Duration(float64(remaining)/speed) * time.Second

			p.task.UpdateProgress(totalGenerated, totalSuccess, totalFailed)
			p.task.UpdateSpeed(speed, eta)

			// 通过回调通知任务管理器更新
			p.mu.RLock()
			onUpdate := p.onUpdate
			p.mu.RUnlock()
			if onUpdate != nil {
				onUpdate(p.task)
			}

			// 如果连续失败过多，标记任务为错误状态
			if totalFailed > 0 && totalSuccess == 0 && totalGenerated >= 100 {
				// 如果前100条全部失败，标记为错误
				p.task.SetStatus(TaskStatusError)
				now := time.Now()
				p.task.EndTime = &now

				// 回滚事务（任务失败）
				_ = p.db.RollbackTransaction()

				p.mu.RLock()
				onComplete := p.onComplete
				p.mu.RUnlock()
				if onComplete != nil {
					onComplete(p.task)
				}
				return
			}

			// 检查是否完成（确保不超过总行数）
			if totalGenerated >= p.config.TotalRows {
				// 如果超过，调整到精确值
				if totalGenerated > p.config.TotalRows {
					excess := totalGenerated - p.config.TotalRows
					if totalSuccess > excess {
						totalSuccess -= excess
					} else {
						totalFailed += excess - totalSuccess
						totalSuccess = 0
					}
					totalGenerated = p.config.TotalRows
					p.task.UpdateProgress(totalGenerated, totalSuccess, totalFailed)
				}

				// 任务完成，提交事务
				if err := p.db.CommitTransaction(); err != nil {
					// 提交失败，标记为错误并回滚
					p.task.SetStatus(TaskStatusError)
					p.task.Error = fmt.Sprintf("提交事务失败: %v", err)
					_ = p.db.RollbackTransaction()
				} else {
					// 提交成功，标记为完成
					p.task.SetStatus(TaskStatusCompleted)
				}

				now := time.Now()
				p.task.EndTime = &now

				// 通过回调通知任务管理器任务完成
				p.mu.RLock()
				onComplete := p.onComplete
				p.mu.RUnlock()
				if onComplete != nil {
					onComplete(p.task)
				}
				return
			}
		}
	}
}
