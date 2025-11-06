<template>
  <div class="ai-generate">
    <el-card title="AI内容生成">
      <el-form :model="aiForm" label-width="120px">
        <el-form-item label="选择模板">
          <el-select v-model="aiForm.template_id" placeholder="选择邮件模板">
            <el-option
              v-for="template in templates"
              :key="template.id"
              :label="template.name"
              :value="template.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="AI提示词">
          <el-input
            v-model="aiForm.prompt"
            type="textarea"
            :rows="6"
            placeholder="输入AI提示词，可以使用变量 {{变量名}}"
          />
        </el-form-item>
        
        <el-form-item label="变量数据">
          <el-input
            v-model="aiForm.dataJson"
            type="textarea"
            :rows="4"
            placeholder='输入JSON格式的变量数据，例如: {"name": "张三", "city": "北京"}'
          />
        </el-form-item>
        
        <el-form-item label="AI服务">
          <el-radio-group v-model="aiForm.service">
            <el-radio label="openai">OpenAI GPT-4</el-radio>
            <el-radio label="custom">自定义API</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="generateContent" :loading="generating">
            生成内容
          </el-button>
          <el-button @click="clearForm">清空</el-button>
        </el-form-item>
      </el-form>
      
      <div v-if="result" class="result-section">
        <h3>生成结果</h3>
        <el-input
          v-model="result"
          type="textarea"
          :rows="8"
          readonly
        />
        <div class="result-actions">
          <el-button @click="copyResult">复制结果</el-button>
          <el-button type="success" @click="useAsTemplate">用作模板</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const templates = ref([])
const generating = ref(false)
const result = ref('')

const aiForm = ref({
  template_id: null,
  prompt: '',
  dataJson: '{"name": "张三", "city": "北京", "company": "ABC公司"}',
  service: 'openai'
})

const loadTemplates = async () => {
  try {
    const response = await api.get('/templates')
    templates.value = response.data
  } catch (error) {
    console.error('加载模板失败:', error)
  }
}

const generateContent = async () => {
  if (!aiForm.value.prompt.trim()) {
    ElMessage.warning('请输入AI提示词')
    return
  }
  
  let data = {}
  try {
    data = JSON.parse(aiForm.value.dataJson)
  } catch (error) {
    ElMessage.error('变量数据格式错误，请输入有效的JSON')
    return
  }
  
  generating.value = true
  try {
    const response = await api.post('/ai/generate', {
      prompt: aiForm.value.prompt,
      data: data,
      service: aiForm.value.service
    })
    result.value = response.data.result
    ElMessage.success('内容生成成功')
  } catch (error) {
    console.error('生成内容失败:', error)
  } finally {
    generating.value = false
  }
}

const clearForm = () => {
  aiForm.value.prompt = ''
  aiForm.value.dataJson = '{"name": "张三", "city": "北京", "company": "ABC公司"}'
  result.value = ''
}

const copyResult = () => {
  navigator.clipboard.writeText(result.value)
  ElMessage.success('结果已复制到剪贴板')
}

const useAsTemplate = () => {
  // 跳转到模板创建页面，并预填充内容
  window.open(`/dashboard/templates?content=${encodeURIComponent(result.value)}`, '_blank')
}

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.ai-generate {
  padding: 20px;
}

.result-section {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color);
}

.result-actions {
  margin-top: 15px;
}
</style>