<template>
  <div class="tasks">
    <div class="page-header">
      <div class="header-content">
        <div class="header-info">
          <h1 class="page-title">
            <el-icon class="title-icon"><Message /></el-icon>
            发送任务管理
          </h1>
          <p class="page-description">管理和监控邮件发送任务的执行状态</p>
        </div>
        
        <div class="header-actions">
          <el-button 
            type="primary" 
            size="large"
            @click="showCreateDialog = true"
            class="create-btn"
          >
            <el-icon><Plus /></el-icon>
            创建新任务
          </el-button>
        </div>
      </div>
    </div>

    <div class="toolbar">
      <div class="filter-section">
        <div class="filter-group">
          <label class="filter-label">状态筛选</label>
          <el-select 
            v-model="statusFilter" 
            placeholder="选择状态" 
            @change="loadTasks"
            class="status-filter"
            clearable
          >
            <el-option label="全部状态" value="" />
            <el-option label="📋 待发送" value="pending" />
            <el-option label="🚀 运行中" value="running" />
            <el-option label="✅ 已完成" value="completed" />
            <el-option label="⏸️ 已暂停" value="paused" />
            <el-option label="❌ 失败" value="failed" />
          </el-select>
        </div>
        
        <div class="filter-group">
          <label class="filter-label">数据源</label>
          <el-select 
            v-model="dataSourceFilter" 
            placeholder="选择数据源" 
            @change="loadTasks"
            class="source-filter"
            clearable
          >
            <el-option label="全部来源" value="" />
            <el-option label="📊 Excel导入" value="excel" />
            <el-option label="🔍 SQL查询" value="sql" />
            <el-option label="✏️ 手动输入" value="manual" />
          </el-select>
        </div>
      </div>
      
      <div class="stats-section">
        <div class="stat-card">
          <div class="stat-number">{{ taskStats.total }}</div>
          <div class="stat-label">总任务数</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">{{ taskStats.running }}</div>
          <div class="stat-label">运行中</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">{{ taskStats.completed }}</div>
          <div class="stat-label">已完成</div>
        </div>
      </div>
    </div>

    <el-table :data="tasks" style="width: 100%">
      <el-table-column prop="name" label="任务名称" />
      <el-table-column prop="template.name" label="邮件模板" />
      <el-table-column prop="data_source" label="数据源">
        <template #default="{ row }">
          <el-tag :type="getDataSourceType(row.data_source)">
            {{ getDataSourceLabel(row.data_source) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="进度">
        <template #default="{ row }">
          <div v-if="row.total_count > 0">
            <el-progress 
              :percentage="Math.round((row.sent_count + row.fail_count) / row.total_count * 100)"
              :status="row.status === 'completed' ? 'success' : ''"
            />
            <div class="progress-text">
              {{ row.sent_count + row.fail_count }} / {{ row.total_count }}
            </div>
          </div>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250">
        <template #default="{ row }">
          <el-button 
            v-if="row.status === 'pending'" 
            size="small" 
            type="success"
            @click="startTask(row.id)"
          >
            开始
          </el-button>
          <el-button 
            v-if="row.status === 'running'" 
            size="small" 
            type="warning"
            @click="pauseTask(row.id)"
          >
            暂停
          </el-button>
          <el-button 
            v-if="row.status === 'paused'" 
            size="small" 
            type="success"
            @click="resumeTask(row.id)"
          >
            恢复
          </el-button>
          <el-button size="small" @click="viewTaskDetail(row)">详情</el-button>
          <el-button size="small" type="danger" @click="deleteTask(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="loadTasks"
      @size-change="loadTasks"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 20px; text-align: right"
    />

    <!-- 创建任务对话框 -->
    <el-dialog 
      v-model="showCreateDialog" 
      title="" 
      width="700px"
      class="create-task-dialog"
      :show-close="false"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title">
            <el-icon class="dialog-icon"><Plus /></el-icon>
            <span>创建发送任务</span>
          </div>
          <el-button 
            @click="showCreateDialog = false" 
            circle 
            class="close-btn"
          >
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </template>
      
      <div class="create-form-container">
        <el-steps :active="currentStep" align-center class="form-steps">
          <el-step title="基本信息" icon="Document" />
          <el-step title="数据配置" icon="Upload" />
          <el-step title="高级设置" icon="Setting" />
        </el-steps>
        
        <el-form :model="taskForm" label-width="120px" class="create-form">
          <!-- 步骤1: 基本信息 -->
          <div v-show="currentStep === 0" class="form-step">
            <div class="step-title">📋 基本信息配置</div>
            
            <el-form-item label="任务名称" required class="form-item">
              <el-input 
                v-model="taskForm.name" 
                placeholder="请输入任务名称"
                class="modern-input"
              >
                <template #prefix>
                  <el-icon><Edit /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item label="邮件模板" required class="form-item">
              <el-select 
                v-model="taskForm.template_id" 
                placeholder="选择邮件模板"
                class="modern-select"
              >
                <el-option
                  v-for="template in templates"
                  :key="template.id"
                  :label="template.name"
                  :value="template.id"
                >
                  <div class="template-option">
                    <span class="template-name">{{ template.name }}</span>
                    <span class="template-subject">{{ template.subject }}</span>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </div>
          
          <!-- 步骤2: 数据配置 -->
          <div v-show="currentStep === 1" class="form-step">
            <div class="step-title">📊 数据源配置</div>
            
            <el-form-item label="数据源" required class="form-item">
              <el-radio-group v-model="taskForm.data_source" class="data-source-group">
                <el-radio-button label="excel">
                  <el-icon><Document /></el-icon>
                  Excel导入
                </el-radio-button>
                <el-radio-button label="sql">
                  <el-icon><Search /></el-icon>
                  SQL查询
                </el-radio-button>
                <el-radio-button label="manual">
                  <el-icon><Edit /></el-icon>
                  手动输入
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
            
            <el-form-item v-if="taskForm.data_source === 'sql'" label="SQL查询" class="form-item">
              <el-input
                v-model="taskForm.data_content"
                type="textarea"
                :rows="6"
                placeholder="请输入SQL查询语句，例如：SELECT email, name FROM users WHERE active = 1"
                class="sql-textarea"
              />
            </el-form-item>
            
            <el-form-item v-if="taskForm.data_source === 'excel'" label="文件上传" class="form-item">
              <el-upload
                class="upload-demo"
                drag
                action="#"
                :auto-upload="false"
              >
                <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                <div class="el-upload__text">
                  将Excel文件拖拽到此处，或<em>点击上传</em>
                </div>
                <template #tip>
                  <div class="el-upload__tip">
                    支持 .xlsx, .csv 格式文件，文件大小不超过10MB
                  </div>
                </template>
              </el-upload>
            </el-form-item>
          </div>
          
          <!-- 步骤3: 高级设置 -->
          <div v-show="currentStep === 2" class="form-step">
            <div class="step-title">⚙️ 高级设置</div>
            
            <el-form-item label="AI提示词" class="form-item">
              <el-input
                v-model="taskForm.ai_prompt"
                type="textarea"
                :rows="4"
                placeholder="可选：输入AI提示词，用于生成个性化内容。例如：根据用户的 {{name}} 和 {{city}} 生成个性化推荐"
                class="ai-textarea"
              >
                <template #prepend>
                  <el-icon><MagicStick /></el-icon>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item label="计划发送" class="form-item">
              <el-date-picker
                v-model="taskForm.scheduled_at"
                type="datetime"
                placeholder="选择发送时间（留空立即发送）"
                format="YYYY-MM-DD HH:mm:ss"
                class="datetime-picker"
              />
            </el-form-item>
            
            <el-form-item label="发送设置" class="form-item">
              <div class="send-settings">
                <el-checkbox v-model="taskForm.enable_retry">启用失败重试</el-checkbox>
                <el-checkbox v-model="taskForm.enable_tracking">启用邮件追踪</el-checkbox>
              </div>
            </el-form-item>
          </div>
        </el-form>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <div class="footer-left">
            <el-button 
              v-if="currentStep > 0" 
              @click="currentStep--"
              class="step-btn"
            >
              <el-icon><ArrowLeft /></el-icon>
              上一步
            </el-button>
          </div>
          
          <div class="footer-right">
            <el-button @click="showCreateDialog = false" class="cancel-btn">
              取消
            </el-button>
            
            <el-button 
              v-if="currentStep < 2" 
              type="primary" 
              @click="currentStep++"
              class="step-btn"
            >
              下一步
              <el-icon><ArrowRight /></el-icon>
            </el-button>
            
            <el-button 
              v-else
              type="primary" 
              @click="createTask"
              class="create-btn"
            >
              <el-icon><Check /></el-icon>
              创建任务
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="任务详情" width="800px">
      <div v-if="selectedTask">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务名称">{{ selectedTask.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedTask.status)">
              {{ getStatusLabel(selectedTask.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="邮件模板">{{ selectedTask.template?.name }}</el-descriptions-item>
          <el-descriptions-item label="数据源">{{ getDataSourceLabel(selectedTask.data_source) }}</el-descriptions-item>
          <el-descriptions-item label="总数量">{{ selectedTask.total_count }}</el-descriptions-item>
          <el-descriptions-item label="已发送">{{ selectedTask.sent_count }}</el-descriptions-item>
          <el-descriptions-item label="失败数">{{ selectedTask.fail_count }}</el-descriptions-item>
          <el-descriptions-item label="成功率">
            {{ selectedTask.total_count > 0 ? Math.round(selectedTask.sent_count / selectedTask.total_count * 100) : 0 }}%
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedTask.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDate(selectedTask.started_at) }}</el-descriptions-item>
        </el-descriptions>
        
        <div v-if="selectedTask.ai_prompt" style="margin-top: 20px">
          <h4>AI提示词</h4>
          <el-input
            :model-value="selectedTask.ai_prompt"
            type="textarea"
            :rows="3"
            readonly
          />
        </div>
        
        <!-- 发送日志 -->
        <div style="margin-top: 20px">
          <h4>发送日志</h4>
          <el-table :data="taskLogs" max-height="300">
            <el-table-column prop="recipient" label="收件人" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="row.status === 'sent' ? 'success' : 'danger'">
                  {{ row.status === 'sent' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="retry_count" label="重试次数" />
            <el-table-column prop="sent_at" label="发送时间">
              <template #default="{ row }">
                {{ formatDate(row.sent_at) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'

const tasks = ref([])
const templates = ref([])
const statusFilter = ref('')
const dataSourceFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const currentStep = ref(0)

const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const selectedTask = ref(null)
const taskLogs = ref([])

const taskForm = ref({
  name: '',
  template_id: null,
  data_source: '',
  data_content: '',
  ai_prompt: '',
  scheduled_at: null,
  enable_retry: true,
  enable_tracking: false,
  project_id: 1
})

const taskStats = computed(() => {
  const stats = {
    total: tasks.value.length,
    running: 0,
    completed: 0,
    pending: 0,
    failed: 0
  }
  
  tasks.value.forEach(task => {
    if (task.status === 'running') stats.running++
    else if (task.status === 'completed') stats.completed++
    else if (task.status === 'pending') stats.pending++
    else if (task.status === 'failed') stats.failed++
  })
  
  return stats
})

const loadTasks = async () => {
  try {
    const response = await api.get('/tasks', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        status: statusFilter.value
      }
    })
    tasks.value = response.data
    total.value = response.total
  } catch (error) {
    console.error('加载任务失败:', error)
  }
}

const loadTemplates = async () => {
  try {
    const response = await api.get('/templates')
    templates.value = response.data
  } catch (error) {
    console.error('加载模板失败:', error)
  }
}

const createTask = async () => {
  try {
    await api.post('/tasks', taskForm.value)
    ElMessage.success('任务创建成功')
    showCreateDialog.value = false
    resetTaskForm()
    loadTasks()
  } catch (error) {
    console.error('创建任务失败:', error)
  }
}

const startTask = async (taskId: number) => {
  try {
    await api.post(`/tasks/${taskId}/start`)
    ElMessage.success('任务已开始')
    loadTasks()
  } catch (error) {
    console.error('启动任务失败:', error)
  }
}

const pauseTask = async (taskId: number) => {
  try {
    await api.post(`/tasks/${taskId}/pause`)
    ElMessage.success('任务已暂停')
    loadTasks()
  } catch (error) {
    console.error('暂停任务失败:', error)
  }
}

const resumeTask = async (taskId: number) => {
  try {
    await api.post(`/tasks/${taskId}/resume`)
    ElMessage.success('任务已恢复')
    loadTasks()
  } catch (error) {
    console.error('恢复任务失败:', error)
  }
}

const deleteTask = async (taskId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个任务吗？', '确认删除')
    await api.delete(`/tasks/${taskId}`)
    ElMessage.success('删除成功')
    loadTasks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除任务失败:', error)
    }
  }
}

const viewTaskDetail = async (task: any) => {
  selectedTask.value = task
  showDetailDialog.value = true
  
  // 加载任务日志
  try {
    const response = await api.get(`/tasks/${task.id}/logs`)
    taskLogs.value = response.data
  } catch (error) {
    console.error('加载任务日志失败:', error)
  }
}

const resetTaskForm = () => {
  taskForm.value = {
    name: '',
    template_id: null,
    data_source: '',
    data_content: '',
    ai_prompt: '',
    scheduled_at: null,
    enable_retry: true,
    enable_tracking: false,
    project_id: 1
  }
  currentStep.value = 0
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    paused: 'warning',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: '待发送',
    running: '运行中',
    completed: '已完成',
    paused: '已暂停',
    failed: '失败'
  }
  return labels[status] || status
}

const getDataSourceType = (source: string) => {
  const types: Record<string, string> = {
    excel: 'success',
    sql: 'warning',
    manual: 'info'
  }
  return types[source] || 'info'
}

const getDataSourceLabel = (source: string) => {
  const labels: Record<string, string> = {
    excel: 'Excel导入',
    sql: 'SQL查询',
    manual: '手动输入'
  }
  return labels[source] || source
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  loadTasks()
  loadTemplates()
})
</script>

<style scoped>
.tasks {
  padding: 24px;
  background: transparent;
}

.page-header {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 32px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  display: flex;
  align-items: center;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0 0 8px 0;
}

.title-icon {
  margin-right: 12px;
  font-size: 32px;
}

.page-description {
  color: #8b949e;
  font-size: 16px;
  margin: 0;
}

.create-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 12px;
  padding: 12px 24px;
  font-weight: 600;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
  transition: all 0.3s ease;
}

