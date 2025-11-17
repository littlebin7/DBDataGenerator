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

    <!-- 模板详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="模板详情" width="80%" :before-close="closeDetail">
      <div v-if="selectedTemplate">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="模板名称">{{ selectedTemplate.name }}</el-descriptions-item>
          <el-descriptions-item label="表名">{{ selectedTemplate.table_name }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ selectedTemplate.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(selectedTemplate.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider>字段配置</el-divider>
        <el-table :data="selectedTemplate.config?.fields || []" border style="margin-top: 20px">
          <el-table-column prop="field_name" label="字段名" width="150" />
          <el-table-column prop="rule_type" label="规则类型" width="150" />
          <el-table-column label="规则配置" min-width="300">
            <template #default="scope">
              <pre style="margin: 0; font-size: 12px">{{ JSON.stringify(scope.row.config, null, 2) }}</pre>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'
import { formatTime } from '../utils/formatters'

const loading = ref(false)
const templates = ref([])
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
  loading.value = true
  try {
    const data = await api.getTemplates()
    templates.value = data.templates || []
  } catch (error) {
    ElMessage.error('获取模板列表失败: ' + (error.formattedMessage || error.message))
  } finally {
    loading.value = false
  }
}

const viewTemplate = async (template) => {
  try {
    const data = await api.getTemplate(template.id)
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
    await api.deleteTemplate(templateId)
    ElMessage.success('模板已删除')
    await loadTemplates()
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

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.template-view {
  padding: 0;
}
</style>

