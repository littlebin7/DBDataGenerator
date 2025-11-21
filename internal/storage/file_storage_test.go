package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, err := NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	if fs == nil {
		t.Fatal("期望创建文件存储实例，但返回 nil")
	}

	if fs.connectionsFile != connectionsFile {
		t.Errorf("连接文件路径不正确: 期望 %s，实际 %s", connectionsFile, fs.connectionsFile)
	}

	if fs.templatesFile != templatesFile {
		t.Errorf("模板文件路径不正确: 期望 %s，实际 %s", templatesFile, fs.templatesFile)
	}

	// 检查目录是否创建
	if _, err := os.Stat(filepath.Dir(connectionsFile)); os.IsNotExist(err) {
		t.Error("连接文件目录未创建")
	}
	if _, err := os.Stat(filepath.Dir(templatesFile)); os.IsNotExist(err) {
		t.Error("模板文件目录未创建")
	}
}

func TestFileStorage_GetDB(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, _ := NewFileStorage(connectionsFile, templatesFile)

	db := fs.GetDB()
	if db != nil {
		t.Error("文件存储应该返回 nil 数据库连接")
	}
}

func TestFileStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, _ := NewFileStorage(connectionsFile, templatesFile)

	err := fs.Close()
	if err != nil {
		t.Errorf("关闭文件存储不应该返回错误: %v", err)
	}
}

func TestFileStorage_InitTables(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, _ := NewFileStorage(connectionsFile, templatesFile)

	err := fs.InitTables()
	if err != nil {
		t.Errorf("初始化表不应该返回错误: %v", err)
	}
}

func TestFileStorage_Type(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fs, _ := NewFileStorage(connectionsFile, templatesFile)

	if fs.Type() != "file" {
		t.Errorf("期望类型为 'file'，实际 %s", fs.Type())
	}
}
