import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    connections: [],
    activeConnection: null,
    loading: false
  }),

  getters: {
    // 获取所有连接
    allConnections: (state) => state.connections,
    
    // 获取活动连接
    currentConnection: (state) => state.activeConnection,
    
    // 根据ID获取连接
    getConnectionById: (state) => (id) => {
      return state.connections.find(conn => conn.id === id)
    }
  },

  actions: {
    // 加载所有连接
    async loadConnections() {
      this.loading = true
      try {
        const data = await api.getConnections()
        this.connections = data.connections || []
      } catch (error) {
        console.error('加载连接列表失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 加载活动连接
    async loadActiveConnection() {
      try {
        const data = await api.getActiveConnection()
        this.activeConnection = data
        return data
      } catch (error) {
        this.activeConnection = null
        return null
      }
    },

    // 切换连接
    async switchConnection(connectionId) {
      try {
        await api.switchConnection(connectionId)
        await this.loadActiveConnection()
        await this.loadConnections()
      } catch (error) {
        console.error('切换连接失败:', error)
        throw error
      }
    },

    // 添加连接
    async addConnection(config) {
      try {
        const data = await api.connect(config)
        await this.loadConnections()
        await this.loadActiveConnection()
        return data
      } catch (error) {
        console.error('添加连接失败:', error)
        throw error
      }
    },

    // 更新连接
    async updateConnection(connectionId, config) {
      try {
        await api.updateConnection(connectionId, config)
        await this.loadConnections()
        if (this.activeConnection?.id === connectionId) {
          await this.loadActiveConnection()
        }
      } catch (error) {
        console.error('更新连接失败:', error)
        throw error
      }
    },

    // 删除连接
    async removeConnection(connectionId) {
      try {
        await api.disconnect(connectionId)
        await this.loadConnections()
        if (this.activeConnection?.id === connectionId) {
          this.activeConnection = null
        }
      } catch (error) {
        console.error('删除连接失败:', error)
        throw error
      }
    }
  }
})

