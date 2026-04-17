<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">角色管理</h1>
      <div class="page-actions">
        <el-button type="primary" @click="showDialog">
          <el-icon><Plus /></el-icon>
          新增角色
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="roles" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column prop="name" label="角色名称" min-width="120"></el-table-column>
        <el-table-column prop="description" label="描述" min-width="180"></el-table-column>
        <el-table-column label="创建时间" width="140" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="editRole(scope.row)">编辑</el-button>
            <el-button type="success" size="small" link @click="showMenusDialog(scope.row)">分配菜单</el-button>
            <el-button type="warning" size="small" link @click="showApisDrawer(scope.row)">分配API</el-button>
            <el-button type="danger" size="small" link @click="deleteRole(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="roleForm.id ? '编辑角色' : '新增角色'" width="480px">
      <el-form :model="roleForm" label-width="80px">
        <el-form-item label="角色名称">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" placeholder="请输入描述"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRole">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menusDialogVisible" title="分配菜单" width="600px" destroy-on-close>
      <el-tree
        v-if="menusDialogVisible"
        ref="menuTreeRef"
        :data="menuTree"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedMenus"
        :props="{ label: 'name', children: 'children' }"
        check-strictly
      ></el-tree>
      <template #footer>
        <el-button @click="menusDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMenus">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="apisDrawerVisible" title="分配API权限" size="50%" direction="rtl">
      <div class="drawer-content">
        <div class="drawer-header">
          <el-input v-model="apiSearchKeyword" placeholder="搜索API路径或描述" clearable style="width: 300px">
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <span class="api-count">已选择 {{ selectedApiCount }} / {{ filteredApis.length }} 个</span>
        </div>
        
        <el-table 
          ref="apiTableRef"
          :data="filteredApis" 
          @selection-change="handleApiSelection" 
          row-key="id"
          height="calc(100vh - 200px)"
          style="width: 100%"
        >
          <el-table-column type="selection" width="55"></el-table-column>
          <el-table-column prop="group" label="分组" width="100">
            <template #default="scope">
              <el-tag size="small" effect="plain">{{ scope.row.group || '默认' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="method" label="方法" width="70" align="center">
            <template #default="scope">
              <el-tag :type="getMethodType(scope.row.method)" size="small">{{ scope.row.method }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路径" min-width="180"></el-table-column>
          <el-table-column prop="description" label="描述" min-width="140"></el-table-column>
        </el-table>
      </div>
      <template #footer>
        <div class="drawer-footer">
          <el-button @click="apisDrawerVisible = false">取消</el-button>
          <el-button type="primary" @click="saveApis">保存</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import request from '../utils/request'

const roles = ref([])
const dialogVisible = ref(false)
const roleForm = ref({
  name: '',
  description: ''
})

const menusDialogVisible = ref(false)
const menuTree = ref([])
const selectedMenus = ref([])
const currentRoleId = ref(0)
const menuTreeRef = ref(null)

const apisDrawerVisible = ref(false)
const apis = ref([])
const selectedApis = ref([])
const currentApiRoleId = ref(0)
const apiTableRef = ref(null)
const apiSearchKeyword = ref('')

const filteredApis = computed(() => {
  if (!apiSearchKeyword.value) {
    return apis.value
  }
  const keyword = apiSearchKeyword.value.toLowerCase()
  return apis.value.filter(api => 
    api.path?.toLowerCase().includes(keyword) ||
    api.description?.toLowerCase().includes(keyword) ||
    api.group?.toLowerCase().includes(keyword)
  )
})

const selectedApiCount = computed(() => {
  return selectedApis.value.length
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

const getMethodType = (method) => {
  const types = {
    'GET': 'success',
    'POST': 'primary',
    'PUT': 'warning',
    'DELETE': 'danger',
    'PATCH': 'info'
  }
  return types[method] || 'info'
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
  roleForm.value = { name: '', description: '' }
  dialogVisible.value = true
}

const editRole = (role) => {
  roleForm.value = { ...role }
  dialogVisible.value = true
}

const saveRole = async () => {
  try {
    if (roleForm.value.id) {
      await request.put(`/role/${roleForm.value.id}`, roleForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/role', roleForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchRoles()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '操作失败')
  }
}

const deleteRole = async (id) => {
  try {
    await request.delete(`/role/${id}`)
    ElMessage.success('删除成功')
    fetchRoles()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const showMenusDialog = async (role) => {
  currentRoleId.value = role.id
  
  try {
    const menuRes = await request.get('/menu/tree')
    if (menuRes.code === 200) {
      menuTree.value = menuRes.data || []
    }
    
    const roleMenusRes = await request.get(`/role/${role.id}/menus`)
    if (roleMenusRes.code === 200) {
      selectedMenus.value = roleMenusRes.data || []
    }
    
    menusDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取菜单数据失败')
  }
}

const saveMenus = async () => {
  try {
    if (!menuTreeRef.value) {
      ElMessage.error('获取菜单选择失败')
      return
    }
    
    const checkedKeys = menuTreeRef.value.getCheckedKeys()
    await request.post(`/role/${currentRoleId.value}/menus`, { menu_ids: checkedKeys })
    ElMessage.success('菜单分配成功')
    menusDialogVisible.value = false
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '分配失败')
  }
}

const showApisDrawer = async (role) => {
  currentApiRoleId.value = role.id
  apiSearchKeyword.value = ''
  selectedApis.value = []
  
  try {
    const apisRes = await request.get('/api', { params: { pageSize: 1000 } })
    apis.value = apisRes.data?.list || []
    
    const roleApisRes = await request.get(`/role/${role.id}/apis`)
    const roleApiIds = roleApisRes.data || []
    
    apisDrawerVisible.value = true
    
    setTimeout(() => {
      if (apiTableRef.value) {
        apis.value.forEach(api => {
          if (roleApiIds.includes(api.id)) {
            apiTableRef.value.toggleRowSelection(api, true)
          }
        })
      }
    }, 100)
  } catch (error) {
    ElMessage.error('获取API数据失败')
    console.error(error)
  }
}

const handleApiSelection = (selection) => {
  selectedApis.value = selection
}

const saveApis = async () => {
  try {
    const apiIds = selectedApis.value.map(api => api.id)
    await request.post(`/role/${currentApiRoleId.value}/apis`, { api_ids: apiIds })
    ElMessage.success('API分配成功')
    apisDrawerVisible.value = false
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '分配失败')
  }
}

onMounted(fetchRoles)
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

.drawer-content {
  padding: 0 20px;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.api-count {
  color: #909399;
  font-size: 14px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>