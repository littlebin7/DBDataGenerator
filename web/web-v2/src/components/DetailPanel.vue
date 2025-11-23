<template>
  <div class="detail-panel">
    <!-- 连接详情 -->
    <div v-if="nodeType === 'connection'" class="connection-detail">
      <div class="detail-header">
        <h3 class="detail-title">{{ nodeData?.connectionName || '连接信息' }}</h3>
        <el-button :icon="Refresh" size="small" text @click="handleRefresh">刷新</el-button>
      </div>
      <div class="detail-content" v-if="connectionInfo">
        <el-descriptions :column="1" border size="small" label-width="120px">
          <el-descriptions-item label="连接名称">
            <el-tag type="primary" size="small">{{ connectionInfo.name }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="数据库类型">
            <el-tag :type="connectionInfo.connected ? 'success' : 'warning'" size="small">
              {{ connectionInfo.type?.toUpperCase() || 'UNKNOWN' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="连接状态">
            <el-tag :type="connectionInfo.connected ? 'success' : 'warning'" size="small">
              {{ connectionInfo.connected ? '已连接' : '未连接' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="主机地址">
            <span>{{ connectionInfo.host || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="端口">
            <span>{{ connectionInfo.port || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="用户名">
            <span>{{ connectionInfo.user || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="默认数据库" v-if="connectionInfo.database && connectionInfo.database !== '-'">
            <el-tag type="info" size="small">{{ connectionInfo.database }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="connectionInfo.type === 'dameng' || connectionInfo.type === 'oracle' ? 'Schema数量' : '数据库数量'">
            <el-tag type="success" size="small">{{ connectionInfo.databaseCount || 0 }} 个</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <!-- 数据库详情 -->
    <div v-else-if="nodeType === 'database'" class="database-detail">
      <div class="detail-header">
        <h3 class="detail-title">{{ nodeData?.database || '数据库信息' }}</h3>
        <el-button :icon="Refresh" size="small" text @click="handleRefresh">刷新</el-button>
      </div>
      <div class="detail-content" v-if="databaseInfo">
        <el-descriptions :column="1" border size="small" label-width="120px">
          <el-descriptions-item label="连接名称">
            <el-tag type="primary" size="small">{{ databaseInfo.connectionName }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="数据库">
            <el-tag type="success" size="small">{{ databaseInfo.database }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="所属用户" v-if="databaseInfo.username">
            <el-tag type="warning" size="small">{{ databaseInfo.username }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="表数量" v-if="databaseInfo.tableCount !== undefined">
            <el-tag type="info" size="small">{{ databaseInfo.tableCount }} 张表</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <!-- 表详情 -->
    <div v-else-if="nodeType === 'table'" class="table-detail">
      <div class="detail-header">
        <h3 class="detail-title">{{ nodeData?.table || '表信息' }}</h3>
        <div class="header-actions">
          <el-button :icon="View" size="small" @click="viewSchema">查看结构</el-button>
          <el-button type="primary" :icon="Edit" size="small" @click="createTask">创建任务</el-button>
          <el-button :icon="Refresh" size="small" text @click="handleRefresh">刷新</el-button>
        </div>
      </div>
      <div class="detail-content" v-if="tableInfo">
        <el-descriptions :column="1" border size="small" label-width="120px" style="margin-bottom: 20px">
          <el-descriptions-item label="连接名称">
            <el-tag type="primary" size="small">{{ tableInfo.connectionName }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="数据库">
            <el-tag type="success" size="small">{{ tableInfo.database }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="表名">
            <el-tag type="warning" size="small">{{ tableInfo.table }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="数据行数">
            <el-tag v-if="tableRowCount !== null" type="info" size="small">
              {{ tableRowCount.toLocaleString() }} 行
            </el-tag>
            <span v-else style="color: #909399;">加载中...</span>
          </el-descriptions-item>
          <el-descriptions-item label="字段数">
            <el-tag v-if="tableSchema && tableSchema.fields" type="success" size="small">
              {{ tableSchema.fields.length }} 个
            </el-tag>
            <span v-else style="color: #909399;">-</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 字段列表 -->
        <div v-if="tableSchema && tableSchema.fields" class="fields-section">
          <h4 class="section-title">字段列表</h4>
          <el-table :data="tableSchema.fields" border size="small" max-height="400">
            <el-table-column prop="name" label="字段名" width="150" fixed="left" />
            <el-table-column prop="type" label="类型" width="120" />
            <el-table-column label="约束" width="200">
              <template #default="scope">
                <el-tag v-if="scope.row.is_primary_key" type="danger" size="small" style="margin-right: 5px">
                  主键
                </el-tag>
                <el-tag v-if="scope.row.is_foreign_key" type="warning" size="small" style="margin-right: 5px">
                  外键
                </el-tag>
                <el-tag v-if="scope.row.is_unique && !scope.row.is_primary_key" type="success" size="small" style="margin-right: 5px">
                  唯一
                </el-tag>
                <el-tag v-if="!scope.row.is_nullable" type="info" size="small">
                  非空
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="comment" label="注释" min-width="200" />
          </el-table>
        </div>
      </div>
    </div>

    <!-- 字段详情 -->
    <div v-else-if="nodeType === 'field'" class="field-detail">
      <div class="detail-header">
        <h3 class="detail-title">{{ nodeData?.field || '字段信息' }}</h3>
      </div>
      <div class="detail-content" v-if="fieldInfo">
        <el-descriptions :column="1" border size="small" label-width="120px">
          <el-descriptions-item label="字段名">
            <el-tag type="primary" size="small">{{ fieldInfo.name }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag type="success" size="small">{{ fieldInfo.type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否主键">
            <el-tag :type="fieldInfo.isPrimaryKey ? 'danger' : 'info'" size="small">
              {{ fieldInfo.isPrimaryKey ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否外键">
            <el-tag :type="fieldInfo.isForeignKey ? 'warning' : 'info'" size="small">
              {{ fieldInfo.isForeignKey ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否唯一">
            <el-tag :type="fieldInfo.isUnique ? 'success' : 'info'" size="small">
              {{ fieldInfo.isUnique ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否可空">
            <el-tag :type="fieldInfo.isNullable ? 'info' : 'warning'" size="small">
              {{ fieldInfo.isNullable ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="注释">
            <span>{{ fieldInfo.comment || '-' }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-detail">
      <el-empty description="请从左侧树中选择一个节点查看详情" :image-size="120" />
    </div>

    <!-- 表结构对话框 -->
    <el-dialog v-model="showSchemaDialog" title="表结构" width="1000px" top="5vh">
      <div v-if="tableSchema && tableSchema.fields" style="margin-bottom: 15px; color: #606266; font-size: 14px;">
        共 <strong style="color: #409EFF;">{{ tableSchema.fields.length }}</strong> 个字段
      </div>
      <el-table v-if="tableSchema && tableSchema.fields" :data="tableSchema.fields" border style="width: 100%" max-height="500">
        <el-table-column prop="name" label="名称" width="150" fixed="left" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column label="大小" width="100">
          <template #default="scope">
            <span v-if="scope.row.max_length">{{ scope.row.max_length }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="比例" width="100">
          <template #default="scope">
            <span v-if="scope.row.precision && scope.row.scale">
              {{ scope.row.precision}},{{ scope.row.scale }}
            </span>
            <span v-else-if="scope.row.scale">{{ scope.row.scale }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="不是 null" width="100" align="center">
          <template #default="scope">
            <el-checkbox :model-value="!scope.row.is_nullable" disabled />
          </template>
        </el-table-column>
        <el-table-column label="键" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.is_primary_key" type="danger" size="small" style="margin-right: 5px">
              主键
            </el-tag>
            <el-tag v-if="scope.row.is_foreign_key" type="warning" size="small" style="margin-right: 5px">
              外键
            </el-tag>
            <el-tag v-if="scope.row.is_unique && !scope.row.is_primary_key" type="success" size="small">
              唯一
            </el-tag>
            <span v-if="!scope.row.is_primary_key && !scope.row.is_foreign_key && !scope.row.is_unique">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="注释" min-width="200" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, View, Edit } from '@element-plus/icons-vue'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const props = defineProps({
  selectedNode: {
    type: Object,
    default: null
  },
  nodeData: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['refresh', 'create-task'])

const connectionStore = useConnectionStore()
const showSchemaDialog = ref(false)
const tableSchema = ref(null)
const tableRowCount = ref(null)

const nodeType = computed(() => {
  return props.nodeData?.type || null
})

const connectionInfo = computed(() => {
  if (nodeType.value !== 'connection') return null
  const connection = connectionStore.allConnections.find(c => c.id === props.nodeData?.connectionId)
  if (!connection) return null
  
  return {
    name: connection.name,
    type: connection.config?.type || props.nodeData?.config?.type,
    connected: connection.connected || props.nodeData?.connected,
    host: connection.config?.host || props.nodeData?.config?.host,
    port: connection.config?.port || props.nodeData?.config?.port,
    user: connection.config?.user || props.nodeData?.config?.user,
    database: connection.config?.database || props.nodeData?.config?.database,
    databaseCount: props.nodeData?.databaseCount || 0
  }
})

const databaseInfo = computed(() => {
  if (nodeType.value !== 'database') return null
  return {
    connectionName: props.nodeData?.connectionName,
    database: props.nodeData?.database,
    username: props.nodeData?.username,
    tableCount: props.nodeData?.tableCount
  }
})

const tableInfo = computed(() => {
  if (nodeType.value !== 'table') return null
  return {
    connectionName: props.nodeData?.connectionName,
    database: props.nodeData?.database,
    table: props.nodeData?.table
  }
})

const fieldInfo = computed(() => {
  if (nodeType.value !== 'field') return null
  return {
    name: props.nodeData?.field,
    type: props.nodeData?.fieldType,
    isPrimaryKey: props.nodeData?.isPrimaryKey,
    isForeignKey: props.nodeData?.isForeignKey,
    isUnique: props.nodeData?.isUnique,
    isNullable: props.nodeData?.isNullable,
    comment: props.nodeData?.comment
  }
})

// 监听表节点变化，加载表信息
watch(() => props.nodeData, async (newData) => {
  if (newData?.type === 'table') {
    await loadTableInfo()
  } else {
    tableSchema.value = null
    tableRowCount.value = null
  }
}, { immediate: true })

// 加载表信息
const loadTableInfo = async () => {
  if (!props.nodeData || props.nodeData.type !== 'table') return

  try {
    // 加载表行数
    const countRes = await api.getTableCount(
      props.nodeData.database,
      props.nodeData.table,
      props.nodeData.connectionId
    )
    tableRowCount.value = countRes.count !== undefined && countRes.count !== null ? countRes.count : null

    // 加载表结构
    const schemaRes = await api.getTableSchema(
      props.nodeData.database,
      props.nodeData.table,
      props.nodeData.connectionId
    )
    let schema = schemaRes.data || schemaRes

    // 兼容处理字段名大小写
    if (schema && schema.Fields && schema.Fields.length > 0) {
      const firstField = schema.Fields[0]
      if (firstField.Name && !firstField.name) {
        schema = {
          table_name: schema.TableName || schema.table_name,
          table_comment: schema.TableComment || schema.table_comment || '',
          tablespace: schema.Tablespace || schema.tablespace || '',
          fields: schema.Fields.map(field => ({
            name: field.Name || field.name,
            type: field.Type || field.type,
            go_type: field.GoType || field.go_type,
            is_primary_key: field.IsPrimaryKey !== undefined ? field.IsPrimaryKey : field.is_primary_key,
            is_foreign_key: field.IsForeignKey !== undefined ? field.IsForeignKey : field.is_foreign_key,
            foreign_table: field.ForeignTable || field.foreign_table,
            is_unique: field.IsUnique !== undefined ? field.IsUnique : field.is_unique,
            is_nullable: field.IsNullable !== undefined ? field.IsNullable : field.is_nullable,
            default_value: field.DefaultValue !== undefined ? field.DefaultValue : field.default_value,
            max_length: field.MaxLength !== undefined ? field.MaxLength : field.max_length,
            precision: field.Precision !== undefined ? field.Precision : field.precision,
            scale: field.Scale !== undefined ? field.Scale : field.scale,
            enum_values: field.EnumValues || field.enum_values || [],
            comment: field.Comment || field.comment || ''
          }))
        }
      }
    }

    tableSchema.value = schema
  } catch (error) {
    console.error('加载表信息失败:', error)
    const errorMsg = error.formattedMessage || error.message || ''
    if (!errorMsg.includes('连接未建立') && !errorMsg.includes('请先连接')) {
      ElMessage.warning('加载表信息失败: ' + errorMsg)
    }
  }
}

// 查看表结构
const viewSchema = async () => {
  if (!tableSchema.value) {
    await loadTableInfo()
  }
  showSchemaDialog.value = true
}

// 创建任务
const createTask = () => {
  if (!props.nodeData || props.nodeData.type !== 'table') return
  emit('create-task', {
    connectionId: props.nodeData.connectionId,
    database: props.nodeData.database,
    table: props.nodeData.table
  })
}

// 刷新
const handleRefresh = () => {
  emit('refresh')
}
</script>

<style scoped>
.detail-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}

.detail-header {
  height: 50px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e6e6e6;
  background: #fafafa;
  flex-shrink: 0;
}

.detail-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.connection-detail,
.database-detail,
.table-detail,
.field-detail {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.fields-section {
  margin-top: 20px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 10px;
}

.empty-detail {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

