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
            <el-tag :type="getStatusType(scope.row.status)">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="200">
          <template #default="scope">
            <el-progress :percentage="Math.round(scope.row.progress)" />
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
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import api from '../api'

const tasks = ref([])
const loading = ref(false)
const searchText = ref('')
const statusFilter = ref('')
const showDetailDialog = ref(false)
const selectedTask = ref(null)
let refreshTimer = null

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

onMounted(async () => {
  await loadTasks()
  // 自动刷新运行中的任务
  refreshTimer = setInterval(() => {
    const hasRunning = tasks.value.some(t => t.status === 'running' || t.status === 'paused')
    if (hasRunning) {
      loadTasks()
    }
  }, 2000) // 每2秒刷新一次
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

const loadTasks = async () => {
  loading.value = true
  try {
    const response = await api.getTasks()
    tasks.value = response.tasks || []
  } catch (error) {
    ElMessage.error('获取任务列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

const getStatusType = (status) => {
  const map = {
    pending: 'info',
    running: 'success',
    paused: 'warning',
    completed: 'success',
    stopped: 'info',
    error: 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = {
    pending: '待开始',
    running: '运行中',
    paused: '已暂停',
    completed: '已完成',
    stopped: '已停止',
    error: '错误'
  }
  return map[status] || status
}

const formatSpeed = (speed) => {
  if (speed < 1000) {
    return `${Math.round(speed)} 行/秒`
  } else if (speed < 1000000) {
    return `${(speed / 1000).toFixed(1)}K 行/秒`
  } else {
    return `${(speed / 1000000).toFixed(2)}M 行/秒`
  }
}

const formatDuration = (seconds) => {
  if (seconds < 60) {
    return `${Math.round(seconds)}秒`
  } else if (seconds < 3600) {
    return `${Math.round(seconds / 60)}分钟`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.round((seconds % 3600) / 60)
    return `${hours}小时${minutes}分钟`
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

const startTask = async (taskId) => {
  try {
    await api.startTask(taskId)
    ElMessage.success('任务已启动')
    await loadTasks()
  } catch (error) {
    ElMessage.error('启动任务失败: ' + (error.response?.data?.error || error.message))
  }
}

const pauseTask = async (taskId) => {
  try {
    await api.pauseTask(taskId)
    ElMessage.success('任务已暂停')
    await loadTasks()
  } catch (error) {
    ElMessage.error('暂停任务失败: ' + (error.response?.data?.error || error.message))
  }
}

const resumeTask = async (taskId) => {
  try {
    await api.resumeTask(taskId)
    ElMessage.success('任务已恢复')
    await loadTasks()
  } catch (error) {
    ElMessage.error('恢复任务失败: ' + (error.response?.data?.error || error.message))
  }
}

const stopTask = async (taskId) => {
  try {
    await ElMessageBox.confirm('确定要停止这个任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await api.stopTask(taskId)
    ElMessage.success('任务已停止')
    await loadTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('停止任务失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

const updateThreadCount = async (taskId, count) => {
  try {
    await api.setThreadCount(taskId, count)
    ElMessage.success('线程数已更新')
  } catch (error) {
    ElMessage.error('更新线程数失败: ' + (error.response?.data?.error || error.message))
    await loadTasks() // 恢复原值
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
    await api.deleteTask(taskId)
    ElMessage.success('任务已删除')
    await loadTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败: ' + (error.response?.data?.error || error.message))
    }
  }
}
</script>
