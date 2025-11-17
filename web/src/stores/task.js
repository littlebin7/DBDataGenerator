import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'

export const useTaskStore = defineStore('task', {
  state: () => ({
    tasks: [],
    selectedTask: null,
    loading: false
  }),

  getters: {
    // 获取所有任务
    allTasks: (state) => state.tasks,
    
    // 获取运行中的任务
    runningTasks: (state) => state.tasks.filter(t => t.status === 'running' || t.status === 'paused'),
    
    // 根据ID获取任务
    getTaskById: (state) => (id) => {
      return state.tasks.find(task => task.id === id)
    },
    
    // 根据状态筛选任务
    getTasksByStatus: (state) => (status) => {
      return state.tasks.filter(task => task.status === status)
    }
  },

  actions: {
    // 加载所有任务
    async loadTasks() {
      this.loading = true
      try {
        const data = await api.getTasks()
        this.tasks = data.tasks || []
      } catch (error) {
        console.error('加载任务列表失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 获取任务详情
    async loadTask(taskId) {
      try {
        const data = await api.getTask(taskId)
        this.selectedTask = data
        // 同时更新任务列表中的任务
        const index = this.tasks.findIndex(t => t.id === taskId)
        if (index !== -1) {
          this.tasks[index] = data
        }
        return data
      } catch (error) {
        console.error('获取任务详情失败:', error)
        throw error
      }
    },

    // 创建任务
    async createTask(name, connectionId, config) {
      try {
        const data = await api.createTask(name, connectionId, config)
        await this.loadTasks()
        return data
      } catch (error) {
        console.error('创建任务失败:', error)
        throw error
      }
    },

    // 启动任务
    async startTask(taskId) {
      try {
        await api.startTask(taskId)
        await this.loadTasks()
      } catch (error) {
        console.error('启动任务失败:', error)
        throw error
      }
    },

    // 暂停任务
    async pauseTask(taskId) {
      try {
        await api.pauseTask(taskId)
        await this.loadTasks()
      } catch (error) {
        console.error('暂停任务失败:', error)
        throw error
      }
    },

    // 恢复任务
    async resumeTask(taskId) {
      try {
        await api.resumeTask(taskId)
        await this.loadTasks()
      } catch (error) {
        console.error('恢复任务失败:', error)
        throw error
      }
    },

    // 停止任务
    async stopTask(taskId) {
      try {
        await api.stopTask(taskId)
        await this.loadTasks()
      } catch (error) {
        console.error('停止任务失败:', error)
        throw error
      }
    },

    // 删除任务
    async deleteTask(taskId) {
      try {
        await api.deleteTask(taskId)
        this.tasks = this.tasks.filter(t => t.id !== taskId)
        if (this.selectedTask?.id === taskId) {
          this.selectedTask = null
        }
      } catch (error) {
        console.error('删除任务失败:', error)
        throw error
      }
    },

    // 设置线程数
    async setThreadCount(taskId, count) {
      try {
        await api.setThreadCount(taskId, count)
        await this.loadTasks()
      } catch (error) {
        console.error('设置线程数失败:', error)
        throw error
      }
    },

    // 更新任务（从外部更新，如 WebSocket）
    updateTask(task) {
      const index = this.tasks.findIndex(t => t.id === task.id)
      if (index !== -1) {
        this.tasks[index] = { ...this.tasks[index], ...task }
      } else {
        this.tasks.push(task)
      }
      
      if (this.selectedTask?.id === task.id) {
        this.selectedTask = { ...this.selectedTask, ...task }
      }
    },

    // 设置选中的任务
    setSelectedTask(task) {
      this.selectedTask = task
    }
  }
})

