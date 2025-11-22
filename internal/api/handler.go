package api

import (
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/scheduler"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/task"
	"DBDataGenerator/internal/websocket"
)

// Handler API 处理器
type Handler struct {
	connMgr        database.ConnectionManagerInterface
	taskManager    *task.Manager
	templateMgr    generator.TemplateManagerInterface
	scheduler      *scheduler.Scheduler
	wsHub          *websocket.Hub
	historyManager *task.HistoryManager
	storage        storage.StorageInterface
	logger         *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(connMgr database.ConnectionManagerInterface, templateMgr generator.TemplateManagerInterface, wsHub *websocket.Hub, storageInstance storage.StorageInterface, logger *zap.Logger) *Handler {
	manager := task.NewManager(connMgr)
	manager.SetWebSocketHub(wsHub) // 设置WebSocket Hub

	// 创建历史管理器（仅当存储支持数据库时）
	var historyManager *task.HistoryManager
	if storageInstance.GetDB() != nil {
		historyManager = task.NewHistoryManager(storageInstance)
		manager.SetHistoryManager(historyManager)

		// 设置任务持久化
		persistence := task.NewTaskPersistence(storageInstance)
		manager.SetPersistence(persistence)

		// 从数据库恢复任务
		if err := manager.LoadTasks(); err != nil {
			logger.Warn("恢复任务失败", zap.Error(err))
		}
	}

	// 创建定时任务调度器
	sched := scheduler.NewScheduler(manager, logger)

	return &Handler{
		connMgr:        connMgr,
		taskManager:    manager,
		templateMgr:    templateMgr,
		scheduler:      sched,
		wsHub:          wsHub,
		historyManager: historyManager,
		storage:        storageInstance,
		logger:         logger,
	}
}
