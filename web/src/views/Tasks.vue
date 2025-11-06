<template>
  <div class="tasks">
    <div class="toolbar">
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建任务
      </el-button>
      
      <el-select v-model="statusFilter" placeholder="状态筛选" @change="loadTasks">
        <el-option label="全部" value="" />
        <el-option label="待发送" value="pending" />
        <el-option label="运行中" value="running" />
        <el-option label="已完成" value="completed" />
        <el-option label="已暂停" value="paused" />
        <el-option label="失败" value="failed" />
      </el-select>
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
    <el-dialog v-model="showCreateDialog" title="创建发送任务" width="600px">
      <el-form :model="taskForm" label-width="120px">
        <el-form-item label="任务名称" required>
          <el-input v-model="taskForm.name" />
        </el-form-item>
        
        <el-form-item label="邮件模板" required>
          <el-select v-model="taskForm.template_id" placeholder="选择模板">
            <el-option
              v-for="template in templates"
              :key="template.id"
              :label="template.name"
              :value="template.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="数据源" required>
          <el-select v-model="taskForm.data_source" placeholder="选择数据源">
            <el-option label="Excel导入" value="excel" />
            <el-option label="SQL查询" value="sql" />
            <el-option label="手动输入" value="manual" />
          </el-select>
        </el-form-item>
        
        <el-form-item v-if="taskForm.data_source === 'sql'" label="SQL查询">
          <el-input
            v-model="taskForm.data_content"
            type="textarea"
            :rows="4"
            placeholder="输入SQL查询语句"
          />
        </el-form-item>
        
        <el-form-item label="AI提示词">
          <el-input
            v-model="taskForm.ai_prompt"
            type="textarea"
            :rows="3"
            placeholder="可选：输入AI提示词，用于生成个性化内容"
          />
        </el-form-item>
        
        <el-form-item label="计划发送时间">
          <el-date-picker
            v-model="taskForm.scheduled_at"
            type="datetime"
            placeholder="选择发送时间"
            format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createTask">创建</el-button>
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'

const tasks = ref([])
const templates = ref([])
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

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
  project_id: 1
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
    project_id: 1
  }
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
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.progress-text {
  font-size: 12px;
  color: var(--el-text-color-regular);
  text-align: center;
  margin-top: 5px;
}
</style>