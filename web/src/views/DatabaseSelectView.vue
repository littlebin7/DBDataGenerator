<template>
  <div class="database-select-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="filterText"
          placeholder="搜索连接、数据库或表名"
          style="width: 300px"
          clearable
          :prefix-icon="Search"
          @input="handleFilter"
        />
      </div>
      <div class="toolbar-right">
        <el-button @click="expandAll" :icon="Plus" :loading="actionLoading.get('expandAll')" :disabled="actionLoading.get('expandAll')">展开全部</el-button>
        <el-button @click="collapseAll" :icon="Minus" :loading="actionLoading.get('collapseAll')" :disabled="actionLoading.get('collapseAll')">折叠全部</el-button>
        <el-button @click="refreshAll" :icon="Refresh" :loading="actionLoading.get('refreshAll')" :disabled="actionLoading.get('refreshAll')">刷新</el-button>
      </div>
    </div>

    <div class="content-wrapper">
      <!-- 左侧树形结构 -->
      <div class="tree-panel">
        <div style="flex: 1; overflow-y: auto; overflow-x: hidden;">
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
            @node-click="handleNodeClick"
            v-loading="loading"
          >
          <template #default="{ node, data }">
            <div class="tree-node-wrapper">
              <span class="tree-node">
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
                  <span class="node-info">{{ data.config?.type?.toUpperCase() }}</span>
                </template>
                
                <!-- 数据库节点 -->
                <template v-else-if="data.type === 'database'">
                  <el-icon class="node-icon database-icon">
                    <Folder />
                  </el-icon>
                  <span class="node-label">{{ node.label }}</span>
                  <span class="node-info" v-if="data.tableCount !== undefined">
                    ({{ data.tableCount }} 表)
                  </span>
                </template>
                
                <!-- 表节点 -->
                <template v-else>
                  <el-icon class="node-icon table-icon">
                    <Document />
                  </el-icon>
                  <span class="node-label">{{ node.label }}</span>
                </template>
              </span>
            </div>
          </template>
          </el-tree>
        </div>
        
        <!-- 空状态 -->
        <div v-if="treeData.length === 0 && !loading" class="empty-state">
          <el-empty description="暂无数据库连接，请先在连接管理中添加连接" />
        </div>
      </div>

      <!-- 右侧详情面板 -->
      <div class="detail-panel">
        <!-- 选中表时的显示 -->
        <div v-if="selectedTable" class="selected-info">
          <h3 class="panel-title">已选择的表</h3>
          <el-descriptions :column="1" border size="small" class="info-descriptions" label-width="100px">
            <el-descriptions-item label="连接名称">
              <el-tag type="primary" size="small">{{ selectedTable.connectionName }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="数据库">
              <el-tag type="success" size="small">{{ selectedTable.database }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="表名">
              <el-tag type="warning" size="small">{{ selectedTable.table }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="数据行数">
              <el-tag v-if="tableRowCount !== null" type="info" size="small">{{ tableRowCount.toLocaleString() }} 行</el-tag>
              <span v-else style="color: #909399;">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="字段数">
              <el-tag v-if="tableSchema && tableSchema.fields" type="success" size="small">{{ tableSchema.fields.length }} 个</el-tag>
              <span v-else style="color: #909399;">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="表注释">
              <span style="color: #606266;">{{ tableSchema && tableSchema.table_comment ? tableSchema.table_comment : '-' }}</span>
            </el-descriptions-item>
          </el-descriptions>
          
          <div class="action-buttons">
            <el-button 
              type="primary" 
              class="action-button"
              @click="viewSchema"
              :icon="View"
              :loading="actionLoading.get('viewSchema')"
              :disabled="actionLoading.get('viewSchema')"
            >
              查看表结构
            </el-button>
            <el-button 
              type="success" 
              class="action-button"
              @click="createTask"
              :icon="Edit"
            >
              创建造数任务
            </el-button>
          </div>
        </div>
        
        <!-- 选中数据库时的显示 -->
        <div v-else-if="selectedDatabase" class="selected-info">
          <h3 class="panel-title">数据库信息</h3>
          <el-descriptions :column="1" border size="small" class="info-descriptions" label-width="100px">
            <el-descriptions-item label="连接名称">
              <el-tag type="primary" size="small">{{ selectedDatabase.connectionName }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="数据库">
              <el-tag type="success" size="small">{{ selectedDatabase.database }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="所属用户" v-if="selectedDatabase.username">
              <el-tag type="warning" size="small">{{ selectedDatabase.username }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="表数量" v-if="selectedDatabase.tableCount !== undefined">
              <el-tag type="info" size="small">{{ selectedDatabase.tableCount }} 张表</el-tag>
            </el-descriptions-item>
          </el-descriptions>
          
          <div class="action-buttons">
            <el-button 
              type="primary" 
              style="width: 100%"
              @click="refreshDatabase"
              :icon="Refresh"
              :loading="actionLoading.get('refreshDatabase')"
              :disabled="actionLoading.get('refreshDatabase')"
            >
              刷新数据库
            </el-button>
          </div>
        </div>
        
        <!-- 未选择任何内容 -->
        <div v-else class="empty-selection">
          <el-empty description="请从左侧树中选择一个数据库或表" />
        </div>
      </div>
    </div>

    <!-- 表结构对话框 -->
    <el-dialog v-model="showSchemaDialog" title="表结构" width="1000px" top="5vh">
      <div v-if="tableSchema && tableSchema.fields" style="margin-bottom: 15px; color: #606266; font-size: 14px;">
        共 <strong style="color: #409EFF;">{{ tableSchema.fields.length }}</strong> 个字段
      </div>
      <div v-if="!tableSchema || !tableSchema.fields || tableSchema.fields.length === 0" style="text-align: center; padding: 20px;">
        <el-empty description="暂无表结构数据" />
      </div>
      <el-table v-else :data="tableSchema.fields" border style="width: 100%" max-height="500" table-layout="auto">
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
            <span v-if="scope.row.precision && scope.row.scale">{{ scope.row.precision}},{{ scope.row.scale }}</span>
            <span v-else-if="scope.row.scale">{{ scope.row.scale }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="不是 null" width="100" align="center">
          <template #default="scope">
            <el-checkbox :model-value="!scope.row.is_nullable" class="schema-checkbox" />
          </template>
        </el-table-column>
        <el-table-column label="键" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.is_primary_key" type="danger" size="small" style="margin-right: 5px">主键</el-tag>
            <el-tag v-if="scope.row.is_foreign_key" type="warning" size="small" style="margin-right: 5px">外键</el-tag>
            <el-tag v-if="scope.row.is_unique && !scope.row.is_primary_key" type="success" size="small">唯一</el-tag>
            <span v-if="!scope.row.is_primary_key && !scope.row.is_foreign_key && !scope.row.is_unique">-</span>
          </template>
        </el-table-column>
        <el-table-column label="注释" min-width="200">
          <template #default="scope">
            <span>{{ scope.row.comment || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { defineOptions } from 'vue'

defineOptions({
  name: 'DatabaseSelect'
})
import { ref, onMounted, watch, nextTick, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, 
  Search, 
  Connection, 
  Folder, 
  Document,
  Plus,
  Minus,
  View,
  Edit
} from '@element-plus/icons-vue'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const connectionStore = useConnectionStore()

const router = useRouter()

// 树形数据
const treeData = ref([])
const treeRef = ref(null)
const filterText = ref('')
const loading = ref(false)

// 已加载的连接数据缓存
const loadedData = ref(new Map())

// 选中的表信息
const selectedTable = ref(null)
// 选中的数据库信息
const selectedDatabase = ref(null)

// 对话框
const showSchemaDialog = ref(false)
const tableSchema = ref(null)
// 表行数
const tableRowCount = ref(null)

// 刷新时间戳（用于限制刷新频率）
const lastRefreshTime = ref(0)
const REFRESH_INTERVAL = 3000 // 3秒间隔

// 操作按钮的 loading 状态
const actionLoading = ref(new Map())

// 树形配置
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

// 处理搜索
const handleFilter = () => {
  treeRef.value?.filter(filterText.value)
}

// 加载所有连接（初始加载）
const loadConnections = async () => {
  loading.value = true
  try {
    // 先清空现有数据
    treeData.value = []
    
    await connectionStore.loadConnections()
    const connections = connectionStore.allConnections
    
    // 确保 connections 是数组
    if (!Array.isArray(connections)) {
      treeData.value = []
      return
    }
    
    if (connections.length === 0) {
      treeData.value = []
      return
    }
    
    // 构建树节点数据
    const nodes = connections.map(conn => ({
      id: `conn-${conn.id}`,
      label: conn.name,
      type: 'connection',
      connectionId: conn.id,
      connectionName: conn.name,
      connected: conn.connected || false,
      config: conn.config,
      children: [], // 初始为空，点击时懒加载
      isLeaf: false,
      loaded: false // 标记是否已加载
    }))
    
    // 直接赋值，确保响应式更新
    treeData.value = nodes
    
    // 确保树组件更新
    await nextTick()
  } catch (error) {
    console.error('加载连接列表失败:', error)
    ElMessage.error('加载连接列表失败: ' + (error.formattedMessage || error.message))
    treeData.value = []
  } finally {
    loading.value = false
  }
}

// 懒加载节点数据
const loadNode = async (node, resolve) => {
  const data = node.data
  
  // 如果是连接节点，加载数据库
  if (data.type === 'connection') {
    if (data.loaded) {
      // 已加载，直接返回缓存的数据
      resolve(data.children || [])
      return
    }
    
    // 如果未连接，自动连接
    if (!data.connected) {
      try {
        await api.reconnect(data.connectionId)
        // 更新连接状态
        data.connected = true
        // 刷新连接列表以获取最新状态
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
        // 处理达梦数据库返回的对象格式（包含 name 和 username）
        const dbName = typeof db === 'string' ? db : (db.name || db)
        const dbUsername = typeof db === 'object' && db.username ? db.username : null
        
        return {
          id: `db-${data.connectionId}-${dbName}`,
          label: dbName, // 树节点只显示模式名，不显示用户名
          type: 'database',
          connectionId: data.connectionId,
          connectionName: data.connectionName,
          database: dbName,
          username: dbUsername, // 保存用户名信息，用于右侧详情面板显示
          children: [], // 初始为空，点击时懒加载
          isLeaf: false,
          loaded: false
        }
      })
      
      // 更新节点数据
      data.children = databaseNodes
      data.loaded = true
      
      resolve(databaseNodes)
    } catch (error) {
      ElMessage.error('加载数据库列表失败: ' + (error.formattedMessage || error.message))
      resolve([])
    }
  } 
  // 如果是数据库节点，加载表
  else if (data.type === 'database') {
    if (data.loaded) {
      resolve(data.children || [])
      return
    }
    
    // 检查连接状态，如果未连接则自动连接
    const connection = connectionStore.allConnections.find(c => c.id === data.connectionId)
    if (connection && !connection.connected) {
      try {
        await api.reconnect(data.connectionId)
        // 刷新连接列表以获取最新状态
        await connectionStore.loadConnections()
        // 更新树节点中的连接状态
        const connNode = findConnectionNode(treeData.value, data.connectionId)
        if (connNode) {
          const updatedConn = connectionStore.allConnections.find(c => c.id === data.connectionId)
          if (updatedConn) {
            connNode.connected = updatedConn.connected || false
          }
        }
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
        isLeaf: true
      }))
      
      // 更新节点数据
      data.children = tableNodes
      data.loaded = true
      data.tableCount = tables.length
      
      // 如果当前选中的数据库是这个节点，更新表数量
      if (selectedDatabase.value && 
          selectedDatabase.value.connectionId === data.connectionId &&
          selectedDatabase.value.database === data.database) {
        selectedDatabase.value.tableCount = tables.length
      }
      
      resolve(tableNodes)
    } catch (error) {
      ElMessage.error('加载表列表失败: ' + (error.formattedMessage || error.message))
      resolve([])
    }
  } else {
    resolve([])
  }
}

// 查找连接节点的辅助函数
const findConnectionNode = (nodes, connectionId) => {
  for (const node of nodes) {
    if (node.type === 'connection' && node.connectionId === connectionId) {
      return node
    }
    if (node.children && node.children.length > 0) {
      const found = findConnectionNode(node.children, connectionId)
      if (found) return found
    }
  }
  return null
}

// 处理节点点击
const handleNodeClick = async (data, node) => {
  if (data.type === 'table') {
    selectedTable.value = {
      connectionId: data.connectionId,
      connectionName: data.connectionName,
      database: data.database,
      table: data.table
    }
    selectedDatabase.value = null
    tableRowCount.value = null
    tableSchema.value = null // 清空之前的表结构
    
    // 加载表行数和表结构（用于显示字段数）
    loadTableCount(data.database, data.table, data.connectionId)
    loadTableSchema(data.database, data.table, data.connectionId)
  } else if (data.type === 'database') {
    selectedDatabase.value = {
      connectionId: data.connectionId,
      connectionName: data.connectionName,
      database: data.database,
      username: data.username, // 保存用户名信息
      tableCount: data.tableCount
    }
    selectedTable.value = null
    
    // 如果节点未展开，展开节点（这会触发 loadNode 加载表数据）
    if (!node.expanded) {
      node.expand()
    }
  } else if (data.type === 'connection') {
    selectedTable.value = null
    selectedDatabase.value = null
    
    // 如果连接节点未加载，则展开并加载数据库
    if (!data.loaded) {
      // 如果节点未展开，先展开（这会触发 loadNode）
      if (!node.expanded) {
        node.expand()
      } else {
        // 如果节点已展开但未加载，手动触发加载
        await loadNode(node, (children) => {
          data.children = children
          data.loaded = true
        })
      }
    } else if (!node.expanded) {
      // 如果已加载但未展开，展开节点显示数据库
      node.expand()
    }
  } else {
    selectedTable.value = null
    selectedDatabase.value = null
  }
}

// 加载表行数
const loadTableCount = async (database, table, connectionId) => {
  try {
    const response = await api.getTableCount(database, table, connectionId)
    // 明确检查 count 是否为 undefined 或 null，0 是有效值
    tableRowCount.value = (response.count !== undefined && response.count !== null) ? response.count : null
  } catch (error) {
    console.error('获取表行数失败:', error)
    tableRowCount.value = null
  }
}

// 加载表结构（用于显示字段数，不打开对话框）
const loadTableSchema = async (database, table, connectionId) => {
  try {
    const response = await api.getTableSchema(database, table, connectionId)
    let schema = response.data || response
    
    // 兼容处理：如果字段名是大写的，转换为小写（向后兼容）
    if (schema && schema.Fields && schema.Fields.length > 0) {
      const firstField = schema.Fields[0]
      if (firstField.Name && !firstField.name) {
        schema = {
          table_name: schema.TableName || schema.table_name,
          table_comment: schema.TableComment || schema.table_comment || '',
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
    
    // 确保表注释字段存在
    if (schema && !schema.table_comment) {
      schema.table_comment = schema.TableComment || ''
    }
    
    tableSchema.value = schema
  } catch (error) {
    console.error('获取表结构失败:', error)
    tableSchema.value = null
  }
}

// 查看表结构
const viewSchema = async () => {
  if (!selectedTable.value) return
  if (actionLoading.value.get('viewSchema')) return
  
  actionLoading.value.set('viewSchema', true)
  try {
    const response = await api.getTableSchema(
      selectedTable.value.database,
      selectedTable.value.table,
      selectedTable.value.connectionId
    )
    // API 返回的数据结构：response.data 包含实际的表结构数据
    // 因为 sendSuccess 包装了数据，所以 response.data 就是 TableSchema 对象
    let schema = response.data || response
    
    // 兼容处理：如果字段名是大写的，转换为小写（向后兼容）
    if (schema && schema.Fields && schema.Fields.length > 0) {
      // 检查第一个字段是否有大写字段名
      const firstField = schema.Fields[0]
      if (firstField.Name && !firstField.name) {
        // 转换字段名为小写格式
        schema = {
          table_name: schema.TableName || schema.table_name,
          table_comment: schema.TableComment || schema.table_comment || '',
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
    console.log('表结构数据:', tableSchema.value)
    // 调试：检查字段注释
    if (tableSchema.value && tableSchema.value.fields) {
      const fieldsWithComment = tableSchema.value.fields.filter(f => f.comment && f.comment !== '')
      console.log('字段总数:', tableSchema.value.fields.length, '有注释的字段数:', fieldsWithComment.length)
      if (fieldsWithComment.length > 0) {
        console.log('有注释的字段:', fieldsWithComment.map(f => ({ name: f.name, comment: f.comment })))
      }
    }
    showSchemaDialog.value = true
  } catch (error) {
    console.error('获取表结构失败:', error)
    ElMessage.error('获取表结构失败: ' + (error.formattedMessage || error.message))
  } finally {
    actionLoading.value.set('viewSchema', false)
  }
}

// 创建任务
const createTask = () => {
  if (!selectedTable.value) return
  if (actionLoading.value.get('createTask')) return
  
  actionLoading.value.set('createTask', true)
  router.push({ 
    name: 'TaskConfig', 
    query: { 
      table: selectedTable.value.table,
      connection_id: selectedTable.value.connectionId,
      database: selectedTable.value.database
    } 
  }).finally(() => {
    // 路由跳转后清除 loading 状态
    actionLoading.value.set('createTask', false)
  })
}

// 展开全部
const expandAll = () => {
  if (actionLoading.value.get('expandAll')) return
  actionLoading.value.set('expandAll', true)
  const expandNode = (nodes) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        treeRef.value?.store.nodesMap[node.id]?.expand()
        expandNode(node.children)
      }
    })
  }
  expandNode(treeData.value)
  actionLoading.value.set('expandAll', false)
}

// 折叠全部
const collapseAll = () => {
  if (actionLoading.value.get('collapseAll')) return
  actionLoading.value.set('collapseAll', true)
  const collapseNode = (nodes) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        treeRef.value?.store.nodesMap[node.id]?.collapse()
        collapseNode(node.children)
      }
    })
  }
  collapseNode(treeData.value)
  actionLoading.value.set('collapseAll', false)
}

// 刷新全部
const refreshAll = async () => {
  if (actionLoading.value.get('refreshAll')) return
  actionLoading.value.set('refreshAll', true)
  try {
    await loadConnections()
    ElMessage.success('刷新成功')
  } finally {
    actionLoading.value.set('refreshAll', false)
  }
}

// 刷新数据库
const refreshDatabase = async () => {
  if (!selectedDatabase.value) return
  if (actionLoading.value.get('refreshDatabase')) return
  
  actionLoading.value.set('refreshDatabase', true)
  try {
    // 检查刷新频率限制
    const now = Date.now()
    const timeSinceLastRefresh = now - lastRefreshTime.value
    if (timeSinceLastRefresh < REFRESH_INTERVAL) {
      const remainingTime = Math.ceil((REFRESH_INTERVAL - timeSinceLastRefresh) / 1000)
      ElMessage.warning(`刷新过于频繁，请等待 ${remainingTime} 秒后再试`)
      return
    }
    
    // 更新最后刷新时间
    lastRefreshTime.value = now
    const { connectionId, database } = selectedDatabase.value
    
    // 找到对应的树节点并重新加载
    const findAndReloadNode = (nodes, connectionId, database) => {
      for (const node of nodes) {
        if (node.type === 'database' && 
            node.connectionId === connectionId && 
            node.database === database) {
          // 重置加载状态和子节点
          node.loaded = false
          node.children = []
          node.tableCount = 0
          
          // 找到树组件中对应的节点
          const nodeId = `db-${connectionId}-${database}`
          const treeNode = treeRef.value?.store?.nodesMap[nodeId]
          
          // 如果树节点存在，清空其子节点
          if (treeNode) {
            treeNode.childNodes = []
            treeNode.data.children = []
          }
          
          // 重新加载表数据
          const reloadTables = async () => {
            try {
              // 检查连接状态
              const connection = connectionStore.allConnections.find(c => c.id === connectionId)
              if (connection && !connection.connected) {
                await api.reconnect(connectionId)
                await connectionStore.loadConnections()
              }
              
              const tablesRes = await api.getTables(database, connectionId)
              const tables = tablesRes.tables || []
              
              const tableNodes = tables.map(table => ({
                id: `table-${connectionId}-${database}-${table}`,
                label: table,
                type: 'table',
                connectionId: connectionId,
                connectionName: selectedDatabase.value.connectionName,
                database: database,
                table: table,
                isLeaf: true
              }))
              
              // 更新节点数据
              node.children = tableNodes
              node.loaded = true
              node.tableCount = tables.length
              
              // 更新选中数据库的表数量
              if (selectedDatabase.value) {
                selectedDatabase.value.tableCount = tables.length
              }
              
              // 如果树节点存在且已展开，需要更新树组件
              if (treeNode && treeNode.expanded) {
                // 通过重新加载节点来更新树结构
                treeNode.expand()
              }
              
              ElMessage.success('刷新成功')
            } catch (error) {
              ElMessage.error('刷新失败: ' + (error.formattedMessage || error.message))
            }
          }
          
          reloadTables()
          return true
        }
        if (node.children && node.children.length > 0) {
          if (findAndReloadNode(node.children, connectionId, database)) {
            return true
          }
        }
      }
      return false
    }
    
    // 重新加载数据库节点
    if (!findAndReloadNode(treeData.value, connectionId, database)) {
      ElMessage.warning('未找到对应的数据库节点')
    }
  } catch (error) {
    ElMessage.error('刷新失败: ' + (error.formattedMessage || error.message))
  }
}

onMounted(async () => {
  // 确保在组件挂载后立即加载连接列表
  await loadConnections()
})

// 当组件被激活时（从其他页面返回时），清除所有 loading 状态
onActivated(() => {
  // 清除所有操作按钮的 loading 状态，防止从其他页面返回时按钮被禁用
  actionLoading.value.clear()
})
</script>

<style scoped>
.database-select-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 10px;
}

.content-wrapper {
  flex: 1;
  display: flex;
  overflow: hidden;
  background: #fff;
  min-height: 0;
}

.tree-panel {
  width: 400px;
  border-right: 1px solid #e6e6e6;
  overflow: hidden;
  padding: 15px;
  background: #fafafa;
  flex-shrink: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.detail-panel {
  flex: 1;
  padding: 20px;
  overflow: hidden;
  background: #fff;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.detail-panel .selected-info,
.detail-panel .empty-selection {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.tree-node-wrapper {
  width: 100%;
  display: flex;
  align-items: center;
}

.tree-node {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
  padding: 2px 0;
}

.node-icon {
  margin-right: 8px;
  font-size: 16px;
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

.node-label {
  flex: 1;
  font-weight: 500;
}

.status-tag {
  margin-left: 8px;
  margin-right: 8px;
}

.node-info {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}

.selected-info {
  max-width: 400px;
}

.panel-title {
  margin-bottom: 15px;
  color: #303133;
  font-size: 16px;
}

.info-descriptions {
  margin-bottom: 20px;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.action-button {
  flex: 1;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

.empty-selection {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

:deep(.el-tree-node__content) {
  height: 36px;
  line-height: 36px;
}

:deep(.el-tree-node__label) {
  font-size: 14px;
}

:deep(.el-tree-node__expand-icon) {
  font-size: 14px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #f5f7fa;
}

/* 表结构对话框中的复选框样式 - 即使禁用也显示正常颜色 */
:deep(.schema-checkbox) {
  pointer-events: none;
}

:deep(.schema-checkbox .el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #409EFF;
  border-color: #409EFF;
}

:deep(.schema-checkbox .el-checkbox__input.is-checked .el-checkbox__inner::after) {
  border-color: #fff;
}

:deep(.schema-checkbox .el-checkbox__input .el-checkbox__inner) {
  background-color: #fff;
  border-color: #DCDFE6;
}
</style>
