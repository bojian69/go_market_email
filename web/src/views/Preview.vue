<template>
  <div class="preview">
    <el-card title="邮件预览">
      <div class="preview-toolbar">
        <div class="toolbar-section">
          <div class="template-selector">
            <label class="selector-label">选择模板</label>
            <el-select 
              v-model="previewForm.template_id" 
              @change="loadTemplate"
              placeholder="请选择邮件模板"
              size="large"
              class="template-select"
            >
              <el-option
                v-for="template in templates"
                :key="template.id"
                :label="template.name"
                :value="template.id"
                class="template-option"
              >
                <div class="template-option-content">
                  <div class="template-name">{{ template.name }}</div>
                  <div class="template-subject">{{ template.subject }}</div>
                </div>
              </el-option>
            </el-select>
          </div>
          
          <div class="action-buttons">
            <el-button type="primary" @click="showDataDialog = true" size="large">
              <el-icon><Setting /></el-icon>
              设置变量数据
            </el-button>
            
            <el-button @click="sendTestEmail" :loading="sending" size="large">
              <el-icon><Message /></el-icon>
              发送测试邮件
            </el-button>
          </div>
        </div>
      </div>

      <div v-if="currentTemplate" class="preview-content">
        <div class="email-preview">
          <div class="email-header">
            <div class="email-field">
              <label>发件人：</label>
              <span>{{ smtpConfig.from_name }} &lt;{{ smtpConfig.username }}&gt;</span>
            </div>
            <div class="email-field">
              <label>收件人：</label>
              <el-input v-model="previewForm.test_email" placeholder="输入测试邮箱" style="width: 300px" />
            </div>
            <div class="email-field">
              <label>主题：</label>
              <span class="subject">{{ renderedSubject }}</span>
            </div>
          </div>
          
          <div class="email-body">
            <div class="content-tabs">
              <el-tabs v-model="activeTab">
                <el-tab-pane label="预览效果" name="preview">
                  <div class="email-content" v-html="renderedContent"></div>
                </el-tab-pane>
                
                <el-tab-pane label="HTML源码" name="html">
                  <el-input
                    :model-value="renderedContent"
                    type="textarea"
                    :rows="20"
                    readonly
                  />
                </el-tab-pane>
                
                <el-tab-pane label="原始模板" name="template">
                  <el-input
                    :model-value="currentTemplate.content"
                    type="textarea"
                    :rows="20"
                    readonly
                  />
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>
        </div>
      </div>
      
      <div v-else class="loading-state">
        <el-empty description="正在加载模板...">
          <template #image>
            <el-icon size="60"><Loading /></el-icon>
          </template>
          <template #description>
            <p>正在加载模板数据，请稍候...</p>
            <el-button type="primary" @click="loadTemplates" style="margin-top: 10px;">
              重新加载
            </el-button>
          </template>
        </el-empty>
      </div>
    </el-card>

    <!-- 变量数据设置对话框 -->
    <el-dialog v-model="showDataDialog" title="设置变量数据" width="600px">
      <div v-if="templateVariables.length > 0">
        <el-form label-width="120px">
          <el-form-item
            v-for="variable in templateVariables"
            :key="variable"
            :label="variable"
          >
            <el-input
              v-model="variableData[variable]"
              :placeholder="`输入 ${variable} 的值`"
              @input="updatePreview"
            />
          </el-form-item>
        </el-form>
        
        <div class="data-templates">
          <h4>快速填充</h4>
          <el-button-group>
            <el-button @click="fillSampleData">示例数据</el-button>
            <el-button @click="clearVariableData">清空数据</el-button>
            <el-button @click="triggerFileUpload">Excel导入</el-button>
            <el-button @click="triggerJsonUpload">JSON导入</el-button>
          </el-button-group>
          <input 
            ref="fileInput" 
            type="file" 
            accept=".xlsx,.xls" 
            @change="handleFileUpload" 
            style="display: none"
          />
          <input 
            ref="jsonInput" 
            type="file" 
            accept=".json" 
            @change="handleJsonUpload" 
            style="display: none"
          />
        </div>
      </div>
      <el-empty v-else description="当前模板没有变量" />
      
      <template #footer>
        <el-button @click="showDataDialog = false">关闭</el-button>
        <el-button type="primary" @click="updatePreview">更新预览</el-button>
      </template>
    </el-dialog>

    <!-- AI生成内容对话框 -->
    <el-dialog v-model="showAIDialog" title="AI生成内容" width="700px">
      <el-form :model="aiForm" label-width="100px">
        <el-form-item label="AI提示词">
          <el-input
            v-model="aiForm.prompt"
            type="textarea"
            :rows="4"
            placeholder="输入AI提示词，可以使用变量 {{变量名}}"
          />
        </el-form-item>
        
        <el-form-item label="AI服务">
          <el-radio-group v-model="aiForm.service">
            <el-radio label="openai">OpenAI GPT-4</el-radio>
            <el-radio label="custom">自定义API</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      
      <div v-if="aiResult" class="ai-result">
        <h4>AI生成结果</h4>
        <el-input
          :model-value="aiResult"
          type="textarea"
          :rows="6"
          readonly
        />
        <el-button @click="useAIResult" type="primary" style="margin-top: 10px">
          使用此结果
        </el-button>
      </div>
      
      <template #footer>
        <el-button @click="showAIDialog = false">关闭</el-button>
        <el-button type="primary" @click="generateAIContent" :loading="aiLoading">
          生成内容
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'
import { Setting, Message } from '@element-plus/icons-vue'
import api from '@/utils/api'

