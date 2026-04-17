<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">用户管理</h1>
      <div class="page-actions">
        <el-button v-if="hasPermission('user:create')" type="primary" @click="showDialog">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column prop="username" label="用户名" min-width="120"></el-table-column>
        <el-table-column prop="real_name" label="姓名" min-width="100"></el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="120"></el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="180"></el-table-column>
        <el-table-column label="角色" min-width="120">
          <template #default="scope">
            <el-tag v-for="role in scope.row.roles" :key="role.id" type="success" size="small" style="margin-right: 6px">
              {{ role.name }}
            </el-tag>
            <el-tag v-if="!scope.row.roles || scope.row.roles.length === 0" type="info" size="small">未分配</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="140" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="scope">
            <el-button v-if="hasPermission('user:edit')" type="primary" size="small" link @click="editUser(scope.row)">
              编辑
            </el-button>
            <el-button v-if="hasPermission('user:delete')" type="danger" size="small" link @click="deleteUser(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="userForm.id ? '编辑用户' : '新增用户'" width="480px">
      <el-form :model="userForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="userForm.username" placeholder="请输入用户名"></el-input>
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
        <el-form-item label="密码" v-if="!userForm.id">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码"></el-input>
        </el-form-item>
        <el-form-item label="角色" v-if="userForm.id">
          <el-select v-model="userForm.role_ids" placeholder="请选择角色" multiple clearable style="width: 100%">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveUser">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '../utils/request'
import { hasPermission } from '../utils/permission'

const users = ref([])
const roles = ref([])
const dialogVisible = ref(false)
const userForm = ref({
  username: '',
  real_name: '',
  phone: '',
  email: '',
  password: '',
  role_ids: []
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

const fetchUsers = async () => {
  try {
    const res = await request.get('/user')
    users.value = res.data || []
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  }
}

const fetchRoles = async () => {
  try {
    const res = await request.get('/role')
    roles.value = res.data || []
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  }
}

const showDialog = () => {
  userForm.value = { username: '', real_name: '', phone: '', email: '', password: '', role_ids: [] }
  dialogVisible.value = true
}

const editUser = async (user) => {
  userForm.value = { ...user, role_ids: [] }
  
  try {
    const rolesRes = await request.get(`/user/${user.id}/roles`)
    if (rolesRes.data && rolesRes.data.length > 0) {
      userForm.value.role_ids = rolesRes.data.map(r => r.id)
    }
  } catch (error) {
    console.error('获取用户角色失败', error)
  }
  
  dialogVisible.value = true
}

const saveUser = async () => {
  try {
    if (userForm.value.id) {
      await request.put(`/user/${userForm.value.id}`, {
        username: userForm.value.username,
        real_name: userForm.value.real_name,
        phone: userForm.value.phone,
        email: userForm.value.email
      })
      
      await request.post(`/user/${userForm.value.id}/roles`, { role_ids: userForm.value.role_ids || [] })
      
      ElMessage.success('更新成功')
    } else {
      await request.post('/user', userForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '操作失败')
  }
}

const deleteUser = async (id) => {
  try {
    await request.delete(`/user/${id}`)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<style scoped>
.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.el-dialog .el-form-item {
  margin-bottom: 20px;
}
</style>