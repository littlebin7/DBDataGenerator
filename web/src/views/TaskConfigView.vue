<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>配置造数规则 - {{ tableName }}</span>
          <div>
            <el-button @click="showLoadTemplateDialog = true">从模板加载</el-button>
            <el-button @click="showSaveTemplateDialog = true">保存为模板</el-button>
            <el-button @click="$router.back()">返回</el-button>
          </div>
        </div>
      </template>

      <!-- 基本信息配置 -->
      <el-form :model="taskConfig" label-width="150px" style="margin-bottom: 30px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="任务名称">
              <el-input v-model="taskConfig.name" placeholder="给任务起个名字" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="生成数量">
              <el-input-number 
                v-model="taskConfig.totalRows" 
                :min="1" 
                :max="100000000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="批次大小">
              <el-input-number 
                v-model="taskConfig.batchSize" 
                :min="1" 
                :max="10000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="线程数">
              <el-input-number 
                v-model="taskConfig.threadCount" 
                :min="1" 
                :max="20"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 字段规则配置 -->
      <el-divider>字段生成规则配置</el-divider>
      
      <!-- 搜索和批量操作 -->
      <div style="margin-bottom: 15px; display: flex; gap: 10px; align-items: center">
        <el-input
          v-model="fieldSearchText"
          placeholder="搜索字段名"
          style="width: 200px"
          clearable
          :prefix-icon="Search"
        />
        <el-button size="small" @click="showBatchDialog = true">批量设置</el-button>
        <el-button size="small" @click="resetAllRules">重置所有</el-button>
      </div>

      <el-table :data="filteredFieldRules" border style="width: 100%">
        <el-table-column prop="fieldName" label="字段名" width="150" />
        <el-table-column prop="fieldType" label="字段类型" width="120" />
        <el-table-column label="规则类型" width="150">
          <template #default="scope">
            <el-select 
              v-model="scope.row.ruleType" 
              placeholder="选择规则"
              @change="onRuleTypeChange(scope.row)"
            >
              <el-option 
                v-for="rule in getAvailableRules(scope.row)" 
                :key="rule.value"
                :label="rule.label" 
                :value="rule.value" 
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="规则配置" min-width="300">
          <template #default="scope">
            <!-- 随机字符串配置 -->
            <div v-if="scope.row.ruleType === 'random_string'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-input-number 
                v-model="scope.row.config.minLength" 
                :min="1" 
                :max="1000"
                placeholder="最小长度"
                size="small"
                style="width: 100px"
              />
              <el-input-number 
                v-model="scope.row.config.maxLength" 
                :min="1" 
                :max="1000"
                placeholder="最大长度"
                size="small"
                style="width: 100px"
              />
              <el-select v-model="scope.row.config.charSet" size="small" style="width: 120px">
                <el-option label="字母" value="letters" />
                <el-option label="数字" value="numbers" />
                <el-option label="中文" value="chinese" />
                <el-option label="特殊字符" value="special" />
                <el-option label="全部" value="all" />
              </el-select>
            </div>

            <!-- 随机数字配置 -->
            <div v-if="scope.row.ruleType === 'random_number'" style="display: flex; gap: 10px">
              <el-input-number 
                v-model="scope.row.config.min" 
                :precision="2"
                placeholder="最小值"
                size="small"
                style="width: 120px"
              />
              <el-input-number 
                v-model="scope.row.config.max" 
                :precision="2"
                placeholder="最大值"
                size="small"
                style="width: 120px"
              />
              <el-checkbox v-model="scope.row.config.isInt" size="small">整数</el-checkbox>
            </div>

            <!-- 随机日期配置 -->
            <div v-if="scope.row.ruleType === 'random_date'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-date-picker
                v-model="scope.row.config.startDate"
                type="datetime"
                placeholder="开始日期"
                size="small"
                style="width: 180px"
                value-format="YYYY-MM-DDTHH:mm:ssZ"
              />
              <el-date-picker
                v-model="scope.row.config.endDate"
                type="datetime"
                placeholder="结束日期"
                size="small"
                style="width: 180px"
                value-format="YYYY-MM-DDTHH:mm:ssZ"
              />
            </div>

            <!-- 固定值配置 -->
            <el-input 
              v-if="scope.row.ruleType === 'fixed'"
              v-model="scope.row.config.value" 
              placeholder="输入固定值"
              size="small"
            />

            <!-- 递增配置 -->
            <div v-if="scope.row.ruleType === 'increment'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-input-number 
                v-model="scope.row.config.startValue" 
                placeholder="起始值"
                size="small"
                style="width: 120px"
              />
              <el-input-number 
                v-model="scope.row.config.step" 
                placeholder="步长"
                size="small"
                style="width: 120px"
              />
              <el-checkbox v-model="scope.row.config.cycle" size="small">循环</el-checkbox>
            </div>

            <!-- 列表配置 -->
            <el-input 
              v-if="scope.row.ruleType === 'list'"
              v-model="scope.row.config.valuesText" 
              placeholder="输入值列表，用逗号分隔"
              size="small"
              @blur="parseListValues(scope.row)"
            />

            <!-- 正则表达式配置 -->
            <el-input 
              v-if="scope.row.ruleType === 'regex'"
              v-model="scope.row.config.pattern" 
              placeholder="输入正则表达式"
              size="small"
            />

            <!-- 函数配置 -->
            <div v-if="scope.row.ruleType === 'function'" style="display: flex; gap: 10px">
              <el-select v-model="scope.row.config.funcName" size="small" style="width: 150px">
                <el-option label="NOW()" value="NOW" />
                <el-option label="TODAY()" value="TODAY" />
                <el-option label="UUID()" value="UUID" />
                <el-option label="RAND()" value="RAND" />
                <el-option label="RAND_INT(min, max)" value="RAND_INT" />
                <el-option label="CONCAT(...)" value="CONCAT" />
              </el-select>
            </div>

            <!-- 模板配置 -->
            <el-input 
              v-if="scope.row.ruleType === 'template'"
              v-model="scope.row.config.template" 
              placeholder="模板，支持 {name}, {date}, {number}, {uuid}"
              size="small"
            />

            <!-- 空值配置 -->
            <el-slider 
              v-if="scope.row.ruleType === 'null'"
              v-model="scope.row.config.probability" 
              :min="0" 
              :max="100"
              :step="1"
              show-input
              size="small"
            />

            <!-- 引用字段配置 -->
            <div v-if="scope.row.ruleType === 'reference'" style="display: flex; flex-direction: column; gap: 10px">
              <el-input 
                v-model="scope.row.config.expression" 
                placeholder="表达式，如 {first_name}_{last_name} 或 {price} * {quantity}"
                size="small"
              />
              <el-input 
                v-model="scope.row.config.fieldsText" 
                placeholder="引用的字段列表（可选，用逗号分隔）"
                size="small"
                @blur="parseReferenceFields(scope.row)"
              />
            </div>

            <!-- 地理数据配置 -->
            <div v-if="scope.row.ruleType === 'geographic'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-select v-model="scope.row.config.type" size="small" style="width: 150px">
                <el-option label="城市" value="city" />
                <el-option label="国家" value="country" />
                <el-option label="地址" value="address" />
                <el-option label="坐标" value="coordinates" />
                <el-option label="纬度" value="latitude" />
                <el-option label="经度" value="longitude" />
                <el-option label="邮编" value="postal_code" />
              </el-select>
              <el-input 
                v-model="scope.row.config.country" 
                placeholder="国家代码（可选，如 CN, US）"
                size="small"
                style="width: 200px"
              />
            </div>

            <!-- 从文件读取配置 -->
            <div v-if="scope.row.ruleType === 'file'" style="display: flex; flex-direction: column; gap: 10px">
              <el-input 
                v-model="scope.row.config.filePath" 
                placeholder="文件路径（绝对路径或相对路径）"
                size="small"
              />
              <div style="display: flex; gap: 10px; flex-wrap: wrap">
                <el-select v-model="scope.row.config.fileType" size="small" style="width: 120px">
                  <el-option label="CSV" value="csv" />
                  <el-option label="TXT" value="txt" />
                  <el-option label="JSON" value="json" />
                </el-select>
                <el-input-number 
                  v-model="scope.row.config.columnIndex" 
                  :min="0"
                  placeholder="列索引（CSV/JSON使用）"
                  size="small"
                  style="width: 150px"
                />
                <el-checkbox v-model="scope.row.config.loop" size="small">循环读取</el-checkbox>
              </div>
            </div>

            <!-- 二进制/图片配置 -->
            <div v-if="scope.row.ruleType === 'binary'" style="display: flex; flex-direction: column; gap: 10px">
              <el-select v-model="scope.row.config.mode" size="small" style="width: 150px">
                <el-option label="生成图片" value="generate" />
                <el-option label="从文件夹读取" value="folder" />
              </el-select>
              
              <!-- 生成图片模式 -->
              <template v-if="scope.row.config.mode === 'generate'">
                <div style="display: flex; gap: 10px; flex-wrap: wrap">
                  <el-input-number 
                    v-model="scope.row.config.width" 
                    :min="1"
                    :max="10000"
                    placeholder="宽度"
                    size="small"
                    style="width: 120px"
                  />
                  <el-input-number 
                    v-model="scope.row.config.height" 
                    :min="1"
                    :max="10000"
                    placeholder="高度"
                    size="small"
                    style="width: 120px"
                  />
                  <el-select v-model="scope.row.config.format" size="small" style="width: 120px">
                    <el-option label="PNG" value="png" />
                    <el-option label="JPEG" value="jpeg" />
                    <el-option label="GIF" value="gif" />
                  </el-select>
                </div>
              </template>
              
              <!-- 从文件夹读取模式 -->
              <template v-if="scope.row.config.mode === 'folder'">
                <el-input 
                  v-model="scope.row.config.folderPath" 
                  placeholder="文件夹路径（绝对路径或相对路径）"
                  size="small"
                />
                <el-input 
                  v-model="scope.row.config.extensionsText" 
                  placeholder="文件扩展名过滤（可选，用逗号分隔，如 jpg,png,gif）"
                  size="small"
                  @blur="parseBinaryExtensions(scope.row)"
                />
                <el-checkbox v-model="scope.row.config.loop" size="small">循环读取</el-checkbox>
              </template>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="约束" width="150">
          <template #default="scope">
            <div style="display: flex; flex-direction: column; gap: 5px">
              <el-tag v-if="scope.row.isPrimaryKey" type="danger" size="small">主键</el-tag>
              <el-tag v-if="scope.row.isForeignKey" type="warning" size="small">外键</el-tag>
              <el-tag v-if="scope.row.isUnique" type="info" size="small">唯一</el-tag>
              <el-tag v-if="!scope.row.isNullable" type="success" size="small">非空</el-tag>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 操作按钮 -->
      <div style="margin-top: 20px; text-align: right">
        <el-button @click="$router.back()">取消</el-button>
        <el-button type="primary" @click="createTask">创建任务</el-button>
      </div>
    </el-card>

    <!-- 保存模板对话框 -->
    <el-dialog v-model="showSaveTemplateDialog" title="保存为模板" width="500px">
      <el-form :model="templateForm" label-width="100px">
        <el-form-item label="模板名称" required>
          <el-input v-model="templateForm.name" placeholder="输入模板名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input 
            v-model="templateForm.description" 
            type="textarea" 
            :rows="3"
            placeholder="输入模板描述（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSaveTemplateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTemplate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 加载模板对话框 -->
    <el-dialog v-model="showLoadTemplateDialog" title="从模板加载" width="700px">
      <el-table :data="templates" style="width: 100%">
        <el-table-column prop="name" label="模板名称" />
        <el-table-column prop="table_name" label="表名" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button size="small" type="primary" @click="loadTemplate(scope.row)">加载</el-button>
            <el-button size="small" type="danger" @click="deleteTemplate(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 批量设置对话框 -->
    <el-dialog v-model="showBatchDialog" title="批量设置规则" width="600px">
      <el-form label-width="120px">
        <el-form-item label="选择字段">
          <el-checkbox-group v-model="selectedFields">
            <el-checkbox 
              v-for="field in fieldRules" 
              :key="field.fieldName" 
              :label="field.fieldName"
            >
              {{ field.fieldName }} ({{ field.fieldType }})
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="batchRuleType" placeholder="选择规则类型" style="width: 100%">
            <el-option label="随机值" value="random_string" />
            <el-option label="随机数字" value="random_number" />
            <el-option label="随机日期" value="random_date" />
            <el-option label="固定值" value="fixed" />
            <el-option label="递增" value="increment" />
            <el-option label="列表选择" value="list" />
            <el-option label="函数" value="function" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则配置" v-if="batchRuleType">
          <div v-if="batchRuleType === 'random_string'">
            <el-input-number v-model="batchConfig.minLength" placeholder="最小长度" style="width: 120px; margin-right: 10px" />
            <el-input-number v-model="batchConfig.maxLength" placeholder="最大长度" style="width: 120px" />
          </div>
          <div v-else-if="batchRuleType === 'random_number'">
            <el-input-number v-model="batchConfig.min" placeholder="最小值" style="width: 120px; margin-right: 10px" />
            <el-input-number v-model="batchConfig.max" placeholder="最大值" style="width: 120px" />
          </div>
          <el-input v-else-if="batchRuleType === 'fixed'" v-model="batchConfig.value" placeholder="固定值" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" @click="applyBatchSettings">应用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import api from '../api'

