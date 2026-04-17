<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">K8s授权管理</h1>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        新增授权
      </el-button>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-form :inline="true" class="search-form">
        <el-form-item label="用户">
          <el-select v-model="searchParams.user_id" clearable placeholder="选择用户" style="width: 150px">
            <el-option v-for="user in users" :key="user.id" :label="user.username" :value="user.id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="集群">
          <el-select v-model="searchParams.cluster_id" clearable placeholder="选择集群" style="width: 150px">
            <el-option label="所有集群" value="0"></el-option>
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select v-model="searchParams.resource" clearable placeholder="选择资源" style="width: 120px">
            <el-option label="所有资源" value="*"></el-option>
            <el-option label="Deployment" value="deployment"></el-option>
            <el-option label="Pod" value="pod"></el-option>
            <el-option label="Service" value="service"></el-option>
            <el-option label="ConfigMap" value="configmap"></el-option>
            <el-option label="Secret" value="secret"></el-option>
            <el-option label="Ingress" value="ingress"></el-option>
            <el-option label="Node" value="node"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchAuthorizations">搜索</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="authorizations" style="width: 100%" v-loading="loading">
        <el-table-column prop="user_id" label="用户ID" width="80"></el-table-column>
        <el-table-column label="用户名" width="120">
          <template #default="scope">{{ getUserName(scope.row.user_id) }}</template>
        </el-table-column>
        <el-table-column label="集群" width="120">
          <template #default="scope">
            <span v-if="scope.row.cluster_id === 0">所有集群</span>
            <span v-else>{{ getClusterName(scope.row.cluster_id) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.namespace === '*'" type="info" size="small">所有</el-tag>
            <span v-else>{{ scope.row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="资源类型" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.resource === '*'" type="info" size="small">所有</el-tag>
            <el-tag v-else type="primary" size="small">{{ scope.row.resource }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="查看" width="60" align="center">
          <template #default="scope">
            <el-icon v-if="scope.row.can_view" class="permission-icon success"><SuccessFilled /></el-icon>
            <el-icon v-else class="permission-icon danger"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="编辑" width="60" align="center">
          <template #default="scope">
            <el-icon v-if="scope.row.can_edit" class="permission-icon success"><SuccessFilled /></el-icon>
            <el-icon v-else class="permission-icon danger"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="删除" width="60" align="center">
          <template #default="scope">
            <el-icon v-if="scope.row.can_delete" class="permission-icon success"><SuccessFilled /></el-icon>
            <el-icon v-else class="permission-icon danger"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="创建" width="60" align="center">
          <template #default="scope">
            <el-icon v-if="scope.row.can_create" class="permission-icon success"><SuccessFilled /></el-icon>
            <el-icon v-else class="permission-icon danger"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="150"></el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="showEditDialog(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteAuthorization(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchAuthorizations"
        @current-change="fetchAuthorizations"
        style="margin-top: 16px; justify-content: flex-end;"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑授权' : '新增授权'" width="500px">
      <el-form :model="authForm" label-width="100px">
        <el-form-item label="用户" required>
          <el-select v-model="authForm.user_id" placeholder="选择用户" style="width: 100%">
            <el-option v-for="user in users" :key="user.id" :label="user.username" :value="user.id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="集群">
          <el-select v-model="authForm.cluster_id" placeholder="选择集群" style="width: 100%">
            <el-option label="所有集群" :value="0"></el-option>
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间" required>
          <el-input v-model="authForm.namespace" placeholder="输入命名空间，* 表示所有"></el-input>
        </el-form-item>
        <el-form-item label="资源类型" required>
          <el-select v-model="authForm.resource" placeholder="选择资源类型" style="width: 100%">
            <el-option label="所有资源" value="*"></el-option>
            <el-option label="Deployment" value="deployment"></el-option>
            <el-option label="Pod" value="pod"></el-option>
            <el-option label="Service" value="service"></el-option>
            <el-option label="ConfigMap" value="configmap"></el-option>
            <el-option label="Secret" value="secret"></el-option>
            <el-option label="Ingress" value="ingress"></el-option>
            <el-option label="Node" value="node"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="权限">
          <el-checkbox v-model="authForm.can_view">查看</el-checkbox>
          <el-checkbox v-model="authForm.can_edit">编辑</el-checkbox>
          <el-checkbox v-model="authForm.can_delete">删除</el-checkbox>
          <el-checkbox v-model="authForm.can_create">创建</el-checkbox>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="authForm.description" type="textarea" placeholder="授权说明"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAuth">{{ isEdit ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import request from '../../utils/request'

const authorizations = ref([])
const users = ref([])
const clusters = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)

const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const searchParams = ref({
  user_id: null,
  cluster_id: null,
  resource: ''
})

const authForm = ref({
  id: null,
  user_id: null,
  cluster_id: 0,
  namespace: '*',
  resource: '*',
  can_view: true,
  can_edit: false,
  can_delete: false,
  can_create: false,
  description: ''
})

const fetchUsers = async () => {
  try {
    const res = await request.get('/user')
    users.value = res.data || []
  } catch (error) {}
}

const fetchClusters = async () => {
  try {
    const res = await request.get('/k8s/cluster')
    clusters.value = res.data || []
  } catch (error) {}
}

const fetchAuthorizations = async () => {
  loading.value = true
  try {
    const res = await request.get('/k8s/auth', {
      params: {
        ...searchParams.value,
        page: page.value,
        pageSize: pageSize.value
      }
    })
    authorizations.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
  } finally {
    loading.value = false
  }
}

const getUserName = (userID) => {
  const user = users.value.find(u => u.id === userID)
  return user?.username || ''
}

const getClusterName = (clusterID) => {
  const cluster = clusters.value.find(c => c.id === clusterID)
  return cluster?.name || ''
}

const showCreateDialog = () => {
  isEdit.value = false
  authForm.value = {
    id: null,
    user_id: null,
    cluster_id: 0,
    namespace: '*',
    resource: '*',
    can_view: true,
    can_edit: false,
    can_delete: false,
    can_create: false,
    description: ''
  }
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  authForm.value = {
    id: row.id,
    user_id: row.user_id,
    cluster_id: row.cluster_id,
    namespace: row.namespace,
    resource: row.resource,
    can_view: row.can_view,
    can_edit: row.can_edit,
    can_delete: row.can_delete,
    can_create: row.can_create,
    description: row.description
  }
  dialogVisible.value = true
}

const submitAuth = async () => {
  if (!authForm.value.user_id) {
    ElMessage.warning('请选择用户')
    return
  }
  if (!authForm.value.namespace) {
    ElMessage.warning('请输入命名空间')
    return
  }
  if (!authForm.value.resource) {
    ElMessage.warning('请选择资源类型')
    return
  }

  try {
    if (isEdit.value) {
      await request.put(`/k8s/auth/${authForm.value.id}`, authForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/k8s/auth', authForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchAuthorizations()
  } catch (error) {}
}

const deleteAuthorization = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除此授权吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await request.delete(`/k8s/auth/${row.id}`)
    ElMessage.success('删除成功')
    fetchAuthorizations()
  } catch (error) {
    if (error !== 'cancel') {}
  }
}

onMounted(() => {
  fetchUsers()
  fetchClusters()
  fetchAuthorizations()
})
</script>

<style scoped>
.search-form {
  margin-bottom: 16px;
}

.permission-icon {
  font-size: 16px;
}

.permission-icon.success {
  color: #67C23A;
}

.permission-icon.danger {
  color: #F56C6C;
}
</style>