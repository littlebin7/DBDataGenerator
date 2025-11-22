package main

import (
	"DBDataGenerator/internal/storage"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	// CORS 已在 api.SetupRoutes 中配置

	// 根据配置创建存储实例
	storageCfg := storage.ConvertConfig(cfg.Storage)
	storageInstance, err := storage.NewStorage(storageCfg)
	if err != nil {
		logger.Fatal("初始化存储失败", zap.Error(err))
	}
	defer storageInstance.Close()

	logger.Info("存储初始化成功", zap.String("type", storageInstance.Type()))

	// 尝试从 JSON 文件迁移数据（如果存在且不是文件存储）
	if storageInstance.Type() != "file" {
		if dbStorage, ok := storageInstance.(*storage.Storage); ok {
			if err := storage.MigrateFromJSON(dbStorage, "./connections.json", "./templates.json"); err != nil {
				logger.Warn("数据迁移失败（可能是首次运行）", zap.Error(err))
			} else {
				logger.Info("数据迁移完成（如果存在旧数据）")
			}
		}
	}

	// 根据存储类型创建连接管理器和模板管理器
	var connMgr database.ConnectionManagerInterface
	var templateMgr generator.TemplateManagerInterface

	if storageInstance.Type() == "file" {
		// 文件存储
		fileStorage := storageInstance.(*storage.FileStorage)
		connMgr, err = database.NewConnectionManagerFile(fileStorage)
		if err != nil {
			logger.Fatal("初始化连接管理器失败", zap.Error(err))
		}

		templateMgr, err = generator.NewTemplateManagerFile(fileStorage)
		if err != nil {
			logger.Fatal("初始化模板管理器失败", zap.Error(err))
		}
	} else {
		// 数据库存储（SQLite/MySQL/PostgreSQL）
		// 需要转换为 *storage.Storage 类型（或其他数据库存储类型）
		if sqliteStorage, ok := storageInstance.(*storage.Storage); ok {
			connMgr, err = database.NewConnectionManagerSQLite(sqliteStorage)
			if err != nil {
				logger.Fatal("初始化连接管理器失败", zap.Error(err))
			}

			templateMgr, err = generator.NewTemplateManagerSQLite(sqliteStorage)
			if err != nil {
				logger.Fatal("初始化模板管理器失败", zap.Error(err))
			}
		} else if mysqlStorage, ok := storageInstance.(*storage.MySQLStorage); ok {
			connMgr, err = database.NewConnectionManagerDB(mysqlStorage)
			if err != nil {
				logger.Fatal("初始化连接管理器失败", zap.Error(err))
			}

			templateMgr, err = generator.NewTemplateManagerDB(mysqlStorage)
			if err != nil {
				logger.Fatal("初始化模板管理器失败", zap.Error(err))
			}
		} else if postgresStorage, ok := storageInstance.(*storage.PostgresStorage); ok {
			connMgr, err = database.NewConnectionManagerDB(postgresStorage)
			if err != nil {
				logger.Fatal("初始化连接管理器失败", zap.Error(err))
			}

			templateMgr, err = generator.NewTemplateManagerDB(postgresStorage)
			if err != nil {
				logger.Fatal("初始化模板管理器失败", zap.Error(err))
			}
		} else {
			logger.Fatal("不支持的存储类型", zap.String("type", storageInstance.Type()))
		}
	}

	// 初始化 WebSocket Hub
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger) // 设置 logger
	go wsHub.Run()

	// 初始化 API 处理器
	handler := api.NewHandler(connMgr, templateMgr, wsHub, storageInstance, logger)
	api.SetupRoutes(router, handler)

	// WebSocket 路由
	router.GET("/ws/task/:id", func(c *gin.Context) {
		wsHub.HandleWebSocket(c.Writer, c.Request)
	})

	// 检查是否为开发模式（优先使用配置文件，环境变量作为备选）
	devMode := cfg.Server.DevMode
	if !devMode {
		// 如果配置文件中未启用，检查环境变量（向后兼容）
		devMode = os.Getenv("DEV_MODE") == "true" || os.Getenv("DEV") == "true"
	}

	viteDevServer := cfg.Server.ViteDevServer
	if viteDevServer == "" {
		// 如果配置文件中未设置，检查环境变量（向后兼容）
		viteDevServer = os.Getenv("VITE_DEV_SERVER")
		if viteDevServer == "" {
			viteDevServer = "http://localhost:5173" // 默认值
		}
	}

	indexPath := "./web/dist/index.html"
	hasFrontendFiles := false
	if _, err := os.Stat(indexPath); err == nil {
		hasFrontendFiles = true
		logger.Info("前端文件已找到", zap.String("path", indexPath))
	}

	// 开发模式：代理到 Vite 开发服务器
	if devMode {
		logger.Info("开发模式已启用，前端请求将代理到 Vite 开发服务器",
			zap.String("vite_server", viteDevServer),
			zap.String("hint", "请确保 Vite 开发服务器正在运行: cd web && npm run dev"),
		)

		// 创建反向代理
		viteURL, err := url.Parse(viteDevServer)
		if err != nil {
			logger.Fatal("解析 Vite 开发服务器地址失败", zap.Error(err))
		}
		proxy := httputil.NewSingleHostReverseProxy(viteURL)

		// 代理所有非 API、非 WebSocket 的请求到 Vite 开发服务器
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// API 请求不代理
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			// WebSocket 请求不代理
			if len(path) >= 3 && path[:3] == "/ws" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			// 其他请求代理到 Vite 开发服务器
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	} else {
		// 生产模式：使用静态文件
		if !hasFrontendFiles {
			logger.Warn("前端文件不存在，请先构建前端",
				zap.String("path", indexPath),
				zap.String("hint", "运行: cd web && npm run build"),
				zap.String("dev_mode_hint", "或设置环境变量 DEV_MODE=true 启用开发模式"),
			)
		}

		// 先设置静态资源（JS、CSS、图片等）
		router.Static("/assets", "./web/dist/assets")
		router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

		// 根路径直接返回 index.html
		router.GET("/", func(c *gin.Context) {
			if !hasFrontendFiles {
				logger.Error("前端文件不存在", zap.String("path", indexPath))
				c.String(http.StatusInternalServerError,
					"前端文件未找到，请先运行: cd web && npm run build\n或设置环境变量 DEV_MODE=true 启用开发模式")
				return
			}
			c.File(indexPath)
		})

		// SPA 路由：所有非 API、非 WebSocket、非静态资源的请求都返回 index.html
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// 如果是 API 请求，返回 404
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			// 如果是 WebSocket 请求，返回 404
			if len(path) >= 3 && path[:3] == "/ws" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			// 如果是静态资源请求，返回 404（应该已经被上面的 Static 处理了）
			if len(path) >= 7 && path[:7] == "/assets" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			// 其他请求返回 index.html（支持前端路由）
			if !hasFrontendFiles {
				logger.Error("前端文件不存在", zap.String("path", indexPath))
				c.String(http.StatusInternalServerError,
					"前端文件未找到，请先运行: cd web && npm run build\n或设置环境变量 DEV_MODE=true 启用开发模式")
				return
			}
			c.File(indexPath)
		})
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		addr := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		if cfg.Server.Host != "0.0.0.0" && cfg.Server.Host != "" {
			addr = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		}

		logger.Info("服务器启动",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port),
			zap.String("url", addr),
			zap.Bool("dev_mode", devMode),
		)

		// 输出访问地址到控制台
		fmt.Println("")
		fmt.Println("========================================")
		fmt.Println("  数据库造数工具 - 服务已启动")
		fmt.Println("========================================")
		fmt.Printf("  访问地址: %s\n", addr)
		if devMode {
			fmt.Printf("  开发模式: 已启用（前端代理到 %s）\n", viteDevServer)
			fmt.Println("  提示: 请确保 Vite 开发服务器正在运行")
			fmt.Println("  启动命令: cd web && npm run dev")
		} else {
			fmt.Println("  模式: 生产模式（使用静态文件）")
		}
		fmt.Println("  按 Ctrl+C 停止服务")
		fmt.Println("========================================")
		fmt.Println("")

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