const route = useRoute()
const router = useRouter()

const tableName = ref(route.query.table || '')
const connectionId = ref(route.query.connection_id || '')
const database = ref(route.query.database || '')
const tableSchema = ref(null)
const fieldRules = ref([])

const taskConfig = ref({
  name: '',
  totalRows: 1000,
  batchSize: 500,
  threadCount: 4
})

onMounted(async () => {
  if (!tableName.value) {
    ElMessage.error('表名不能为空')
    router.back()
    return
  }

  await loadTableSchema()
  await loadTemplates()
})

const loadTableSchema = async () => {
  try {
    const params = { database: database.value }
    if (connectionId.value) {
      params.connection_id = connectionId.value
    }
    const response = await api.getTableSchema(database.value, tableName.value, connectionId.value)
    tableSchema.value = response
    
    // 初始化字段规则
    fieldRules.value = response.fields.map(field => ({
      fieldName: field.name || field.Name,
      fieldType: field.type || field.Type,
      goType: (field.go_type || field.GoType || '').toLowerCase(),
      ruleType: getDefaultRuleType(field),
      config: getDefaultConfig(field),
      isPrimaryKey: field.is_primary_key || field.IsPrimaryKey || false,
      isForeignKey: field.is_foreign_key || field.IsForeignKey || false,
      isUnique: field.is_unique || field.IsUnique || false,
      isNullable: field.is_nullable !== false && field.IsNullable !== false,
      defaultValue: field.default_value || field.DefaultValue
    }))
  } catch (error) {
    ElMessage.error('加载表结构失败: ' + (error.response?.data?.error || error.message))
  }
}

