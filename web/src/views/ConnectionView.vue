<template>
  <div>
    <!-- 连接列表 -->
    <el-card style="margin-bottom: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>已保存的连接</span>
          <div style="display: flex; gap: 10px;">
            <el-button @click="showImportDialog = true">导入连接</el-button>
            <el-button 
              type="success" 
              @click="exportConnections"
              :disabled="selectedConnections.length === 0"
            >
              导出选中 ({{ selectedConnections.length }})
            </el-button>
            <el-button type="primary" @click="showAddDialog = true">添加连接</el-button>
          </div>
        </div>
      </template>
      <el-table 
        :data="connections" 
        style="width: 100%" 
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="连接名称" />
        <el-table-column prop="config.type" label="类型" />
        <el-table-column prop="config.host" label="主机" />
        <el-table-column prop="config.port" label="端口" />
        <el-table-column prop="config.database" label="数据库" />
        <el-table-column label="连接状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.connected ? 'success' : 'info'">
              {{ scope.row.connected ? '已连接' : '未连接' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="活动状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'info'">
              {{ scope.row.is_active ? '活动' : '非活动' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300">
          <template #default="scope">
            <el-button 
              v-if="!scope.row.connected"
              size="small" 
              type="success"
              @click="reconnect(scope.row.id)"
            >
              连接
            </el-button>
            <el-button 
              v-if="!scope.row.is_active && scope.row.connected" 
              size="small" 
              @click="switchConnection(scope.row.id)"
            >
              设为活动
            </el-button>
            <el-button 
              size="small" 
              @click="editConnection(scope.row)"
            >
              编辑
            </el-button>
            <el-button 
              size="small" 
              @click="testConnection(scope.row)"
            >
              测试
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="removeConnection(scope.row.id)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无连接，请点击上方按钮添加连接" />
        </template>
      </el-table>
    </el-card>

    <!-- 添加/编辑连接对话框 -->
    <el-dialog 
      v-model="showDialog" 
      :title="editingConnection ? '编辑连接' : '添加数据库连接'" 
      width="600px"
    >
      <el-form :model="form" label-width="130px" ref="formRef">
        <el-form-item label="连接名称" required>
          <el-input v-model="form.name" placeholder="给连接起个名字" />
        </el-form-item>
        <el-form-item label="数据库类型" required>
          <el-select v-model="form.type" placeholder="请选择" style="width: 100%" @change="onTypeChange">
            <el-option label="PostgreSQL" value="postgres" />
            <el-option label="MySQL" value="mysql" />
            <el-option label="MariaDB" value="mariadb" />
            <el-option label="达梦数据库" value="dameng" />
            <el-option label="SQLite" value="sqlite" />
            <el-option label="SQL Server" value="mssql" />
            <el-option label="Oracle" value="oracle" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机" required v-if="form.type !== 'sqlite'">
          <el-input v-model="form.host" placeholder="localhost" />
        </el-form-item>
        <el-form-item label="端口" required v-if="form.type !== 'sqlite'">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item 
          :label="form.type === 'dameng' ? '数据库（可选）' : '数据库/文件路径'" 
          :required="form.type !== 'dameng' && form.type !== 'sqlite'"
          style="white-space: nowrap;"
        >
          <el-input 
            v-model="form.database" 
            :placeholder="form.type === 'sqlite' ? '例如: /path/to/database.db 或 database.db' : (form.type === 'dameng' ? '留空则连接后获取所有数据库' : '数据库名')" 
          />
          <div v-if="form.type === 'sqlite'" style="font-size: 12px; color: #909399; margin-top: 5px">
            支持相对路径或绝对路径，如不指定路径则会在当前目录创建
          </div>
          <div v-if="form.type === 'dameng'" style="font-size: 12px; color: #909399; margin-top: 5px">
            留空则连接后可以获取所有数据库列表
          </div>
        </el-form-item>
        <el-form-item label="用户名" required v-if="form.type !== 'sqlite'">
          <el-input v-model="form.user" />
          <div v-if="form.type === 'mssql'" style="font-size: 12px; color: #909399; margin-top: 5px">
            留空则使用 Windows 认证
          </div>
        </el-form-item>
        <el-form-item label="密码" required v-if="form.type !== 'sqlite'">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item v-if="form.type !== 'sqlite'">
          <el-checkbox v-model="form.save_password">记住密码</el-checkbox>
        </el-form-item>
        <el-form-item v-if="form.type === 'postgres'" label="SSL模式">
          <el-select v-model="form.ssl_mode" placeholder="请选择" style="width: 100%">
            <el-option label="disable" value="disable" />
            <el-option label="require" value="require" />
            <el-option label="verify-ca" value="verify-ca" />
            <el-option label="verify-full" value="verify-full" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.type === 'mysql' || form.type === 'mariadb'" label="字符集">
          <el-select v-model="form.charset" placeholder="请选择" style="width: 100%">
            <el-option label="utf8mb4" value="utf8mb4" />
            <el-option label="utf8" value="utf8" />
            <el-option label="gbk" value="gbk" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="connectionTested">
          <el-alert
            type="success"
            :closable="false"
            show-icon
          >
            <template #title>
              <div>
                <div>连接测试成功，可以保存</div>
                <div v-if="connectionVersion" style="margin-top: 5px; font-size: 12px; color: #67C23A;">
                  数据库版本: {{ connectionVersion }}
                </div>
              </div>
            </template>
          </el-alert>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button @click="testConnectionBeforeSave" :loading="testing">测试连接</el-button>
        <el-button type="primary" @click="saveConnection" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 导入连接对话框 -->
    <el-dialog 
      v-model="showImportDialog" 
      title="导入连接" 
      width="800px"
    >
      <div v-if="!importFileSelected">
        <el-upload
          ref="uploadRef"
          :auto-upload="false"
          :on-change="handleImportFileChange"
          :limit="1"
          accept=".json"
          drag
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            将文件拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              支持 JSON 格式的加密连接配置文件
            </div>
          </template>
        </el-upload>
      </div>
      
      <div v-else>
        <el-alert
          type="info"
          :closable="false"
          style="margin-bottom: 15px"
        >
          <template #title>
            <span>已选择文件：{{ importFileName }}</span>
            <el-button 
              type="text" 
              size="small" 
              @click="resetImport"
              style="margin-left: 10px"
            >
              重新选择
            </el-button>
          </template>
        </el-alert>
        
        <div v-if="importPreviewLoading" style="text-align: center; padding: 20px">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span style="margin-left: 10px">正在解析文件...</span>
        </div>
        
        <div v-else-if="importPreviewConnections.length > 0">
          <div style="margin-bottom: 15px; display: flex; justify-content: space-between; align-items: center">
            <span>共 {{ importPreviewConnections.length }} 个连接</span>
            <div>
              <el-button size="small" @click="selectAllImport">全选</el-button>
              <el-button size="small" @click="clearAllImport">清空</el-button>
            </div>
          </div>
          
          <el-table 
            :data="importPreviewConnections" 
            max-height="400"
            @selection-change="handleImportSelectionChange"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="name" label="连接名称" />
            <el-table-column prop="config.type" label="类型" width="100" />
            <el-table-column prop="config.host" label="主机" />
            <el-table-column prop="config.port" label="端口" width="80" />
            <el-table-column prop="config.database" label="数据库" />
            <el-table-column label="冲突状态" width="120">
              <template #default="scope">
                <el-tag 
                  v-if="scope.row.has_conflict" 
                  type="warning"
                  size="small"
                >
                  {{ scope.row.conflict_reason || '冲突' }}
                </el-tag>
                <el-tag v-else type="success" size="small">无冲突</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <div v-else-if="importFileSelected && !importPreviewLoading">
          <el-empty description="文件中没有找到连接配置" />
        </div>
      </div>
      
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button 
          v-if="importFileSelected && !importPreviewLoading" 
          @click="loadImportPreview"
          :loading="importPreviewLoading"
        >
          重新解析
        </el-button>
        <el-button 
          v-if="importFileSelected && importPreviewConnections.length > 0"
          type="primary" 
          @click="doImport"
          :loading="importing"
          :disabled="selectedImportConnections.length === 0"
        >
          导入选中 ({{ selectedImportConnections.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, Loading } from '@element-plus/icons-vue'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const router = useRouter()
const connectionStore = useConnectionStore()

const connections = computed(() => connectionStore.allConnections)
const loading = computed(() => connectionStore.loading)

const showDialog = ref(false)
const showAddDialog = ref(false)
const showImportDialog = ref(false)
const saving = ref(false)
const testing = ref(false)
const connectionTested = ref(false)
const connectionVersion = ref('')
const editingConnection = ref(null)
const formRef = ref(null)

// 多选相关
const selectedConnections = ref([])

// 处理表格选择变化
const handleSelectionChange = (selection) => {
  selectedConnections.value = selection
}

// 导出选中的连接
const exportConnections = async () => {
  if (selectedConnections.value.length === 0) {
    ElMessage.warning('请先选择要导出的连接')
    return
  }
  
  try {
    const connectionIds = selectedConnections.value.map(conn => conn.id)
    const response = await api.exportConnections(connectionIds)
    
    // 后端返回格式：{ data: "base64编码的加密数据", count: 1, encrypted: true }
    // 需要提取 data 字段（base64 编码的字符串）并保存
    const encryptedData = response.data
    
    if (!encryptedData || typeof encryptedData !== 'string') {
      ElMessage.error('导出数据格式错误')
      return
    }
    
    // 直接保存 base64 编码的加密数据（这是加密后的二进制数据的 base64 编码）
    // 创建下载链接
    const blob = new Blob([encryptedData], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `connections_${new Date().getTime()}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success(`成功导出 ${selectedConnections.value.length} 个连接`)
    // 清空选择
    selectedConnections.value = []
  } catch (error) {
    ElMessage.error('导出失败: ' + (error.formattedMessage || error.message))
  }
}

// 导入相关函数
const handleImportFileChange = (file) => {
  importFile.value = file.raw
  importFileName.value = file.name
  importFileSelected.value = true
  // 自动加载预览
  loadImportPreview()
}

const resetImport = () => {
  importFile.value = null
  importFileName.value = ''
  importFileSelected.value = false
  importPreviewConnections.value = []
  selectedImportConnections.value = []
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
}

const loadImportPreview = async () => {
  if (!importFile.value) {
    ElMessage.warning('请先选择文件')
    return
  }
  
  importPreviewLoading.value = true
  try {
    const response = await api.importConnectionsPreview(importFile.value)
    importPreviewConnections.value = response.connections || []
    selectedImportConnections.value = []
  } catch (error) {
    ElMessage.error('解析文件失败: ' + (error.formattedMessage || error.message))
    importPreviewConnections.value = []
  } finally {
    importPreviewLoading.value = false
  }
}

const handleImportSelectionChange = (selection) => {
  selectedImportConnections.value = selection
}

const selectAllImport = () => {
  // 通过表格引用全选
  if (importPreviewConnections.value.length > 0) {
    selectedImportConnections.value = [...importPreviewConnections.value]
  }
}

const clearAllImport = () => {
  selectedImportConnections.value = []
}

const doImport = async () => {
  if (selectedImportConnections.value.length === 0) {
    ElMessage.warning('请先选择要导入的连接')
    return
  }
  
  importing.value = true
  try {
    // 确保 selectedImportConnections 是数组
    if (!Array.isArray(selectedImportConnections.value)) {
      ElMessage.error('选择的数据格式错误')
      return
    }
    
    // 准备导入数据
    const connectionsToImport = selectedImportConnections.value.map(conn => ({
      id: conn.id,
      name: conn.name,
      config: conn.config,
      has_conflict: conn.has_conflict,
      conflict_reason: conn.conflict_reason
    }))
    
    // 处理冲突：如果有冲突，询问用户如何处理
    const hasConflicts = connectionsToImport.some(conn => conn.has_conflict)
    let conflictResolution = 'skip' // 默认跳过冲突
    
    if (hasConflicts) {
      try {
        await ElMessageBox.confirm(
          '检测到部分连接与现有连接名称冲突，是否覆盖现有连接？',
          '冲突提示',
          {
            confirmButtonText: '覆盖',
            cancelButtonText: '跳过',
            type: 'warning'
          }
        )
        conflictResolution = 'overwrite'
      } catch {
        conflictResolution = 'skip'
      }
    }
    
    // 确保 connectionsToImport 是数组
    console.log('准备导入的连接数据:', connectionsToImport)
    console.log('数据类型:', Array.isArray(connectionsToImport) ? '数组' : '非数组')
    
    await api.importConnections({
      connections: connectionsToImport,
      conflict_resolution: conflictResolution
    })
    
    ElMessage.success(`成功导入 ${selectedImportConnections.value.length} 个连接`)
    
    // 刷新连接列表
    await connectionStore.loadConnections()
    
    // 关闭对话框并重置
    showImportDialog.value = false
    resetImport()
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.formattedMessage || error.message))
  } finally {
    importing.value = false
  }
}

// 导入相关
const uploadRef = ref(null)
const importFileSelected = ref(false)
const importFileName = ref('')
const importFile = ref(null)
const importPreviewLoading = ref(false)
const importPreviewConnections = ref([])
const selectedImportConnections = ref([])
const importing = ref(false)

const form = ref({
  name: '',
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  user: '',
  password: '',
  database: '',
  ssl_mode: 'disable',
  charset: 'utf8mb4',
  save_password: true // 默认记住密码
})

onMounted(async () => {
  await connectionStore.loadConnections()
})

// 数据库类型切换时的处理
const onTypeChange = (type) => {
  // 根据数据库类型设置默认端口
  const defaultPorts = {
    postgres: 5432,
    mysql: 3306,
    mariadb: 3306,
    dameng: 5236,
    sqlite: 0, // SQLite 不需要端口
    mssql: 1433,
    oracle: 1521
  }
  if (defaultPorts[type] !== undefined) {
    form.value.port = defaultPorts[type]
  }
  
  // SQLite 不需要用户名和密码
  if (type === 'sqlite') {
    form.value.user = ''
    form.value.password = ''
    form.value.host = ''
  }
  
  // 重置连接测试状态
  connectionTested.value = false
}


// 测试连接（保存前）
const testConnectionBeforeSave = async () => {
  // SQLite 只需要数据库文件路径
  if (form.value.type === 'sqlite') {
    if (!form.value.name || !form.value.database) {
      ElMessage.warning('请填写连接名称和数据库文件路径')
      return
    }
  } else {
    // 其他数据库需要完整信息
    // 达梦数据库的数据库名是可选的
    if (!form.value.name || !form.value.host) {
      ElMessage.warning('请填写必填项')
      return
    }
    if (form.value.type !== 'dameng' && !form.value.database) {
      ElMessage.warning('请填写数据库名')
      return
    }
    // SQL Server 支持 Windows 认证，用户名可以为空
    if (form.value.type !== 'mssql' && !form.value.user) {
      ElMessage.warning('请填写用户名')
      return
    }
  }

  testing.value = true
  connectionTested.value = false
  connectionVersion.value = ''
  try {
    const response = await api.testConnection(form.value)
    ElMessage.success('连接测试成功')
    connectionTested.value = true
    // 如果有版本信息，显示出来
    if (response && response.version) {
      connectionVersion.value = response.version
    }
  } catch (error) {
    ElMessage.error('连接测试失败: ' + (error.formattedMessage || error.message))
    connectionTested.value = false
    connectionVersion.value = ''
  } finally {
    testing.value = false
  }
}

const saveConnection = async () => {
  // SQLite 只需要数据库文件路径
  if (form.value.type === 'sqlite') {
    if (!form.value.name || !form.value.database) {
      ElMessage.warning('请填写连接名称和数据库文件路径')
      return
    }
  } else {
    // 其他数据库需要完整信息
    // 达梦数据库的数据库名是可选的
    if (!form.value.name || !form.value.host) {
      ElMessage.warning('请填写必填项')
      return
    }
    if (form.value.type !== 'dameng' && !form.value.database) {
      ElMessage.warning('请填写数据库名')
      return
    }
    // SQL Server 支持 Windows 认证，用户名可以为空
    if (form.value.type !== 'mssql' && !form.value.user) {
      ElMessage.warning('请填写用户名')
      return
    }
  }

  saving.value = true
  try {
    if (editingConnection.value) {
      // 编辑连接
      await connectionStore.updateConnection(editingConnection.value.id, form.value)
      ElMessage.success('连接更新成功')
      showDialog.value = false
      resetForm()
      editingConnection.value = null
    } else {
      // 新建连接
      await connectionStore.addConnection(form.value)
      ElMessage.success('连接保存成功')
      showDialog.value = false
      resetForm()
    }
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.formattedMessage || error.message))
    connectionTested.value = false
  } finally {
    saving.value = false
  }
}

const editConnection = (conn) => {
  editingConnection.value = conn
  // 如果连接配置中有密码（已保存），则默认勾选记住密码并使用已保存的密码
  const hasPassword = conn.config.password && conn.config.password !== ''
  form.value = {
    name: conn.name,
    type: conn.config.type,
    host: conn.config.host,
    port: conn.config.port,
    user: conn.config.user,
    password: hasPassword ? conn.config.password : '', // 如果有已保存的密码，使用它
    database: conn.config.database,
    ssl_mode: conn.config.ssl_mode || 'disable',
    charset: conn.config.charset || 'utf8mb4',
    save_password: hasPassword // 如果已有密码，默认勾选记住密码
  }
  connectionTested.value = false // 编辑时需要重新测试
  showDialog.value = true
}

const testConnection = async (conn) => {
  try {
    // 获取连接配置
    const config = {
      type: conn.config.type,
      host: conn.config.host,
      port: conn.config.port,
      user: conn.config.user,
      password: conn.config.password || '', // 密码可能为空
      database: conn.config.database,
      ssl_mode: conn.config.ssl_mode || 'disable',
      charset: conn.config.charset || 'utf8mb4'
    }
    
    await api.testConnection(config)
    ElMessage.success('连接测试成功')
  } catch (error) {
    ElMessage.error('测试失败: ' + (error.formattedMessage || error.message))
  }
}

const reconnect = async (connId) => {
  try {
    await api.reconnect(connId)
    ElMessage.success('重新连接成功')
    // 刷新连接列表
    if (connectionStore && typeof connectionStore.loadConnections === 'function') {
      await connectionStore.loadConnections()
    }
    if (connectionStore && typeof connectionStore.loadActiveConnection === 'function') {
      await connectionStore.loadActiveConnection()
    }
  } catch (error) {
    console.error('重新连接错误:', error)
    ElMessage.error('连接失败: ' + (error.formattedMessage || error.message))
  }
}

const switchConnection = async (connId) => {
  try {
    await connectionStore.switchConnection(connId)
    ElMessage.success('已切换连接')
  } catch (error) {
    ElMessage.error('切换失败: ' + (error.formattedMessage || error.message))
  }
}

const removeConnection = async (connId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个连接吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await connectionStore.removeConnection(connId)
    ElMessage.success('已删除连接')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.formattedMessage || error.message))
    }
  }
}

const resetForm = () => {
  form.value = {
    name: '',
    type: 'postgres',
    host: 'localhost',
    port: 5432,
    user: '',
    password: '',
    database: '',
    ssl_mode: 'disable',
    charset: 'utf8mb4',
    save_password: true // 默认记住密码
  }
  editingConnection.value = null
  connectionTested.value = false
  connectionVersion.value = ''
}

watch(showAddDialog, (val) => {
  if (val) {
    resetForm()
    showDialog.value = true
    showAddDialog.value = false
  }
})

// 监听表单字段变化，重置测试状态
watch(() => [
  form.value.type,
  form.value.host,
  form.value.port,
  form.value.user,
  form.value.password,
  form.value.database,
  form.value.ssl_mode,
  form.value.charset
], () => {
  // 如果已经测试过，但配置改变了，重置测试状态
  if (connectionTested.value) {
    connectionTested.value = false
    connectionVersion.value = ''
  }
}, { deep: true })
</script>

