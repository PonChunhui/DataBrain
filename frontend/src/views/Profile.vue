<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">个人信息</h1>
    </div>

    <el-row :gutter="24">
      <el-col :span="12">
        <el-card class="content-card" shadow="hover">
          <template #header>
            <div class="card-header-title">基本信息</div>
          </template>
          
          <el-form :model="userForm" label-width="90px">
            <el-form-item label="用户名">
              <el-input v-model="userForm.username" disabled></el-input>
            </el-form-item>
            <el-form-item label="姓名">
              <el-input v-model="userForm.real_name" placeholder="请输入姓名"></el-input>
            </el-form-item>
            <el-form-item label="手机号">
              <el-input v-model="userForm.phone" placeholder="请输入手机号"></el-input>
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="userForm.email" placeholder="请输入邮箱"></el-input>
            </el-form-item>
            <el-form-item label="角色">
              <el-tag v-for="role in userForm.roles" :key="role.id" type="success" size="small" style="margin-right: 6px">
                {{ role.name }}
              </el-tag>
              <el-tag v-if="!userForm.roles || userForm.roles.length === 0" type="info" size="small">未分配</el-tag>
            </el-form-item>
            <el-form-item label="创建时间">
              <el-input :value="formatTime(userForm.created_at)" disabled></el-input>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="updateProfile">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="content-card" shadow="hover">
          <template #header>
            <div class="card-header-title">修改密码</div>
          </template>
          
          <el-form :model="passwordForm" label-width="90px">
            <el-form-item label="原密码">
              <el-input v-model="passwordForm.old_password" type="password" show-password placeholder="请输入原密码"></el-input>
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="请输入新密码"></el-input>
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input v-model="passwordForm.confirm_password" type="password" show-password placeholder="请确认新密码"></el-input>
            </el-form-item>
            <el-form-item>
              <el-button type="warning" @click="changePassword">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="24" style="margin-top: 24px;">
      <el-col :span="24">
        <el-card class="content-card" shadow="hover">
          <template #header>
            <div class="card-header-title" style="display: flex; justify-content: space-between; align-items: center;">
              <span>Webhook Token</span>
              <el-button type="primary" size="small" @click="showCreateDialog">新增</el-button>
            </div>
          </template>

          <el-table :data="webhooks" style="width: 100%">
            <el-table-column prop="name" label="名称" min-width="100"></el-table-column>
            <el-table-column prop="token" label="Token" min-width="200">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <span style="font-family: monospace; font-size: 12px;">{{ row.token.substring(0, 16) }}...</span>
                  <el-button size="small" text @click="copyToken(row.token)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="expires_days" label="有效期(天)" width="100" align="center"></el-table-column>
            <el-table-column prop="expires_at" label="过期时间" width="160">
              <template #default="{ row }">
                <span :style="{ color: row.is_expired ? '#F56C6C' : '' }">{{ formatTime(row.expires_at) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.is_expired ? 'danger' : 'success'" size="small">
                  {{ row.is_expired ? '已过期' : '有效' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="primary" text @click="refreshWebhook(row)">刷新</el-button>
                <el-button size="small" type="danger" text @click="deleteWebhook(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="createDialogVisible" title="创建Webhook Token" width="400px">
      <el-form :model="webhookForm" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="webhookForm.name" placeholder="请输入名称"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="webhookForm.description" type="textarea" placeholder="请输入描述"></el-input>
        </el-form-item>
        <el-form-item label="有效期">
          <el-input-number v-model="webhookForm.expires_days" :min="1" :max="365" :step="1"></el-input-number>
          <span style="margin-left: 10px; color: #909399;">天</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createWebhook">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import request from '../utils/request'

const userForm = ref({
  username: '',
  real_name: '',
  phone: '',
  email: '',
  roles: [],
  created_at: ''
})

const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const userId = ref(0)
const webhooks = ref([])
const createDialogVisible = ref(false)
const webhookForm = ref({
  name: '',
  description: '',
  expires_days: 30
})

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const fetchUserInfo = async () => {
  try {
    const userStr = localStorage.getItem('user')
    if (!userStr) return
    
    const user = JSON.parse(userStr)
    userId.value = user.id
    
    const res = await request.get(`/user/${user.id}`)
    if (res.code === 200) {
      userForm.value = res.data || {}
      
      const rolesRes = await request.get(`/user/${user.id}/roles`)
      if (rolesRes.code === 200) {
        userForm.value.roles = rolesRes.data || []
      }
    }
  } catch (error) {
    ElMessage.error('获取用户信息失败')
  }
}

const fetchWebhooks = async () => {
  try {
    const res = await request.get('/webhook')
    if (res.code === 200) {
      webhooks.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('获取Webhook列表失败')
  }
}

const updateProfile = async () => {
  try {
    const res = await request.put(`/user/${userId.value}`, {
      username: userForm.value.username,
      real_name: userForm.value.real_name,
      phone: userForm.value.phone,
      email: userForm.value.email
    })
    
    if (res.code === 200) {
      ElMessage.success('更新成功')
      localStorage.setItem('user', JSON.stringify(userForm.value))
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '更新失败')
  }
}

const changePassword = async () => {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password || !passwordForm.value.confirm_password) {
    ElMessage.warning('请填写完整的密码信息')
    return
  }
  
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    ElMessage.warning('新密码和确认密码不一致')
    return
  }
  
  if (passwordForm.value.new_password.length < 6) {
    ElMessage.warning('密码长度至少6位')
    return
  }
  
  try {
    const res = await request.put(`/user/${userId.value}/password`, {
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    
    if (res.code === 200) {
      ElMessage.success('密码修改成功')
      passwordForm.value = {
        old_password: '',
        new_password: '',
        confirm_password: ''
      }
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '密码修改失败')
  }
}

const showCreateDialog = () => {
  webhookForm.value = {
    name: '',
    description: '',
    expires_days: 30
  }
  createDialogVisible.value = true
}

const createWebhook = async () => {
  if (!webhookForm.value.name) {
    ElMessage.warning('请输入名称')
    return
  }
  
  try {
    const res = await request.post('/webhook', webhookForm.value)
    if (res.code === 200) {
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      fetchWebhooks()
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '创建失败')
  }
}

const copyToken = async (token) => {
  try {
    await navigator.clipboard.writeText(token)
    ElMessage.success('Token已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const refreshWebhook = async (row) => {
  try {
    await ElMessageBox.confirm('刷新将生成新的Token，原Token将失效，是否继续?', '确认刷新', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await request.put(`/webhook/${row.id}/refresh`, { expires_days: row.expires_days })
    if (res.code === 200) {
      ElMessage.success('刷新成功')
      fetchWebhooks()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '刷新失败')
    }
  }
}

const deleteWebhook = async (row) => {
  try {
    await ElMessageBox.confirm('删除后Token将立即失效，是否继续?', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await request.delete(`/webhook/${row.id}`)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchWebhooks()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '删除失败')
    }
  }
}

onMounted(() => {
  fetchUserInfo()
  fetchWebhooks()
})
</script>

<style scoped>
.card-header-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.el-form-item {
  margin-bottom: 20px;
}
</style>