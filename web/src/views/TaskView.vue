<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>任务管理</span>
          <el-button @click="loadTasks" :icon="Refresh">刷新</el-button>
        </div>
      </template>

      <!-- 搜索和筛选 -->
      <div style="margin-bottom: 20px; display: flex; gap: 10px">
        <el-input
          v-model="searchText"
          placeholder="搜索任务名或表名"
          style="width: 300px"
          clearable
          :prefix-icon="Search"
        />
        <el-select v-model="statusFilter" placeholder="筛选状态" clearable style="width: 150px">
          <el-option label="待开始" value="pending" />
          <el-option label="运行中" value="running" />
          <el-option label="已暂停" value="paused" />
          <el-option label="已完成" value="completed" />
          <el-option label="已停止" value="stopped" />
          <el-option label="错误" value="error" />
        </el-select>
      </div>

      <el-table :data="filteredTasks" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="任务名" width="200" />
        <el-table-column prop="table" label="表名" width="150" />
        <el-table-column prop="connection_id" label="连接ID" width="200" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <StatusTag :status="scope.row.status" />
          </template>
        </el-table-column>
        <el-table-column label="进度" width="200">
          <template #default="scope">
            <div style="display: flex; align-items: center; gap: 8px">
              <el-progress :percentage="Math.round(scope.row.progress)" style="flex: 1" />
              <!-- WebSocket 连接状态指示器 -->
              <el-tooltip 
                v-if="scope.row.status === 'running' || scope.row.status === 'paused'"
                :content="getConnectionStatusText(scope.row.id)"
                placement="top"
              >
                <el-icon 
                  :style="{ color: getConnectionStatusColor(scope.row.id) }"
                  :size="14"
                >
                  <component :is="getConnectionStatusIcon(scope.row.id)" />
                </el-icon>
              </el-tooltip>
            </div>
            <div style="font-size: 12px; color: #909399; margin-top: 5px">
              {{ scope.row.generated_rows || 0 }} / {{ scope.row.total_rows || 0 }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="统计信息" width="200">
          <template #default="scope">
            <div style="font-size: 12px">
              <div>成功: <span style="color: #67c23a">{{ scope.row.success_rows || 0 }}</span></div>
              <div>失败: <span style="color: #f56c6c">{{ scope.row.failed_rows || 0 }}</span></div>
              <div v-if="scope.row.speed > 0">速度: {{ formatSpeed(scope.row.speed) }}</div>
              <div v-if="scope.row.eta">预计剩余: {{ formatDuration(scope.row.eta) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="线程数" width="120">
          <template #default="scope">
            <el-input-number
              v-model="scope.row.thread_count"
              :min="1"
              :max="20"
              size="small"
              @change="updateThreadCount(scope.row.id, scope.row.thread_count)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="scope">
            <el-button 
              v-if="scope.row.status === 'pending' || scope.row.status === 'stopped'"
              size="small" 
              type="success"
              @click="startTask(scope.row.id)"
            >
              开始
            </el-button>
            <el-button 
              v-if="scope.row.status === 'running'"
              size="small" 
              @click="pauseTask(scope.row.id)"
            >
              暂停
            </el-button>
            <el-button 
              v-if="scope.row.status === 'paused'"
              size="small" 
              type="warning"
              @click="resumeTask(scope.row.id)"
            >
              恢复
            </el-button>
            <el-button 
              v-if="scope.row.status === 'running' || scope.row.status === 'paused'"
              size="small" 
              type="danger"
              @click="stopTask(scope.row.id)"
            >
              停止
            </el-button>
            <el-button 
              size="small" 
              @click="viewTaskDetail(scope.row)"
            >
              详情
            </el-button>
            <el-button 
              size="small" 
              type="danger"
              @click="deleteTask(scope.row.id)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="任务详情" width="800px">
      <el-descriptions :column="2" border v-if="selectedTask">
        <el-descriptions-item label="任务ID">{{ selectedTask.id }}</el-descriptions-item>
        <el-descriptions-item label="任务名称">{{ selectedTask.name }}</el-descriptions-item>
        <el-descriptions-item label="表名">{{ selectedTask.table }}</el-descriptions-item>
        <el-descriptions-item label="数据库">{{ selectedTask.database }}</el-descriptions-item>
        <el-descriptions-item label="连接ID">{{ selectedTask.connection_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(selectedTask.status)">
            {{ getStatusText(selectedTask.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总行数">{{ selectedTask.total_rows }}</el-descriptions-item>
        <el-descriptions-item label="已生成">{{ selectedTask.generated_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功">{{ selectedTask.success_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="失败">{{ selectedTask.failed_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="进度">{{ Math.round(selectedTask.progress || 0) }}%</el-descriptions-item>
        <el-descriptions-item label="线程数">{{ selectedTask.thread_count }}</el-descriptions-item>
        <el-descriptions-item label="生成速度" v-if="selectedTask.speed > 0">
          {{ formatSpeed(selectedTask.speed) }}
        </el-descriptions-item>
        <el-descriptions-item label="预计剩余" v-if="selectedTask.eta">
          {{ formatDuration(selectedTask.eta) }}
        </el-descriptions-item>
        <el-descriptions-item label="开始时间" v-if="selectedTask.start_time">
          {{ formatTime(selectedTask.start_time) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间" v-if="selectedTask.end_time">
          {{ formatTime(selectedTask.end_time) }}
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="selectedTask.error" :span="2">
          <el-alert :title="selectedTask.error" type="error" :closable="false" />
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Connection, CircleCheck, Loading } from '@element-plus/icons-vue'
import api from '../api'
import { formatTime, formatDuration, formatSpeed } from '../utils/formatters'
import { TASK_STATUS_TYPE, TASK_STATUS_TEXT, WS_CONNECTION_STATUS, WS_CONNECTION_STATUS_TEXT, WS_CONNECTION_STATUS_COLOR, RECONNECT_CONFIG } from '../utils/constants'
import StatusTag from '../components/StatusTag.vue'
import { useTaskStore } from '../stores/task'

const taskStore = useTaskStore()

const tasks = computed(() => taskStore.allTasks)
const loading = computed(() => taskStore.loading)
const searchText = ref('')
const statusFilter = ref('')
const showDetailDialog = ref(false)
const selectedTask = ref(null)
let refreshTimer = null
let wsConnections = new Map() // WebSocket连接映射
let reconnectTimers = new Map() // 重连定时器映射
const wsConnectionStatus = ref(new Map()) // WebSocket连接状态映射：'connected' | 'disconnected' | 'reconnecting'
const MAX_RECONNECT_ATTEMPTS = RECONNECT_CONFIG.MAX_ATTEMPTS
const RECONNECT_DELAY = RECONNECT_CONFIG.DELAY

const filteredTasks = computed(() => {
  let result = tasks.value
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(task => 
      task.name?.toLowerCase().includes(search) || 
      task.table?.toLowerCase().includes(search)
    )
  }
  if (statusFilter.value) {
    result = result.filter(task => task.status === statusFilter.value)
  }
  return result
})

// 监听 store 中的任务变化，更新 WebSocket 连接
watch(() => taskStore.allTasks, (newTasks) => {
  newTasks.forEach(task => {
    if ((task.status === 'running' || task.status === 'paused')) {
      const existingWs = wsConnections.get(task.id)
      // 如果连接不存在或已关闭，建立新连接
      if (!existingWs || existingWs.readyState !== WebSocket.OPEN) {
        connectWebSocket(task.id)
      }
    } else {
      // 任务已停止，关闭WebSocket连接
      const ws = wsConnections.get(task.id)
      if (ws) {
        if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
          ws.close()
        }
        wsConnections.delete(task.id)
      }
      // 清除重连定时器
      if (reconnectTimers.has(task.id)) {
        clearTimeout(reconnectTimers.get(task.id))
        reconnectTimers.delete(task.id)
      }
      // 更新连接状态
      wsConnectionStatus.value.set(task.id, 'disconnected')
    }
  })
}, { deep: true })

onMounted(async () => {
  await taskStore.loadTasks()
  // 为运行中的任务建立WebSocket连接
  setupWebSocketConnections()
  // 保留轮询作为后备（每10秒刷新一次，仅用于非运行中的任务）
  refreshTimer = setInterval(() => {
    const hasNonRunning = tasks.value.some(t => t.status !== 'running' && t.status !== 'paused')
    if (hasNonRunning) {
      taskStore.loadTasks()
    }
  }, 10000) // 每10秒刷新一次（仅用于非运行中的任务）
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  // 关闭所有WebSocket连接
  wsConnections.forEach((ws, taskId) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close()
    }
  })
  wsConnections.clear()
  
  // 清除所有重连定时器
  reconnectTimers.forEach((timer) => {
    clearTimeout(timer)
  })
  reconnectTimers.clear()
})

// 设置WebSocket连接
const setupWebSocketConnections = () => {
  tasks.value.forEach(task => {
    if ((task.status === 'running' || task.status === 'paused') && !wsConnections.has(task.id)) {
      connectWebSocket(task.id)
    }
  })
}

// 连接WebSocket
const connectWebSocket = (taskId, reconnectAttempt = 0) => {
  // 如果已经存在连接，先关闭
  if (wsConnections.has(taskId)) {
    const existingWs = wsConnections.get(taskId)
    if (existingWs && existingWs.readyState === WebSocket.OPEN) {
      return // 连接已存在且正常，不需要重新连接
    }
    wsConnections.delete(taskId)
  }

  // 清除之前的重连定时器
  if (reconnectTimers.has(taskId)) {
    clearTimeout(reconnectTimers.get(taskId))
    reconnectTimers.delete(taskId)
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/task/${taskId}?task_id=${taskId}`
  
  try {
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log(`WebSocket连接已建立: ${taskId}`)
      // 连接成功后清除重连计数
      reconnectTimers.delete(taskId)
      // 更新连接状态
      wsConnectionStatus.value.set(taskId, WS_CONNECTION_STATUS.CONNECTED)
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'task_update' && data.task) {
          // 使用 store 更新任务状态
          taskStore.updateTask(data.task)
        }
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }

    ws.onerror = (error) => {
      console.error(`WebSocket错误 (${taskId}):`, error)
      wsConnectionStatus.value.set(taskId, 'disconnected')
    }

    ws.onclose = (event) => {
      console.log(`WebSocket连接已关闭: ${taskId}`, event.code, event.reason)
      wsConnections.delete(taskId)
      
      // 如果任务仍在运行，尝试重连
      const task = tasks.value.find(t => t.id === taskId)
      if (task && (task.status === 'running' || task.status === 'paused')) {
        if (reconnectAttempt < MAX_RECONNECT_ATTEMPTS) {
          // 更新连接状态为重连中
          wsConnectionStatus.value.set(taskId, WS_CONNECTION_STATUS.RECONNECTING)
          const delay = RECONNECT_DELAY * (reconnectAttempt + 1) // 指数退避
          console.log(`将在 ${delay}ms 后重连 (${reconnectAttempt + 1}/${MAX_RECONNECT_ATTEMPTS})`)
          const timer = setTimeout(() => {
            connectWebSocket(taskId, reconnectAttempt + 1)
          }, delay)
          reconnectTimers.set(taskId, timer)
        } else {
          console.error(`WebSocket重连失败，已达到最大重连次数: ${taskId}`)
          wsConnectionStatus.value.set(taskId, WS_CONNECTION_STATUS.DISCONNECTED)
          ElMessage.warning(`任务 ${task.name} 的实时更新连接失败，请刷新页面`)
        }
      } else {
        // 任务已停止，更新状态为断开
        wsConnectionStatus.value.set(taskId, 'disconnected')
      }
    }

    wsConnections.set(taskId, ws)
  } catch (error) {
    console.error(`创建WebSocket连接失败 (${taskId}):`, error)
    // 如果创建连接失败，也尝试重连
    const task = tasks.value.find(t => t.id === taskId)
    if (task && (task.status === 'running' || task.status === 'paused')) {
      if (reconnectAttempt < MAX_RECONNECT_ATTEMPTS) {
        const delay = RECONNECT_DELAY * (reconnectAttempt + 1)
        const timer = setTimeout(() => {
          connectWebSocket(taskId, reconnectAttempt + 1)
        }, delay)
        reconnectTimers.set(taskId, timer)
      }
    }
  }
}

const loadTasks = async () => {
  try {
    await taskStore.loadTasks()
  } catch (error) {
    ElMessage.error('获取任务列表失败: ' + (error.formattedMessage || error.message))
  }
}

// 使用工具函数，移除重复代码

// WebSocket 连接状态相关函数
const getConnectionStatus = (taskId) => {
  return wsConnectionStatus.value.get(taskId) || WS_CONNECTION_STATUS.DISCONNECTED
}

const getConnectionStatusColor = (taskId) => {
  const status = getConnectionStatus(taskId)
  return WS_CONNECTION_STATUS_COLOR[status] || '#909399'
}

const getConnectionStatusIcon = (taskId) => {
  const status = getConnectionStatus(taskId)
  switch (status) {
    case WS_CONNECTION_STATUS.CONNECTED:
      return CircleCheck
    case WS_CONNECTION_STATUS.RECONNECTING:
      return Loading
    case WS_CONNECTION_STATUS.DISCONNECTED:
      return Connection
    default:
      return Connection
  }
}

const getConnectionStatusText = (taskId) => {
  const status = getConnectionStatus(taskId)
  return WS_CONNECTION_STATUS_TEXT[status] || '未知状态'
}

const startTask = async (taskId) => {
  try {
    await taskStore.startTask(taskId)
    ElMessage.success('任务已启动')
  } catch (error) {
    ElMessage.error('启动任务失败: ' + (error.formattedMessage || error.message))
  }
}

const pauseTask = async (taskId) => {
  try {
    await taskStore.pauseTask(taskId)
    ElMessage.success('任务已暂停')
  } catch (error) {
    ElMessage.error('暂停任务失败: ' + (error.formattedMessage || error.message))
  }
}

const resumeTask = async (taskId) => {
  try {
    await taskStore.resumeTask(taskId)
    ElMessage.success('任务已恢复')
  } catch (error) {
    ElMessage.error('恢复任务失败: ' + (error.formattedMessage || error.message))
  }
}

const stopTask = async (taskId) => {
  try {
    await ElMessageBox.confirm('确定要停止这个任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await taskStore.stopTask(taskId)
    ElMessage.success('任务已停止')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('停止任务失败: ' + (error.formattedMessage || error.message))
    }
  }
}

const updateThreadCount = async (taskId, count) => {
  try {
    await taskStore.setThreadCount(taskId, count)
    ElMessage.success('线程数已更新')
  } catch (error) {
    ElMessage.error('更新线程数失败: ' + (error.formattedMessage || error.message))
    await taskStore.loadTasks() // 恢复原值
  }
}

const viewTaskDetail = (task) => {
  selectedTask.value = task
  showDetailDialog.value = true
}

const deleteTask = async (taskId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await taskStore.deleteTask(taskId)
    ElMessage.success('任务已删除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败: ' + (error.formattedMessage || error.message))
    }
  }
}
</script>
