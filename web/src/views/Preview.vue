<template>
  <div class="preview">
    <el-card title="邮件预览">
      <div class="preview-toolbar">
        <el-form :model="previewForm" :inline="true">
          <el-form-item label="选择模板">
            <el-select v-model="previewForm.template_id" @change="loadTemplate">
              <el-option
                v-for="template in templates"
                :key="template.id"
                :label="template.name"
                :value="template.id"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="showDataDialog = true">
              设置变量数据
            </el-button>
          </el-form-item>
          
          <el-form-item>
            <el-button @click="sendTestEmail" :loading="sending">
              发送测试邮件
            </el-button>
          </el-form-item>
        </el-form>
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
      
      <el-empty v-else description="请选择邮件模板" />
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
          </el-button-group>
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
    templates.value = response.data
    
    // 如果URL中有template_id参数，自动选择
    const templateId = route.query.template_id
    if (templateId) {
      previewForm.value.template_id = parseInt(templateId as string)
      await loadTemplate()
    }
  } catch (error) {
    console.error('加载模板失败:', error)
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
    templateVariables.value = variablesResponse.data
    
    // 初始化变量数据
    const newVariableData = {}
    templateVariables.value.forEach(variable => {
      newVariableData[variable] = variableData.value[variable] || ''
    })
    variableData.value = newVariableData
    
  } catch (error) {
    console.error('加载模板失败:', error)
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

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.preview {
  padding: 20px;
}

.preview-toolbar {
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--el-border-color);
}

.email-preview {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  overflow: hidden;
}

.email-header {
  background: var(--el-bg-color-page);
  padding: 15px;
  border-bottom: 1px solid var(--el-border-color);
}

.email-field {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.email-field label {
  font-weight: bold;
  width: 80px;
  color: var(--el-text-color-regular);
}

.subject {
  font-weight: bold;
  color: var(--el-text-color-primary);
}

.email-body {
  background: white;
}

.content-tabs {
  min-height: 400px;
}

.email-content {
  padding: 20px;
  line-height: 1.6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.data-templates {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid var(--el-border-color);
}

.ai-result {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid var(--el-border-color);
}
</style>