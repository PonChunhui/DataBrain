<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <h1 class="error-code">404</h1>
      <h2 class="error-title">页面不存在</h2>
      <p class="error-message">抱歉，您访问的页面不存在或已被移除</p>
      <el-button type="primary" @click="goHome" size="large">返回首页</el-button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.not-found-content {
  text-align: center;
  padding: 40px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.error-code {
  font-size: 120px;
  font-weight: bold;
  color: #667eea;
  margin: 0;
  line-height: 1;
}

.error-title {
  font-size: 32px;
  color: #333;
  margin: 20px 0 10px;
}

.error-message {
  font-size: 16px;
  color: #666;
  margin-bottom: 30px;
}

.el-button {
  padding: 12px 40px;
}
</style>