import { createRouter, createWebHistory } from 'vue-router'

import userRoutes from './modules/user'
import rbacRoutes from './modules/rbac'
import k8sRoutes from './modules/k8s'
import auditRoutes from './modules/audit'
import aiopsRoutes from './modules/aiops'

const baseRoutes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true, title: '首页' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/NotFound.vue'),
    meta: { title: '页面不存在' }
  }
]

const routes = [
  ...baseRoutes,
  ...userRoutes,
  ...rbacRoutes,
  ...k8sRoutes,
  ...auditRoutes,
  ...aiopsRoutes
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const isLoggedIn = localStorage.getItem('isLoggedIn')
  
  document.title = to.meta.title || 'DevOps管理平台'
  
  if (to.meta.requiresAuth) {
    if (!token || !isLoggedIn) {
      next('/login')
    } else {
      next()
    }
  } else {
    if (to.path === '/login' && token && isLoggedIn) {
      next('/dashboard')
    } else {
      next()
    }
  }
})

export default router