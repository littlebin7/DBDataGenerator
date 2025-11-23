<template>
  <div class="database-tree">
    <div class="tree-header">
      <span class="tree-title">我的连接</span>
      <div class="tree-actions">
        <el-button
          :icon="Plus"
          size="small"
          text
          @click="expandAll"
          title="展开全部"
        />
        <el-button
          :icon="Minus"
          size="small"
          text
          @click="collapseAll"
          title="折叠全部"
        />
      </div>
    </div>
    <div class="tree-content" v-loading="loading">
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="treeProps"
        :filter-node-method="filterNode"
        :expand-on-click-node="false"
        :lazy="true"
        :load="loadNode"
        node-key="id"
        highlight-current
        :default-expand-all="false"
        @node-click="handleNodeClick"
      >
        <template #default="{ node, data }">
          <div class="tree-node-wrapper">
            <!-- 连接节点 -->
            <template v-if="data.type === 'connection'">
              <el-icon class="node-icon connection-icon">
                <Connection />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <el-tag
                v-if="data.connected"
                type="success"
                size="small"
                class="status-tag"
              >
                已连接
              </el-tag>
              <el-tag
                v-else
                type="warning"
                size="small"
                class="status-tag"
              >
                未连接
              </el-tag>
              <span class="node-type">{{ data.config?.type?.toUpperCase() || 'UNKNOWN' }}</span>
            </template>

            <!-- 数据库节点 -->
            <template v-else-if="data.type === 'database'">
              <el-icon class="node-icon database-icon">
                <Folder />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <span v-if="data.tableCount !== undefined" class="node-count">
                ({{ data.tableCount }} 表)
              </span>
            </template>

            <!-- 表节点 -->
            <template v-else-if="data.type === 'table'">
              <el-icon class="node-icon table-icon">
                <Document />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
            </template>

            <!-- 字段分组节点 -->
            <template v-else-if="data.type === 'fields-group'">
              <el-icon class="node-icon fields-group-icon">
                <List />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <span class="node-count" v-if="data.children && data.children.length > 0">
                ({{ data.children.length }})
              </span>
            </template>

            <!-- 索引分组节点 -->
            <template v-else-if="data.type === 'indexes-group'">
              <el-icon class="node-icon indexes-group-icon">
                <Sort />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <span class="node-count" v-if="data.children && data.children.length > 0">
                ({{ data.children.length }})
              </span>
            </template>

            <!-- 字段节点 -->
            <template v-else-if="data.type === 'field'">
              <el-icon class="node-icon field-icon">
                <Key />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <el-tag
                v-if="data.isPrimaryKey"
                type="danger"
                size="small"
                class="field-tag"
              >
                PK
              </el-tag>
              <el-tag
                v-if="data.isForeignKey"
                type="warning"
                size="small"
                class="field-tag"
              >
                FK
              </el-tag>
            </template>

            <!-- 索引节点 -->
            <template v-else-if="data.type === 'index'">
              <el-icon class="node-icon index-icon">
                <Sort />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
            </template>
          </div>
        </template>
      </el-tree>

      <!-- 空状态 -->
      <div v-if="treeData.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无数据库连接" :image-size="100">
          <el-button type="primary" size="small" @click="$emit('new-connection')">
            新建连接
          </el-button>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Connection,
  Folder,
  Document,
  Key,
  Sort,
  Plus,
  Minus,
  List
} from '@element-plus/icons-vue'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const props = defineProps({
  searchText: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['node-click', 'connection-updated'])

const connectionStore = useConnectionStore()
const treeRef = ref(null)
const treeData = ref([])
const loading = ref(false)

const treeProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf'
}

// 过滤节点
const filterNode = (value, data) => {
  if (!value) return true
  const searchText = value.toLowerCase()
  return data.label.toLowerCase().includes(searchText)
}

// 监听搜索文本变化
watch(() => props.searchText, (val) => {
  treeRef.value?.filter(val)
})

// 加载连接列表
const loadConnections = async () => {
  loading.value = true
  try {
    await connectionStore.loadConnections()
    const connections = connectionStore.allConnections

    if (!Array.isArray(connections) || connections.length === 0) {
      treeData.value = []
      return
    }

    treeData.value = connections.map(conn => ({
      id: `conn-${conn.id}`,
      label: conn.name,
      type: 'connection',
      connectionId: conn.id,
      connectionName: conn.name,
      connected: conn.connected || false,
      config: conn.config,
      children: [],
      isLeaf: false,
      loaded: false
    }))

    await nextTick()
  } catch (error) {
    console.error('加载连接列表失败:', error)
    ElMessage.error('加载连接列表失败: ' + (error.formattedMessage || error.message))
    treeData.value = []
  } finally {
    loading.value = false
  }
}

