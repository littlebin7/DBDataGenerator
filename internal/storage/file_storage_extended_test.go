package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage_SaveConnections(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 测试数据
	connections := []map[string]interface{}{
		{
			"id":   "conn1",
			"name": "连接1",
			"type": "sqlite",
		},
		{
			"id":   "conn2",
			"name": "连接2",
			"type": "mysql",
		},
	}

	// 保存连接
	err = fs.SaveConnections(connections)
	if err != nil {
		t.Fatalf("保存连接失败: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(connectionsFile); os.IsNotExist(err) {
		t.Error("期望连接配置文件已创建，但文件不存在")
	}

	// 验证文件内容
	data, err := os.ReadFile(connectionsFile)
	if err != nil {
		t.Fatalf("读取连接配置文件失败: %v", err)
	}

	var loadedConnections []map[string]interface{}
	if err := json.Unmarshal(data, &loadedConnections); err != nil {
		t.Fatalf("解析连接配置失败: %v", err)
	}

	if len(loadedConnections) != 2 {
		t.Errorf("期望有 2 个连接，实际 %d", len(loadedConnections))
	}
}

func TestFileStorage_LoadConnections(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 测试数据
	connections := []map[string]interface{}{
		{
			"id":   "conn1",
			"name": "连接1",
			"type": "sqlite",
		},
	}

	// 先保存
	err = fs.SaveConnections(connections)
	if err != nil {
		t.Fatalf("保存连接失败: %v", err)
	}

	// 加载连接
	data, err := fs.LoadConnections()
	if err != nil {
		t.Fatalf("加载连接失败: %v", err)
	}

	if len(data) == 0 {
		t.Error("期望加载到数据，但返回空")
	}

	// 验证内容
	var loadedConnections []map[string]interface{}
	if err := json.Unmarshal(data, &loadedConnections); err != nil {
		t.Fatalf("解析连接配置失败: %v", err)
	}

	if len(loadedConnections) != 1 {
		t.Errorf("期望有 1 个连接，实际 %d", len(loadedConnections))
	}
}

func TestFileStorage_LoadConnections_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "nonexistent.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 加载不存在的文件应该返回空数组
	data, err := fs.LoadConnections()
	if err != nil {
		t.Fatalf("加载连接失败: %v", err)
	}

	// 应该返回空数组 JSON
	if string(data) != "[]" {
		t.Errorf("期望返回空数组 '[]'，实际 %s", string(data))
	}
}

func TestFileStorage_SaveTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 测试数据
	templates := map[string]interface{}{
		"template1": map[string]interface{}{
			"id":         "template1",
			"name":       "模板1",
			"table_name": "users",
		},
		"template2": map[string]interface{}{
			"id":         "template2",
			"name":       "模板2",
			"table_name": "orders",
		},
	}

	// 保存模板
	err = fs.SaveTemplates(templates)
	if err != nil {
		t.Fatalf("保存模板失败: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		t.Error("期望模板配置文件已创建，但文件不存在")
	}

	// 验证文件内容
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		t.Fatalf("读取模板配置文件失败: %v", err)
	}

	var loadedTemplates map[string]interface{}
	if err := json.Unmarshal(data, &loadedTemplates); err != nil {
		t.Fatalf("解析模板配置失败: %v", err)
	}

	if len(loadedTemplates) != 2 {
		t.Errorf("期望有 2 个模板，实际 %d", len(loadedTemplates))
	}
}

func TestFileStorage_LoadTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 测试数据
	templates := map[string]interface{}{
		"template1": map[string]interface{}{
			"id":         "template1",
			"name":       "模板1",
			"table_name": "users",
		},
	}

	// 先保存
	err = fs.SaveTemplates(templates)
	if err != nil {
		t.Fatalf("保存模板失败: %v", err)
	}

	// 加载模板
	data, err := fs.LoadTemplates()
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	if len(data) == 0 {
		t.Error("期望加载到数据，但返回空")
	}

	// 验证内容
	var loadedTemplates map[string]interface{}
	if err := json.Unmarshal(data, &loadedTemplates); err != nil {
		t.Fatalf("解析模板配置失败: %v", err)
	}

	if len(loadedTemplates) != 1 {
		t.Errorf("期望有 1 个模板，实际 %d", len(loadedTemplates))
	}
}

func TestFileStorage_LoadTemplates_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "nonexistent.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 加载不存在的文件应该返回空对象
	data, err := fs.LoadTemplates()
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 应该返回空对象 JSON
	if string(data) != "{}" {
		t.Errorf("期望返回空对象 '{}'，实际 %s", string(data))
	}
}

func TestFileStorage_SaveConnections_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 创建一个无法序列化的数据（循环引用）
	type Circular struct {
		Self *Circular
	}
	circular := &Circular{}
	circular.Self = circular

	// 保存应该失败
	err = fs.SaveConnections(circular)
	if err == nil {
		t.Error("期望保存无效数据时返回错误")
	}
}

func TestFileStorage_SaveTemplates_InvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 创建一个无法序列化的数据（循环引用）
	type Circular struct {
		Self *Circular
	}
	circular := &Circular{}
	circular.Self = circular

	// 保存应该失败
	err = fs.SaveTemplates(circular)
	if err == nil {
		t.Error("期望保存无效数据时返回错误")
	}
}

func TestFileStorage_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 并发保存和加载
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go func(id int) {
			connections := []map[string]interface{}{
				{
					"id":   "conn" + string(rune(id)),
					"name": "连接" + string(rune(id)),
				},
			}
			fs.SaveConnections(connections)
			fs.LoadConnections()
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证最终状态
	data, err := fs.LoadConnections()
	if err != nil {
		t.Fatalf("加载连接失败: %v", err)
	}

	// 应该至少有一个连接（取决于并发执行顺序）
	if len(data) == 0 {
		t.Error("期望至少有一个连接")
	}
}
