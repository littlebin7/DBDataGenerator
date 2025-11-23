<template>
  <div class="schedule-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">定时任务管理</span>
          <el-button type="primary" :icon="Refresh" @click="loadSchedules" :loading="loading">刷新</el-button>
        </div>
      </template>

      <div v-loading="loading">
        <el-table :data="schedules" stripe border v-if="schedules.length > 0">
          <el-table-column prop="task_id" label="任务ID" width="200" />
          <el-table-column prop="task_name" label="任务名称" width="200" />
          <el-table-column prop="cron_expr" label="Cron 表达式" width="150" />
          <el-table-column prop="next_run_time" label="下次执行时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.next_run_time) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.enabled ? 'success' : 'info'">
                {{ scope.row.enabled ? '已启用' : '已禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="scope">
              <div style="display: flex; gap: 4px;">
                <el-button
                  v-if="scope.row.enabled"
                  type="warning"
                  size="small"
                  @click="disableSchedule(scope.row.task_id)"
                >
                  禁用
                </el-button>
                <el-button
                  v-else
                  type="success"
                  size="small"
                  @click="enableSchedule(scope.row.task_id)"
                >
                  启用
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="removeSchedule(scope.row.task_id)"
                >
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无定时任务" />
      </div>
    </el-card>

    <!-- 设置定时任务对话框 -->
    <el-dialog v-model="showScheduleDialog" title="设置定时任务" width="500px">
      <el-form :model="scheduleForm" label-width="120px">
        <el-form-item label="任务ID">
          <el-input v-model="scheduleForm.taskId" disabled />
        </el-form-item>
        <el-form-item label="Cron 表达式" required>
          <el-input v-model="scheduleForm.cronExpr" placeholder="例如: 0 0 * * * (每天0点执行)" />
          <div style="font-size: 12px; color: #909399; margin-top: 5px">
            Cron 表达式格式: 秒 分 时 日 月 周
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showScheduleDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSchedule">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'
import { formatTime } from '../utils/formatters'

const loading = ref(false)
const schedules = ref([])
const showScheduleDialog = ref(false)
const scheduleForm = ref({
  taskId: '',
  cronExpr: ''
})

const loadSchedules = async () => {
  loading.value = true
  try {
    const data = await api.getAllSchedules()
    schedules.value = data.schedules || []
  } catch (error) {
    ElMessage.error('获取定时任务列表失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
}

const enableSchedule = async (taskId) => {
  try {
    await api.enableSchedule(taskId)
    ElMessage.success('定时任务已启用')
    await loadSchedules()
  } catch (error) {
    ElMessage.error('启用失败: ' + (error.formattedMessage || error.message))
  }
}

const disableSchedule = async (taskId) => {
  try {
    await api.disableSchedule(taskId)
    ElMessage.success('定时任务已禁用')
    await loadSchedules()
  } catch (error) {
    ElMessage.error('禁用失败: ' + (error.formattedMessage || error.message))
  }
}

const removeSchedule = async (taskId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个定时任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await api.removeSchedule(taskId)
    ElMessage.success('定时任务已删除')
    await loadSchedules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.formattedMessage || error.message))
    }
  }
}

const saveSchedule = async () => {
  if (!scheduleForm.value.cronExpr) {
    ElMessage.warning('请输入 Cron 表达式')
    return
  }

  try {
    await api.scheduleTask(scheduleForm.value.taskId, scheduleForm.value.cronExpr)
    ElMessage.success('定时任务设置成功')
    showScheduleDialog.value = false
    await loadSchedules()
  } catch (error) {
    ElMessage.error('设置失败: ' + (error.formattedMessage || error.message))
  }
}

onMounted(() => {
  loadSchedules()
})
</script>

<style scoped>
.schedule-view {
  padding: 0;
}
</style>

