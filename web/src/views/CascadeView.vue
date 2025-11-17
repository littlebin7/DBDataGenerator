<template>
  <div class="cascade-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">级联生成</span>
          <el-button type="primary" :icon="Search" @click="analyzeRelations" :loading="analyzing" :disabled="!canAnalyze">分析表关系</el-button>
        </div>
      </template>

      <!-- 选择连接和数据库 -->
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
        </el-row>
      </el-form>

      <!-- 表关系图 -->
      <el-card v-if="relationsGraph" shadow="hover" style="margin-bottom: 20px">
        <template #header>
          <span style="font-size: 16px; font-weight: 500">表关系图</span>
        </template>
        <div style="padding: 20px; background: #f5f5f5; border-radius: 4px; min-height: 200px">
          <div v-if="relationsGraph.nodes && relationsGraph.nodes.length > 0">
            <div style="display: flex; flex-wrap: wrap; gap: 10px">
              <el-tag
                v-for="node in relationsGraph.nodes"
                :key="node.name"
                size="large"
                style="padding: 10px 15px"
              >
                {{ node.name }}
              </el-tag>
            </div>
            <div v-if="relationsGraph.edges && relationsGraph.edges.length > 0" style="margin-top: 20px">
              <div style="font-weight: 500; margin-bottom: 10px">表关系:</div>
              <div v-for="edge in relationsGraph.edges" :key="`${edge.from}-${edge.to}`" style="margin-bottom: 5px; font-size: 14px">
                {{ edge.from }} → {{ edge.to }} ({{ edge.type }})
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无表关系数据" />
        </div>
      </el-card>

      <!-- 级联生成配置 -->
      <el-card v-if="relationsGraph" shadow="hover">
        <template #header>
          <span style="font-size: 16px; font-weight: 500">级联生成配置</span>
        </template>
        <el-form :model="cascadeForm" label-width="150px">
          <el-form-item label="选择表" required>
            <el-checkbox-group v-model="cascadeForm.selectedTables">
              <el-checkbox
                v-for="node in relationsGraph.nodes"
                :key="node.name"
                :label="node.name"
              >
                {{ node.name }}
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="generateCascade" :loading="generating" :disabled="cascadeForm.selectedTables.length === 0">
              开始级联生成
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <div v-else-if="!analyzing" style="text-align: center; padding: 40px; color: #909399">
        <el-empty description="请选择连接和数据库后分析表关系" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const analyzing = ref(false)
const generating = ref(false)
const connections = ref([])
const databases = ref([])
const form = ref({
  connectionId: '',
  database: ''
})
const relationsGraph = ref(null)
const cascadeForm = ref({
  selectedTables: []
})

const canAnalyze = computed(() => {
  return form.value.connectionId && form.value.database
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
  databases.value = []
  relationsGraph.value = null
  cascadeForm.value.selectedTables = []

  if (!form.value.connectionId) return

  try {
    const data = await api.getDatabases(form.value.connectionId)
    databases.value = data.databases || []
  } catch (error) {
    ElMessage.error('获取数据库列表失败: ' + (error.formattedMessage || error.message))
  }
}

const onDatabaseChange = () => {
  relationsGraph.value = null
  cascadeForm.value.selectedTables = []
}

const analyzeRelations = async () => {
  if (!canAnalyze.value) {
    ElMessage.warning('请先选择连接和数据库')
    return
  }

  analyzing.value = true
  try {
    const data = await api.getTableRelations(form.value.connectionId, form.value.database)
    relationsGraph.value = data
    ElMessage.success('表关系分析完成')
  } catch (error) {
    ElMessage.error('分析失败: ' + (error.formattedMessage || error.message))
  } finally {
    analyzing.value = false
  }
}

const generateCascade = async () => {
  if (cascadeForm.value.selectedTables.length === 0) {
    ElMessage.warning('请至少选择一个表')
    return
  }

  generating.value = true
  try {
    // 构建表配置（简化处理，实际应该让用户配置每个表的生成规则）
    const tableConfigs = {}
    cascadeForm.value.selectedTables.forEach(table => {
      tableConfigs[table] = {
        database: form.value.database,
        table_name: table,
        fields: []
      }
    })

    const config = {
      connection_id: form.value.connectionId,
      database: form.value.database,
      tables: cascadeForm.value.selectedTables,
      table_configs: tableConfigs
    }

    const data = await api.generateCascade(config)
    ElMessage.success(`级联生成已启动，已创建 ${data.task_ids?.length || 0} 个任务`)
    
    // 重置表单
    cascadeForm.value.selectedTables = []
  } catch (error) {
    ElMessage.error('级联生成失败: ' + (error.formattedMessage || error.message))
  } finally {
    generating.value = false
  }
}

loadConnections()
</script>

<style scoped>
.cascade-view {
  padding: 0;
}
</style>

