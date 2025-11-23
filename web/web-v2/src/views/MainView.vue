<template>
  <div class="main-view">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h1 class="app-title">数据库造数工具</h1>
        <div v-if="activeConnection" class="connection-status">
          <el-icon class="status-icon"><SuccessFilled /></el-icon>
          <span>已连接: {{ activeConnection.name }}</span>
        </div>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchText"
          placeholder="搜索连接、数据库或表名"
          style="width: 300px; margin-right: 10px"
          clearable
          :prefix-icon="Search"
          @input="handleSearch"
        />
        <el-button :icon="Plus" @click="showConnectionDialog = true">新建连接</el-button>
        <el-button :icon="Refresh" @click="refreshAll" :loading="refreshing">刷新</el-button>
        <el-button :icon="Setting" @click="showSettings = true">设置</el-button>
      </div>
    </div>

    <!-- 主内容区：左右分栏 -->
    <div class="content-area">
      <!-- 左侧树形导航 -->
      <div class="tree-panel">
        <DatabaseTree
          ref="treeRef"
          :search-text="searchText"
          @node-click="handleNodeClick"
          @connection-updated="handleConnectionUpdated"
        />
      </div>

      <!-- 右侧内容面板 -->
      <div class="detail-panel">
        <DetailPanel
          :selected-node="selectedNode"
          :node-data="selectedNodeData"
          @refresh="handleRefresh"
          @create-task="handleCreateTask"
        />
      </div>
    </div>

    <!-- 新建连接对话框 -->
    <ConnectionDialog
      v-model="showConnectionDialog"
      @success="handleConnectionSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { SuccessFilled, Search, Plus, Refresh, Setting } from '@element-plus/icons-vue'
import { useConnectionStore } from '../stores/connection'
import DatabaseTree from '../components/DatabaseTree.vue'
import DetailPanel from '../components/DetailPanel.vue'
import ConnectionDialog from '../components/ConnectionDialog.vue'
import api from '../api'

const router = useRouter()
const connectionStore = useConnectionStore()

const treeRef = ref(null)
const searchText = ref('')
const refreshing = ref(false)
const showConnectionDialog = ref(false)
const showSettings = ref(false)
const selectedNode = ref(null)
const selectedNodeData = ref(null)

const activeConnection = computed(() => connectionStore.currentConnection)

// 搜索处理
const handleSearch = () => {
  treeRef.value?.filter(searchText.value)
}

// 刷新全部
const refreshAll = async () => {
  refreshing.value = true
  try {
    await connectionStore.loadConnections()
    await connectionStore.loadActiveConnection()
    treeRef.value?.refresh()
    ElMessage.success('刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败: ' + (error.formattedMessage || error.message))
  } finally {
    refreshing.value = false
  }
}

// 节点点击处理
const handleNodeClick = (node, data) => {
  selectedNode.value = node
  selectedNodeData.value = data
}

// 连接更新处理
const handleConnectionUpdated = () => {
  connectionStore.loadConnections()
  connectionStore.loadActiveConnection()
}

// 连接成功处理
const handleConnectionSuccess = () => {
  showConnectionDialog.value = false
  refreshAll()
}

// 刷新处理
const handleRefresh = () => {
  refreshAll()
}

// 创建任务
const handleCreateTask = (tableInfo) => {
  // TODO: 实现任务创建功能
  // 可以打开任务配置对话框或跳转到任务配置页面
  ElMessage.info('任务创建功能待实现')
  console.log('创建任务:', tableInfo)
}

onMounted(async () => {
  await connectionStore.loadConnections()
  await connectionStore.loadActiveConnection()
})
</script>

<style scoped>
.main-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.toolbar {
  height: 50px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.app-title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #67c23a;
  font-size: 14px;
}

.status-icon {
  font-size: 16px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.content-area {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

.tree-panel {
  width: 350px;
  background: #fff;
  border-right: 1px solid #e6e6e6;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.detail-panel {
  flex: 1;
  background: #fff;
  overflow: hidden;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
</style>

