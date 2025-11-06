<template>
  <div class="settings">
    <el-tabs v-model="activeTab" type="card">
      <!-- 认证设置 -->
      <el-tab-pane label="认证设置" name="auth">
        <el-card title="认证令牌配置">
          <el-form label-width="120px">
            <el-form-item label="当前令牌">
              <el-input
                v-model="currentToken"
                type="password"
                show-password
                readonly
              />
            </el-form-item>
            
            <el-form-item label="新令牌">
              <el-input
                v-model="newToken"
                placeholder="输入新的认证令牌"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="updateToken">更新令牌</el-button>
              <el-button @click="generateToken">生成随机令牌</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 邮件测试 -->
      <el-tab-pane label="邮件测试" name="email">
        <el-card title="发送测试邮件">
          <el-form :model="testForm" label-width="120px">
            <el-form-item label="收件人邮箱">
              <el-input v-model="testForm.email" placeholder="输入测试邮箱地址" />
            </el-form-item>
            
            <el-form-item label="邮件主题">
              <el-input v-model="testForm.subject" placeholder="输入邮件主题" />
            </el-form-item>
            
            <el-form-item label="邮件内容">
              <el-input
                v-model="testForm.content"
                type="textarea"
                :rows="8"
                placeholder="输入邮件内容"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="sendTestEmail" :loading="sending">
                发送测试邮件
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 系统信息 -->
      <el-tab-pane label="系统信息" name="system">
        <el-card title="系统状态">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="系统版本">v1.0.0</el-descriptions-item>
            <el-descriptions-item label="Go版本">1.21</el-descriptions-item>
            <el-descriptions-item label="Vue版本">3.3.8</el-descriptions-item>
            <el-descriptions-item label="数据库">MySQL 8.0</el-descriptions-item>
            <el-descriptions-item label="缓存">Redis</el-descriptions-item>
            <el-descriptions-item label="运行状态">
              <el-tag type="success">正常运行</el-tag>
            </el-descriptions-item>
          </el-descriptions>
          
          <div style="margin-top: 20px;">
            <el-button @click="checkHealth">检查系统健康状态</el-button>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 配置管理 -->
      <el-tab-pane label="配置管理" name="config">
        <el-card title="环境配置">
          <el-alert
            title="配置说明"
            description="这些配置需要在服务器端的配置文件或环境变量中设置，前端仅用于查看"
            type="info"
            show-icon
            :closable="false"
          />
          
          <div class="config-section">
            <h4>数据库配置</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="主机">localhost:3306</el-descriptions-item>
              <el-descriptions-item label="数据库">go_market_email</el-descriptions-item>
            </el-descriptions>
          </div>
          
          <div class="config-section">
            <h4>SMTP配置</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="服务器">smtp.partner.outlook.cn:587</el-descriptions-item>
              <el-descriptions-item label="发送频率">25封/分钟</el-descriptions-item>
              <el-descriptions-item label="重试次数">2次</el-descriptions-item>
            </el-descriptions>
          </div>
          
          <div class="config-section">
            <h4>AI配置</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="OpenAI模型">GPT-4</el-descriptions-item>
              <el-descriptions-item label="自定义API">已配置</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const activeTab = ref('auth')
const currentToken = ref('')
const newToken = ref('')
const sending = ref(false)

const testForm = ref({
  email: '',
  subject: '测试邮件',
  content: '这是一封测试邮件，用于验证邮件发送功能是否正常。'
})

const updateToken = () => {
  if (!newToken.value.trim()) {
    ElMessage.warning('请输入新令牌')
    return
  }
  
  localStorage.setItem('token', newToken.value)
  currentToken.value = newToken.value
  newToken.value = ''
  ElMessage.success('令牌已更新')
}

const generateToken = () => {
  const token = 'gme-' + Math.random().toString(36).substr(2, 9) + '-' + Date.now().toString(36)
  newToken.value = token
}

const sendTestEmail = async () => {
  if (!testForm.value.email || !testForm.value.subject || !testForm.value.content) {
    ElMessage.warning('请填写完整的邮件信息')
    return
  }
  
  sending.value = true
  try {
    // 这里需要创建一个简单的测试模板
    const response = await api.post('/emails/test', {
      template_id: 1, // 使用默认模板ID
      email: testForm.value.email,
      data: {
        subject: testForm.value.subject,
        content: testForm.value.content
      }
    })
    ElMessage.success('测试邮件发送成功')
  } catch (error) {
    console.error('发送测试邮件失败:', error)
  } finally {
    sending.value = false
  }
}

const checkHealth = async () => {
  try {
    const response = await fetch('/health')
    if (response.ok) {
      ElMessage.success('系统运行正常')
    } else {
      ElMessage.error('系统状态异常')
    }
  } catch (error) {
    ElMessage.error('无法连接到服务器')
  }
}

onMounted(() => {
  currentToken.value = localStorage.getItem('token') || ''
})
</script>

<style scoped>
.settings {
  padding: 20px;
}

.config-section {
  margin-top: 20px;
}

.config-section h4 {
  margin-bottom: 10px;
  color: var(--el-text-color-primary);
}
</style>