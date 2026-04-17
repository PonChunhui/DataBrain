<template>
  <div id="app">
    <template v-if="isLoggedIn">
      <el-container>
        <el-aside width="240px" class="sidebar">
          <div class="sidebar-header">
            <div class="logo-wrapper">
              <div class="logo-icon-wrapper">
                <el-icon class="logo-icon"><Monitor /></el-icon>
              </div>
              <div class="logo-text-wrapper">
                <span class="logo-text">数智运维</span>
                <span class="logo-sub">DevOps Platform</span>
              </div>
            </div>
          </div>
          
          <el-scrollbar class="sidebar-scrollbar">
            <el-menu
              :default-active="activeMenu"
              router
              class="menu"
            >
              <template v-for="menu in menus" :key="menu.id">
                <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="menu.path">
                  <template #title>
                    <el-icon><component :is="getIconComponent(menu.icon)" /></el-icon>
                    <span>{{ menu.name }}</span>
                  </template>
                  <el-menu-item v-for="child in menu.children" :key="child.id" :index="child.path">
                    <el-icon><component :is="getIconComponent(child.icon)" /></el-icon>
                    <span>{{ child.name }}</span>
                  </el-menu-item>
                </el-sub-menu>
                <el-menu-item v-else :index="menu.path">
                  <el-icon><component :is="getIconComponent(menu.icon)" /></el-icon>
                  <span>{{ menu.name }}</span>
                </el-menu-item>
              </template>
            </el-menu>
          </el-scrollbar>
        </el-aside>
          
        <el-main class="main">
          <router-view />
        </el-main>
      </el-container>
    </template>
    
    <template v-else>
      <router-view />
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as Icons from '@element-plus/icons-vue'
import { Monitor } from '@element-plus/icons-vue'

const route = useRoute()

const isLoggedInRef = ref(false)
const menus = ref([])

const getIconComponent = (iconName) => {
  return Icons[iconName] || Icons.Menu
}

const isLoggedIn = computed(() => isLoggedInRef.value)

const activeMenu = computed(() => {
  return route.path
})

const checkLoginStatus = async () => {
  isLoggedInRef.value = localStorage.getItem('isLoggedIn') === 'true'
  
  if (isLoggedInRef.value) {
    await fetchMenus()
  }
}

const fetchMenus = async () => {
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    if (!user.id) return
    
    const menusStr = localStorage.getItem('menus')
    if (menusStr) {
      menus.value = JSON.parse(menusStr)
    }
  } catch (error) {
    console.error('获取菜单失败:', error)
  }
}

onMounted(() => {
  checkLoginStatus()
})

watch(() => route.path, () => {
  checkLoginStatus()
})
</script>

<style>
:root {
  --primary-color: #3B82F6;
  --primary-light: #60A5FA;
  --primary-lighter: #EFF6FF;
  --success-color: #10B981;
  --warning-color: #F59E0B;
  --danger-color: #EF4444;
  --info-color: #6366F1;
  
  --bg-color: #F8FAFC;
  --sidebar-bg: #1E293B;
  --sidebar-header-bg: #0F172A;
  --card-bg: #FFFFFF;
  
  --text-primary: #1E293B;
  --text-secondary: #475569;
  --text-muted: #94A3B8;
  --text-sidebar: #E2E8F0;
  --text-sidebar-muted: #94A3B8;
  
  --border-color: #E2E8F0;
  --border-light: #F1F5F9;
  --sidebar-border: #334155;
  
  --space-xs: 8px;
  --space-sm: 12px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;
  
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;
  
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-card: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
  --shadow-hover: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
}

* {
  box-sizing: border-box;
}

#app {
  font-family: 'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-primary);
  height: 100vh;
}

body {
  margin: 0;
  padding: 0;
  background: var(--bg-color);
}

.el-container {
  height: 100vh;
}

.sidebar {
  background: var(--sidebar-bg);
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 var(--space-lg);
  background: var(--sidebar-header-bg);
  border-bottom: 1px solid var(--sidebar-border);
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.logo-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: rgba(59, 130, 246, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-icon {
  font-size: 22px;
  color: var(--primary-light);
}

.logo-text-wrapper {
  display: flex;
  flex-direction: column;
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-sidebar);
  letter-spacing: 0.5px;
}

.logo-sub {
  font-size: 10px;
  color: var(--text-sidebar-muted);
  margin-top: 2px;
}

.sidebar-scrollbar {
  flex: 1;
  overflow: hidden;
  padding: var(--space-sm) 0;
}

.sidebar-scrollbar .el-scrollbar__bar.is-vertical {
  width: 4px;
  background: var(--sidebar-border);
}

.sidebar-scrollbar .el-scrollbar__thumb {
  background: var(--text-sidebar-muted);
  border-radius: 2px;
}

.menu {
  border-right: none;
  background: transparent;
}

.menu .el-menu-item {
  color: var(--text-sidebar);
  height: 44px;
  line-height: 44px;
  margin: 2px 8px;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
  font-weight: 500;
  position: relative;
}

.menu .el-menu-item:hover {
  background: rgba(59, 130, 246, 0.1);
  color: var(--primary-light);
}

.menu .el-menu-item.is-active {
  background: rgba(59, 130, 246, 0.15);
  color: var(--primary-light);
  font-weight: 600;
}

.menu .el-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 24px;
  background: var(--primary-color);
  border-radius: 0 2px 2px 0;
}

