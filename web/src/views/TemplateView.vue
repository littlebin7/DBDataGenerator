<template>
  <div class="template-view">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 18px; font-weight: 500">模板管理</span>
          <div style="display: flex; gap: 10px">
            <el-input
              v-model="searchText"
              placeholder="搜索模板名称或表名"
              style="width: 250px"
              clearable
              :prefix-icon="Search"
            />
            <el-button type="primary" :icon="Refresh" @click="loadTemplates" :loading="loading">刷新</el-button>
          </div>
        </div>
      </template>

      <div v-loading="loading">
        <el-table :data="filteredTemplates" stripe border v-if="filteredTemplates.length > 0">
          <el-table-column prop="name" label="模板名称" width="200" />
          <el-table-column prop="table_name" label="表名" width="150" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button
                type="primary"
                size="small"
                @click="viewTemplate(scope.row)"
              >
                查看
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="deleteTemplate(scope.row.id)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无模板" />
      </div>
    </el-card>

    <!-- 模板详情对话框（只读） -->
    <el-dialog 
      v-model="showDetailDialog" 
      title="模板详情（只读）" 
      width="80%" 
      :before-close="closeDetail"
      :close-on-click-modal="false"
    >
      <div v-if="selectedTemplate">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="模板名称">{{ selectedTemplate.name }}</el-descriptions-item>
          <el-descriptions-item label="表名">{{ selectedTemplate.table_name }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ selectedTemplate.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(selectedTemplate.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider>字段配置</el-divider>
        <el-table 
          :data="selectedTemplate.config?.field_rules || []" 
          border 
          style="margin-top: 20px"
          :row-class-name="() => 'readonly-row'"
        >
          <el-table-column prop="field_name" label="字段名" width="150" />
          <el-table-column label="规则类型" width="150">
            <template #default="scope">
              <el-tag size="small" type="info">{{ getRuleTypeLabel(scope.row.rule_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="规则配置" min-width="300">
            <template #default="scope">
              <div style="font-size: 13px; line-height: 1.6; color: #606266">
                {{ formatConfig(scope.row.rule_type, scope.row.config) }}
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <template #footer>
        <div style="text-align: right">
          <el-button @click="closeDetail">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatTime } from '../utils/formatters'
import { useTemplateStore } from '../stores/template'

const templateStore = useTemplateStore()

const loading = computed(() => templateStore.loading)
const templates = computed(() => templateStore.allTemplates)
const searchText = ref('')
const showDetailDialog = ref(false)
const selectedTemplate = ref(null)

const filteredTemplates = computed(() => {
  if (!searchText.value) return templates.value
  const search = searchText.value.toLowerCase()
  return templates.value.filter(t => 
    t.name?.toLowerCase().includes(search) || 
    t.table_name?.toLowerCase().includes(search)
  )
})

const loadTemplates = async () => {
  try {
    await templateStore.loadTemplates()
  } catch (error) {
    ElMessage.error('获取模板列表失败: ' + (error.formattedMessage || error.message))
  }
}

const viewTemplate = async (template) => {
  try {
    const data = await templateStore.loadTemplate(template.id)
    selectedTemplate.value = data
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error('获取模板详情失败: ' + (error.formattedMessage || error.message))
  }
}

const deleteTemplate = async (templateId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个模板吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await templateStore.deleteTemplate(templateId)
    ElMessage.success('模板已删除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.formattedMessage || error.message))
    }
  }
}

const closeDetail = () => {
  showDetailDialog.value = false
  selectedTemplate.value = null
}

// 获取规则类型标签（与TaskConfigView保持一致）
const getRuleTypeLabel = (ruleType) => {
  const labels = {
    'random_string': '随机值',
    'random_number': '随机数字',
    'random_date': '随机日期',
    'fixed': '固定值',
    'increment': '递增',
    'list': '列表选择',
    'regex': '正则表达式',
    'function': '函数',
    'template': '模板',
    'null': '空值',
    'reference': '引用字段',
    'geographic': '地理数据',
    'file': '从文件读取',
    'binary': '二进制/图片',
    'foreign': '外键引用'
  }
  return labels[ruleType] || ruleType
}

// 格式化配置显示（与TaskConfigView保持一致）
const formatConfig = (ruleType, config) => {
  if (!config) return '-'
  
  // 处理不同的规则类型
  switch (ruleType) {
    case 'random_string':
      return `长度: ${config.min_length || 10}-${config.max_length || 50}, 字符集: ${getCharSetLabel(config.char_set || 'all')}`
    case 'random_number':
      return `范围: ${config.min || 0} ~ ${config.max || 1000}${config.is_int === false ? ' (浮点数)' : ' (整数)'}`
    case 'random_date':
      return `日期范围: ${config.start_date || '无限制'} ~ ${config.end_date || '无限制'}${config.format ? `, 格式: ${config.format}` : ''}`
    case 'fixed':
      return `固定值: ${config.value || '-'}`
    case 'increment':
      return `起始值: ${config.start_value || 1}, 步长: ${config.step || 1}${config.cycle ? ', 循环' : ''}`
    case 'list':
      const values = config.values || []
      return `可选值: ${values.length > 0 ? values.slice(0, 3).join(', ') + (values.length > 3 ? `... (共${values.length}个)` : '') : '-'}`
    case 'regex':
      return `正则表达式: ${config.pattern || '-'}`
    case 'function':
      const funcName = config.func_name || config.funcName || 'UUID'
      const params = config.params || []
      return `函数: ${funcName}${params.length > 0 ? `(${params.join(', ')})` : '()'}`
    case 'template':
      return `模板: ${config.template || '-'}`
    case 'null':
      return `空值概率: ${((config.probability || 0) * 100).toFixed(0)}%`
    case 'reference':
      return `表达式: ${config.expression || '-'}${config.fields && config.fields.length > 0 ? `, 引用字段: ${config.fields.join(', ')}` : ''}`
    case 'geographic':
      return `类型: ${config.type || 'city'}${config.country ? `, 国家: ${config.country}` : ''}`
    case 'file':
      return `文件路径: ${config.file_path || '-'}, 类型: ${config.file_type || 'txt'}, 列索引: ${config.column_index || 0}${config.loop ? ', 循环读取' : ''}`
    case 'binary':
      if (config.mode === 'generate') {
        return `生成模式: ${config.width || 100}x${config.height || 100}, 格式: ${config.format || 'png'}`
      } else {
        return `文件路径: ${config.folder_path || '-'}, 扩展名: ${(config.extensions || []).join(', ') || '全部'}${config.loop ? ', 循环读取' : ''}`
      }
    case 'foreign':
      return `外键表: ${config.foreign_table || '-'}`
    default:
      // 如果无法识别，返回格式化的JSON（但更简洁）
      try {
        const keys = Object.keys(config)
        if (keys.length === 0) return '-'
        return keys.map(key => `${key}: ${config[key]}`).join(', ')
      } catch {
        return JSON.stringify(config)
      }
  }
}

// 获取字符集标签
const getCharSetLabel = (charSet) => {
  const labels = {
    'letters': '字母',
    'numbers': '数字',
    'chinese': '中文',
    'special': '特殊字符',
    'all': '全部'
  }
  return labels[charSet] || charSet
}

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.template-view {
  padding: 0;
}

/* 只读行样式 */
:deep(.readonly-row) {
  background-color: #fafafa;
}

:deep(.readonly-row:hover) {
  background-color: #f5f5f5;
}
</style>

