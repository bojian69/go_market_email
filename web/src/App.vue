<template>
  <el-config-provider :locale="locale">
    <div id="app" :class="{ 'dark': isDark }">
      <router-view />
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

const locale = zhCn
const isDark = ref(false)

// 监听主题切换
const toggleTheme = () => {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
}

// 暴露给全局使用
window.toggleTheme = toggleTheme
</script>

<style>
#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100vh;
}

html.dark {
  color-scheme: dark;
}
</style>