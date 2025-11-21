<template>
  <div class="rollback-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">数据回滚管理</span>
          <el-button type="primary" :icon="Refresh" @click="loadTasks" :loading="loading">刷新任务列表</el-button>
        </div>
      </template>

      <!-- 任务选择 -->
      <el-form :model="form" label-width="120px" style="margin-bottom: 20px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="选择任务">
              <el-select v-model="form.taskId" placeholder="请选择任务" @change="loadRollbackRecord" style="width: 100%">
                <el-option
                  v-for="task in tasks"
                  :key="task.id"
                  :label="`${task.name} (${task.table})`"
                  :value="task.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 回滚记录 -->
      <el-card v-if="rollbackRecord" shadow="hover" style="margin-bottom: 20px">
        <template #header>
          <span style="font-size: 16px; font-weight: 500">回滚记录</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">{{ rollbackRecord.task_id }}</el-descriptions-item>
          <el-descriptions-item label="表名">{{ rollbackRecord.table_name }}</el-descriptions-item>
          <el-descriptions-item label="开始ID">{{ rollbackRecord.start_id }}</el-descriptions-item>
          <el-descriptions-item label="结束ID">{{ rollbackRecord.end_id }}</el-descriptions-item>
          <el-descriptions-item label="回滚行数">{{ formatNumber(rollbackRecord.rollback_count) }}</el-descriptions-item>
          <el-descriptions-item label="回滚时间">{{ formatTime(rollbackRecord.rollback_time) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 回滚操作 -->
      <el-card v-if="form.taskId" shadow="hover">
        <template #header>
          <span style="font-size: 16px; font-weight: 500">回滚操作</span>
        </template>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="完全回滚" name="full">
            <el-alert
              title="警告"
              description="完全回滚将删除该任务生成的所有数据，此操作不可恢复！"
              type="warning"
              :closable="false"
              style="margin-bottom: 20px"
            />
            <el-button type="danger" @click="rollbackFull" :loading="rollingBack">执行完全回滚</el-button>
          </el-tab-pane>
          <el-tab-pane label="部分回滚" name="partial">
            <el-form :model="partialForm" label-width="150px" style="margin-top: 20px">
              <el-form-item label="回滚方式">
                <el-radio-group v-model="partialForm.type">
                  <el-radio label="row_count">按行数</el-radio>
                  <el-radio label="time_range">按时间范围</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="partialForm.type === 'row_count'" label="回滚行数">
                <el-input-number
                  v-model="partialForm.rowCount"
                  :min="1"
                  :max="1000000"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item v-if="partialForm.type === 'time_range'" label="时间范围">
                <el-input
                  v-model="partialForm.timeRange"
                  placeholder="例如: 1h (1小时), 30m (30分钟), 2h30m (2小时30分钟)"
                  style="width: 100%"
                />
                <div style="font-size: 12px; color: #909399; margin-top: 5px">
                  支持格式: 数字 + 单位 (h=小时, m=分钟, s=秒)
                </div>
              </el-form-item>
              <el-form-item>
                <el-button type="danger" @click="rollbackPartial" :loading="rollingBack">执行部分回滚</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-card>

      <div v-else style="text-align: center; padding: 40px; color: #909399">
        <el-empty description="请先选择一个任务" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'
import { formatTime, formatNumber } from '../utils/formatters'
import { useTaskStore } from '../stores/task'

const taskStore = useTaskStore()

const loading = ref(false)
const rollingBack = ref(false)
const tasks = ref([])
const form = ref({
  taskId: ''
})
const rollbackRecord = ref(null)
const activeTab = ref('full')
const partialForm = ref({
  type: 'row_count',
  rowCount: 100,
  timeRange: '1h'
})

const loadTasks = async () => {
  loading.value = true
  try {
    await taskStore.loadTasks()
    tasks.value = taskStore.allTasks.filter(t => t.status === 'completed' || t.status === 'stopped')
  } catch (error) {
    ElMessage.error('获取任务列表失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
}

const loadRollbackRecord = async () => {
  if (!form.value.taskId) {
    rollbackRecord.value = null
    return
  }

  try {
    const data = await api.getRollbackRecord(form.value.taskId)
    rollbackRecord.value = data
  } catch (error) {
    // 如果没有回滚记录，这是正常的
    rollbackRecord.value = null
  }
}

const rollbackFull = async () => {
  if (!form.value.taskId) {
    ElMessage.warning('请先选择任务')
    return
  }

  try {
    await ElMessageBox.confirm(
      '确定要完全回滚该任务生成的所有数据吗？此操作不可恢复！',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    rollingBack.value = true
    await api.rollbackTask(form.value.taskId)
    ElMessage.success('回滚成功')
    await loadRollbackRecord()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('回滚失败: ' + (error.formattedMessage || error.message))
    }
  } finally {
    rollingBack.value = false
  }
}

const rollbackPartial = async () => {
  if (!form.value.taskId) {
    ElMessage.warning('请先选择任务')
    return
  }

  if (partialForm.value.type === 'row_count' && !partialForm.value.rowCount) {
    ElMessage.warning('请输入回滚行数')
    return
  }

  if (partialForm.value.type === 'time_range' && !partialForm.value.timeRange) {
    ElMessage.warning('请输入时间范围')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要执行部分回滚吗？${partialForm.value.type === 'row_count' ? `将回滚 ${partialForm.value.rowCount} 行数据` : `将回滚 ${partialForm.value.timeRange} 内的数据`}`,
      '确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    rollingBack.value = true
    const options = {}
    if (partialForm.value.type === 'row_count') {
      options.row_count = partialForm.value.rowCount
    } else {
      options.time_range = partialForm.value.timeRange
    }

    await api.rollbackPartial(form.value.taskId, options)
    ElMessage.success('部分回滚成功')
    await loadRollbackRecord()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('回滚失败: ' + (error.formattedMessage || error.message))
    }
  } finally {
    rollingBack.value = false
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.rollback-view {
  padding: 0;
}
</style>

