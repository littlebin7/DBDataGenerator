package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"DBDataGenerator/internal/api"
	"DBDataGenerator/internal/config"
	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/websocket"
	"DBDataGenerator/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger := utils.InitLogger(cfg.Log.Level, cfg.Log.Output, cfg.Log.FilePath)
	defer logger.Sync()

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// 使用文件存储（简单、无需 CGO）
	// 连接配置保存在 connections.json
	connMgr := database.NewConnectionManagerFile("./connections.json")

	// 模板配置保存在 templates.json
	templateMgr := generator.NewTemplateManager("./templates.json")

	// 初始化 WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 初始化 API 处理器
	handler := api.NewHandler(connMgr, templateMgr, wsHub, logger)
	api.SetupRoutes(router, handler)

	// WebSocket 路由
	router.GET("/ws/task/:id", func(c *gin.Context) {
		wsHub.HandleWebSocket(c.Writer, c.Request)
	})

	// 静态文件服务（前端）
	// 先设置静态资源（JS、CSS、图片等）
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// SPA 路由：所有非 API、非 WebSocket、非静态资源的请求都返回 index.html
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是 API 或 WebSocket 请求，返回 404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		if len(path) >= 3 && path[:3] == "/ws" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		// 其他请求返回 index.html（支持前端路由）
		c.File("./web/dist/index.html")
	})

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info("服务器启动",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}