const getDefaultRuleType = (field) => {
  if (field.is_primary_key || field.IsPrimaryKey) {
    return 'function' // 默认使用 UUID
  }
  if (field.is_foreign_key || field.IsForeignKey) {
    return 'foreign'
  }
  
  const type = (field.go_type || field.GoType || '').toLowerCase()
  if (type.includes('int')) {
    return 'random_number'
  } else if (type.includes('float') || type.includes('decimal')) {
    return 'random_number'
  } else if (type.includes('time') || type.includes('date')) {
    return 'random_date'
  } else if (type === 'bool') {
    return 'list'
  } else {
    return 'random_string'
  }
}

const getDefaultConfig = (field) => {
  const ruleType = field.ruleType || ''
  
  // 根据规则类型返回默认配置
  if (ruleType === 'reference') {
    return { expression: '', fields: [], fieldsText: '' }
  } else if (ruleType === 'geographic') {
    return { type: 'city', country: '' }
  } else if (ruleType === 'file') {
    return { filePath: '', fileType: 'txt', columnIndex: 0, loop: false }
  } else if (ruleType === 'binary') {
    return { mode: 'generate', width: 100, height: 100, format: 'png', folderPath: '', extensions: [], extensionsText: '', loop: false }
  }
  
  // 根据字段类型返回默认配置
  const type = (field.go_type || field.GoType || '').toLowerCase()
  
  if (type.includes('int')) {
    return { min: 1, max: 1000000, isInt: true }
  } else if (type.includes('float') || type.includes('decimal')) {
    return { min: 0, max: 10000, isInt: false }
  } else if (type.includes('time') || type.includes('date')) {
    return { startDate: '', endDate: '', format: '' }
  } else if (type === 'bool') {
    return { values: [true, false], valuesText: 'true,false' }
  } else {
    return { minLength: 5, maxLength: 20, charSet: 'all' }
  }
}

