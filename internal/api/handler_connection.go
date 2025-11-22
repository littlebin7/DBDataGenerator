package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/storage"
)

// TestConnection 测试数据库连接
func (h *Handler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("测试连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	config := &database.ConnectionConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
		Charset:  req.Charset,
	}

	// 创建带超时的上下文（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建数据库实例
	db, err := database.NewDatabase(config.Type)
	if err != nil {
		h.logger.Error("创建数据库实例失败", zap.Error(err), zap.String("type", config.Type))
		h.sendError(c, http.StatusBadRequest, ErrCodeUnsupportedDBType, "不支持的数据库类型", config.Type)
		return
	}

	// 连接数据库
	if err := db.Connect(config); err != nil {
		h.logger.Warn("数据库连接失败", zap.Error(err), zap.String("host", config.Host), zap.Int("port", config.Port))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "连接失败", err.Error())
		return
	}
	defer db.Disconnect()

	// 测试连接（带超时）
	done := make(chan error, 1)
	go func() {
		done <- db.TestConnection()
	}()

	select {
	case <-ctx.Done():
		h.logger.Warn("测试连接超时", zap.String("host", config.Host), zap.Int("port", config.Port))
		h.sendError(c, http.StatusRequestTimeout, ErrCodeConnectionTimeout, "连接测试超时")
		return
	case err := <-done:
		if err != nil {
			h.logger.Warn("测试连接失败", zap.Error(err), zap.String("host", config.Host), zap.Int("port", config.Port))
			h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "连接测试失败", err.Error())
			return
		}
	}

	// 获取数据库版本
	version := ""
	if versionStr, err := db.GetVersion(); err == nil {
		version = versionStr
		h.logger.Info("获取数据库版本成功", zap.String("version", version))
	} else {
		h.logger.Warn("获取数据库版本失败", zap.Error(err))
		// 版本获取失败不影响连接测试结果
	}

	h.logger.Info("测试连接成功", zap.String("host", config.Host), zap.Int("port", config.Port), zap.String("version", version))
	h.sendSuccess(c, gin.H{"version": version}, "连接成功")
}

// Connect 连接数据库
func (h *Handler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Name == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接名称不能为空")
		return
	}

	// 如果不保存密码，则清空密码
	password := req.Password
	if !req.SavePassword {
		password = ""
	}

	config := &database.ConnectionConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
		Charset:  req.Charset,
	}

	connID, err := h.connMgr.AddConnection(req.Name, config)
	if err != nil {
		h.logger.Error("添加连接失败", zap.Error(err), zap.String("name", req.Name), zap.String("host", req.Host))
		h.sendError(c, http.StatusInternalServerError, ErrCodeConnectionFailed, "添加连接失败", err.Error())
		return
	}

	// AddConnection 会保存配置，即使连接失败也会保存配置
	// 连接状态会在返回的连接信息中体现（connected 字段）
	h.logger.Info("连接配置保存成功", zap.String("conn_id", connID), zap.String("name", req.Name))
	h.sendSuccess(c, gin.H{"connection_id": connID}, "连接配置已保存")
}

// UpdateConnection 更新连接
func (h *Handler) UpdateConnection(c *gin.Context) {
	connID := c.Param("id")
	if connID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID不能为空")
		return
	}

	var req UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新连接请求参数错误", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	// 处理密码逻辑：
	// 1. 如果不保存密码，则清空密码
	// 2. 如果保存密码但密码为空，保留原来的密码
	password := req.Password
	if !req.SavePassword {
		password = ""
	} else if password == "" {
		// 如果保存密码但密码为空，获取原来的密码
		conn, err := h.connMgr.GetConnection(connID)
		if err == nil && conn != nil && conn.Config != nil {
			password = conn.Config.Password // 使用原来的密码（已解密）
		}
	}

	config := &database.ConnectionConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
		Charset:  req.Charset,
	}

	if err := h.connMgr.UpdateConnection(connID, req.Name, config); err != nil {
		h.logger.Error("更新连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeConnectionFailed, "更新连接失败", err.Error())
		return
	}

	// UpdateConnection 会保存配置，即使连接失败也会保存配置
	h.logger.Info("连接配置更新成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接配置已更新")
}

// Disconnect 断开连接
func (h *Handler) Disconnect(c *gin.Context) {
	connID := c.Param("id")
	if err := h.connMgr.RemoveConnection(connID); err != nil {
		h.logger.Warn("断开连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "断开连接失败", err.Error())
		return
	}

	h.logger.Info("断开连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接已断开")
}

// GetConnections 获取所有连接
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.connMgr.GetAllConnections()
	h.sendSuccess(c, gin.H{"connections": connections})
}

