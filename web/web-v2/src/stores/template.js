import { defineStore } from 'pinia'
import api from '../api'

export const useTemplateStore = defineStore('template', {
  state: () => ({
    templates: [],
    loading: false
  }),

  getters: {
    // 获取所有模板
    allTemplates: (state) => state.templates,
    
    // 根据表名获取模板
    getTemplatesByTable: (state) => (tableName) => {
      return state.templates.filter(t => t.table_name === tableName)
    },
    
    // 根据ID获取模板
    getTemplateById: (state) => (id) => {
      return state.templates.find(t => t.id === id)
    }
  },

  actions: {
    // 加载所有模板
    async loadTemplates(tableName = '') {
      this.loading = true
      try {
        const data = await api.getTemplates(tableName)
        this.templates = data.templates || []
      } catch (error) {
        console.error('加载模板列表失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 获取模板详情
    async loadTemplate(templateId) {
      try {
        const data = await api.getTemplate(templateId)
        return data
      } catch (error) {
        console.error('获取模板详情失败:', error)
        throw error
      }
    },

    // 保存模板
    async saveTemplate(name, description, tableName, config) {
      try {
        const data = await api.saveTemplate({
          name,
          description,
          table_name: tableName,
          config
        })
        await this.loadTemplates()
        return data
      } catch (error) {
        console.error('保存模板失败:', error)
        throw error
      }
    },

    // 删除模板
    async deleteTemplate(templateId) {
      try {
        await api.deleteTemplate(templateId)
        this.templates = this.templates.filter(t => t.id !== templateId)
      } catch (error) {
        console.error('删除模板失败:', error)
        throw error
      }
    }
  }
})