const onRuleTypeChange = (field) => {
  // 切换规则类型时重置配置
  field.config = getDefaultConfig(field)
}

const parseListValues = (field) => {
  if (field.config.valuesText) {
    field.config.values = field.config.valuesText.split(',').map(v => v.trim())
  }
}

const parseReferenceFields = (field) => {
  if (field.config.fieldsText) {
    field.config.fields = field.config.fieldsText.split(',').map(v => v.trim()).filter(v => v)
  } else {
    field.config.fields = []
  }
}

const parseBinaryExtensions = (field) => {
  if (field.config.extensionsText) {
    field.config.extensions = field.config.extensionsText.split(',').map(v => v.trim()).filter(v => v)
  } else {
    field.config.extensions = []
  }
}

// 根据字段类型获取可用的规则选项
const getAvailableRules = (field) => {
  const allRules = [
    { label: '随机值', value: 'random_string' },
    { label: '随机数字', value: 'random_number' },
    { label: '随机日期', value: 'random_date' },
    { label: '固定值', value: 'fixed' },
    { label: '递增', value: 'increment' },
    { label: '列表选择', value: 'list' },
    { label: '正则表达式', value: 'regex' },
    { label: '函数', value: 'function' },
    { label: '模板', value: 'template' },
    { label: '空值', value: 'null' },
    { label: '引用字段', value: 'reference' },
    { label: '地理数据', value: 'geographic' },
    { label: '从文件读取', value: 'file' },
    { label: '二进制/图片', value: 'binary' }
  ]

  const goType = (field.goType || '').toLowerCase()
  const fieldType = (field.fieldType || '').toLowerCase()

  // 如果是外键，只显示外键相关规则
  if (field.isForeignKey) {
    return [
      { label: '外键引用', value: 'foreign' },
      { label: '固定值', value: 'fixed' },
      { label: '空值', value: 'null' }
    ]
  }

  // 根据 Go 类型过滤规则
  if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
    // 数字类型
    return [
      { label: '随机数字', value: 'random_number' },
      { label: '固定值', value: 'fixed' },
      { label: '递增', value: 'increment' },
      { label: '列表选择', value: 'list' },
      { label: '函数', value: 'function' },
      { label: '引用字段', value: 'reference' },
      { label: '空值', value: 'null' }
    ]
  } else if (goType.includes('time') || goType.includes('date')) {
    // 日期时间类型
    return [
      { label: '随机日期', value: 'random_date' },
      { label: '固定值', value: 'fixed' },
      { label: '函数', value: 'function' },
      { label: '引用字段', value: 'reference' },
      { label: '空值', value: 'null' }
    ]
  } else if (goType === 'bool' || goType === 'boolean') {
    // 布尔类型
    return [
      { label: '列表选择', value: 'list' },
      { label: '固定值', value: 'fixed' },
      { label: '空值', value: 'null' }
    ]
  } else if (goType.includes('[]byte') || goType.includes('bytea') || 
             fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
    // 二进制类型
    return [
      { label: '二进制/图片', value: 'binary' },
      { label: '固定值', value: 'fixed' },
      { label: '空值', value: 'null' }
    ]
  } else {
    // 字符串类型（默认）
    return [
      { label: '随机值', value: 'random_string' },
      { label: '固定值', value: 'fixed' },
      { label: '列表选择', value: 'list' },
      { label: '正则表达式', value: 'regex' },
      { label: '函数', value: 'function' },
      { label: '模板', value: 'template' },
      { label: '引用字段', value: 'reference' },
      { label: '地理数据', value: 'geographic' },
      { label: '从文件读取', value: 'file' },
      { label: '空值', value: 'null' }
    ]
  }
}