// SwitchConnection 切换连接
func (h *Handler) SwitchConnection(c *gin.Context) {
	connID := c.Param("id")
	if err := h.connMgr.SwitchConnection(connID); err != nil {
		h.logger.Warn("切换连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "切换连接失败", err.Error())
		return
	}

	h.logger.Info("切换连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接已切换")
}

// GetActiveConnection 获取活动连接
func (h *Handler) GetActiveConnection(c *gin.Context) {
	conn, err := h.connMgr.GetActiveConnection()
	if err != nil {
		// 没有活动连接时返回 200 和 null，而不是 404
		// 这样前端可以正常处理，不需要捕获错误
		h.sendSuccess(c, nil)
		return
	}
	h.sendSuccess(c, conn)
}

// Reconnect 重新连接
func (h *Handler) Reconnect(c *gin.Context) {
	connID := c.Param("id")
	if connID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID不能为空")
		return
	}

	if err := h.connMgr.Reconnect(connID); err != nil {
		h.logger.Warn("重新连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "重新连接失败", err.Error())
		return
	}

	h.logger.Info("重新连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接成功")
}

// ExportConnections 导出连接配置
func (h *Handler) ExportConnections(c *gin.Context) {
	var req struct {
		ConnectionIDs []string `json:"connection_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("导出连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if len(req.ConnectionIDs) == 0 {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请选择要导出的连接")
		return
	}

	// 获取所有连接
	allConnections := h.connMgr.GetAllConnections()

	// 筛选要导出的连接
	exportConnections := make([]*database.ConnectionInfo, 0)
	for _, conn := range allConnections {
		for _, id := range req.ConnectionIDs {
			if conn.ID == id {
				exportConnections = append(exportConnections, conn)
				break
			}
		}
	}

	if len(exportConnections) == 0 {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "未找到要导出的连接")
		return
	}

	// 准备导出数据（包含密码，但已加密）
	exportData := make([]map[string]interface{}, 0, len(exportConnections))
	for _, conn := range exportConnections {
		exportData = append(exportData, map[string]interface{}{
			"id":        conn.ID,
			"name":      conn.Name,
			"config":    conn.Config,
			"is_active": conn.IsActive,
		})
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		h.logger.Error("序列化导出数据失败", zap.Error(err))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "序列化数据失败", err.Error())
		return
	}

	// 加密整个文件内容
	encryptedData, err := storage.EncryptData(jsonData)
	if err != nil {
		h.logger.Error("加密导出数据失败", zap.Error(err))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "加密数据失败", err.Error())
		return
	}

	// Base64编码以便传输
	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	h.logger.Info("导出连接成功", zap.Int("count", len(exportData)))
	h.sendSuccess(c, gin.H{
		"data":      encodedData,
		"count":     len(exportData),
		"encrypted": true,
	}, "导出成功")
}

// ImportConnectionsPreview 预览导入的连接（解析文件但不导入）
func (h *Handler) ImportConnectionsPreview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Warn("导入连接文件错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请选择要导入的文件", err.Error())
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		h.logger.Warn("打开文件失败", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "打开文件失败", err.Error())
		return
	}
	defer src.Close()

	// 读取文件内容
	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		h.logger.Warn("读取文件失败", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "读取文件失败", err.Error())
		return
	}

	// 导出格式：base64 编码的加密数据（字符串）
	// 1. 文件内容是 base64 编码的字符串
	// 2. Base64 解码得到加密的二进制数据
	// 3. 解密得到 JSON 数组
	// 4. 解析 JSON 数组

	// Base64 解码
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		h.logger.Warn("Base64 解码失败", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "文件格式错误，请确保是有效的导出文件", err.Error())
		return
	}

	// 解密
	jsonData, err := storage.DecryptData(decodedData)
	if err != nil {
		h.logger.Warn("解密文件失败", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "文件解密失败，请确保是有效的导出文件", err.Error())
		return
	}

	// 解析为数组
	var connections []map[string]interface{}
	if err := json.Unmarshal(jsonData, &connections); err != nil {
		h.logger.Warn("解析文件失败", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "文件格式错误，请确保是有效的导出文件", err.Error())
		return
	}

	// 获取现有连接，用于检测冲突
	existingConnections := h.connMgr.GetAllConnections()
	existingNames := make(map[string]bool)
	existingIDs := make(map[string]bool)
	for _, conn := range existingConnections {
		existingNames[conn.Name] = true
		existingIDs[conn.ID] = true
	}

	// 处理连接数据，标记冲突
	previewData := make([]map[string]interface{}, 0, len(connections))
	for _, conn := range connections {
		connName, _ := conn["name"].(string)
		connID, _ := conn["id"].(string)

		// 检测冲突：同名或同ID
		hasNameConflict := existingNames[connName]
		hasIDConflict := existingIDs[connID]
		hasConflict := hasNameConflict || hasIDConflict

		conflictReason := ""
		if hasNameConflict && hasIDConflict {
			conflictReason = "名称和ID都已存在"
		} else if hasNameConflict {
			conflictReason = "名称已存在"
		} else if hasIDConflict {
			conflictReason = "ID已存在"
		}

		previewData = append(previewData, map[string]interface{}{
			"id":              connID,
			"name":            connName,
			"config":          conn["config"],
			"is_active":       conn["is_active"],
			"has_conflict":    hasConflict,
			"conflict_reason": conflictReason,
		})
	}

	h.logger.Info("预览导入连接成功", zap.Int("count", len(previewData)))
	h.sendSuccess(c, gin.H{"connections": previewData}, "预览成功")
}

// ImportConnections 批量导入连接
func (h *Handler) ImportConnections(c *gin.Context) {
	var req struct {
		Connections        []map[string]interface{} `json:"connections"`
		ConflictResolution string                   `json:"conflict_resolution"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("导入连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if len(req.Connections) == 0 {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请选择要导入的连接")
		return
	}

	// 默认冲突处理方式为跳过
	if req.ConflictResolution == "" {
		req.ConflictResolution = "skip"
	}

	// 获取现有连接
	existingConnections := h.connMgr.GetAllConnections()
	existingNames := make(map[string]bool)
	existingIDs := make(map[string]bool)
	for _, conn := range existingConnections {
		existingNames[conn.Name] = true
		existingIDs[conn.ID] = true
	}

	imported := 0
	skipped := 0
	updated := 0
	errors := make([]string, 0)

	for _, connData := range req.Connections {
		connName, _ := connData["name"].(string)
		connID, _ := connData["id"].(string)

		if connName == "" {
			errors = append(errors, fmt.Sprintf("连接名称不能为空"))
			skipped++
			continue
		}

		// 检查冲突
		hasNameConflict := existingNames[connName]
		hasIDConflict := existingIDs[connID]

		// 如果ID冲突，尝试更新现有连接
		if hasIDConflict {
			// 更新现有连接
			configData, ok := connData["config"].(map[string]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("连接 %s 配置格式错误", connName))
				skipped++
				continue
			}

			config := &database.ConnectionConfig{
				Type:     getString(configData, "type"),
				Host:     getString(configData, "host"),
				Port:     getInt(configData, "port"),
				User:     getString(configData, "user"),
				Password: getString(configData, "password"),
				Database: getString(configData, "database"),
				SSLMode:  getString(configData, "ssl_mode"),
				Charset:  getString(configData, "charset"),
			}

			// 尝试解密密码（如果是加密的）
			if config.Password != "" {
				decryptedPassword, err := storage.DecryptPassword(config.Password)
				if err == nil {
					config.Password = decryptedPassword
				}
			}

			// 更新连接
			if err := h.connMgr.UpdateConnection(connID, connName, config); err != nil {
				errors = append(errors, fmt.Sprintf("更新连接 %s 失败: %v", connName, err))
				skipped++
				continue
			}

			updated++
			continue
		}

		// 如果名称冲突但ID不冲突，生成新名称
		if hasNameConflict {
			// 生成新名称（添加序号）
			counter := 1
			newName := fmt.Sprintf("%s_%d", connName, counter)
			for existingNames[newName] {
				counter++
				newName = fmt.Sprintf("%s_%d", connName, counter)
			}
			connName = newName
		}

		// 解析配置
		configData, ok := connData["config"].(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("连接 %s 配置格式错误", connName))
			skipped++
			continue
		}

		config := &database.ConnectionConfig{
			Type:     getString(configData, "type"),
			Host:     getString(configData, "host"),
			Port:     getInt(configData, "port"),
			User:     getString(configData, "user"),
			Password: getString(configData, "password"),
			Database: getString(configData, "database"),
			SSLMode:  getString(configData, "ssl_mode"),
			Charset:  getString(configData, "charset"),
		}

		// 尝试解密密码（如果是加密的）
		if config.Password != "" {
			decryptedPassword, err := storage.DecryptPassword(config.Password)
			if err == nil {
				config.Password = decryptedPassword
			}
			// 如果解密失败，保持原值（可能是未加密的旧密码）
		}

		// 导入连接（不强制连接）
		_, err := h.connMgr.AddConnection(connName, config)
		if err != nil {
			errors = append(errors, fmt.Sprintf("导入连接 %s 失败: %v", connName, err))
			skipped++
			continue
		}

		imported++
		existingNames[connName] = true // 标记为已存在，避免重复导入
		if connID != "" {
			existingIDs[connID] = true
		}
	}

	h.logger.Info("导入连接完成", zap.Int("imported", imported), zap.Int("updated", updated), zap.Int("skipped", skipped))

	message := fmt.Sprintf("导入完成：新增 %d 个，更新 %d 个，跳过 %d 个", imported, updated, skipped)
	if len(errors) > 0 {
		message += fmt.Sprintf("，错误 %d 个", len(errors))
	}

	h.sendSuccess(c, gin.H{
		"imported": imported,
		"updated":  updated,
		"skipped":  skipped,
		"errors":   errors,
	}, message)
}

// 辅助函数：从 map 中获取字符串
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从 map 中获取整数
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}
