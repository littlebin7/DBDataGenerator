package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, handler *Handler) {
	api := router.Group("/api")
	{
		// 连接管理
		api.POST("/connect/test", handler.TestConnection)
		api.POST("/connect", handler.Connect)
		api.GET("/connections", handler.GetConnections)
		api.GET("/connection/active", handler.GetActiveConnection)
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
	}
}
