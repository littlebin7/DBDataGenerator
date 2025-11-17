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
        <el-button @click="expandAll" :icon="Plus">展开全部</el-button>
        <el-button @click="collapseAll" :icon="Minus">折叠全部</el-button>
        <el-button @click="refreshAll" :icon="Refresh">刷新</el-button>
      </div>
    </div>

    <div class="content-wrapper">
      <!-- 左侧树形结构 -->
      <div class="tree-panel">
        <div v-if="treeData.length > 0" style="flex: 1; overflow-y: auto; overflow-x: hidden;">
          <el-tree
            ref="treeRef"
            :data="treeData"
            :props="treeProps"
            :filter-node-method="filterNode"
            :expand-on-click-node="false"
            :lazy="false"
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
        <div v-if="selectedTable" class="selected-info">
          <h3 class="panel-title">已选择的表</h3>
          <el-descriptions :column="1" border size="small" class="info-descriptions">
            <el-descriptions-item label="连接名称">
              <el-tag type="primary" size="small">{{ selectedTable.connectionName }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="数据库">
              <el-tag type="success" size="small">{{ selectedTable.database }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="表名">
              <el-tag type="warning" size="small">{{ selectedTable.table }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
          
          <div class="action-buttons">
            <el-button 
              type="primary" 
              style="width: 100%"
              @click="viewSchema"
              :icon="View"
            >
              查看表结构
            </el-button>
            <el-button 
              type="success" 
              style="width: 100%; margin-top: 10px"
              @click="createTask"
              :icon="Edit"
            >
              创建造数任务
            </el-button>
          </div>
        </div>
        
        <div v-else class="empty-selection">
          <el-empty description="请从左侧树中选择一个表" />
        </div>
      </div>
    </div>

    <!-- 表结构对话框 -->
    <el-dialog v-model="showSchemaDialog" title="表结构" width="1000px" top="5vh">
      <el-table :data="tableSchema?.fields" border style="width: 100%" max-height="500">
        <el-table-column prop="name" label="字段名" width="150" fixed="left" />
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
import { ref, onMounted, watch, nextTick } from 'vue'
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

// 对话框
const showSchemaDialog = ref(false)
const tableSchema = ref(null)

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
    const connectionsRes = await api.getConnections()
    const connections = connectionsRes.connections || []
    
    treeData.value = connections.map(conn => ({
      id: `conn-${conn.id}`,
      label: conn.name,
      type: 'connection',
      connectionId: conn.id,
      connectionName: conn.name,
      connected: !!conn.database,
      config: conn.config,
      children: [], // 初始为空，点击时懒加载
      isLeaf: false,
      loaded: false // 标记是否已加载
    }))
  } catch (error) {
    ElMessage.error('加载连接列表失败: ' + (error.formattedMessage || error.message))
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
    
    if (!data.connected) {
      ElMessage.warning('该连接未建立，请先在连接管理界面连接数据库')
      resolve([])
      return
    }
    
    try {
      const databasesRes = await api.getDatabases(data.connectionId)
      const databases = databasesRes.databases || []
      
      const databaseNodes = databases.map(db => ({
        id: `db-${data.connectionId}-${db}`,
        label: db,
        type: 'database',
        connectionId: data.connectionId,
        connectionName: data.connectionName,
        database: db,
        children: [], // 初始为空，点击时懒加载
        isLeaf: false,
        loaded: false
      }))
      
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
      
      resolve(tableNodes)
    } catch (error) {
      ElMessage.error('加载表列表失败: ' + (error.formattedMessage || error.message))
      resolve([])
    }
  } else {
    resolve([])
  }
}

// 处理节点点击
const handleNodeClick = (data) => {
  if (data.type === 'table') {
    selectedTable.value = {
      connectionId: data.connectionId,
      connectionName: data.connectionName,
      database: data.database,
      table: data.table
    }
  } else {
    selectedTable.value = null
  }
}

// 查看表结构
const viewSchema = async () => {
  if (!selectedTable.value) return
  
  try {
    const response = await api.getTableSchema(
      selectedTable.value.database,
      selectedTable.value.table,
      selectedTable.value.connectionId
    )
    tableSchema.value = response
    showSchemaDialog.value = true
  } catch (error) {
    ElMessage.error('获取表结构失败: ' + (error.formattedMessage || error.message))
  }
}

// 创建任务
const createTask = () => {
  if (!selectedTable.value) return
  
  router.push({ 
    name: 'TaskConfig', 
    query: { 
      table: selectedTable.value.table,
      connection_id: selectedTable.value.connectionId,
      database: selectedTable.value.database
    } 
  })
}

// 展开全部
const expandAll = () => {
  const expandNode = (nodes) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        treeRef.value?.store.nodesMap[node.id]?.expand()
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
        treeRef.value?.store.nodesMap[node.id]?.collapse()
        collapseNode(node.children)
      }
    })
  }
  collapseNode(treeData.value)
}

// 刷新全部
const refreshAll = async () => {
  await loadConnections()
  ElMessage.success('刷新成功')
}

onMounted(() => {
  loadConnections()
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
</style>
