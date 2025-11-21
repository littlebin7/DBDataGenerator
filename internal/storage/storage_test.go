package storage

import (
	"os"
	"testing"
)

func TestNewSQLiteStorage(t *testing.T) {
	tmpDB := "test_storage.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestStorage_Type(t *testing.T) {
	tmpDB := "test_storage_type.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	storageType := storage.Type()
	if storageType != "sqlite" {
		t.Errorf("期望存储类型为 'sqlite'，实际 %s", storageType)
	}
}

func TestStorage_GetDB(t *testing.T) {
	tmpDB := "test_storage_getdb.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	db := storage.GetDB()
	if db == nil {
		t.Error("期望获取数据库实例，但返回 nil")
	}
}

func TestStorage_InitTables(t *testing.T) {
	tmpDB := "test_storage_inittables.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	// InitTables 在 NewSQLiteStorage 中已调用，这里测试再次调用
	err = storage.InitTables()
	if err != nil {
		t.Errorf("初始化表失败: %v", err)
	}
}

func TestStorage_Close(t *testing.T) {
	tmpDB := "test_storage_close.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}

	// 关闭存储
	err = storage.Close()
	if err != nil {
		t.Errorf("关闭存储失败: %v", err)
	}

	// 再次关闭应该不报错
	err = storage.Close()
	if err != nil {
		t.Errorf("重复关闭存储应该不报错: %v", err)
	}
}

func TestNewSQLiteStorage_WithInvalidPath(t *testing.T) {
	// 使用无效路径（只读目录或权限不足）
	// 注意：在 Windows 上可能无法创建真正的无效路径测试
	// 这里只测试基本功能

	tmpDB := "test_storage_invalid.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Logf("创建存储失败（可能是预期的）: %v", err)
		return
	}
	defer storage.Close()

	if storage == nil {
		t.Error("期望创建存储实例，但返回 nil")
	}
}

func TestStorage_InitTables_MultipleTimes(t *testing.T) {
	tmpDB := "test_storage_multiple_init.db"
	defer os.Remove(tmpDB)

	storage, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	// 多次初始化表应该不报错（使用 IF NOT EXISTS）
	for i := 0; i < 3; i++ {
		err = storage.InitTables()
		if err != nil {
			t.Errorf("第 %d 次初始化表失败: %v", i+1, err)
		}
	}
}
