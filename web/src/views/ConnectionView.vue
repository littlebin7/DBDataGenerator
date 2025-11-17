<template>
  <div>
    <!-- 连接列表 -->
    <el-card style="margin-bottom: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>已保存的连接</span>
          <el-button type="primary" @click="showAddDialog = true">添加连接</el-button>
        </div>
      </template>
      <el-table :data="connections" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="连接名称" />
        <el-table-column prop="config.type" label="类型" />
        <el-table-column prop="config.host" label="主机" />
        <el-table-column prop="config.port" label="端口" />
        <el-table-column prop="config.database" label="数据库" />
        <el-table-column label="连接状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.database ? 'success' : 'info'">
              {{ scope.row.database ? '已连接' : '未连接' }}
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
              v-if="!scope.row.database"
              size="small" 
              type="success"
              @click="reconnect(scope.row.id)"
            >
              连接
            </el-button>
            <el-button 
              v-if="!scope.row.is_active && scope.row.database" 
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
        <el-form-item label="数据库/文件路径" required style="white-space: nowrap;">
          <el-input 
            v-model="form.database" 
            :placeholder="form.type === 'sqlite' ? '例如: /path/to/database.db 或 database.db' : '数据库名'" 
          />
          <div v-if="form.type === 'sqlite'" style="font-size: 12px; color: #909399; margin-top: 5px">
            支持相对路径或绝对路径，如不指定路径则会在当前目录创建
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
              <span>连接测试成功，可以保存</span>
            </template>
          </el-alert>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button @click="testConnectionBeforeSave" :loading="testing">测试连接</el-button>
        <el-button type="primary" @click="saveConnection" :loading="saving" :disabled="!connectionTested">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const router = useRouter()
const connections = ref([])
const showDialog = ref(false)
const showAddDialog = ref(false)
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const connectionTested = ref(false)
const editingConnection = ref(null)
const formRef = ref(null)

const form = ref({
  name: '',
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  user: '',
  password: '',
  database: '',
  ssl_mode: 'disable',
  charset: 'utf8mb4'
})

onMounted(async () => {
  await loadConnections()
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

const loadConnections = async () => {
  loading.value = true
  try {
    const response = await api.getConnections()
    connections.value = response.connections || []
    console.log('加载的连接列表:', connections.value)
  } catch (error) {
    console.error('加载连接列表失败:', error)
    ElMessage.error('加载连接列表失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
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
    if (!form.value.name || !form.value.host || !form.value.database) {
      ElMessage.warning('请填写必填项')
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
  try {
    await api.testConnection(form.value)
    ElMessage.success('连接测试成功')
    connectionTested.value = true
  } catch (error) {
    ElMessage.error('连接测试失败: ' + (error.formattedMessage || error.message))
    connectionTested.value = false
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
    if (!form.value.name || !form.value.host || !form.value.database) {
      ElMessage.warning('请填写必填项')
      return
    }
    // SQL Server 支持 Windows 认证，用户名可以为空
    if (form.value.type !== 'mssql' && !form.value.user) {
      ElMessage.warning('请填写用户名')
      return
    }
  }

  // 必须先测试连接成功才能保存
  if (!connectionTested.value) {
    ElMessage.warning('请先测试连接，确保连接成功后再保存')
    return
  }

  saving.value = true
  try {
    if (editingConnection.value) {
      // 编辑连接
      await api.updateConnection(editingConnection.value.id, form.value)
      ElMessage.success('连接更新成功')
      showDialog.value = false
      resetForm()
      editingConnection.value = null
      await loadConnections()
    } else {
      // 新建连接
      const response = await api.connect(form.value)
      ElMessage.success('连接保存成功')
      showDialog.value = false
      resetForm()
      await loadConnections()
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
  form.value = {
    name: conn.name,
    type: conn.config.type,
    host: conn.config.host,
    port: conn.config.port,
    user: conn.config.user,
    password: '', // 不显示密码，需要重新输入
    database: conn.config.database,
    ssl_mode: conn.config.ssl_mode || 'disable',
    charset: conn.config.charset || 'utf8mb4'
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
    // TODO: 实现重新连接API
    ElMessage.info('重新连接功能待实现')
    await loadConnections()
  } catch (error) {
    ElMessage.error('连接失败: ' + (error.formattedMessage || error.message))
  }
}

const switchConnection = async (connId) => {
  try {
    await api.switchConnection(connId)
    ElMessage.success('已切换连接')
    await loadConnections()
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
    await api.disconnect(connId)
    ElMessage.success('已删除连接')
    await loadConnections()
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
    charset: 'utf8mb4'
  }
  editingConnection.value = null
  connectionTested.value = false
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
  }
}, { deep: true })
</script>