const route = useRoute()
const templates = ref([])
const currentTemplate = ref(null)
const templateVariables = ref([])
const variableData = ref({})
const activeTab = ref('preview')
const sending = ref(false)

const showDataDialog = ref(false)
const showAIDialog = ref(false)
const aiLoading = ref(false)
const aiResult = ref('')
const fileInput = ref(null)
const jsonInput = ref(null)

const previewForm = ref({
  template_id: null,
  test_email: ''
})

const aiForm = ref({
  prompt: '',
  service: 'openai'
})

const smtpConfig = ref({
  from_name: 'AI Marketing',
  username: 'noreply@example.com'
})

// 渲染后的内容
const renderedSubject = computed(() => {
  if (!currentTemplate.value) return ''
  return replaceVariables(currentTemplate.value.subject, variableData.value)
})

const renderedContent = computed(() => {
  if (!currentTemplate.value) return ''
  return replaceVariables(currentTemplate.value.content, variableData.value)
})

const loadTemplates = async () => {
  try {
    const response = await api.get('/templates')
    templates.value = response.data || []
    
    // 如果URL中有template_id参数，自动选择
    const templateId = route.query.template_id
    if (templateId) {
      previewForm.value.template_id = parseInt(templateId as string)
      await loadTemplate()
    }
  } catch (error) {
    console.error('加载模板失败:', error)
    // 如果没有模板，创建一个默认模板
    templates.value = [{
      id: 1,
      name: '默认模板',
      subject: '欢迎 {{name}} 加入我们！',
      content: '亲爱的 {{name}}，<br><br>欢迎来到 {{company}}！<br><br>我们很高兴您能加入我们的团队。'
    }]
  }
}

const loadTemplate = async () => {
  if (!previewForm.value.template_id) return
  
  try {
    const response = await api.get(`/templates/${previewForm.value.template_id}`)
    currentTemplate.value = response.data
    
    // 提取变量
    const variablesResponse = await api.post('/templates/extract-variables', {
      content: currentTemplate.value.content,
      subject: currentTemplate.value.subject
    })
    templateVariables.value = variablesResponse.data || []
    
    // 初始化变量数据
    const newVariableData = {}
    templateVariables.value.forEach(variable => {
      newVariableData[variable] = variableData.value[variable] || ''
    })
    variableData.value = newVariableData
    
  } catch (error) {
    console.error('加载模板失败:', error)
    // 使用默认模板
    if (templates.value.length > 0) {
      currentTemplate.value = templates.value[0]
      templateVariables.value = ['name', 'company']
      variableData.value = {
        name: '张三',
        company: 'ABC公司'
      }
    }
  }
}

const replaceVariables = (template: string, data: Record<string, any>) => {
  let result = template
  Object.keys(data).forEach(key => {
    const placeholder = `{{${key}}}`
    const value = data[key] || `[${key}]`
    result = result.replace(new RegExp(placeholder.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), value)
  })
  return result
}

const updatePreview = () => {
  showDataDialog.value = false
}

const fillSampleData = () => {
  const sampleData = {
    name: '张三',
    email: 'zhangsan@example.com',
    company: 'ABC公司',
    city: '北京',
    product: '智能营销系统',
    date: new Date().toLocaleDateString()
  }
  
  templateVariables.value.forEach(variable => {
    if (sampleData[variable]) {
      variableData.value[variable] = sampleData[variable]
    }
  })
}

const clearVariableData = () => {
  templateVariables.value.forEach(variable => {
    variableData.value[variable] = ''
  })
}

const sendTestEmail = async () => {
  if (!previewForm.value.test_email) {
    ElMessage.warning('请输入测试邮箱')
    return
  }
  
  if (!currentTemplate.value) {
    ElMessage.warning('请选择邮件模板')
    return
  }
  
  sending.value = true
  try {
    await api.post('/emails/test', {
      template_id: previewForm.value.template_id,
      email: previewForm.value.test_email,
      data: variableData.value
    })
    ElMessage.success('测试邮件发送成功')
  } catch (error) {
    console.error('发送测试邮件失败:', error)
  } finally {
    sending.value = false
  }
}

const generateAIContent = async () => {
  if (!aiForm.value.prompt.trim()) {
    ElMessage.warning('请输入AI提示词')
    return
  }
  
  aiLoading.value = true
  try {
    const response = await api.post('/ai/generate', {
      prompt: aiForm.value.prompt,
      data: variableData.value,
      service: aiForm.value.service
    })
    aiResult.value = response.data.result
  } catch (error) {
    console.error('AI生成失败:', error)
  } finally {
    aiLoading.value = false
  }
}