const createTask = async () => {
  try {
    // 构建字段规则
    const rules = fieldRules.value.map(field => {
      const rule = {
        field_name: field.fieldName,
        field_type: field.fieldType,
        rule_type: field.ruleType,
        config: {},
        is_primary_key: field.isPrimaryKey,
        is_foreign_key: field.isForeignKey,
        is_unique: field.isUnique,
        is_nullable: field.isNullable,
        default_value: field.defaultValue
      }

      // 根据规则类型构建配置
      switch (field.ruleType) {
        case 'random_string':
          rule.config = {
            min_length: field.config.minLength || 10,
            max_length: field.config.maxLength || 50,
            char_set: field.config.charSet || 'all'
          }
          break
        case 'random_number':
          rule.config = {
            min: field.config.min || 0,
            max: field.config.max || 1000,
            is_int: field.config.isInt || false
          }
          break
        case 'random_date':
          rule.config = {
            start_date: field.config.startDate || '',
            end_date: field.config.endDate || '',
            format: field.config.format || ''
          }
          break
        case 'fixed':
          rule.config = { value: field.config.value || '' }
          break
        case 'increment':
          rule.config = {
            start_value: field.config.startValue || 1,
            step: field.config.step || 1,
            cycle: field.config.cycle || false
          }
          break
        case 'list':
          rule.config = {
            values: field.config.values || [],
            allow_repeat: true
          }
          break
        case 'regex':
          rule.config = { pattern: field.config.pattern || '' }
          break
        case 'function':
          rule.config = {
            func_name: field.config.funcName || 'UUID',
            params: []
          }
          break
        case 'template':
          rule.config = { template: field.config.template || '' }
          break
        case 'null':
          rule.config = {
            probability: (field.config.probability || 0) / 100
          }
          break
        case 'reference':
          rule.config = {
            expression: field.config.expression || '',
            fields: field.config.fields || []
          }
          break
        case 'geographic':
          rule.config = {
            type: field.config.type || 'city',
            country: field.config.country || ''
          }
          break
        case 'file':
          rule.config = {
            file_path: field.config.filePath || '',
            file_type: field.config.fileType || 'txt',
            column_index: field.config.columnIndex || 0,
            loop: field.config.loop || false
          }
          break
        case 'binary':
          rule.config = {
            mode: field.config.mode || 'generate',
            width: field.config.width || 100,
            height: field.config.height || 100,
            format: field.config.format || 'png',
            folder_path: field.config.folderPath || '',
            extensions: field.config.extensions || [],
            loop: field.config.loop || false
          }
          break
      }

      return rule
    })

    // 构建任务配置
    const config = {
      table_name: tableName.value,
      database: database.value,
      total_rows: taskConfig.value.totalRows,
      batch_size: taskConfig.value.batchSize,
      field_rules: rules,
      use_transaction: true,
      on_error: 'skip',
      retry_times: 3
    }

    const taskName = taskConfig.value.name || `${tableName.value}_${Date.now()}`
    
    await api.createTask(taskName, connectionId.value, config)
    ElMessage.success('任务创建成功')
    router.push('/tasks')
  } catch (error) {
    ElMessage.error('创建任务失败: ' + (error.response?.data?.error || error.message))
  }
}