// 懒加载节点
const loadNode = async (node, resolve) => {
  const data = node.data

  if (data.type === 'connection') {
    if (data.loaded) {
      resolve(data.children || [])
      return
    }

    // 如果未连接，自动连接
    if (!data.connected) {
      try {
        await api.reconnect(data.connectionId)
        data.connected = true
        await connectionStore.loadConnections()
        const updatedConn = connectionStore.allConnections.find(c => c.id === data.connectionId)
        if (updatedConn) {
          data.connected = updatedConn.connected || false
        }
      } catch (error) {
        ElMessage.error('自动连接失败: ' + (error.formattedMessage || error.message))
        resolve([])
        return
      }
    }

    try {
      const databasesRes = await api.getDatabases(data.connectionId)
      const databases = databasesRes.databases || []

      const databaseNodes = databases.map(db => {
        const dbName = typeof db === 'string' ? db : (db.name || db)
        const dbUsername = typeof db === 'object' && db.username ? db.username : null

        return {
          id: `db-${data.connectionId}-${dbName}`,
          label: dbName,
          type: 'database',
          connectionId: data.connectionId,
          connectionName: data.connectionName,
          database: dbName,
          username: dbUsername,
          children: [],
          isLeaf: false,
          loaded: false
        }
      })

      data.children = databaseNodes
      data.loaded = true
      resolve(databaseNodes)
    } catch (error) {
      ElMessage.error('加载数据库列表失败: ' + (error.formattedMessage || error.message))
      resolve([])
    }
  } else if (data.type === 'database') {
    if (data.loaded) {
      resolve(data.children || [])
      return
    }

    // 检查连接状态
    const connection = connectionStore.allConnections.find(c => c.id === data.connectionId)
    if (connection && !connection.connected) {
      try {
        await api.reconnect(data.connectionId)
        await connectionStore.loadConnections()
      } catch (error) {
        ElMessage.error('自动连接失败: ' + (error.formattedMessage || error.message))
        resolve([])
        return
      }
    }

    try {
      const tablesRes = await api.getTables(data.database, data.connectionId)
      const tables = tablesRes.tables || []

      const tableNodes = tables.map(table => ({
        id: `table-${data.connectionId}-${data.database}-${table}`,
        label: table,
        type: 'table',
        connectionId: data.connectionId,
        connectionName: data.connectionName,
        database: data.database,
        table: table,
        children: [],
        isLeaf: false,
        loaded: false
      }))

      data.children = tableNodes
      data.loaded = true
      data.tableCount = tables.length
      resolve(tableNodes)
    } catch (error) {
      const errorMsg = error.formattedMessage || error.message || ''
      if (!errorMsg.includes('连接失败') && !errorMsg.includes('连接未建立')) {
        ElMessage.warning('加载表列表失败: ' + errorMsg)
      }
      resolve([])
    }
  } else if (data.type === 'table') {
    if (data.loaded) {
      resolve(data.children || [])
      return
    }
  } else if (data.type === 'fields-group' || data.type === 'indexes-group') {
    // 分组节点已加载，直接返回子节点
    resolve(data.children || [])
    return

    try {
      const schemaRes = await api.getTableSchema(data.database, data.table, data.connectionId)
      let schema = schemaRes.data || schemaRes

      // 兼容处理字段名大小写
      if (schema && schema.Fields && schema.Fields.length > 0) {
        const firstField = schema.Fields[0]
        if (firstField.Name && !firstField.name) {
          schema = {
            fields: schema.Fields.map(field => ({
              name: field.Name || field.name,
              type: field.Type || field.type,
              is_primary_key: field.IsPrimaryKey !== undefined ? field.IsPrimaryKey : field.is_primary_key,
              is_foreign_key: field.IsForeignKey !== undefined ? field.IsForeignKey : field.is_foreign_key,
              is_unique: field.IsUnique !== undefined ? field.IsUnique : field.is_unique,
              is_nullable: field.IsNullable !== undefined ? field.IsNullable : field.is_nullable,
              comment: field.Comment || field.comment || ''
            }))
          }
        }
      }

      const children = []
      
      // 添加字段分组节点
      if (schema && schema.fields && schema.fields.length > 0) {
        children.push({
          id: `fields-${data.connectionId}-${data.database}-${data.table}`,
          label: '字段',
          type: 'fields-group',
          connectionId: data.connectionId,
          database: data.database,
          table: data.table,
          children: schema.fields.map(field => ({
            id: `field-${data.connectionId}-${data.database}-${data.table}-${field.name}`,
            label: field.name,
            type: 'field',
            connectionId: data.connectionId,
            database: data.database,
            table: data.table,
            field: field.name,
            fieldData: field,
            isPrimaryKey: field.is_primary_key,
            isForeignKey: field.is_foreign_key,
            isUnique: field.is_unique,
            isNullable: field.is_nullable,
            fieldType: field.type,
            comment: field.comment,
            isLeaf: true
          })),
          isLeaf: false,
          loaded: true
        })
      }

      // 添加索引分组节点（暂时为空，后续可以添加索引加载逻辑）
      // 如果表结构中有索引信息，可以在这里添加
      children.push({
        id: `indexes-${data.connectionId}-${data.database}-${data.table}`,
        label: '索引',
        type: 'indexes-group',
        connectionId: data.connectionId,
        database: data.database,
        table: data.table,
        children: [], // 暂时为空，后续可以从 API 获取索引信息
        isLeaf: false,
        loaded: true
      })

      data.children = children
      data.loaded = true
      resolve(children)
    } catch (error) {
      const errorMsg = error.formattedMessage || error.message || ''
      if (!errorMsg.includes('连接失败') && !errorMsg.includes('连接未建立')) {
        ElMessage.warning('加载表结构失败: ' + errorMsg)
      }
      resolve([])
    }
  } else {
    resolve([])
  }
}

