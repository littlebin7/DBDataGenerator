//go:build 386 || arm || freebsd || openbsd || netbsd
// +build 386 arm freebsd openbsd netbsd

// 在不支持达梦数据库的平台上提供存根实现

package database

import "fmt"

type DamengDB struct {
	db     interface{}
	config *ConnectionConfig
}

func NewDamengDB() *DamengDB {
	return &DamengDB{}
}

func (db *DamengDB) Connect(config *ConnectionConfig) error {
	return fmt.Errorf("达梦数据库在此平台不支持（32位平台和BSD平台不支持）")
}

func (db *DamengDB) Disconnect() error {
	return nil
}

func (db *DamengDB) TestConnection() error {
	return fmt.Errorf("达梦数据库在此平台不支持")
}

func (db *DamengDB) GetDatabases() ([]string, error) {
	return nil, fmt.Errorf("达梦数据库在此平台不支持")
}

func (db *DamengDB) GetTables(database string) ([]string, error) {
	return nil, fmt.Errorf("达梦数据库在此平台不支持")
}

func (db *DamengDB) GetTableSchema(database, table string) (*TableSchema, error) {
	return nil, fmt.Errorf("达梦数据库在此平台不支持")
}

func (db *DamengDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return fmt.Errorf("达梦数据库在此平台不支持")
}

func (db *DamengDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, fmt.Errorf("达梦数据库在此平台不支持")
}
