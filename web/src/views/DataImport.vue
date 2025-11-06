<template>
  <div class="data-import">
    <el-tabs v-model="activeTab" type="card">
      <!-- Excel导入 -->
      <el-tab-pane label="Excel导入" name="excel">
        <el-upload
          class="upload-demo"
          drag
          :action="uploadUrl"
          :headers="uploadHeaders"
          :on-success="handleExcelSuccess"
          :on-error="handleUploadError"
          :before-upload="beforeUpload"
          accept=".xlsx,.xls,.csv"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">
            将Excel文件拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              支持 .xlsx, .xls, .csv 格式，文件大小不超过50MB
            </div>
          </template>
        </el-upload>
      </el-tab-pane>

      <!-- SQL查询 -->
      <el-tab-pane label="SQL查询" name="sql">
        <div class="sql-editor">
          <el-form :model="sqlForm" label-width="100px">
            <el-form-item label="SQL语句">
              <el-input
                v-model="sqlForm.query"
                type="textarea"
                :rows="8"
                placeholder="输入SQL查询语句（仅支持SELECT和INSERT）"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="executeSql" :loading="sqlLoading">
                执行查询
              </el-button>
              <el-button @click="clearSql">清空</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <!-- 手动输入 -->
      <el-tab-pane label="手动输入" name="manual">
        <div class="manual-input">
          <div class="toolbar">
            <el-button @click="addRow">添加行</el-button>
            <el-button @click="addColumn">添加列</el-button>
            <el-button type="danger" @click="clearData">清空数据</el-button>
          </div>
          
          <el-table :data="manualData" border style="width: 100%">
            <el-table-column
              v-for="(column, index) in columns"
              :key="index"
              :label="column"
              :prop="column"
            >
              <template #header>
                <el-input
                  v-model="columns[index]"
                  size="small"
                  placeholder="列名"
                />
              </template>
              <template #default="{ row, $index }">
                <el-input
                  v-model="row[column]"
                  size="small"
                  @input="updateManualData"
                />
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="80">
              <template #default="{ $index }">
                <el-button
                  size="small"
                  type="danger"
                  @click="removeRow($index)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 数据预览 -->
    <el-card v-if="previewData.length > 0" title="数据预览" class="preview-card">
      <div class="preview-toolbar">
        <span>共 {{ previewData.length }} 条记录</span>
        <el-button type="primary" @click="saveData">保存数据</el-button>
      </div>
      
      <el-table :data="previewData.slice(0, 10)" border style="width: 100%">
        <el-table-column
          v-for="key in Object.keys(previewData[0] || {})"
          :key="key"
          :prop="key"
          :label="key"
          show-overflow-tooltip
        />
      </el-table>
      
      <div v-if="previewData.length > 10" class="preview-tip">
        仅显示前10条记录，实际共 {{ previewData.length }} 条
      </div>
    </el-card>

    <!-- 保存数据对话框 -->
    <el-dialog v-model="showSaveDialog" title="保存数据" width="500px">
      <el-form :model="saveForm" label-width="100px">
        <el-form-item label="关联任务">
          <el-select v-model="saveForm.taskId" placeholder="选择任务">
            <el-option
              v-for="task in tasks"
              :key="task.id"
              :label="task.name"
              :value="task.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="数据描述">
          <el-input
            v-model="saveForm.description"
            type="textarea"
            placeholder="可选：描述这批数据的用途"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showSaveDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmSaveData">确认保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const activeTab = ref('excel')
const previewData = ref([])
const tasks = ref([])

// Excel上传
const uploadUrl = '/api/v1/data/upload'
const uploadHeaders = {
  Authorization: `Bearer ${localStorage.getItem('token')}`
}

// SQL查询
const sqlForm = ref({
  query: ''
})
const sqlLoading = ref(false)

// 手动输入
const columns = ref(['email', 'name'])
const manualData = ref([
  { email: '', name: '' }
])

// 保存数据
const showSaveDialog = ref(false)
const saveForm = ref({
  taskId: null,
  description: ''
})

const handleExcelSuccess = (response: any) => {
  if (response.data) {
    previewData.value = response.data
    ElMessage.success('文件上传成功')
  }
}

const handleUploadError = () => {
  ElMessage.error('文件上传失败')
}

const beforeUpload = (file: File) => {
  const isValidType = ['application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 
                      'application/vnd.ms-excel', 'text/csv'].includes(file.type)
  const isLt50M = file.size / 1024 / 1024 < 50

  if (!isValidType) {
    ElMessage.error('只支持 Excel 和 CSV 文件')
    return false
  }
  if (!isLt50M) {
    ElMessage.error('文件大小不能超过 50MB')
    return false
  }
  return true
}

const executeSql = async () => {
  if (!sqlForm.value.query.trim()) {
    ElMessage.warning('请输入SQL语句')
    return
  }

  sqlLoading.value = true
  try {
    const response = await api.post('/data/sql', {
      query: sqlForm.value.query
    })
    previewData.value = response.data
    ElMessage.success('查询执行成功')
  } catch (error) {
    console.error('SQL查询失败:', error)
  } finally {
    sqlLoading.value = false
  }
}

const clearSql = () => {
  sqlForm.value.query = ''
}

const addRow = () => {
  const newRow: any = {}
  columns.value.forEach(col => {
    newRow[col] = ''
  })
  manualData.value.push(newRow)
}

const addColumn = () => {
  const newColumn = `column_${columns.value.length + 1}`
  columns.value.push(newColumn)
  manualData.value.forEach(row => {
    row[newColumn] = ''
  })
}

const removeRow = (index: number) => {
  manualData.value.splice(index, 1)
}

const clearData = () => {
  manualData.value = [{}]
  columns.value.forEach(col => {
    manualData.value[0][col] = ''
  })
}

const updateManualData = () => {
  previewData.value = manualData.value.filter(row => {
    return Object.values(row).some(val => val !== '')
  })
}

const saveData = () => {
  showSaveDialog.value = true
}

const confirmSaveData = async () => {
  if (!saveForm.value.taskId) {
    ElMessage.warning('请选择关联任务')
    return
  }

  try {
    await api.post('/data/save', {
      task_id: saveForm.value.taskId,
      data: previewData.value,
      description: saveForm.value.description
    })
    
    ElMessage.success('数据保存成功')
    showSaveDialog.value = false
    previewData.value = []
  } catch (error) {
    console.error('保存数据失败:', error)
  }
}

const loadTasks = async () => {
  try {
    const response = await api.get('/tasks')
    tasks.value = response.data
  } catch (error) {
    console.error('加载任务失败:', error)
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.data-import {
  padding: 20px;
}

.upload-demo {
  margin: 20px 0;
}

.sql-editor {
  padding: 20px;
}

.manual-input {
  padding: 20px;
}

.toolbar {
  margin-bottom: 20px;
}

.preview-card {
  margin-top: 20px;
}

.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.preview-tip {
  text-align: center;
  color: var(--el-text-color-regular);
  margin-top: 10px;
}
</style>