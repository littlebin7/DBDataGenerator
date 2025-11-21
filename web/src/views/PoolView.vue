<template>
  <div class="pool-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">连接池监控</span>
          <div style="display: flex; gap: 10px">
            <el-select v-model="selectedConnectionId" placeholder="选择连接" clearable style="width: 200px" @change="loadPoolStatus">
              <el-option
                v-for="conn in connections"
                :key="conn.id"
                :label="conn.name"
                :value="conn.id"
              />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="loadPoolStatus" :loading="loading">刷新</el-button>
          </div>
        </div>
      </template>

      <div v-loading="loading">
        <div v-if="!selectedConnectionId && poolStatuses.length === 0" style="text-align: center; padding: 40px; color: #909399">
          <el-empty description="请选择连接或查看所有连接池状态" />
        </div>

        <!-- 单个连接池状态 -->
        <div v-if="selectedConnectionId && poolStatus">
          <el-card shadow="hover" style="margin-bottom: 20px">
            <template #header>
              <span style="font-size: 16px; font-weight: 500">连接池状态</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="连接ID">{{ poolStatus.connection_id }}</el-descriptions-item>
              <el-descriptions-item label="数据库类型">{{ poolStatus.db_type }}</el-descriptions-item>
              <el-descriptions-item label="最大连接数">{{ poolStatus.max_connections }}</el-descriptions-item>
              <el-descriptions-item label="当前连接数">{{ poolStatus.current_connections }}</el-descriptions-item>
              <el-descriptions-item label="活跃连接数">{{ poolStatus.active_connections }}</el-descriptions-item>
              <el-descriptions-item label="空闲连接数">{{ poolStatus.idle_connections }}</el-descriptions-item>
              <el-descriptions-item label="等待连接数">{{ poolStatus.waiting_connections }}</el-descriptions-item>
              <el-descriptions-item label="连接使用率">
                <el-progress :percentage="poolStatus.usage_percentage" :color="getUsageColor(poolStatus.usage_percentage)" />
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 连接池建议 -->
          <el-card shadow="hover" v-if="recommendations && recommendations.length > 0">
            <template #header>
              <span style="font-size: 16px; font-weight: 500">优化建议</span>
            </template>
            <el-alert
              v-for="(rec, index) in recommendations"
              :key="index"
              :title="rec.title"
              :description="rec.description"
              :type="rec.type || 'info'"
              :closable="false"
              style="margin-bottom: 10px"
            />
          </el-card>
        </div>

        <!-- 所有连接池状态 -->
        <el-table v-if="!selectedConnectionId && poolStatuses.length > 0" :data="poolStatuses" stripe border>
          <el-table-column prop="connection_id" label="连接ID" width="200" />
          <el-table-column prop="db_type" label="数据库类型" width="120" />
          <el-table-column prop="max_connections" label="最大连接数" width="120" />
          <el-table-column prop="current_connections" label="当前连接数" width="120" />
          <el-table-column prop="active_connections" label="活跃连接数" width="120" />
          <el-table-column prop="idle_connections" label="空闲连接数" width="120" />
          <el-table-column prop="waiting_connections" label="等待连接数" width="120" />
          <el-table-column label="连接使用率" width="150">
            <template #default="scope">
              <el-progress :percentage="scope.row.usage_percentage" :color="getUsageColor(scope.row.usage_percentage)" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const connectionStore = useConnectionStore()

const loading = ref(false)
const connections = ref([])
const selectedConnectionId = ref('')
const poolStatus = ref(null)
const recommendations = ref([])
const poolStatuses = ref([])

const loadConnections = async () => {
  try {
    await connectionStore.loadConnections()
    connections.value = connectionStore.allConnections
  } catch (error) {
    console.error('获取连接列表失败:', error)
  }
}

const loadPoolStatus = async () => {
  loading.value = true
  try {
    if (selectedConnectionId.value) {
      const data = await api.getPoolStatus(selectedConnectionId.value)
      poolStatus.value = data.status
      recommendations.value = data.recommendations || []
      poolStatuses.value = []
    } else {
      const data = await api.getPoolStatus()
      poolStatuses.value = data.pools || []
      poolStatus.value = null
      recommendations.value = []
    }
  } catch (error) {
    console.error('获取连接池状态失败:', error)
  } finally {
    loading.value = false
  }
}

const getUsageColor = (percentage) => {
  if (percentage < 50) return '#67C23A'
  if (percentage < 80) return '#E6A23C'
  return '#F56C6C'
}

onMounted(async () => {
  await loadConnections()
  await loadPoolStatus()
})
</script>

<style scoped>
.pool-view {
  padding: 0;
}
</style>

