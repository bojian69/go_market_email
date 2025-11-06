<template>
  <div class="error-boundary">
    <el-result
      icon="warning"
      title="页面加载失败"
      sub-title="请检查网络连接或刷新页面重试"
    >
      <template #extra>
        <el-button type="primary" @click="reload">刷新页面</el-button>
        <el-button @click="goHome">返回首页</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const reload = () => {
  window.location.reload()
}

const goHome = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.error-boundary {
  padding: 50px;
  text-align: center;
}
</style>