const useAIResult = () => {
  // 将AI结果添加到变量数据中
  variableData.value['ai_result'] = aiResult.value
  showAIDialog.value = false
  ElMessage.success('AI结果已添加到变量数据')
}

const triggerFileUpload = () => {
  fileInput.value?.click()
}

const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  const formData = new FormData()
  formData.append('file', file)
  
  try {
    const response = await api.post('/data/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    const excelData = response.data
    if (excelData && excelData.length > 1) {
      // 第一行是字段名，第二行是数据
      const headers = excelData[0]
      const values = excelData[1]
      
      // 将Excel数据映射到变量数据
      headers.forEach((header, index) => {
        if (templateVariables.value.includes(header) && values[index]) {
          variableData.value[header] = values[index]
        }
      })
      
      ElMessage.success('Excel数据导入成功')
    } else {
      ElMessage.warning('Excel文件格式不正确，请确保第一行为字段名，第二行为数据')
    }
  } catch (error) {
    console.error('Excel导入失败:', error)
    ElMessage.error('Excel导入失败，请检查文件格式')
  }
  
  // 清空文件输入
  event.target.value = ''
}

const triggerJsonUpload = () => {
  jsonInput.value?.click()
}

const handleJsonUpload = (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const jsonData = JSON.parse(e.target.result)
      
      // 将JSON数据映射到变量数据
      Object.keys(jsonData).forEach(key => {
        if (templateVariables.value.includes(key)) {
          variableData.value[key] = jsonData[key]
        }
      })
      
      ElMessage.success('JSON数据导入成功')
    } catch (error) {
      console.error('JSON解析失败:', error)
      ElMessage.error('JSON文件格式错误，请检查文件内容')
    }
  }
  
  reader.readAsText(file)
  // 清空文件输入
  event.target.value = ''
}

onMounted(async () => {
  try {
    await loadTemplates()
    // 如果有template_id参数但没有加载成功，尝试加载默认模板
    if (route.query.template_id && !currentTemplate.value && templates.value.length > 0) {
      previewForm.value.template_id = templates.value[0].id
      await loadTemplate()
    }
  } catch (error) {
    console.error('初始化失败:', error)
  }
})
</script>

<style scoped>
.preview {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

:deep(.el-card) {
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  border: none;
  overflow: hidden;
}

:deep(.el-card__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  font-size: 18px;
  padding: 20px 24px;
  border-bottom: none;
}

:deep(.el-card__body) {
  padding: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

.preview-toolbar {
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(102, 126, 234, 0.2);
}

.toolbar-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.template-selector {
  flex: 1;
  max-width: 400px;
}

.selector-label {
  display: block;
  color: white;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.template-select {
  width: 100%;
}

:deep(.template-select .el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

:deep(.template-select .el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.6);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
}

:deep(.template-select .el-input__wrapper.is-focus) {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.template-option-content {
  padding: 4px 0;
}

.template-name {
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 2px;
}

.template-subject {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

:deep(.action-buttons .el-button) {
  border-radius: 10px;
  font-weight: 500;
  padding: 12px 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.1);
  color: white;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

:deep(.action-buttons .el-button:hover) {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}

:deep(.action-buttons .el-button--primary) {
  background: rgba(64, 158, 255, 0.8);
  border-color: rgba(64, 158, 255, 0.6);
}

:deep(.action-buttons .el-button--primary:hover) {
  background: rgba(64, 158, 255, 1);
  border-color: rgba(64, 158, 255, 0.8);
}

.email-preview {
  margin: 24px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  background: white;
}

.email-header {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.email-field {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 0;
}

.email-field:last-child {
  margin-bottom: 0;
}

.email-field label {
  font-weight: 600;
  width: 100px;
  color: #495057;
  font-size: 14px;
}

.subject {
  font-weight: 600;
  color: #212529;
  font-size: 16px;
}

.email-body {
  background: white;
}

.content-tabs {
  min-height: 400px;
}

:deep(.content-tabs .el-tabs__header) {
  margin: 0;
  padding: 0 24px;
  background: #f8f9fa;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

:deep(.content-tabs .el-tabs__nav-wrap) {
  padding: 12px 0;
}

:deep(.content-tabs .el-tabs__item) {
  font-weight: 500;
  color: #6c757d;
  border-radius: 8px 8px 0 0;
  margin-right: 4px;
}

:deep(.content-tabs .el-tabs__item.is-active) {
  color: #495057;
  background: white;
  border-bottom-color: white;
}

.email-content {
  padding: 24px;
  line-height: 1.6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: white;
}

.data-templates {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.data-templates h4 {
  margin: 0 0 12px 0;
  color: #495057;
  font-weight: 600;
}

.ai-result {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.ai-result h4 {
  margin: 0 0 12px 0;
  color: #495057;
  font-weight: 600;
}

:deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
}

:deep(.el-dialog__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 20px 24px;
  margin: 0;
}

:deep(.el-dialog__title) {
  color: white;
  font-weight: 600;
}

:deep(.el-dialog__headerbtn .el-dialog__close) {
  color: white;
}

:deep(.el-dialog__headerbtn) {
  display: none;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px 24px;
  background: #f8f9fa;
}
</style>