const loadTemplates = async () => {
  try {
    const response = await api.getTemplates(tableName.value)
    templates.value = response.templates || []
  } catch (error) {
    console.error('加载模板列表失败:', error)
  }
}

const saveTemplate = async () => {
  if (!templateForm.value.name) {
    ElMessage.warning('请输入模板名称')
    return
  }

  try {
    // 构建配置
    const rules = fieldRules.value.map(field => {
      const rule = {
        field_name: field.fieldName,
        field_type: field.fieldType,
        rule_type: field.ruleType,
        config: {},
        is_primary_key: field.isPrimaryKey,
        is_foreign_key: field.isForeignKey,
        is_unique: field.isUnique,
        is_nullable: field.isNullable,
        default_value: field.defaultValue
      }

      // 根据规则类型构建配置（复用 createTask 中的逻辑）
      switch (field.ruleType) {
        case 'random_string':
          rule.config = {
            min_length: field.config.minLength || 10,
            max_length: field.config.maxLength || 50,
            char_set: field.config.charSet || 'all'
          }
          break
        case 'random_number':
          rule.config = {
            min: field.config.min || 0,
            max: field.config.max || 1000,
            is_int: field.config.isInt || false
          }
          break
        case 'random_date':
          rule.config = {
            start_date: field.config.startDate || '',
            end_date: field.config.endDate || '',
            format: field.config.format || ''
          }
          break
        case 'fixed':
          rule.config = { value: field.config.value || '' }
          break
        case 'increment':
          rule.config = {
            start_value: field.config.startValue || 1,
            step: field.config.step || 1,
            cycle: field.config.cycle || false
          }
          break
        case 'list':
          rule.config = {
            values: field.config.values || [],
            allow_repeat: true
          }
          break
        case 'regex':
          rule.config = { pattern: field.config.pattern || '' }
          break
        case 'function':
          rule.config = {
            func_name: field.config.funcName || 'UUID',
            params: []
          }
          break
        case 'template':
          rule.config = { template: field.config.template || '' }
          break
        case 'null':
          rule.config = {
            probability: (field.config.probability || 0) / 100
          }
          break
        case 'reference':
          rule.config = {
            expression: field.config.expression || '',
            fields: field.config.fields || []
          }
          break
        case 'geographic':
          rule.config = {
            type: field.config.type || 'city',
            country: field.config.country || ''
          }
          break
        case 'file':
          rule.config = {
            file_path: field.config.filePath || '',
            file_type: field.config.fileType || 'txt',
            column_index: field.config.columnIndex || 0,
            loop: field.config.loop || false
          }
          break
        case 'binary':
          rule.config = {
            mode: field.config.mode || 'generate',
            width: field.config.width || 100,
            height: field.config.height || 100,
            format: field.config.format || 'png',
            folder_path: field.config.folderPath || '',
            extensions: field.config.extensions || [],
            loop: field.config.loop || false
          }
          break
      }

      return rule
    })

    const config = {
      table_name: tableName.value,
      database: database.value,
      total_rows: taskConfig.value.totalRows,
      batch_size: taskConfig.value.batchSize,
      field_rules: rules,
      use_transaction: true,
      on_error: 'skip',
      retry_times: 3
    }

    await api.saveTemplate({
      name: templateForm.value.name,
      description: templateForm.value.description,
      table_name: tableName.value,
      config
    })

    ElMessage.success('模板保存成功')
    showSaveTemplateDialog.value = false
    templateForm.value = { name: '', description: '' }
    await loadTemplates()
  } catch (error) {
    ElMessage.error('保存模板失败: ' + (error.response?.data?.error || error.message))
  }
}