.menu .el-menu-item .el-icon {
  font-size: 18px;
  margin-right: var(--space-sm);
  color: var(--text-sidebar-muted);
}

.menu .el-menu-item:hover .el-icon {
  color: var(--primary-light);
}

.menu .el-menu-item.is-active .el-icon {
  color: var(--primary-light);
}

.menu .el-sub-menu {
  margin: 2px 8px;
}

.menu .el-sub-menu__title {
  height: 44px;
  line-height: 44px;
  color: var(--text-sidebar);
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
  font-weight: 500;
}

.menu .el-sub-menu__title:hover {
  background: rgba(59, 130, 246, 0.1);
  color: var(--primary-light);
}

.menu .el-sub-menu__title .el-icon {
  font-size: 18px;
  margin-right: var(--space-sm);
  color: var(--text-sidebar-muted);
}

.menu .el-sub-menu__title:hover .el-icon {
  color: var(--primary-light);
}

.menu .el-sub-menu .el-menu-item {
  padding-left: 48px !important;
  height: 40px;
  line-height: 40px;
  background: transparent !important;
  margin: 1px 8px;
  font-size: 14px;
}

.menu .el-sub-menu .el-menu {
  background: transparent !important;
}

.menu .el-sub-menu .el-menu-item:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  color: var(--primary-light) !important;
}

.menu .el-sub-menu .el-menu-item.is-active {
  background: rgba(59, 130, 246, 0.15) !important;
  color: var(--primary-light) !important;
  font-weight: 600;
}

.menu .el-sub-menu .el-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--primary-color);
  border-radius: 0 2px 2px 0;
}

.el-sub-menu.is-opened > .el-sub-menu__title {
  background: rgba(59, 130, 246, 0.1);
  color: var(--primary-light);
}

.el-sub-menu.is-opened > .el-sub-menu__title .el-icon {
  color: var(--primary-light);
}

.main {
  background: var(--bg-color);
  padding: var(--space-lg);
  height: 100vh;
  overflow-y: auto;
}

.page-container {
  min-height: calc(100vh - 48px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-lg);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid var(--border-color);
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.page-actions {
  display: flex;
  gap: var(--space-sm);
}

.header-right-actions {
  display: flex;
  gap: var(--space-sm);
  align-items: center;
}

.content-card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-color);
}

.section-gap {
  margin-top: var(--space-lg);
}

.search-card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: var(--space-md) var(--space-lg);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-color);
}

.el-button--primary {
  background: var(--primary-color);
  border: none;
}

.el-button--primary:hover {
  background: var(--primary-light);
}

.el-button--success {
  background: var(--success-color);
  border: none;
}

.el-button--warning {
  background: var(--warning-color);
  border: none;
}

.el-button--danger {
  background: var(--danger-color);
  border: none;
}

.el-table {
  --el-table-border-color: var(--border-color);
  --el-table-header-bg-color: var(--border-light);
  --el-table-row-hover-bg-color: var(--primary-lighter);
}

.el-table th.el-table__cell {
  background-color: var(--border-light);
  color: var(--text-primary);
  font-weight: 600;
}

.el-tag--success {
  background-color: #D1FAE5;
  border-color: #A7F3D0;
  color: #065F46;
}

.el-tag--warning {
  background-color: #FEF3C7;
  border-color: #FDE68A;
  color: #92400E;
}

.el-tag--danger {
  background-color: #FEE2E2;
  border-color: #FECACA;
  color: #991B1B;
}

.el-tag--info {
  background-color: #E0E7FF;
  border-color: #C7D2FE;
  color: #3730A3;
}

.el-tag--primary {
  background-color: #DBEAFE;
  border-color: #BFDBFE;
  color: #1E40AF;
}

.el-dialog {
  border-radius: var(--radius-lg);
}

.el-dialog .el-dialog__header {
  border-bottom: 1px solid var(--border-color);
  padding: var(--space-md) var(--space-lg);
}

.el-dialog .el-dialog__body {
  padding: var(--space-lg);
}

.el-dialog .el-dialog__footer {
  border-top: 1px solid var(--border-color);
  padding: var(--space-md) var(--space-lg);
}

.el-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
}

.el-input__wrapper {
  border-radius: var(--radius-md);
}

.el-select .el-input__wrapper {
  border-radius: var(--radius-md);
}

.el-form-item__label {
  color: var(--text-secondary);
  font-weight: 500;
}

.el-message {
  border-radius: var(--radius-md);
}

.el-message--success {
  background: #D1FAE5;
  border-color: #A7F3D0;
}

.el-message--warning {
  background: #FEF3C7;
  border-color: #FDE68A;
}

.el-message--error {
  background: #FEE2E2;
  border-color: #FECACA;
}
</style>