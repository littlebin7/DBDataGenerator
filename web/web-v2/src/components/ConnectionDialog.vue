<template>
  <el-dialog
    v-model="dialogVisible"
    title="新建数据库连接"
    width="600px"
    @close="handleClose"
  >
    <el-form :model="form" label-width="130px" ref="formRef" :rules="rules">
      <el-form-item label="连接名称" prop="name">
        <el-input v-model="form.name" placeholder="给连接起个名字" />
      </el-form-item>
      <el-form-item label="数据库类型" prop="type">
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
      <el-form-item label="主机" prop="host" v-if="form.type !== 'sqlite'">
        <el-input v-model="form.host" placeholder="localhost" />
      </el-form-item>
      <el-form-item label="端口" prop="port" v-if="form.type !== 'sqlite'">
        <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
      </el-form-item>
      <el-form-item
        :label="form.type === 'dameng' ? '数据库（可选）' : '数据库/文件路径'"
        :required="form.type !== 'dameng' && form.type !== 'sqlite'"
        prop="database"
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
      <el-form-item label="用户名" prop="user" v-if="form.type !== 'sqlite'">
        <el-input v-model="form.user" />
        <div v-if="form.type === 'mssql'" style="font-size: 12px; color: #909399; margin-top: 5px">
          留空则使用 Windows 认证
        </div>
      </el-form-item>
      <el-form-item label="密码" prop="password" v-if="form.type !== 'sqlite'">
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
        <el-alert type="success" :closable="false" show-icon>
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
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button @click="testConnection" :loading="testing">测试连接</el-button>
      <el-button type="primary" @click="saveConnection" :loading="saving">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'
import { useConnectionStore } from '../stores/connection'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const connectionStore = useConnectionStore()
const formRef = ref(null)
const testing = ref(false)
const saving = ref(false)
const connectionTested = ref(false)
const connectionVersion = ref('')

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const form = ref({
  name: '',
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  database: '',
  user: '',
  password: '',
  save_password: true,
  ssl_mode: 'disable',
  charset: 'utf8mb4'
})

const rules = {
  name: [{ required: true, message: '请输入连接名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择数据库类型', trigger: 'change' }],
  host: [
    {
      required: true,
      message: '请输入主机地址',
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'sqlite') {
          callback()
        } else if (!value) {
          callback(new Error('请输入主机地址'))
        } else {
          callback()
        }
      }
    }
  ],
  port: [
    {
      required: true,
      message: '请输入端口',
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'sqlite') {
          callback()
        } else if (!value) {
          callback(new Error('请输入端口'))
        } else {
          callback()
        }
      }
    }
  ],
  database: [
    {
      required: true,
      message: '请输入数据库名或文件路径',
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'dameng' || form.value.type === 'sqlite') {
          callback()
        } else if (!value) {
          callback(new Error('请输入数据库名或文件路径'))
        } else {
          callback()
        }
      }
    }
  ],
  user: [
    {
      required: true,
      message: '请输入用户名',
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'sqlite' || (form.value.type === 'mssql' && !value)) {
          callback()
        } else if (!value) {
          callback(new Error('请输入用户名'))
        } else {
          callback()
        }
      }
    }
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'sqlite' || (form.value.type === 'mssql' && !form.value.user)) {
          callback()
        } else if (!value) {
          callback(new Error('请输入密码'))
        } else {
          callback()
        }
      }
    }
  ]
}

// 类型变化处理
const onTypeChange = () => {
  connectionTested.value = false
  connectionVersion.value = ''
  
  // 设置默认端口
  const defaultPorts = {
    postgres: 5432,
    mysql: 3306,
    mariadb: 3306,
    dameng: 5236,
    mssql: 1433,
    oracle: 1521
  }
  if (defaultPorts[form.value.type]) {
    form.value.port = defaultPorts[form.value.type]
  }
}

// 测试连接
const testConnection = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    testing.value = true
    connectionTested.value = false
    connectionVersion.value = ''
    
    try {
      const config = {
        name: form.value.name,
        type: form.value.type,
        host: form.value.host,
        port: form.value.port,
        database: form.value.database,
        user: form.value.user,
        password: form.value.password,
        save_password: form.value.save_password
      }
      
      if (form.value.type === 'postgres' && form.value.ssl_mode) {
        config.ssl_mode = form.value.ssl_mode
      }
      
      if ((form.value.type === 'mysql' || form.value.type === 'mariadb') && form.value.charset) {
        config.charset = form.value.charset
      }
      
      const result = await api.testConnection(config)
      connectionTested.value = true
      if (result.version) {
        connectionVersion.value = result.version
      }
      ElMessage.success('连接测试成功')
    } catch (error) {
      ElMessage.error('连接测试失败: ' + (error.formattedMessage || error.message))
    } finally {
      testing.value = false
    }
  })
}

// 保存连接
const saveConnection = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    saving.value = true
    
    try {
      const config = {
        name: form.value.name,
        type: form.value.type,
        host: form.value.host,
        port: form.value.port,
        database: form.value.database,
        user: form.value.user,
        password: form.value.password,
        save_password: form.value.save_password
      }
      
      if (form.value.type === 'postgres' && form.value.ssl_mode) {
        config.ssl_mode = form.value.ssl_mode
      }
      
      if ((form.value.type === 'mysql' || form.value.type === 'mariadb') && form.value.charset) {
        config.charset = form.value.charset
      }
      
      await connectionStore.addConnection(config)
      ElMessage.success('连接保存成功')
      emit('success')
      handleClose()
    } catch (error) {
      ElMessage.error('保存连接失败: ' + (error.formattedMessage || error.message))
    } finally {
      saving.value = false
    }
  })
}

// 关闭对话框
const handleClose = () => {
  form.value = {
    name: '',
    type: 'postgres',
    host: 'localhost',
    port: 5432,
    database: '',
    user: '',
    password: '',
    save_password: true,
    ssl_mode: 'disable',
    charset: 'utf8mb4'
  }
  connectionTested.value = false
  connectionVersion.value = ''
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}
</script>

<style scoped>
</style>