const loadTemplate = (template) => {
  try {
    // 加载模板配置
    const config = template.config

    // 恢复基本信息
    taskConfig.value.totalRows = config.total_rows || 1000
    taskConfig.value.batchSize = config.batch_size || 500
    taskConfig.value.threadCount = 4

    // 恢复字段规则
    fieldRules.value = config.field_rules.map(rule => {
      const field = tableSchema.value.fields.find(f => 
        (f.name || f.Name) === rule.field_name
      ) || {}

      const fieldRule = {
        fieldName: rule.field_name,
        fieldType: rule.field_type,
        goType: (field.go_type || field.GoType || '').toLowerCase(),
        ruleType: rule.rule_type,
        config: {},
        isPrimaryKey: rule.is_primary_key || false,
        isForeignKey: rule.is_foreign_key || false,
        isUnique: rule.is_unique || false,
        isNullable: rule.is_nullable !== false,
        defaultValue: rule.default_value
      }

      // 恢复配置
      const ruleConfig = rule.config || {}
      switch (rule.rule_type) {
        case 'random_string':
          fieldRule.config = {
            minLength: ruleConfig.min_length || 10,
            maxLength: ruleConfig.max_length || 50,
            charSet: ruleConfig.char_set || 'all'
          }
          break
        case 'random_number':
          fieldRule.config = {
            min: ruleConfig.min || 0,
            max: ruleConfig.max || 1000,
            isInt: ruleConfig.is_int || false
          }
          break
        case 'random_date':
          fieldRule.config = {
            startDate: ruleConfig.start_date || '',
            endDate: ruleConfig.end_date || '',
            format: ruleConfig.format || ''
          }
          break
        case 'fixed':
          fieldRule.config = { value: ruleConfig.value || '' }
          break
        case 'increment':
          fieldRule.config = {
            startValue: ruleConfig.start_value || 1,
            step: ruleConfig.step || 1,
            cycle: ruleConfig.cycle || false
          }
          break
        case 'list':
          fieldRule.config = {
            values: ruleConfig.values || [],
            valuesText: (ruleConfig.values || []).join(',')
          }
          break
        case 'regex':
          fieldRule.config = { pattern: ruleConfig.pattern || '' }
          break
        case 'function':
          fieldRule.config = {
            funcName: ruleConfig.func_name || 'UUID',
            params: ruleConfig.params || []
          }
          break
        case 'template':
          fieldRule.config = { template: ruleConfig.template || '' }
          break
        case 'null':
          fieldRule.config = {
            probability: (ruleConfig.probability || 0) * 100
          }
          break
        case 'reference':
          fieldRule.config = {
            expression: ruleConfig.expression || '',
            fields: ruleConfig.fields || [],
            fieldsText: (ruleConfig.fields || []).join(',')
          }
          break
        case 'geographic':
          fieldRule.config = {
            type: ruleConfig.type || 'city',
            country: ruleConfig.country || ''
          }
          break
        case 'file':
          fieldRule.config = {
            filePath: ruleConfig.file_path || '',
            fileType: ruleConfig.file_type || 'txt',
            columnIndex: ruleConfig.column_index || 0,
            loop: ruleConfig.loop || false
          }
          break
        case 'binary':
          fieldRule.config = {
            mode: ruleConfig.mode || 'generate',
            width: ruleConfig.width || 100,
            height: ruleConfig.height || 100,
            format: ruleConfig.format || 'png',
            folderPath: ruleConfig.folder_path || '',
            extensions: ruleConfig.extensions || [],
            extensionsText: (ruleConfig.extensions || []).join(','),
            loop: ruleConfig.loop || false
          }
          break
        default:
          fieldRule.config = getDefaultConfig(field)
      }

      return fieldRule
    })

    ElMessage.success('模板加载成功')
    showLoadTemplateDialog.value = false
  } catch (error) {
    ElMessage.error('加载模板失败: ' + (error.response?.data?.error || error.message))
  }
}

