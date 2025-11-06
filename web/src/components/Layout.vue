<template>
  <el-container class="layout-container">
    <el-aside width="250px" class="sidebar">
      <div class="logo">
        <el-icon><Message /></el-icon>
        <span>AI邮件营销</span>
      </div>
      
      <el-menu
        :default-active="currentPath"
        router
        class="sidebar-menu"
        :collapse="false"
      >
        <el-menu-item
          v-for="route in menuRoutes"
          :key="route.path"
          :index="'/dashboard/' + route.path"
        >
          <el-icon><component :is="route.meta.icon" /></el-icon>
          <span>{{ route.meta.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <h2>{{ currentTitle }}</h2>
          <el-breadcrumb separator="/" style="margin-top: 8px;">
            <el-breadcrumb-item>AI邮件营销系统</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-button @click="toggleTheme" circle>
            <el-icon><Moon /></el-icon>
          </el-button>
          <el-dropdown>
            <el-avatar :size="32">U</el-avatar>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人设置</el-dropdown-item>
                <el-dropdown-item divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const menuRoutes = computed(() => {
  const dashboardRoute = router.getRoutes().find(route => route.path === '/dashboard')
  return dashboardRoute?.children || []
})

const currentTitle = computed(() => {
  return route.meta?.title || '监控面板'
})

const currentPath = computed(() => {
  return route.path
})

const toggleTheme = () => {
  window.toggleTheme?.()
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background: var(--el-bg-color-page);
  border-right: 1px solid var(--el-border-color);
}

.logo {
  display: flex;
  align-items: center;
  padding: 20px;
  font-size: 18px;
  font-weight: bold;
  color: var(--el-color-primary);
}

.logo .el-icon {
  margin-right: 10px;
  font-size: 24px;
}

.sidebar-menu {
  border: none;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.main-content {
  background: var(--el-bg-color-page);
  padding: 20px;
}
</style>