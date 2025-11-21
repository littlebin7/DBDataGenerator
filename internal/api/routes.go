package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, handler *Handler) {
	// 添加 CORS 中间件
	router.Use(CORSMiddleware())

	// 添加审计日志中间件
	router.Use(AuditMiddleware(handler.logger))

	// 创建频率限制器（每分钟 100 个请求）
	rateLimiter := NewRateLimiter(100, time.Minute, handler.logger)

	api := router.Group("/api")
	// 对 API 路由应用频率限制（排除 WebSocket 升级）
	api.Use(rateLimiter.Limit())
	{
		// 连接管理
		api.POST("/connect/test", handler.TestConnection)
		api.POST("/connect", handler.Connect)
		api.GET("/connections", handler.GetConnections)
		api.GET("/connection/active", handler.GetActiveConnection) // 必须在 /connection/:id 之前
		api.PUT("/connection/:id", handler.UpdateConnection)
		api.POST("/connection/:id/switch", handler.SwitchConnection)
		api.DELETE("/connection/:id", handler.Disconnect)

		// 数据库操作
		api.GET("/databases", handler.GetDatabases)
		api.GET("/tables", handler.GetTables)
		api.GET("/table/:name/schema", handler.GetTableSchema)

		// 任务管理
		api.POST("/task/create", handler.CreateTask)
		api.GET("/tasks", handler.GetTasks)
		api.GET("/task/:id", handler.GetTask)
		api.POST("/task/:id/start", handler.StartTask)
		api.POST("/task/:id/pause", handler.PauseTask)
		api.POST("/task/:id/resume", handler.ResumeTask)
		api.POST("/task/:id/stop", handler.StopTask)
		api.DELETE("/task/:id", handler.DeleteTask)
		api.PUT("/task/:id/threads", handler.SetThreadCount)

		// 配置模板管理
		api.POST("/template/save", handler.SaveTemplate)
		api.GET("/templates", handler.GetTemplates)
		api.GET("/template/:id", handler.GetTemplate)
		api.DELETE("/template/:id", handler.DeleteTemplate)

		// 数据预览
		api.POST("/generator/preview", handler.PreviewData)

		// 任务历史
		api.GET("/tasks/history", handler.GetTaskHistory)
		api.GET("/task/:id/history", handler.GetTaskHistoryByTaskID)
		api.DELETE("/task/history/:id", handler.DeleteTaskHistory)

		// 数据导入导出
		api.GET("/export/:connection_id/:database/:table", handler.ExportData)
		api.POST("/import/:connection_id/:database/:table", handler.ImportData)

		// 定时任务
		api.POST("/task/:id/schedule", handler.ScheduleTask)
		api.GET("/task/:id/schedule", handler.GetSchedule)
		api.GET("/schedules", handler.GetAllSchedules)
		api.POST("/task/:id/schedule/enable", handler.EnableSchedule)
		api.POST("/task/:id/schedule/disable", handler.DisableSchedule)
		api.DELETE("/task/:id/schedule", handler.RemoveSchedule)

		// 表关系分析
		api.GET("/relations", handler.GetTableRelations)
		api.GET("/table/:name/relations", handler.GetTableRelationsByTable)
		api.POST("/cascade/generate", handler.GenerateCascade)

		// 预设模板
		api.GET("/presets", handler.GetPresetTemplates)
		api.GET("/preset/:id", handler.GetPresetTemplate)
		api.POST("/preset/apply", handler.ApplyPresetTemplate)

		// 批量任务管理
		api.POST("/tasks/batch/create", handler.BatchCreateTasks)
		api.POST("/tasks/batch/start", handler.BatchStartTasks)
		api.POST("/tasks/batch/stop", handler.BatchStopTasks)
		api.POST("/tasks/batch/delete", handler.BatchDeleteTasks)

		// 数据质量检查
		api.GET("/quality/check", handler.CheckDataQuality)

		// 性能监控
		api.GET("/monitor/metrics", handler.GetSystemMetrics)

		// 数据回滚
		api.POST("/task/:id/rollback", handler.RollbackTask)
		api.POST("/task/:id/rollback/partial", handler.RollbackPartial)
		api.GET("/task/:id/rollback", handler.GetRollbackRecord)

		// 任务复制
		api.POST("/task/:id/clone", handler.CloneTask)

		// 连接池管理
		api.GET("/pool/status", handler.GetPoolStatus)
	}
}
