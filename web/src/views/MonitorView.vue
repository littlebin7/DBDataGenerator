<template>
  <div class="monitor-view">
    <el-card shadow="hover" style="margin-bottom: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">系统监控</span>
          <el-button type="primary" :icon="Refresh" @click="loadMetrics" :loading="loading">刷新</el-button>
        </div>
      </template>

      <div>
        <!-- 主机级系统指标 -->
        <el-card shadow="hover" style="margin-bottom: 20px">
          <template #header>
            <span style="font-size: 16px; font-weight: 500">主机性能监控</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #409EFF">{{ systemMetrics.host_cpu_usage?.toFixed(1) || 0 }}%</div>
                  <div style="color: #909399; margin-top: 5px">主机 CPU 使用率</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #67C23A">{{ formatFileSize(systemMetrics.host_memory_used || 0) }}</div>
                  <div style="color: #909399; margin-top: 5px">主机内存使用</div>
                  <div style="font-size: 12px; color: #909399; margin-top: 5px">
                    总内存: {{ formatFileSize(systemMetrics.host_memory_total || 0) }} ({{ systemMetrics.host_memory_usage?.toFixed(1) || 0 }}%)
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #E6A23C">{{ formatFileSize(systemMetrics.host_network_sent || 0) }}/s</div>
                  <div style="color: #909399; margin-top: 5px">网络上传速度</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #F56C6C">{{ formatFileSize(systemMetrics.host_network_recv || 0) }}/s</div>
                  <div style="color: #909399; margin-top: 5px">网络下载速度</div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </el-card>

        <!-- 应用级系统指标 -->
        <el-card shadow="hover" style="margin-bottom: 20px">
          <template #header>
            <span style="font-size: 16px; font-weight: 500">应用性能监控</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #409EFF">{{ systemMetrics.cpu_usage?.toFixed(1) || 0 }}%</div>
                  <div style="color: #909399; margin-top: 5px">应用 CPU 使用率</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #67C23A">{{ formatFileSize(systemMetrics.memory_used || 0) }}</div>
                  <div style="color: #909399; margin-top: 5px">应用内存使用</div>
                  <div style="font-size: 12px; color: #909399; margin-top: 5px">
                    总内存: {{ formatFileSize(systemMetrics.memory_total || 0) }} ({{ systemMetrics.memory_usage?.toFixed(1) || 0 }}%)
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #E6A23C">{{ systemMetrics.goroutines || 0 }}</div>
                  <div style="color: #909399; margin-top: 5px">Goroutines</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card shadow="hover" :body-style="{ padding: '15px' }">
                <div style="text-align: center">
                  <div style="font-size: 24px; font-weight: bold; color: #F56C6C">{{ taskMetrics.length || 0 }}</div>
                  <div style="color: #909399; margin-top: 5px">运行中任务</div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </el-card>

        <!-- 任务性能指标 -->
        <el-card shadow="hover" style="margin-top: 20px">
          <template #header>
            <span style="font-size: 16px; font-weight: 500">任务性能指标</span>
          </template>
          <el-table :data="taskMetrics" stripe border>
            <el-table-column prop="task_id" label="任务ID" width="200" />
            <el-table-column prop="table_name" label="表名" width="150" />
            <el-table-column label="状态" width="100">
              <template #default="scope">
                <StatusTag :status="scope.row.status" />
              </template>
            </el-table-column>
            <el-table-column prop="speed" label="速度" width="120">
              <template #default="scope">
                {{ formatSpeed(scope.row.speed) }}
              </template>
            </el-table-column>
            <el-table-column prop="progress" label="进度" width="150">
              <template #default="scope">
                <el-progress :percentage="Math.round(scope.row.progress)" />
              </template>
            </el-table-column>
            <el-table-column prop="generated_rows" label="已生成" width="120">
              <template #default="scope">
                {{ formatNumber(scope.row.generated_rows) }}
              </template>
            </el-table-column>
            <el-table-column prop="success_rows" label="成功" width="100">
              <template #default="scope">
                <span style="color: #67C23A">{{ formatNumber(scope.row.success_rows) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="failed_rows" label="失败" width="100">
              <template #default="scope">
                <span style="color: #F56C6C">{{ formatNumber(scope.row.failed_rows) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="eta" label="预计剩余时间" width="150">
              <template #default="scope">
                {{ formatDuration(scope.row.eta) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import api from '../api'
import { formatFileSize, formatSpeed, formatDuration, formatNumber } from '../utils/formatters'
import StatusTag from '../components/StatusTag.vue'

const loading = ref(false)
const systemMetrics = ref({})
const taskMetrics = ref([])
let refreshTimer = null

const loadMetrics = async () => {
  loading.value = true
  try {
    const data = await api.getSystemMetrics()
    systemMetrics.value = data.system || {}
    taskMetrics.value = data.tasks || []
  } catch (error) {
    console.error('获取监控数据失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadMetrics()
  // 每5秒自动刷新
  refreshTimer = setInterval(() => {
    loadMetrics()
  }, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.monitor-view {
  padding: 0;
}
</style>

