package task

import (
	"context"
	"sync"
	"time"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
)

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
	id   int
	pool *WorkerPool
}

// NewWorkerPool 创建协程池
func NewWorkerPool(taskID string, threadCount int, db database.Database, engine *generator.Engine, config *generator.TableConfig, task *Task) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		taskID:      taskID,
		threadCount: threadCount,
		taskChan:    make(chan *WorkItem, 100),
		resultChan:  make(chan *WorkResult, 100),
		ctx:         ctx,
		cancel:      cancel,
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
		db:          db,
		engine:      engine,
		config:      config,
		task:        task,
	}
}

// Start 启动协程池
func (p *WorkerPool) Start() {
	p.workers = make([]*Worker, p.threadCount)
	for i := 0; i < p.threadCount; i++ {
		worker := &Worker{
			id:   i,
			pool: p,
		}
		p.workers[i] = worker
		go worker.run()
	}

	// 启动结果收集协程
	go p.collectResults()
}

// Stop 停止协程池
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.taskChan)
	close(p.resultChan)
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
			worker := &Worker{
				id:   i,
				pool: p,
			}
			p.workers = append(p.workers, worker)
			go worker.run()
		}
	} else if count < p.threadCount {
		// 减少工作协程（通过取消上下文）
		// 实际实现中可能需要更优雅的方式
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
			}
		}
	}
}

// execute 执行工作项
func (w *Worker) execute(item *WorkItem) *WorkResult {
	// 生成数据
	rows, err := w.pool.engine.GenerateBatch(w.pool.ctx, w.pool.config, item.StartIndex, item.BatchSize)
	if err != nil {
		return &WorkResult{
			FailedCount: int64(item.BatchSize),
			Error:       err,
		}
	}

	// 插入数据库
	err = w.pool.db.BatchInsert(w.pool.config.Database, w.pool.config.TableName, rows)
	if err != nil {
		return &WorkResult{
			SuccessCount: 0,
			FailedCount:  int64(len(rows)),
			Error:        err,
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

			// 更新任务进度
			elapsed := time.Since(startTime)
			speed := float64(totalGenerated) / elapsed.Seconds()
			remaining := p.config.TotalRows - totalGenerated
			eta := time.Duration(float64(remaining)/speed) * time.Second

			p.task.UpdateProgress(totalGenerated, totalSuccess, totalFailed)
			p.task.UpdateSpeed(speed, eta)

			// 检查是否完成
			if totalGenerated >= p.config.TotalRows {
				p.task.SetStatus(TaskStatusCompleted)
				now := time.Now()
				p.task.EndTime = &now
				return
			}
		}
	}
}
