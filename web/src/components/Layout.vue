<template>
  <el-container class="layout-container">
    <el-aside :width="sidebarCollapsed ? '64px' : '280px'" class="sidebar">
      <div class="logo" @click="toggleSidebar">
        <el-icon class="logo-icon"><Message /></el-icon>
        <transition name="fade">
          <span v-show="!sidebarCollapsed" class="logo-text">AI邮件营销</span>
        </transition>
      </div>
      
      <div class="nav-section">
        <div v-if="!sidebarCollapsed" class="section-title">主要功能</div>
        <el-menu
          :default-active="currentPath"
          router
          class="sidebar-menu"
          :collapse="sidebarCollapsed"
        >
          <el-menu-item
            v-for="route in menuRoutes"
            :key="route.path"
            :index="'/dashboard/' + route.path"
            class="nav-item"
          >
            <el-icon class="nav-icon"><component :is="route.meta.icon" /></el-icon>
            <span class="nav-text">{{ route.meta.title }}</span>
          </el-menu-item>
        </el-menu>
      </div>

      <div class="sidebar-footer">
        <el-button 
          @click="toggleSidebar" 
          :icon="sidebarCollapsed ? 'Expand' : 'Fold'"
          circle
          class="collapse-btn"
        />
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <div class="page-info">
            <h2 class="page-title">{{ currentTitle }}</h2>
            <el-breadcrumb separator=">" class="breadcrumb">
              <el-breadcrumb-item>
                <el-icon><House /></el-icon>
                AI邮件营销系统
              </el-breadcrumb-item>
              <el-breadcrumb-item class="current-page">{{ currentTitle }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
        </div>
        
        <div class="header-right">
          <div class="header-actions">
            <el-tooltip content="切换主题" placement="bottom">
              <el-button @click="toggleTheme" circle class="action-btn">
                <el-icon><Moon /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip content="通知" placement="bottom">
              <el-badge :value="3" class="notification-badge">
                <el-button circle class="action-btn">
                  <el-icon><Bell /></el-icon>
                </el-button>
              </el-badge>
            </el-tooltip>
            
            <el-dropdown class="user-dropdown">
              <div class="user-info">
                <el-avatar :size="36" class="user-avatar">
                  <el-icon><User /></el-icon>
                </el-avatar>
                <div v-if="!sidebarCollapsed" class="user-details">
                  <div class="username">管理员</div>
                  <div class="user-role">系统管理员</div>
                </div>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item>
                    <el-icon><User /></el-icon>
                    个人设置
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-icon><Setting /></el-icon>
                    系统设置
                  </el-dropdown-item>
                  <el-dropdown-item divided>
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const sidebarCollapsed = ref(false)

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

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.sidebar {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 4px 0 20px rgba(0, 0, 0, 0.1);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  display: flex;
  flex-direction: column;
}

.logo {
  display: flex;
  align-items: center;
  padding: 20px;
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin: 16px 16px 24px 16px;
  border-radius: 16px;
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3);
  cursor: pointer;
  transition: all 0.3s ease;
  min-height: 60px;
  justify-content: center;
}

.logo:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 35px rgba(102, 126, 234, 0.4);
}

.logo-icon {
  font-size: 28px;
  margin-right: 12px;
}

.logo-text {
  white-space: nowrap;
}

.nav-section {
  flex: 1;
  padding: 0 16px;
}

.section-title {
  font-size: 12px;
  font-weight: 600;
  color: #8b949e;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0 8px 12px 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(139, 148, 158, 0.2);
}

.sidebar-menu {
  border: none;
  background: transparent;
}

.nav-item {
  margin: 6px 0;
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.nav-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  width: 4px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  transform: scaleY(0);
  transition: transform 0.3s ease;
}

.nav-item:hover::before {
  transform: scaleY(1);
}

.nav-icon {
  font-size: 20px;
  margin-right: 12px;
}

.nav-text {
  font-weight: 500;
}

:deep(.el-menu-item) {
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

:deep(.el-menu-item:hover) {
  background: rgba(102, 126, 234, 0.08) !important;
  color: #667eea !important;
  transform: translateX(4px);
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.15) 0%, rgba(118, 75, 162, 0.15) 100%) !important;
  color: #667eea !important;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
}

.sidebar-footer {
  padding: 16px;
  display: flex;
  justify-content: center;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
}

.collapse-btn {
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.2);
  color: #667eea;
  transition: all 0.3s ease;
}

.collapse-btn:hover {
  background: rgba(102, 126, 234, 0.2);
  transform: scale(1.1);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 32px;
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  height: 80px;
}

.page-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: 700;
  font-size: 24px;
  margin: 0;
}

.breadcrumb {
  font-size: 14px;
}

:deep(.breadcrumb .el-breadcrumb__item) {
  color: #8b949e;
}

:deep(.breadcrumb .current-page) {
  color: #667eea;
  font-weight: 500;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-btn {
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.2);
  color: #667eea;
  transition: all 0.3s ease;
}

.action-btn:hover {
  background: rgba(102, 126, 234, 0.2);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.notification-badge {
  cursor: pointer;
}

.user-dropdown {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.user-info:hover {
  background: rgba(102, 126, 234, 0.1);
}

.user-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.username {
  font-weight: 600;
  font-size: 14px;
  color: #2c3e50;
}

.user-role {
  font-size: 12px;
  color: #8b949e;
}

.main-content {
  background: transparent;
  padding: 0;
  overflow: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>