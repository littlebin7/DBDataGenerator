//go:build !cgo

package database

import "fmt"

type OracleDB struct {
	db     interface{}
	config *ConnectionConfig
}

func NewOracleDB() *OracleDB {
	return &OracleDB{}
}

func (db *OracleDB) Connect(config *ConnectionConfig) error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持，当前编译环境不支持")
}

func (db *OracleDB) Disconnect() error {
	return nil
}

func (db *OracleDB) TestConnection() error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetDatabases() ([]string, error) {
	return nil, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetTables(database string) ([]string, error) {
	return nil, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetTableSchema(database, table string) (*TableSchema, error) {
	return nil, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetTableCount(database, table string) (int64, error) {
	return 0, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetDBType() string {
	return "oracle"
}

func (db *OracleDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) GetVersion() (string, error) {
	return "", fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) BeginTransaction() error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) CommitTransaction() error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) RollbackTransaction() error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}

func (db *OracleDB) BatchInsertInTransaction(database, table string, rows []map[string]interface{}) error {
	return fmt.Errorf("Oracle 数据库需要 CGO 和 Oracle Instant Client 库支持")
}
