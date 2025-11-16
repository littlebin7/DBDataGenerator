import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

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
  }
}