.create-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.filter-section {
  display: flex;
  gap: 24px;
  align-items: center;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: #8b949e;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-filter,
.source-filter {
  width: 180px;
}

.stats-section {
  display: flex;
  gap: 16px;
}

.stat-card {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-radius: 12px;
  padding: 16px 20px;
  text-align: center;
  min-width: 80px;
  border: 1px solid rgba(102, 126, 234, 0.2);
}

.stat-number {
  font-size: 24px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #8b949e;
  font-weight: 500;
}

.progress-text {
  font-size: 12px;
  color: var(--el-text-color-regular);
  text-align: center;
  margin-top: 5px;
}

/* 创建任务对话框样式 */
:deep(.create-task-dialog) {
  border-radius: 16px;
  overflow: hidden;
}

:deep(.create-task-dialog .el-dialog__header) {
  padding: 0;
  margin: 0;
}

:deep(.create-task-dialog .el-dialog__body) {
  padding: 0;
}

:deep(.create-task-dialog .el-dialog__footer) {
  padding: 0;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.dialog-title {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: 600;
}

.dialog-icon {
  margin-right: 12px;
  font-size: 24px;
}

.close-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.create-form-container {
  padding: 32px;
}

.form-steps {
  margin-bottom: 32px;
}

:deep(.form-steps .el-step__title) {
  font-weight: 600;
}

.form-step {
  min-height: 300px;
}

.step-title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 2px solid rgba(102, 126, 234, 0.2);
}

.form-item {
  margin-bottom: 24px;
}

.modern-input,
.modern-select {
  border-radius: 8px;
}

.template-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.template-name {
  font-weight: 600;
  color: #2c3e50;
}

.template-subject {
  font-size: 12px;
  color: #8b949e;
}

.data-source-group {
  display: flex;
  gap: 12px;
}

:deep(.data-source-group .el-radio-button__inner) {
  border-radius: 8px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.sql-textarea,
.ai-textarea {
  border-radius: 8px;
}

.send-settings {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
}

.footer-left,
.footer-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.step-btn {
  border-radius: 8px;
  padding: 10px 20px;
  font-weight: 500;
}

.cancel-btn {
  border-radius: 8px;
  padding: 10px 20px;
}

:deep(.el-table) {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 12px;
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
</style>