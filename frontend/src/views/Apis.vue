<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">API管理</h1>
      <div class="page-actions">
        <el-button type="primary" @click="showDialog">
          <el-icon><Plus /></el-icon>
          新增API
        </el-button>
      </div>
    </div>

    <el-card class="search-card" shadow="hover">
      <el-form :model="searchForm" inline>
        <el-form-item label="路径">
          <el-input v-model="searchForm.path" placeholder="搜索路径" clearable style="width: 180px"></el-input>
        </el-form-item>
        <el-form-item label="方法">
          <el-select v-model="searchForm.method" placeholder="选择方法" clearable style="width: 100px">
            <el-option label="GET" value="GET"></el-option>
            <el-option label="POST" value="POST"></el-option>
            <el-option label="PUT" value="PUT"></el-option>
            <el-option label="DELETE" value="DELETE"></el-option>
            <el-option label="PATCH" value="PATCH"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="searchForm.group" placeholder="搜索分组" clearable style="width: 120px"></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 100px">
            <el-option label="启用" :value="true"></el-option>
            <el-option label="禁用" :value="false"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="content-card section-gap" shadow="hover">
      <el-table :data="apis" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column prop="path" label="路径" min-width="200"></el-table-column>
        <el-table-column prop="method" label="方法" width="80" align="center">
          <template #default="scope">
            <el-tag :type="getMethodType(scope.row.method)" size="small">{{ scope.row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="160"></el-table-column>
        <el-table-column prop="group" label="分组" width="100"></el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.status" type="success" size="small">启用</el-tag>
            <el-tag v-else type="info" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="editApi(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteApi(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px; justify-content: flex-end"
      ></el-pagination>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="apiForm.id ? '编辑API' : '新增API'" width="500px">
      <el-form :model="apiForm" label-width="80px">
        <el-form-item label="路径">
          <el-input v-model="apiForm.path" placeholder="/api/example"></el-input>
        </el-form-item>
        <el-form-item label="方法">
          <el-select v-model="apiForm.method" placeholder="请选择方法" style="width: 100%">
            <el-option label="GET" value="GET"></el-option>
            <el-option label="POST" value="POST"></el-option>
            <el-option label="PUT" value="PUT"></el-option>
            <el-option label="DELETE" value="DELETE"></el-option>
            <el-option label="PATCH" value="PATCH"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="apiForm.description" placeholder="API描述"></el-input>
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="apiForm.group" placeholder="用户管理"></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="apiForm.status"></el-switch>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveApi">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '../utils/request'

const apis = ref([])
const dialogVisible = ref(false)
const apiForm = ref({
  path: '',
  method: 'GET',
  description: '',
  group: '',
  status: true
})

const searchForm = ref({
  path: '',
  method: '',
  group: '',
  status: null
})

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
})

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

const fetchApis = async () => {
  try {
    const params = {
      ...searchForm.value,
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    }
    
    if (params.status === null || params.status === '') {
      params.status = undefined
    }
    
    const res = await request.get('/api', { params })
    if (res.code === 200) {
      apis.value = res.data.list || []
      pagination.value.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取API列表失败')
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  fetchApis()
}

const handleReset = () => {
  searchForm.value = {
    path: '',
    method: '',
    group: '',
    status: null
  }
  pagination.value.page = 1
  fetchApis()
}

const handleSizeChange = (val) => {
  pagination.value.pageSize = val
  fetchApis()
}

const handleCurrentChange = (val) => {
  pagination.value.page = val
  fetchApis()
}

const showDialog = () => {
  apiForm.value = { path: '', method: 'GET', description: '', group: '', status: true }
  dialogVisible.value = true
}

const editApi = (api) => {
  apiForm.value = { ...api }
  dialogVisible.value = true
}

const saveApi = async () => {
  try {
    if (apiForm.value.id) {
      await request.put(`/api/${apiForm.value.id}`, apiForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/api', apiForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchApis()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '操作失败')
  }
}

const deleteApi = async (id) => {
  try {
    await request.delete(`/api/${id}`)
    ElMessage.success('删除成功')
    fetchApis()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(fetchApis)
</script>

<style scoped>
.search-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px 24px;
}

.search-card .el-form-item {
  margin-bottom: 0;
}

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