// 节点点击处理
const handleNodeClick = (data, node) => {
  emit('node-click', node, data)
}

// 展开全部
const expandAll = () => {
  const expandNode = (nodes) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        const nodeId = node.id || node.data?.id
        if (nodeId) {
          treeRef.value?.store.nodesMap[nodeId]?.expand()
        }
        expandNode(node.children)
      }
    })
  }
  expandNode(treeData.value)
}

// 折叠全部
const collapseAll = () => {
  const collapseNode = (nodes) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        const nodeId = node.id || node.data?.id
        if (nodeId) {
          treeRef.value?.store.nodesMap[nodeId]?.collapse()
        }
        collapseNode(node.children)
      }
    })
  }
  collapseNode(treeData.value)
}

// 刷新
const refresh = () => {
  loadConnections()
}

// 过滤
const filter = (text) => {
  treeRef.value?.filter(text)
}

defineExpose({
  refresh,
  filter
})

onMounted(() => {
  loadConnections()
})
</script>

<style scoped>
.database-tree {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.tree-header {
  height: 40px;
  padding: 0 15px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e6e6e6;
  background: #fafafa;
  flex-shrink: 0;
}

.tree-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.tree-actions {
  display: flex;
  gap: 5px;
}

.tree-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 10px;
}

.tree-node-wrapper {
  display: flex;
  align-items: center;
  width: 100%;
  font-size: 13px;
}

.node-icon {
  margin-right: 6px;
  font-size: 14px;
}

.connection-icon {
  color: #409EFF;
}

.database-icon {
  color: #67C23A;
}

.table-icon {
  color: #E6A23C;
}

.fields-group-icon {
  color: #606266;
}

.indexes-group-icon {
  color: #606266;
}

.field-icon {
  color: #909399;
}

.index-icon {
  color: #909399;
}

.node-label {
  flex: 1;
  font-weight: 400;
}

.status-tag {
  margin-left: 8px;
  margin-right: 8px;
}

.node-type {
  color: #909399;
  font-size: 11px;
  margin-left: 5px;
}

.node-count {
  color: #909399;
  font-size: 11px;
  margin-left: 5px;
}

.field-tag {
  margin-left: 5px;
  font-size: 10px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

:deep(.el-tree-node__content) {
  height: 32px;
  line-height: 32px;
}

:deep(.el-tree-node__label) {
  font-size: 13px;
}

:deep(.el-tree-node__expand-icon) {
  font-size: 12px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #f5f7fa;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #ecf5ff;
}
</style>

