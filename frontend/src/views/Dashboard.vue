<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">系统概览</h1>
      <el-dropdown @command="handleCommand" trigger="hover">
        <div class="user-info">
          <div class="user-avatar">
            <el-icon><UserFilled /></el-icon>
          </div>
          <div class="user-details">
            <div class="user-name">{{ currentUser?.username || '用户' }}</div>
            <div class="login-time-text">上次登录：{{ loginTime }}</div>
          </div>
          <el-icon class="dropdown-icon"><ArrowUp /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">
              <el-icon><User /></el-icon>
              <span>个人信息</span>
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <el-icon><SwitchButton /></el-icon>
              <span>退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div class="stats-grid">
      <div class="stat-card stat-card-primary">
        <div class="stat-icon">
          <el-icon><User /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.userCount }}</div>
          <div class="stat-label">用户总数</div>
        </div>
      </div>
      <div class="stat-card stat-card-success">
        <div class="stat-icon">
          <el-icon><Avatar /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.roleCount }}</div>
          <div class="stat-label">角色总数</div>
        </div>
      </div>
      <div class="stat-card stat-card-warning">
        <div class="stat-icon">
          <el-icon><Lock /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.permissionCount }}</div>
          <div class="stat-label">权限总数</div>
        </div>
      </div>
      <div class="stat-card stat-card-danger">
        <div class="stat-icon">
          <el-icon><Connection /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.apiCount }}</div>
          <div class="stat-label">API接口</div>
        </div>
      </div>
    </div>

    <el-row :gutter="24" class="section-gap">
      <el-col :span="16">
        <el-card class="content-card" shadow="hover">
          <template #header>
            <div class="card-header-title">快捷操作</div>
          </template>
          <div class="quick-grid">
            <div class="quick-item" @click="$router.push('/users')">
              <el-icon class="quick-icon"><User /></el-icon>
              <span>用户管理</span>
            </div>
            <div class="quick-item" @click="$router.push('/roles')">
              <el-icon class="quick-icon"><Avatar /></el-icon>
              <span>角色管理</span>
            </div>
            <div class="quick-item" @click="$router.push('/menus')">
              <el-icon class="quick-icon"><Menu /></el-icon>
              <span>菜单管理</span>
            </div>
            <div class="quick-item" @click="$router.push('/apis')">
              <el-icon class="quick-icon"><Connection /></el-icon>
              <span>API管理</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="content-card system-card" shadow="hover">
          <template #header>
            <div class="card-header-title">系统信息</div>
          </template>
          <div class="system-list">
            <div class="system-item">
              <span class="system-label">系统名称</span>
              <span class="system-value">数智运维平台</span>
            </div>
            <div class="system-item">
              <span class="system-label">版本</span>
              <span class="system-value">v1.0.0</span>
            </div>
            <div class="system-item">
              <span class="system-label">技术栈</span>
              <span class="system-value">Go + Vue3</span>
            </div>
            <div class="system-item">
              <span class="system-label">数据库</span>
              <span class="system-value">MySQL 8.0</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { User, Avatar, Lock, Connection, Menu, UserFilled, ArrowUp, SwitchButton } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../utils/request'

const router = useRouter()

const stats = ref({
  userCount: 0,
  roleCount: 0,
  permissionCount: 0,
  apiCount: 12
})

const currentUser = ref(null)
const loginTime = ref('')

const fetchStats = async () => {
  try {
    const [usersRes, rolesRes, buttonsRes] = await Promise.all([
      request.get('/user'),
      request.get('/role'),
      request.get('/menu-button')
    ])

    stats.value.userCount = usersRes.data?.length || 0
    stats.value.roleCount = rolesRes.data?.length || 0
    stats.value.permissionCount = buttonsRes.data?.length || 0
  } catch (error) {
    ElMessage.error('获取统计数据失败')
  }
}

const handleCommand = (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      handleLogout()
    }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/profile')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  localStorage.removeItem('isLoggedIn')
  localStorage.removeItem('loginTime')
  localStorage.removeItem('menus')
  localStorage.removeItem('apis')
  
  ElMessage.success('已退出登录')
  router.push('/login')
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
  }
  
  const loginTimeStr = localStorage.getItem('loginTime')
  if (loginTimeStr) {
    loginTime.value = new Date(loginTimeStr).toLocaleString('zh-CN')
  } else {
    loginTime.value = new Date().toLocaleString('zh-CN')
  }
  
  fetchStats()
})
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 24px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
}

.stat-icon .el-icon {
  font-size: 28px;
}

.stat-card-primary .stat-icon {
  background: rgba(64, 158, 255, 0.15);
  color: #409EFF;
}

.stat-card-success .stat-icon {
  background: rgba(103, 194, 58, 0.15);
  color: #67C23A;
}

.stat-card-warning .stat-icon {
  background: rgba(230, 162, 60, 0.15);
  color: #E6A23C;
}

.stat-card-danger .stat-icon {
  background: rgba(245, 108, 108, 0.15);
  color: #F56C6C;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.card-header-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  border-radius: 12px;
  background: #f5f7fa;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-item:hover {
  background: #e8eaed;
  transform: translateY(-2px);
}

.quick-icon {
  font-size: 32px;
  color: #409EFF;
  margin-bottom: 12px;
}

.quick-item span {
  font-size: 14px;
  color: #606266;
}

.system-card .system-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.system-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.system-label {
  font-size: 14px;
  color: #909399;
}

.system-value {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.user-info {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  border-radius: 10px;
  background: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.2);
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-info:hover {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.3);
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #3B82F6 0%, #6366F1 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.user-avatar .el-icon {
  font-size: 18px;
  color: #FFFFFF;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: #1E293B;
}

.login-time-text {
  font-size: 12px;
  color: #64748B;
  margin-top: 2px;
}

.dropdown-icon {
  font-size: 12px;
  color: #64748B;
  margin-left: 8px;
  transition: transform 0.2s;
}

.user-info:hover .dropdown-icon {
  color: #3B82F6;
}
</style>