const deleteTemplate = async (templateId) => {
  try {
    await api.deleteTemplate(templateId)
    ElMessage.success('模板已删除')
    await loadTemplates()
  } catch (error) {
    ElMessage.error('删除模板失败: ' + (error.response?.data?.error || error.message))
  }
}

const selectedFields = ref([])
const batchRuleType = ref('')
const batchConfig = ref({
  minLength: 10,
  maxLength: 50,
  min: 0,
  max: 1000,
  value: ''
})

const applyBatchSettings = () => {
  if (selectedFields.value.length === 0) {
    ElMessage.warning('请选择要设置的字段')
    return
  }
  if (!batchRuleType.value) {
    ElMessage.warning('请选择规则类型')
    return
  }

  selectedFields.value.forEach(fieldName => {
    const field = fieldRules.value.find(f => f.fieldName === fieldName)
    if (field) {
      field.ruleType = batchRuleType.value
      onRuleTypeChange(field)
      
      // 应用配置
      if (batchRuleType.value === 'random_string') {
        field.config.minLength = batchConfig.value.minLength
        field.config.maxLength = batchConfig.value.maxLength
      } else if (batchRuleType.value === 'random_number') {
        field.config.min = batchConfig.value.min
        field.config.max = batchConfig.value.max
      } else if (batchRuleType.value === 'fixed') {
        field.config.value = batchConfig.value.value
      }
    }
  })

  ElMessage.success(`已批量设置 ${selectedFields.value.length} 个字段`)
  showBatchDialog.value = false
  selectedFields.value = []
  batchRuleType.value = ''
}

const resetAllRules = () => {
  fieldRules.value.forEach(field => {
    field.ruleType = getDefaultRuleType(field)
    field.config = getDefaultConfig(field)
  })
  ElMessage.success('已重置所有字段规则')
}
</script>

<style scoped>
.el-table {
  margin-top: 20px;
}
</style>

