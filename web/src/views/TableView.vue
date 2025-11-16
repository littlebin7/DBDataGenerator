<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>表列表</span>
          <div>
            <el-select 
              v-model="selectedDatabase" 
              placeholder="选择数据库" 
              style="width: 200px; margin-right: 10px"
              @change="loadTables"
            >
              <el-option 
                v-for="db in databases" 
                :key="db" 
                :label="db" 
                :value="db" 
              />
            </el-select>
            <el-button @click="loadTables" :icon="Refresh">刷新</el-button>
          </div>
        </div>
      </template>
      
      <el-input
        v-model="searchText"
        placeholder="搜索表名"
        style="margin-bottom: 20px"
        clearable
        :prefix-icon="Search"
      />

      <el-table :data="filteredTables" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="表名" />
        <el-table-column label="操作" width="250">
          <template #default="scope">
            <el-button size="small" @click="viewSchema(scope.row.name)">查看结构</el-button>
            <el-button type="primary" size="small" @click="createTask(scope.row.name)">创建任务</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 表结构对话框 -->
    <el-dialog v-model="showSchemaDialog" title="表结构" width="900px">
      <el-table :data="tableSchema?.fields" border style="width: 100%">
        <el-table-column prop="name" label="字段名" width="150" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="go_type" label="Go类型" width="100" />
        <el-table-column label="约束" width="200">
          <template #default="scope">
            <el-tag v-if="scope.row.is_primary_key" type="danger" size="small" style="margin-right: 5px">主键</el-tag>
            <el-tag v-if="scope.row.is_foreign_key" type="warning" size="small" style="margin-right: 5px">外键</el-tag>
            <el-tag v-if="scope.row.is_unique" type="info" size="small" style="margin-right: 5px">唯一</el-tag>
            <el-tag v-if="!scope.row.is_nullable" type="success" size="small">非空</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="max_length" label="最大长度" width="100" />
        <el-table-column prop="default_value" label="默认值" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import api from '../api'

const router = useRouter()
const tables = ref([])
const databases = ref([])
const selectedDatabase = ref('')
const searchText = ref('')
const loading = ref(false)
const showSchemaDialog = ref(false)
const tableSchema = ref(null)

const filteredTables = computed(() => {
  if (!searchText.value) {
    return tables.value
  }
  return tables.value.filter(table => 
    table.name.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

onMounted(async () => {
  await loadDatabases()
  await loadTables()
})

const loadDatabases = async () => {
  try {
    const activeConn = await api.getActiveConnection()
    if (activeConn && activeConn.id && activeConn.config?.database) {
      selectedDatabase.value = activeConn.config.database
      databases.value = [activeConn.config.database]
    }
  } catch (error) {
    console.error('获取数据库列表失败:', error)
  }
}

const loadTables = async () => {
  loading.value = true
  try {
    const response = await api.getTables(selectedDatabase.value || '')
    tables.value = response.tables.map(name => ({ name }))
  } catch (error) {
    ElMessage.error('获取表列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

const viewSchema = async (tableName) => {
  try {
    const activeConn = await api.getActiveConnection()
    const response = await api.getTableSchema(
      selectedDatabase.value || activeConn?.config?.database || '',
      tableName,
      activeConn?.id
    )
    tableSchema.value = response
    showSchemaDialog.value = true
  } catch (error) {
    ElMessage.error('获取表结构失败: ' + (error.response?.data?.error || error.message))
  }
}

const createTask = (tableName) => {
  api.getActiveConnection().then(res => {
    router.push({ 
      name: 'TaskConfig', 
      query: { 
        table: tableName,
        connection_id: res.id,
        database: selectedDatabase.value || res.config?.database || ''
      } 
    })
  }).catch(() => {
    router.push({ 
      name: 'TaskConfig', 
      query: { 
        table: tableName,
        database: selectedDatabase.value || ''
      } 
    })
  })
}
</script>
