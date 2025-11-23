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
            <el-button 
              type="danger" 
              :disabled="selectedHistoryIds.length === 0 || batchDeleteLoading"
              :loading="batchDeleteLoading"
              @click="batchDeleteHistory"
              style="margin-right: 10px"
            >
              批量删除 ({{ selectedHistoryIds.length }})
            </el-button>
            <el-button :icon="Refresh" @click="loadHistory" :loading="refreshLoading">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table 
        :data="history" 
        v-loading="loading" 
        border 
        stripe 
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
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
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button 
              size="small" 
              :loading="getActionLoading(scope.row.id, 'detail')"
              :disabled="getActionLoading(scope.row.id)"
              @click="viewDetail(scope.row)"
            >
              查看详情
            </el-button>
            <el-button 
              size="small" 
              type="danger"
              :loading="getActionLoading(scope.row.id, 'delete')"
              :disabled="getActionLoading(scope.row.id)"
              @click="deleteHistory(scope.row.id)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="任务执行详情" width="900px">
      <el-descriptions :column="2" border v-if="selectedHistory">
        <el-descriptions-item label="任务ID">{{ selectedHistory.task_id }}</el-descriptions-item>
        <el-descriptions-item label="任务名称">{{ selectedHistory.task_name }}</el-descriptions-item>
        <el-descriptions-item label="表名">{{ selectedHistory.table_name }}</el-descriptions-item>
        <el-descriptions-item label="数据库">{{ selectedHistory.database }}</el-descriptions-item>
        <el-descriptions-item label="连接ID">{{ selectedHistory.connection_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <StatusTag :status="selectedHistory.status" />
        </el-descriptions-item>
        <el-descriptions-item label="总行数">{{ selectedHistory.total_rows }}</el-descriptions-item>
        <el-descriptions-item label="已生成">{{ selectedHistory.generated_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功">{{ selectedHistory.success_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="失败">{{ selectedHistory.failed_rows || 0 }}</el-descriptions-item>
        <el-descriptions-item label="开始时间" v-if="selectedHistory.start_time">
          {{ formatTime(selectedHistory.start_time) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间" v-if="selectedHistory.end_time">
          {{ formatTime(selectedHistory.end_time) }}
        </el-descriptions-item>
        <el-descriptions-item label="持续时间" v-if="selectedHistory.duration">
          {{ formatDuration(selectedHistory.duration) }}
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="selectedHistory.error_message" :span="2">
          <el-alert 
            :title="selectedHistory.error_message" 
            type="error" 
            :closable="false"
            show-icon
            style="word-break: break-all; white-space: pre-wrap;"
          />
        </el-descriptions-item>
      </el-descriptions>

      <!-- 字段规则配置（单独显示） -->
      <div v-if="selectedHistory.config && selectedHistory.config.field_rules" style="margin-top: 20px">
        <el-divider>字段规则配置</el-divider>
        <el-table 
          :data="selectedHistory.config.field_rules" 
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
    </el-dialog>

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
import { ref, onMounted, onActivated } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import api from '../api'
import { formatTime, formatDuration } from '../utils/formatters'
import StatusTag from '../components/StatusTag.vue'

const route = useRoute()
const history = ref([])
const loading = ref(false)
const statusFilter = ref('')
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const showDetailDialog = ref(false)
const selectedHistory = ref(null)
let isFirstLoad = true

// 批量选择
const selectedHistoryIds = ref([])
const batchDeleteLoading = ref(false)

// 操作按钮的 loading 状态
const actionLoading = ref(new Map()) // 存储每个历史记录的按钮 loading 状态
const refreshLoading = ref(false) // 刷新按钮 loading

// 获取操作按钮的 loading 状态
const getActionLoading = (historyId, action = null) => {
  const key = action ? `${historyId}_${action}` : historyId
  return actionLoading.value.get(key) || false
}

// 设置操作按钮的 loading 状态
const setActionLoading = (historyId, action, loading) => {
  const key = `${historyId}_${action}`
  if (loading) {
    actionLoading.value.set(key, true)
  } else {
    actionLoading.value.delete(key)
  }
}

onMounted(() => {
  loadHistory()
  isFirstLoad = false
})

// 只在首次激活时加载，避免从其他页面切换过来时自动刷新
onActivated(() => {
  if (isFirstLoad) {
    isFirstLoad = false
    return
  }
  // 不自动刷新，让用户手动点击刷新按钮
})

const loadHistory = async () => {
  if (refreshLoading.value) return
  refreshLoading.value = true
  loading.value = true
  try {
    const offset = (currentPage.value - 1) * pageSize.value
    const response = await api.getTaskHistory(pageSize.value, offset, statusFilter.value)
    // 确保 history 是数组，即使后端返回 null
    history.value = Array.isArray(response.history) ? response.history : (response.history || [])
    total.value = response.total || 0
  } catch (error) {
    ElMessage.error('加载历史记录失败: ' + (error.formattedMessage || error.message))
    history.value = []
    total.value = 0
  } finally {
    loading.value = false
    refreshLoading.value = false
  }
}

const deleteHistory = async (historyId) => {
  if (getActionLoading(historyId)) return
  setActionLoading(historyId, 'delete', true)
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
  } finally {
    setActionLoading(historyId, 'delete', false)
  }
}

// 处理表格选择变化
const handleSelectionChange = (selection) => {
  selectedHistoryIds.value = selection.map(item => item.id)
}

// 批量删除历史记录
const batchDeleteHistory = async () => {
  if (selectedHistoryIds.value.length === 0) {
    ElMessage.warning('请选择要删除的历史记录')
    return
  }

  if (batchDeleteLoading.value) return
  batchDeleteLoading.value = true

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedHistoryIds.value.length} 条历史记录吗？`,
      '确认批量删除',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )

    const response = await api.batchDeleteTaskHistory(selectedHistoryIds.value)
    
    const successCount = response.count || 0
    const errorCount = response.errors ? response.errors.length : 0

    if (errorCount === 0) {
      ElMessage.success(`成功删除 ${successCount} 条历史记录`)
    } else {
      ElMessage.warning(`成功删除 ${successCount} 条，失败 ${errorCount} 条`)
      if (response.errors && response.errors.length > 0) {
        console.error('批量删除错误:', response.errors)
      }
    }

    // 清空选择
    selectedHistoryIds.value = []
    // 重新加载列表
    await loadHistory()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败: ' + (error.formattedMessage || error.message))
    }
  } finally {
    batchDeleteLoading.value = false
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

const viewDetail = (item) => {
  if (getActionLoading(item.id, 'detail')) return
  setActionLoading(item.id, 'detail', true)
  selectedHistory.value = item
  showDetailDialog.value = true
  // 查看详情是同步操作，立即释放
  setTimeout(() => {
    setActionLoading(item.id, 'detail', false)
  }, 100)
}

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

<style scoped>
.task-history-view {
  padding: 20px;
  height: 100%;
  overflow-y: auto;
}
</style>

