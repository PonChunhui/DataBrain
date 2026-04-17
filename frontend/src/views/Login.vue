<template>
  <div class="login-container">
    <div class="login-left">
      <div class="brand-content">
        <div class="logo-wrapper">
          <div class="logo-icon">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
          </div>
          <h1 class="brand-title">数智运维平台</h1>
        </div>
        <p class="brand-subtitle">DevOps Platform</p>
        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/></svg>
            </div>
            <span>多集群统一管理</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/></svg>
            </div>
            <span>精细化权限控制</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3.5 18.49l6-6.01 4 4L22 6.92l-1.41-1.41-7.09 7.97-4-4L2 16.99z"/></svg>
            </div>
            <span>实时监控告警</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM19 18H6c-2.21 0-4-1.79-4-4s1.79-4 4-4h.71C7.37 7.69 9.48 6 12 6c3.04 0 5.5 2.46 5.5 5.5v.5H19c1.66 0 3 1.34 3 3s-1.34 3-3 3z"/></svg>
            </div>
            <span>自动化运维流程</span>
          </div>
        </div>
      </div>
      <div class="decoration-circles">
        <div class="circle circle-1"></div>
        <div class="circle circle-2"></div>
        <div class="circle circle-3"></div>
      </div>
    </div>
    
    <div class="login-right">
      <div class="login-form-wrapper">
        <div class="form-header">
          <h2>欢迎登录</h2>
          <p>请输入您的账号信息</p>
        </div>
        
        <el-form :model="form" class="login-form">
          <el-form-item>
            <el-input 
              v-model="form.username" 
              placeholder="请输入用户名" 
              prefix-icon="User"
              size="large"
              clearable
            />
          </el-form-item>
          <el-form-item>
            <el-input 
              v-model="form.password" 
              type="password" 
              placeholder="请输入密码" 
              prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button 
              type="primary" 
              size="large" 
              :loading="loading" 
              @click="handleLogin"
              class="login-btn"
            >
              <span v-if="!loading">登 录</span>
              <span v-else>登录中...</span>
            </el-button>
          </el-form-item>
        </el-form>
        
        <div class="form-footer">
          <p>数智运维平台 © 2024</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '../utils/request'
import { fetchUserPermissions } from '../utils/permission'

const router = useRouter()
const form = ref({
  username: '',
  password: ''
})
const loading = ref(false)

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    const res = await request.post('/user/login', form.value)
    if (res.code === 200) {
      const { token, user } = res.data
      
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
      localStorage.setItem('isLoggedIn', 'true')
      localStorage.setItem('loginTime', new Date().toISOString())
      
      try {
        const menuRes = await request.get(`/user/${user.id}/menus`)
        if (menuRes.code === 200) {
          localStorage.setItem('menus', JSON.stringify(menuRes.data || []))
          await fetchUserPermissions()
        }
        
        const apiRes = await request.get(`/user/${user.id}/apis`)
        if (apiRes.code === 200) {
          localStorage.setItem('apis', JSON.stringify(apiRes.data || []))
        }
      } catch (error) {
        console.error('获取权限数据失败:', error)
      }
      
      ElMessage.success('登录成功')
      router.push('/dashboard')
    }
  } catch (error) {
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

.login-left {
  width: 55%;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
}

.brand-content {
  z-index: 10;
  text-align: center;
  padding: 40px;
}

.logo-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 16px;
}

.logo-icon {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s ease-in-out infinite;
}

.logo-icon svg {
  width: 36px;
  height: 36px;
  color: white;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.4);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 0 20px 10px rgba(102, 126, 234, 0.2);
  }
}

.brand-title {
  font-size: 36px;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  letter-spacing: 2px;
}

.brand-subtitle {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.6);
  margin: 0;
  letter-spacing: 4px;
  text-transform: uppercase;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-top: 48px;
  padding: 0 60px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateX(8px);
}

.feature-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.feature-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.feature-item span {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.9);
  font-weight: 500;
}

.decoration-circles {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  z-index: 1;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.3) 0%, rgba(118, 75, 162, 0.3) 100%);
}

.circle-1 {
  width: 300px;
  height: 300px;
  top: -100px;
  left: -100px;
  animation: float 6s ease-in-out infinite;
}

.circle-2 {
  width: 200px;
  height: 200px;
  bottom: -50px;
  right: -50px;
  animation: float 8s ease-in-out infinite reverse;
}

.circle-3 {
  width: 150px;
  height: 150px;
  top: 50%;
  left: 10%;
  animation: float 10s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-30px) rotate(180deg);
  }
}

.login-right {
  width: 45%;
  background: #f8fafc;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
}

.login-right::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  width: 1px;
  height: 100%;
  background: linear-gradient(to bottom, transparent, rgba(102, 126, 234, 0.3), transparent);
}

.login-form-wrapper {
  width: 400px;
  padding: 40px;
}

.form-header {
  text-align: center;
  margin-bottom: 40px;
}

.form-header h2 {
  font-size: 28px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 12px 0;
}

.form-header p {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}

.login-form {
  margin-top: 0;
}

.login-form .el-form-item {
  margin-bottom: 24px;
}

.login-form .el-input {
  height: 48px;
}

.login-form .el-input__wrapper {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #e2e8f0;
  transition: all 0.3s ease;
}

.login-form .el-input__wrapper:hover {
  border-color: #cbd5e1;
}

.login-form .el-input__wrapper.is-focus {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.login-form .el-input__inner {
  height: 48px;
  line-height: 48px;
  font-size: 15px;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  letter-spacing: 4px;
  transition: all 0.3s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.login-btn:active {
  transform: translateY(0);
}

.form-footer {
  text-align: center;
  margin-top: 40px;
}

.form-footer p {
  font-size: 12px;
  color: #94a3b8;
  margin: 0;
}

@media (max-width: 1024px) {
  .login-left {
    width: 40%;
  }
  .login-right {
    width: 60%;
  }
  .brand-title {
    font-size: 28px;
  }
  .feature-list {
    padding: 0 40px;
  }
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }
  .login-left {
    width: 100%;
    height: 40vh;
    padding: 20px;
  }
  .login-right {
    width: 100%;
    height: 60vh;
  }
  .brand-title {
    font-size: 24px;
  }
  .feature-list {
    display: none;
  }
  .login-form-wrapper {
    width: 100%;
    padding: 20px 40px;
  }
}
</style>