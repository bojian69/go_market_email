<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon template">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.templateCount }}</div>
              <div class="stat-label">邮件模板</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon pending">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.pendingCount }}</div>
              <div class="stat-label">待发送邮件</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon success">
              <el-icon><Check /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.sentCount }}</div>
              <div class="stat-label">发送成功</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon failed">
              <el-icon><Close /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.failedCount }}</div>
              <div class="stat-label">发送失败</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card title="发送趋势">
          <v-chart :option="trendOption" style="height: 300px" />
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card title="发送状态分布">
          <v-chart :option="pieOption" style="height: 300px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 实时任务 -->
    <el-card title="实时任务状态" class="tasks-card">
      <el-table :data="runningTasks" style="width: 100%">
        <el-table-column prop="name" label="任务名称" />
        <el-table-column prop="progress" label="进度">
          <template #default="{ row }">
            <el-progress :percentage="row.progress" />
          </template>
        </el-table-column>
        <el-table-column prop="sent_count" label="已发送" />
        <el-table-column prop="total_count" label="总数" />
        <el-table-column prop="estimated_remaining" label="预计剩余时间" />
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === 'running'" 
              @click="pauseTask(row.id)"
              size="small"
            >
              暂停
            </el-button>
            <el-button 
              v-else 
              @click="resumeTask(row.id)"
              size="small" 
              type="success"
            >
              恢复
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import api from '@/utils/api'

use([CanvasRenderer, LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

const stats = ref({
  templateCount: 0,
  pendingCount: 0,
  sentCount: 0,
  failedCount: 0
})

const runningTasks = ref([])
let wsConnection: WebSocket | null = null

const trendOption = ref({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: [] },
  yAxis: { type: 'value' },
  series: [{
    name: '发送数量',
    type: 'line',
    data: [],
    smooth: true
  }]
})

const pieOption = ref({
  tooltip: { trigger: 'item' },
  legend: { orient: 'vertical', left: 'left' },
  series: [{
    type: 'pie',
    radius: '50%',
    data: [
      { value: 0, name: '成功' },
      { value: 0, name: '失败' },
      { value: 0, name: '待发送' }
    ]
  }]
})

const loadStats = async () => {
  try {
    const response = await api.get('/stats')
    stats.value = response.data
    
    // 更新饼图数据
    pieOption.value.series[0].data = [
      { value: stats.value.sentCount, name: '成功' },
      { value: stats.value.failedCount, name: '失败' },
      { value: stats.value.pendingCount, name: '待发送' }
    ]
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

const loadRunningTasks = async () => {
  try {
    const response = await api.get('/tasks/running')
    runningTasks.value = response.data || []
  } catch (error) {
    console.error('加载运行任务失败:', error)
  }
}

const pauseTask = async (taskId: number) => {
  try {
    await api.post(`/tasks/${taskId}/pause`)
    await loadRunningTasks()
  } catch (error) {
    console.error('暂停任务失败:', error)
  }
}

const resumeTask = async (taskId: number) => {
  try {
    await api.post(`/tasks/${taskId}/resume`)
    await loadRunningTasks()
  } catch (error) {
    console.error('恢复任务失败:', error)
  }
}

const connectWebSocket = () => {
  const wsUrl = `ws://${location.host}/ws/stats`
  wsConnection = new WebSocket(wsUrl)
  
  wsConnection.onmessage = (event) => {
    const data = JSON.parse(event.data)
    stats.value = data.stats
    runningTasks.value = data.tasks || []
  }
  
  wsConnection.onclose = () => {
    setTimeout(connectWebSocket, 5000) // 重连
  }
}

onMounted(() => {
  loadStats()
  loadRunningTasks()
  connectWebSocket()
})

onUnmounted(() => {
  if (wsConnection) {
    wsConnection.close()
  }
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  font-size: 24px;
  color: white;
}

.stat-icon.template { background: #409eff; }
.stat-icon.pending { background: #e6a23c; }
.stat-icon.success { background: #67c23a; }
.stat-icon.failed { background: #f56c6c; }

.stat-number {
  font-size: 28px;
  font-weight: bold;
  color: var(--el-text-color-primary);
}

.stat-label {
  color: var(--el-text-color-regular);
  margin-top: 5px;
}

.charts-row {
  margin-bottom: 20px;
}

.tasks-card {
  margin-top: 20px;
}
</style>