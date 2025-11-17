import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 响应拦截器：统一处理错误格式和成功响应格式
api.interceptors.response.use(
  (response) => {
    // 处理新的成功响应格式 { message, data }
    // 如果响应包含 data 字段，直接返回 data
    if (response.data && response.data.data !== undefined) {
      return { ...response, data: response.data.data }
    }
    return response
  },
  (error) => {
    // 统一处理错误格式
    if (error.response?.data?.error) {
      const errorData = error.response.data.error
      if (typeof errorData === 'object' && errorData.message) {
        // 新格式：{ code, message, details }
        error.formattedMessage = errorData.message
        if (errorData.details) {
          error.formattedMessage += ` (${errorData.details})`
        }
        error.errorCode = errorData.code
        error.errorDetails = errorData.details
      } else {
        // 旧格式：字符串（向后兼容）
        error.formattedMessage = errorData
      }
    } else {
      error.formattedMessage = error.message || '请求失败'
    }
    return Promise.reject(error)
  }
)

export default {
  async testConnection(config) {
    const response = await api.post('/connect/test', config, {
      timeout: 10000 // 10秒超时（包含网络传输时间）
    })
    return response.data
  },

  async connect(config) {
    const response = await api.post('/connect', config)
    return response.data
  },

  async updateConnection(connId, config) {
    const response = await api.put(`/connection/${connId}`, config)
    return response.data
  },

  async getConnections() {
    const response = await api.get('/connections')
    return response.data
  },

  async getActiveConnection() {
    const response = await api.get('/connection/active')
    return response.data
  },

  async switchConnection(connId) {
    const response = await api.post(`/connection/${connId}/switch`)
    return response.data
  },

  async disconnect(connId) {
    const response = await api.delete(`/connection/${connId}`)
    return response.data
  },

  async getDatabases(connectionId) {
    const params = {}
    if (connectionId) {
      params.connection_id = connectionId
    }
    const response = await api.get('/databases', { params })
    return response.data
  },

  async getTables(database, connectionId) {
    const params = { database }
    if (connectionId) {
      params.connection_id = connectionId
    }
    const response = await api.get('/tables', { params })
    return response.data
  },

  async getTableSchema(database, table, connectionId) {
    const params = { database }
    if (connectionId) {
      params.connection_id = connectionId
    }
    const response = await api.get(`/table/${table}/schema`, { params })
    return response.data
  },

  async createTask(name, connectionId, config) {
    const response = await api.post('/task/create', {
      name,
      connection_id: connectionId,
      config
    })
    return response.data
  },

  async getTasks() {
    const response = await api.get('/tasks')
    return response.data
  },

  async getTask(taskId) {
    const response = await api.get(`/task/${taskId}`)
    return response.data
  },

  async startTask(taskId) {
    const response = await api.post(`/task/${taskId}/start`)
    return response.data
  },

  async pauseTask(taskId) {
    const response = await api.post(`/task/${taskId}/pause`)
    return response.data
  },

  async resumeTask(taskId) {
    const response = await api.post(`/task/${taskId}/resume`)
    return response.data
  },

  async stopTask(taskId) {
    const response = await api.post(`/task/${taskId}/stop`)
    return response.data
  },

  async deleteTask(taskId) {
    const response = await api.delete(`/task/${taskId}`)
    return response.data
  },

  async setThreadCount(taskId, count) {
    const response = await api.put(`/task/${taskId}/threads`, { count })
    return response.data
  },

  // 模板管理
  async saveTemplate(data) {
    const response = await api.post('/template/save', data)
    return response.data
  },

  async getTemplates(tableName) {
    const params = {}
    if (tableName) {
      params.table_name = tableName
    }
    const response = await api.get('/templates', { params })
    return response.data
  },

  async getTemplate(templateId) {
    const response = await api.get(`/template/${templateId}`)
    return response.data
  },

  async deleteTemplate(templateId) {
    const response = await api.delete(`/template/${templateId}`)
    return response.data
  },

  // 数据预览
  async previewData(connectionId, config, count = 10) {
    const response = await api.post('/generator/preview', {
      connection_id: connectionId,
      config,
      count
    })
    return response.data
  },

  // 任务历史
  async getTaskHistory(limit = 50, offset = 0, status = '') {
    const params = { limit, offset }
    if (status) {
      params.status = status
    }
    const response = await api.get('/tasks/history', { params })
    return response.data
  },

  async getTaskHistoryByTaskId(taskId) {
    const response = await api.get(`/task/${taskId}/history`)
    return response.data
  },

  async deleteTaskHistory(historyId) {
    const response = await api.delete(`/task/history/${historyId}`)
    return response.data
  },

  // 监控相关
  async getSystemMetrics() {
    const response = await api.get('/monitor/metrics')
    return response.data
  },

  // 连接池相关
  async getPoolStatus(connectionId) {
    const params = {}
    if (connectionId) {
      params.connection_id = connectionId
    }
    const response = await api.get('/pool/status', { params })
    return response.data
  },

  // 数据质量检查
  async checkDataQuality(connectionId, database, table) {
    const params = {
      connection_id: connectionId,
      database,
      table
    }
    const response = await api.get('/quality/check', { params })
    return response.data
  },

  // 定时任务相关
  async scheduleTask(taskId, cronExpr) {
    const response = await api.post(`/task/${taskId}/schedule`, { cron_expr: cronExpr })
    return response.data
  },

  async getSchedule(taskId) {
    const response = await api.get(`/task/${taskId}/schedule`)
    return response.data
  },

  async getAllSchedules() {
    const response = await api.get('/schedules')
    return response.data
  },

  async enableSchedule(taskId) {
    const response = await api.post(`/task/${taskId}/schedule/enable`)
    return response.data
  },

  async disableSchedule(taskId) {
    const response = await api.post(`/task/${taskId}/schedule/disable`)
    return response.data
  },

  async removeSchedule(taskId) {
    const response = await api.delete(`/task/${taskId}/schedule`)
    return response.data
  },

  // 表关系分析
  async getTableRelations(connectionId, database) {
    const params = {
      connection_id: connectionId,
      database
    }
    const response = await api.get('/relations', { params })
    return response.data
  },

  async getTableRelationsByTable(connectionId, database, tableName) {
    const params = {
      connection_id: connectionId,
      database
    }
    const response = await api.get(`/table/${tableName}/relations`, { params })
    return response.data
  },

  async generateCascade(config) {
    const response = await api.post('/cascade/generate', config)
    return response.data
  },

  // 预设模板
  async getPresetTemplates(category, fieldType, keyword) {
    const params = {}
    if (category) params.category = category
    if (fieldType) params.field_type = fieldType
    if (keyword) params.keyword = keyword
    const response = await api.get('/presets', { params })
    return response.data
  },

  async getPresetTemplate(presetId) {
    const response = await api.get(`/preset/${presetId}`)
    return response.data
  },

  async applyPresetTemplate(taskId, presetId) {
    const response = await api.post('/preset/apply', {
      task_id: taskId,
      preset_id: presetId
    })
    return response.data
  },

  // 批量任务管理
  async batchCreateTasks(tasks) {
    const response = await api.post('/tasks/batch/create', { tasks })
    return response.data
  },

  async batchStartTasks(taskIds) {
    const response = await api.post('/tasks/batch/start', { task_ids: taskIds })
    return response.data
  },

  async batchStopTasks(taskIds) {
    const response = await api.post('/tasks/batch/stop', { task_ids: taskIds })
    return response.data
  },

  async batchDeleteTasks(taskIds) {
    const response = await api.post('/tasks/batch/delete', { task_ids: taskIds })
    return response.data
  },

  // 数据回滚
  async rollbackTask(taskId) {
    const response = await api.post(`/task/${taskId}/rollback`)
    return response.data
  },

  async rollbackPartial(taskId, options) {
    const response = await api.post(`/task/${taskId}/rollback/partial`, options)
    return response.data
  },

  async getRollbackRecord(taskId) {
    const response = await api.get(`/task/${taskId}/rollback`)
    return response.data
  },

  // 任务复制
  async cloneTask(taskId) {
    const response = await api.post(`/task/${taskId}/clone`)
    return response.data
  },

  // 数据导入导出
  async exportData(connectionId, database, table) {
    const response = await api.get(`/export/${connectionId}/${database}/${table}`, {
      responseType: 'blob'
    })
    return response.data
  },

  async importData(connectionId, database, table, data) {
    const response = await api.post(`/import/${connectionId}/${database}/${table}`, { data })
    return response.data
  }
}

