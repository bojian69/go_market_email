import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/components/Layout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: '监控面板', icon: 'DataAnalysis' }
        },
        {
          path: 'templates',
          name: 'Templates',
          component: () => import('@/views/Templates.vue'),
          meta: { title: '模板管理', icon: 'Document' }
        },
        {
          path: 'data-import',
          name: 'DataImport',
          component: () => import('@/views/DataImport.vue'),
          meta: { title: '数据导入', icon: 'Upload' }
        },
        {
          path: 'tasks',
          name: 'Tasks',
          component: () => import('@/views/Tasks.vue'),
          meta: { title: '发送任务', icon: 'Message' }
        },
        {
          path: 'preview',
          name: 'Preview',
          component: () => import('@/views/Preview.vue'),
          meta: { title: '邮件预览', icon: 'View' }
        }
      ]
    }
  ]
})

export default router