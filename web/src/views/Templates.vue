<template>
  <div class="templates">
    <div class="toolbar">
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建模板
      </el-button>
      
      <div class="search-box">
        <el-input
          v-model="searchText"
          placeholder="搜索模板..."
          @input="handleSearch"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <el-table :data="templates" style="width: 100%">
      <el-table-column prop="name" label="模板名称" />
      <el-table-column prop="subject" label="邮件主题" show-overflow-tooltip />
      <el-table-column prop="version" label="版本" width="80" />
      <el-table-column prop="variables" label="变量">
        <template #default="{ row }">
          <el-tag
            v-for="variable in JSON.parse(row.variables || '[]')"
            :key="variable"
            size="small"
            style="margin-right: 5px"
          >
            {{ variable }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="editTemplate(row)">编辑</el-button>
          <el-button size="small" @click="previewTemplate(row)">预览</el-button>
          <el-button size="small" type="danger" @click="deleteTemplate(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="loadTemplates"
      @size-change="loadTemplates"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 20px; text-align: right"
    />

    <!-- 创建/编辑模板对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingTemplate ? '编辑模板' : '创建模板'"
      width="800px"
    >
      <el-form :model="templateForm" label-width="100px">
        <el-form-item label="模板名称" required>
          <el-input v-model="templateForm.name" />
        </el-form-item>
        
        <el-form-item label="邮件主题" required>
          <el-input v-model="templateForm.subject" />
        </el-form-item>
        
        <el-form-item label="邮件内容" required>
          <div class="editor-container">
            <div class="editor-toolbar">
              <el-button size="small" @click="insertVariable">插入变量</el-button>
              <el-button size="small" @click="extractVariables">提取变量</el-button>
            </div>
            <el-input
              v-model="templateForm.content"
              type="textarea"
              :rows="15"
              placeholder="输入邮件内容，使用 {{变量名}} 格式插入变量"
            />
          </div>
        </el-form-item>
        
        <el-form-item label="检测到的变量">
          <el-tag
            v-for="variable in detectedVariables"
            :key="variable"
            style="margin-right: 5px"
          >
            {{ variable }}
          </el-tag>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTemplate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 变量插入对话框 -->
    <el-dialog v-model="showVariableDialog" title="插入变量" width="400px">
      <el-form>
        <el-form-item label="变量名">
          <el-input v-model="newVariable" placeholder="输入变量名" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showVariableDialog = false">取消</el-button>
        <el-button type="primary" @click="insertVariableToContent">插入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'

const templates = ref([])
const searchText = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const showCreateDialog = ref(false)
const showVariableDialog = ref(false)
const editingTemplate = ref(null)
const newVariable = ref('')

const templateForm = ref({
  name: '',
  subject: '',
  content: '',
  project_id: 1
})

const detectedVariables = ref([])

const loadTemplates = async () => {
  try {
    const response = await api.get('/templates', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        search: searchText.value
      }
    })
    templates.value = response.data
    total.value = response.total
  } catch (error) {
    console.error('加载模板失败:', error)
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadTemplates()
}

const editTemplate = (template: any) => {
  editingTemplate.value = template
  templateForm.value = { ...template }
  extractVariables()
  showCreateDialog.value = true
}

const deleteTemplate = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个模板吗？', '确认删除')
    await api.delete(`/templates/${id}`)
    ElMessage.success('删除成功')
    loadTemplates()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模板失败:', error)
    }
  }
}

const saveTemplate = async () => {
  try {
    if (editingTemplate.value) {
      await api.put(`/templates/${editingTemplate.value.id}`, templateForm.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/templates', templateForm.value)
      ElMessage.success('创建成功')
    }
    
    showCreateDialog.value = false
    resetForm()
    loadTemplates()
  } catch (error) {
    console.error('保存模板失败:', error)
  }
}

const extractVariables = async () => {
  try {
    const response = await api.post('/templates/extract-variables', {
      content: templateForm.value.content,
      subject: templateForm.value.subject
    })
    detectedVariables.value = response.data
  } catch (error) {
    console.error('提取变量失败:', error)
  }
}

const insertVariable = () => {
  showVariableDialog.value = true
}

const insertVariableToContent = () => {
  if (newVariable.value) {
    templateForm.value.content += `{{${newVariable.value}}}`
    newVariable.value = ''
    showVariableDialog.value = false
    extractVariables()
  }
}

const previewTemplate = (template: any) => {
  // 跳转到预览页面
  window.open(`/dashboard/preview?template_id=${template.id}`, '_blank')
}

const resetForm = () => {
  templateForm.value = {
    name: '',
    subject: '',
    content: '',
    project_id: 1
  }
  detectedVariables.value = []
  editingTemplate.value = null
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

// 监听内容变化自动提取变量
watch([() => templateForm.value.content, () => templateForm.value.subject], () => {
  if (templateForm.value.content || templateForm.value.subject) {
    extractVariables()
  }
}, { debounce: 500 })

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.templates {
  padding: 24px;
  background: transparent;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
}

.search-box {
  width: 320px;
}

.editor-container {
  width: 100%;
}

.editor-toolbar {
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
}

:deep(.el-table) {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

:deep(.el-pagination) {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  background: rgba(255, 255, 255, 0.95);
  padding: 16px;
  border-radius: 12px;
  backdrop-filter: blur(10px);
}

:deep(.el-button) {
  border-radius: 8px;
  font-weight: 500;
}

:deep(.el-button--small) {
  padding: 6px 12px;
}
</style>