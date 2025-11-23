/**
 * 常量定义
 */

// 任务状态映射
export const TASK_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
  STOPPED: 'stopped',
  ERROR: 'error'
}

// 任务状态文本映射
export const TASK_STATUS_TEXT = {
  [TASK_STATUS.PENDING]: '待开始',
  [TASK_STATUS.RUNNING]: '运行中',
  [TASK_STATUS.PAUSED]: '已暂停',
  [TASK_STATUS.COMPLETED]: '已完成',
  [TASK_STATUS.STOPPED]: '已停止',
  [TASK_STATUS.ERROR]: '错误'
}

// 任务状态类型映射（用于 Element Plus Tag）
export const TASK_STATUS_TYPE = {
  [TASK_STATUS.PENDING]: 'info',
  [TASK_STATUS.RUNNING]: 'success',
  [TASK_STATUS.PAUSED]: 'warning',
  [TASK_STATUS.COMPLETED]: 'success',
  [TASK_STATUS.STOPPED]: 'info',
  [TASK_STATUS.ERROR]: 'danger'
}

// 数据库类型映射
export const DB_TYPE = {
  POSTGRES: 'postgres',
  MYSQL: 'mysql',
  MARIADB: 'mariadb',
  SQLITE: 'sqlite',
  SQLSERVER: 'mssql',
  ORACLE: 'oracle',
  DAMENG: 'dameng'
}

// 数据库类型文本映射
export const DB_TYPE_TEXT = {
  [DB_TYPE.POSTGRES]: 'PostgreSQL',
  [DB_TYPE.MYSQL]: 'MySQL',
  [DB_TYPE.MARIADB]: 'MariaDB',
  [DB_TYPE.SQLITE]: 'SQLite',
  [DB_TYPE.SQLSERVER]: 'SQL Server',
  [DB_TYPE.ORACLE]: 'Oracle',
  [DB_TYPE.DAMENG]: '达梦数据库'
}

// WebSocket 连接状态
export const WS_CONNECTION_STATUS = {
  CONNECTED: 'connected',
  DISCONNECTED: 'disconnected',
  RECONNECTING: 'reconnecting'
}

// WebSocket 连接状态文本
export const WS_CONNECTION_STATUS_TEXT = {
  [WS_CONNECTION_STATUS.CONNECTED]: '实时更新已连接',
  [WS_CONNECTION_STATUS.DISCONNECTED]: '实时更新已断开',
  [WS_CONNECTION_STATUS.RECONNECTING]: '正在重连...'
}

// WebSocket 连接状态颜色
export const WS_CONNECTION_STATUS_COLOR = {
  [WS_CONNECTION_STATUS.CONNECTED]: '#67c23a', // 绿色
  [WS_CONNECTION_STATUS.RECONNECTING]: '#e6a23c', // 橙色
  [WS_CONNECTION_STATUS.DISCONNECTED]: '#f56c6c' // 红色
}

// 默认配置
export const DEFAULT_CONFIG = {
  THREAD_COUNT: 4,
  BATCH_SIZE: 500,
  TOTAL_ROWS: 1000,
  PREVIEW_COUNT: 10,
  PAGE_SIZE: 20
}

// 分页配置
export const PAGE_SIZES = [10, 20, 50, 100, 200]

// 重连配置
export const RECONNECT_CONFIG = {
  MAX_ATTEMPTS: 5,
  DELAY: 3000 // 毫秒
}

