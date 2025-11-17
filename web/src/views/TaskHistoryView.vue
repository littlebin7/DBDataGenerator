<template>
  <div class="task-history-view">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>任务历史记录</span>
          <div>
            <el-select v-model="statusFilter" placeholder="筛选状态" clearable style="width: 150px; margin-right: 10px" @change="loadHistory">
              <el-option label="全部" value="" />
              <el-option label="待开始" value="pending" />
              <el-option label="运行中" value="running" />
              <el-option label="已暂停" value="paused" />
              <el-option label="已完成" value="completed" />
              <el-option label="已停止" value="stopped" />
              <el-option label="错误" value="error" />
            </el-select>
            <el-button :icon="Refresh" @click="loadHistory">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table :data="history" v-loading="loading" border stripe style="width: 100%">
        <el-table-column prop="task_name" label="任务名称" min-width="150" />
        <el-table-column prop="table_name" label="表名" width="150" />
        <el-table-column prop="database" label="数据库" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <StatusTag :status="scope.row.status" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="进度" width="150">
          <template #default="scope">
            <el-progress 
              :percentage="scope.row.total_rows > 0 ? Math.round((scope.row.generated_rows / scope.row.total_rows) * 100) : 0"
              :status="scope.row.status === 'completed' ? 'success' : scope.row.status === 'error' ? 'exception' : undefined"
            />
          </template>
        </el-table-column>
        <el-table-column label="生成统计" width="200">
          <template #default="scope">
            <div style="font-size: 12px">
              <div>总数: {{ scope.row.total_rows }}</div>
              <div>成功: <span style="color: #67c23a">{{ scope.row.success_rows }}</span></div>
              <div>失败: <span style="color: #f56c6c">{{ scope.row.failed_rows }}</span></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="thread_count" label="线程数" width="80" />
        <el-table-column label="开始时间" width="160">
          <template #default="scope">
            {{ scope.row.start_time ? formatTime(scope.row.start_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" width="160">
          <template #default="scope">
            {{ scope.row.end_time ? formatTime(scope.row.end_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="持续时间" width="120">
          <template #default="scope">
            {{ scope.row.duration ? formatDuration(scope.row.duration) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.error_message" style="color: #f56c6c">
              {{ scope.row.error_message }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button size="small" type="danger" @click="deleteHistory(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 20px; display: flex; justify-content: space-between; align-items: center">
        <div>
          共 {{ total }} 条记录
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import api from '../api'
import { formatTime, formatDuration } from '../utils/formatters'
import StatusTag from '../components/StatusTag.vue'

const history = ref([])
const loading = ref(false)
const statusFilter = ref('')
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)

onMounted(() => {
  loadHistory()
})

const loadHistory = async () => {
  loading.value = true
  try {
    const offset = (currentPage.value - 1) * pageSize.value
    const response = await api.getTaskHistory(pageSize.value, offset, statusFilter.value)
    history.value = response.history || []
    total.value = response.total || 0
  } catch (error) {
    ElMessage.error('加载历史记录失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
}

const deleteHistory = async (historyId) => {
  try {
    await ElMessageBox.confirm('确定要删除这条历史记录吗？', '确认删除', {
      type: 'warning'
    })
    
    await api.deleteTaskHistory(historyId)
    ElMessage.success('删除成功')
    await loadHistory()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.formattedMessage || error.message))
    }
  }
}

// 使用工具函数，移除重复代码

const handleSizeChange = () => {
  currentPage.value = 1
  loadHistory()
}

const handlePageChange = () => {
  loadHistory()
}
</script>

<style scoped>
.task-history-view {
  padding: 20px;
  height: 100%;
  overflow-y: auto;
}
</style>

