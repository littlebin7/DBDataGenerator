<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>任务管理</span>
          <el-button @click="loadTasks" :icon="Refresh" :loading="refreshLoading">刷新</el-button>
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
        <el-table-column label="统计信息" width="250">
          <template #default="scope">
            <div style="font-size: 12px">
              <div>成功: <span style="color: #67c23a">{{ scope.row.success_rows || 0 }}</span></div>
              <div>失败: <span style="color: #f56c6c">{{ scope.row.failed_rows || 0 }}</span></div>
              <div v-if="scope.row.speed > 0">速度: {{ formatSpeed(scope.row.speed) }}</div>
              <div v-if="scope.row.eta">预计剩余: {{ formatDuration(scope.row.eta) }}</div>
              <div v-if="scope.row.error || scope.row.status === 'error'" style="margin-top: 5px">
                <el-tooltip 
                  :content="scope.row.error || '任务执行失败'" 
                  placement="top" 
                  effect="dark"
                  :disabled="!scope.row.error"
                >
                  <el-tag type="danger" size="small" style="cursor: pointer">
                    <el-icon style="margin-right: 4px"><Warning /></el-icon>
                    有错误
                  </el-tag>
                </el-tooltip>
              </div>
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
        <el-table-column label="操作" width="400" fixed="right">
          <template #default="scope">
            <el-button 
              v-if="scope.row.status === 'pending' || scope.row.status === 'stopped'"
              size="small" 
              type="success"
              :loading="getActionLoading(scope.row.id, 'start')"
              :disabled="getActionLoading(scope.row.id)"
              @click="startTask(scope.row.id)"
            >
              开始
            </el-button>
            <el-button 
              v-if="scope.row.status === 'error' || scope.row.status === 'stopped'"
              size="small" 
              type="warning"
              :loading="getActionLoading(scope.row.id, 'retry')"
              :disabled="getActionLoading(scope.row.id)"
              @click="retryTask(scope.row.id)"
            >
              重试
            </el-button>
            <el-button 
              v-if="scope.row.status === 'running'"
              size="small" 
              :loading="getActionLoading(scope.row.id, 'pause')"
              :disabled="getActionLoading(scope.row.id)"
              @click="pauseTask(scope.row.id)"
            >
              暂停
            </el-button>
            <el-button 
              v-if="scope.row.status === 'paused'"
              size="small" 
              type="warning"
              :loading="getActionLoading(scope.row.id, 'resume')"
              :disabled="getActionLoading(scope.row.id)"
              @click="resumeTask(scope.row.id)"
            >
              恢复
            </el-button>
            <el-button 
              v-if="scope.row.status === 'running' || scope.row.status === 'paused'"
              size="small" 
              type="danger"
              :loading="getActionLoading(scope.row.id, 'stop')"
              :disabled="getActionLoading(scope.row.id)"
              @click="stopTask(scope.row.id)"
            >
              停止
            </el-button>
            <el-button 
              size="small" 
              :loading="getActionLoading(scope.row.id, 'detail')"
              :disabled="getActionLoading(scope.row.id)"
              @click="viewTaskDetail(scope.row)"
            >
              详情
            </el-button>
            <el-button 
              size="small" 
              type="danger"
              :loading="getActionLoading(scope.row.id, 'delete')"
              :disabled="getActionLoading(scope.row.id)"
              @click="deleteTask(scope.row.id)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="任务详情" width="900px">
      <div v-if="selectedTask">
        <el-descriptions :column="2" border>
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
          <el-descriptions-item label="错误信息" v-if="selectedTask.error || selectedTask.status === 'error'" :span="2">
            <el-alert 
              :title="selectedTask.error || '任务执行失败，请查看详情'" 
              type="error" 
              :closable="false"
              show-icon
              style="word-break: break-all; white-space: pre-wrap;"
            />
          </el-descriptions-item>
        </el-descriptions>

        <!-- 字段规则配置（单独显示） -->
        <div v-if="selectedTask.config && selectedTask.config.field_rules" style="margin-top: 20px">
          <el-divider>字段规则配置</el-divider>
          <el-table 
            :data="selectedTask.config.field_rules" 
            border 
            stripe
            max-height="400"
            style="width: 100%"
          >
            <el-table-column prop="field_name" label="字段名" width="150" />
            <el-table-column label="规则类型" width="150">
              <template #default="scope">
                <el-tag size="small" type="info">{{ getRuleTypeLabel(scope.row.rule_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="规则配置" min-width="300">
              <template #default="scope">
                <div style="font-size: 13px; line-height: 1.6; color: #606266">
                  {{ formatConfig(scope.row.rule_type, scope.row.config) }}
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Connection, CircleCheck, Loading, Warning } from '@element-plus/icons-vue'
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

// 操作按钮的 loading 状态
const actionLoading = ref(new Map()) // 存储每个任务的按钮 loading 状态
const refreshLoading = ref(false) // 刷新按钮 loading

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
  // 保留轮询作为后备（每5秒刷新一次，用于确保任务列表同步）
  refreshTimer = setInterval(() => {
    // 定期刷新任务列表，确保数据同步
    taskStore.loadTasks().catch(err => {
      console.error('刷新任务列表失败:', err)
    })
  }, 5000) // 每5秒刷新一次
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
        console.log(`[WebSocket消息] 任务 ${taskId}:`, data)
        if (data.type === 'task_update' && data.task) {
          // 使用 store 更新任务状态
          taskStore.updateTask(data.task)
          // 如果详情对话框打开且是当前任务，更新详情
          if (showDetailDialog.value && selectedTask.value?.id === data.task.id) {
            selectedTask.value = { ...selectedTask.value, ...data.task }
          }
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

// 获取操作按钮的 loading 状态
const getActionLoading = (taskId, action = null) => {
  const key = action ? `${taskId}_${action}` : taskId
  return actionLoading.value.get(key) || false
}

// 设置操作按钮的 loading 状态
const setActionLoading = (taskId, action, loading) => {
  const key = `${taskId}_${action}`
  if (loading) {
    actionLoading.value.set(key, true)
  } else {
    actionLoading.value.delete(key)
  }
}

const loadTasks = async () => {
  if (refreshLoading.value) return
  refreshLoading.value = true
  try {
    await taskStore.loadTasks()
  } catch (error) {
    ElMessage.error('获取任务列表失败: ' + (error.formattedMessage || error.message))
  } finally {
    refreshLoading.value = false
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
  if (getActionLoading(taskId)) return
  setActionLoading(taskId, 'start', true)
  try {
    await taskStore.startTask(taskId)
    ElMessage.success('任务已启动')
    // 启动后立即建立 WebSocket 连接以获取实时进度
    setTimeout(() => {
      connectWebSocket(taskId)
    }, 500) // 延迟500ms确保任务状态已更新
  } catch (error) {
    ElMessage.error('启动任务失败: ' + (error.formattedMessage || error.message))
  } finally {
    setActionLoading(taskId, 'start', false)
  }
}

const retryTask = async (taskId) => {
  if (getActionLoading(taskId)) return
  setActionLoading(taskId, 'retry', true)
  try {
    await ElMessageBox.confirm('确定要重试这个任务吗？任务进度将被重置。', '确认重试', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await taskStore.retryTask(taskId)
    ElMessage.success('任务已重新启动')
    // 延迟连接 WebSocket，确保任务状态已更新
    setTimeout(() => {
      connectWebSocket(taskId)
    }, 500)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重试任务失败: ' + (error.formattedMessage || error.message))
    }
  } finally {
    setActionLoading(taskId, 'retry', false)
  }
}

const pauseTask = async (taskId) => {
  if (getActionLoading(taskId)) return
  setActionLoading(taskId, 'pause', true)
  try {
    await taskStore.pauseTask(taskId)
    ElMessage.success('任务已暂停')
  } catch (error) {
    ElMessage.error('暂停任务失败: ' + (error.formattedMessage || error.message))
  } finally {
    setActionLoading(taskId, 'pause', false)
  }
}

const resumeTask = async (taskId) => {
  if (getActionLoading(taskId)) return
  setActionLoading(taskId, 'resume', true)
  try {
    await taskStore.resumeTask(taskId)
    ElMessage.success('任务已恢复')
  } catch (error) {
    ElMessage.error('恢复任务失败: ' + (error.formattedMessage || error.message))
  } finally {
    setActionLoading(taskId, 'resume', false)
  }
}

const stopTask = async (taskId) => {
  if (getActionLoading(taskId)) return
  setActionLoading(taskId, 'stop', true)
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
  } finally {
    setActionLoading(taskId, 'stop', false)
  }
}

const updateThreadCount = async (taskId, count) => {
  if (getActionLoading(taskId, 'thread')) return
  setActionLoading(taskId, 'thread', true)
  try {
    await taskStore.setThreadCount(taskId, count)
    ElMessage.success('线程数已更新')
  } catch (error) {
    ElMessage.error('更新线程数失败: ' + (error.formattedMessage || error.message))
    await taskStore.loadTasks() // 恢复原值
  } finally {
    setActionLoading(taskId, 'thread', false)
  }
}

const viewTaskDetail = async (task) => {
  try {
    // 获取最新的任务详情
    const taskDetail = await taskStore.loadTask(task.id)
    selectedTask.value = taskDetail || task
    showDetailDialog.value = true
  } catch (error) {
    console.error('获取任务详情失败:', error)
    // 如果获取失败，使用当前任务信息
    selectedTask.value = task
    showDetailDialog.value = true
    ElMessage.warning('获取任务详情失败，显示缓存信息')
  }
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

// 获取状态类型（用于 Element Plus Tag）
const getStatusType = (status) => {
  return TASK_STATUS_TYPE[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  return TASK_STATUS_TEXT[status] || status || '未知'
}

// 获取规则类型标签（与TemplateView保持一致）
const getRuleTypeLabel = (ruleType) => {
  const labels = {
    'random_string': '随机值',
    'random_number': '随机数字',
    'random_date': '随机日期',
    'fixed': '固定值',
    'increment': '递增',
    'list': '列表选择',
    'regex': '正则表达式',
    'function': '函数',
    'template': '模板',
    'null': '空值',
    'reference': '引用字段',
    'geographic': '地理数据',
    'file': '从文件读取',
    'binary': '二进制/图片',
    'foreign': '外键引用'
  }
  return labels[ruleType] || ruleType
}

// 格式化配置显示（与TemplateView保持一致）
const formatConfig = (ruleType, config) => {
  if (!config) return '-'
  
  // 处理不同的规则类型
  switch (ruleType) {
    case 'random_string':
      return `长度: ${config.min_length || 10}-${config.max_length || 50}, 字符集: ${getCharSetLabel(config.char_set || 'all')}`
    case 'random_number':
      return `范围: ${config.min || 0} ~ ${config.max || 1000}${config.is_int === false ? ' (浮点数)' : ' (整数)'}`
    case 'random_date':
      return `日期范围: ${config.start_date || '无限制'} ~ ${config.end_date || '无限制'}${config.format ? `, 格式: ${config.format}` : ''}`
    case 'fixed':
      return `固定值: ${config.value || '-'}`
    case 'increment':
      return `起始值: ${config.start_value || 1}, 步长: ${config.step || 1}${config.cycle ? ', 循环' : ''}`
    case 'list':
      const values = config.values || []
      return `可选值: ${values.length > 0 ? values.slice(0, 3).join(', ') + (values.length > 3 ? `... (共${values.length}个)` : '') : '-'}`
    case 'regex':
      return `正则表达式: ${config.pattern || '-'}`
    case 'function':
      const funcName = config.func_name || config.funcName || 'UUID'
      const params = config.params || []
      return `函数: ${funcName}${params.length > 0 ? `(${params.join(', ')})` : '()'}`
    case 'template':
      return `模板: ${config.template || '-'}`
    case 'null':
      return `空值概率: ${((config.probability || 0) * 100).toFixed(0)}%`
    case 'reference':
      return `表达式: ${config.expression || '-'}${config.fields && config.fields.length > 0 ? `, 引用字段: ${config.fields.join(', ')}` : ''}`
    case 'geographic':
      return `类型: ${config.type || 'city'}${config.country ? `, 国家: ${config.country}` : ''}`
    case 'file':
      return `文件路径: ${config.file_path || '-'}, 类型: ${config.file_type || 'txt'}, 列索引: ${config.column_index || 0}${config.loop ? ', 循环读取' : ''}`
    case 'binary':
      if (config.mode === 'generate') {
        return `生成模式: ${config.width || 100}x${config.height || 100}, 格式: ${config.format || 'png'}`
      } else {
        return `文件路径: ${config.folder_path || '-'}, 扩展名: ${(config.extensions || []).join(', ') || '全部'}${config.loop ? ', 循环读取' : ''}`
      }
    case 'foreign':
      return `外键表: ${config.foreign_table || '-'}`
    default:
      // 如果无法识别，返回格式化的JSON（但更简洁）
      try {
        const keys = Object.keys(config)
        if (keys.length === 0) return '-'
        return keys.map(key => `${key}: ${config[key]}`).join(', ')
      } catch {
        return JSON.stringify(config)
      }
  }
}

// 获取字符集标签
const getCharSetLabel = (charSet) => {
  const labels = {
    'letters': '字母',
    'numbers': '数字',
    'chinese': '中文',
    'special': '特殊字符',
    'all': '全部'
  }
  return labels[charSet] || charSet
}
</script>
