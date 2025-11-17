<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div style="display: flex; justify-content: space-between; align-items: center; height: 100%">
        <h1 style="margin: 0">数据库造数工具</h1>
        <div v-if="activeConnection" style="color: #67c23a; font-size: 14px">
          <el-icon><SuccessFilled /></el-icon>
          已连接: {{ activeConnection.name }}
        </div>
      </div>
    </el-header>
    <el-container class="app-body">
      <el-aside width="200px" class="app-aside">
        <NavMenu />
      </el-aside>
      <el-main class="app-main">
        <el-breadcrumb separator="/" style="margin-bottom: 20px">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item v-if="breadcrumbItems.length > 0">
            {{ breadcrumbItems[breadcrumbItems.length - 1] }}
          </el-breadcrumb-item>
        </el-breadcrumb>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { SuccessFilled } from '@element-plus/icons-vue'
import NavMenu from './components/NavMenu.vue'
import api from './api'

const route = useRoute()
const activeConnection = ref(null)

const breadcrumbItems = computed(() => {
  const path = route.path
  const items = []
  if (path === '/connection') items.push('连接配置')
  else if (path === '/database-select') items.push('选择数据库')
  else if (path === '/tasks') items.push('任务列表')
  else if (path === '/task/config') items.push('配置造数规则')
  else if (path === '/task-history') items.push('任务历史')
  else if (path === '/monitor') items.push('系统监控')
  else if (path === '/pool') items.push('连接池监控')
  else if (path === '/quality') items.push('数据质量检查')
  else if (path === '/schedule') items.push('定时任务管理')
  else if (path === '/cascade') items.push('级联生成')
  else if (path === '/template') items.push('模板管理')
  else if (path === '/rollback') items.push('数据回滚')
  return items
})

const loadActiveConnection = async () => {
  try {
    const res = await api.getActiveConnection()
    // 检查是否有有效的连接
    if (res && res.id) {
      activeConnection.value = res
    } else {
      activeConnection.value = null
    }
  } catch (error) {
    activeConnection.value = null
  }
}

onMounted(() => {
  loadActiveConnection()
})

watch(() => route.path, () => {
  loadActiveConnection()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  overflow: hidden;
}

#app {
  height: 100vh;
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  background-color: #409EFF;
  color: white;
  line-height: 60px;
  height: 60px !important;
  flex-shrink: 0;
}

.app-header h1 {
  margin: 0;
  font-size: 24px;
}

.app-body {
  flex: 1;
  overflow: hidden;
  display: flex;
}

.app-aside {
  background: #f5f5f5;
  border-right: 1px solid #e6e6e6;
  overflow-y: auto;
  flex-shrink: 0;
}

.app-main {
  flex: 1;
  overflow: hidden;
  padding: 20px;
  background: #fff;
  display: flex;
  flex-direction: column;
}
</style>

