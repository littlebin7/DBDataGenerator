<template>
  <div class="quality-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">数据质量检查</span>
          <el-button type="primary" :icon="Search" @click="checkQuality" :loading="loading" :disabled="!canCheck">开始检查</el-button>
        </div>
      </template>

      <!-- 选择连接和表 -->
      <el-form :model="form" label-width="120px" style="margin-bottom: 20px">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="连接">
              <el-select v-model="form.connectionId" placeholder="请选择连接" @change="onConnectionChange" style="width: 100%">
                <el-option
                  v-for="conn in connections"
                  :key="conn.id"
                  :label="conn.name"
                  :value="conn.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="数据库">
              <el-select v-model="form.database" placeholder="请选择数据库" @change="onDatabaseChange" :disabled="!form.connectionId" style="width: 100%">
                <el-option
                  v-for="db in databases"
                  :key="db"
                  :label="db"
                  :value="db"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="表名">
              <el-select v-model="form.table" placeholder="请选择表" :disabled="!form.database" style="width: 100%">
                <el-option
                  v-for="table in tables"
                  :key="table"
                  :label="table"
                  :value="table"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 检查结果 -->
      <div v-if="report" v-loading="loading">
        <el-card shadow="hover">
          <template #header>
            <span style="font-size: 16px; font-weight: 500">检查结果</span>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="表名">{{ report.table_name }}</el-descriptions-item>
            <el-descriptions-item label="总行数">{{ formatNumber(report.total_rows) }}</el-descriptions-item>
            <el-descriptions-item label="空值检查">
              <el-tag :type="report.null_check?.passed ? 'success' : 'danger'">
                {{ report.null_check?.passed ? '通过' : '失败' }}
              </el-tag>
              <span v-if="report.null_check?.failed_fields" style="margin-left: 10px; color: #F56C6C">
                失败字段: {{ report.null_check.failed_fields.join(', ') }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="唯一性检查">
              <el-tag :type="report.uniqueness_check?.passed ? 'success' : 'danger'">
                {{ report.uniqueness_check?.passed ? '通过' : '失败' }}
              </el-tag>
              <span v-if="report.uniqueness_check?.failed_fields" style="margin-left: 10px; color: #F56C6C">
                失败字段: {{ report.uniqueness_check.failed_fields.join(', ') }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="外键检查">
              <el-tag :type="report.foreign_key_check?.passed ? 'success' : 'danger'">
                {{ report.foreign_key_check?.passed ? '通过' : '失败' }}
              </el-tag>
              <span v-if="report.foreign_key_check?.failed_relations" style="margin-left: 10px; color: #F56C6C">
                失败关系: {{ report.foreign_key_check.failed_relations?.length || 0 }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="检查时间">{{ formatTime(report.check_time) }}</el-descriptions-item>
          </el-descriptions>

          <!-- 详细结果 -->
          <div v-if="report.null_check?.details" style="margin-top: 20px">
            <el-card shadow="hover">
              <template #header>
                <span style="font-size: 14px; font-weight: 500">空值检查详情</span>
              </template>
              <el-table :data="Object.entries(report.null_check.details)" border>
                <el-table-column prop="0" label="字段名" width="200" />
                <el-table-column prop="1" label="空值数量" width="150">
                  <template #default="scope">
                    {{ formatNumber(scope.row[1]) }}
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </div>
        </el-card>
      </div>

      <div v-else-if="!loading" style="text-align: center; padding: 40px; color: #909399">
        <el-empty description="请选择连接、数据库和表后开始检查" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '../api'
import { formatNumber, formatTime } from '../utils/formatters'

const loading = ref(false)
const connections = ref([])
const databases = ref([])
const tables = ref([])
const form = ref({
  connectionId: '',
  database: '',
  table: ''
})
const report = ref(null)

const canCheck = computed(() => {
  return form.value.connectionId && form.value.database && form.value.table
})

const loadConnections = async () => {
  try {
    const data = await api.getConnections()
    connections.value = data.connections || []
  } catch (error) {
    console.error('获取连接列表失败:', error)
  }
}

const onConnectionChange = async () => {
  form.value.database = ''
  form.value.table = ''
  databases.value = []
  tables.value = []
  report.value = null

  if (!form.value.connectionId) return

  try {
    const data = await api.getDatabases(form.value.connectionId)
    databases.value = data.databases || []
  } catch (error) {
    ElMessage.error('获取数据库列表失败: ' + (error.formattedMessage || error.message))
  }
}

const onDatabaseChange = async () => {
  form.value.table = ''
  tables.value = []
  report.value = null

  if (!form.value.database || !form.value.connectionId) return

  try {
    const data = await api.getTables(form.value.database, form.value.connectionId)
    tables.value = data.tables || []
  } catch (error) {
    ElMessage.error('获取表列表失败: ' + (error.formattedMessage || error.message))
  }
}

const checkQuality = async () => {
  if (!canCheck.value) {
    ElMessage.warning('请先选择连接、数据库和表')
    return
  }

  loading.value = true
  try {
    const data = await api.checkDataQuality(form.value.connectionId, form.value.database, form.value.table)
    report.value = data
    ElMessage.success('数据质量检查完成')
  } catch (error) {
    ElMessage.error('检查失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
}

loadConnections()
</script>

<style scoped>
.quality-view {
  padding: 0;
}
</style>

