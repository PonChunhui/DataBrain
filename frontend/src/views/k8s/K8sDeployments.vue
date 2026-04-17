<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Deployment管理</h1>
      <div class="header-right">
        <K8sResourceFilter 
          ref="resourceFilter"
          @query="handleQuery"
          @cluster-change="handleClusterChange"
          @ready="handleReady"
        />
        <el-button type="primary" @click="$router.push('/k8s/deployment/create')">
          <el-icon><Plus /></el-icon>
          创建Deployment
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="deployments" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="140">
          <template #default="scope">
            <el-link type="primary" @click="goToDetail(scope.row)">{{ scope.row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="副本数" width="140" align="center">
          <template #default="scope">
            <span>{{ scope.row.ready_replicas }}/{{ scope.row.replicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120"></el-table-column>
        <el-table-column label="镜像" min-width="200">
          <template #default="scope">
            <el-tag v-for="(image, idx) in scope.row.images" :key="idx" size="small" style="margin-right: 4px">
              {{ image }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="scaleDialog(scope.row)">扩缩容</el-button>
            <el-button type="warning" size="small" link @click="restartDeployment(scope.row)">重启</el-button>
            <el-button type="danger" size="small" link @click="deleteDeployment(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="400px">
      <el-form label-width="80px">
        <el-form-item label="当前副本">
          <span style="font-size: 16px; font-weight: 500;">{{ currentDeployment.value?.replicas || 0 }}</span>
        </el-form-item>
        <el-form-item label="目标副本">
          <el-input-number v-model="targetReplicas" :min="0" :max="100"></el-input-number>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="scaleDeployment">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '../../utils/request'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const router = useRouter()
const resourceFilter = ref(null)
const deployments = ref([])
const loading = ref(false)
const searchParams = ref({
  cluster: null,
  namespace: 'default'
})

const scaleDialogVisible = ref(false)
const currentDeployment = ref({})
const targetReplicas = ref(1)

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getStatusType = (status) => {
  const statusMap = {
    'Running': 'success',
    'Updating': 'warning',
    'Starting': 'info',
    'Paused': 'info',
    'ScaledDown': 'info',
    'Degraded': 'warning',
    'Failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const goToDetail = (row) => {
  router.push({
    path: `/k8s/deployment/${row.name}`,
    query: {
      cluster: searchParams.value.cluster,
      namespace: row.namespace || searchParams.value.namespace
    }
  })
}

const handleQuery = (params) => {
  searchParams.value = params
  fetchDeployments()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchDeployments()
}

const handleClusterChange = () => {
  deployments.value = []
}

const fetchDeployments = async () => {
  if (!searchParams.value.cluster) {
    return
  }
  
  loading.value = true
  try {
    const res = await request.get('/k8s/deployment', {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    deployments.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Deployment列表失败')
  } finally {
    loading.value = false
  }
}

const scaleDialog = (deploy) => {
  currentDeployment.value = deploy
  targetReplicas.value = deploy.replicas
  scaleDialogVisible.value = true
}

const scaleDeployment = async () => {
  try {
    await request.post(`/k8s/deployment/${currentDeployment.value.name}/scale`, {
      replicas: targetReplicas.value
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('扩缩容成功')
    scaleDialogVisible.value = false
    fetchDeployments()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '扩缩容失败')
  }
}

const restartDeployment = async (deploy) => {
  try {
    await ElMessageBox.confirm(`确定要重启 ${deploy.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
await request.post(`/k8s/deployment/${deploy.name}/restart`, null, {
    params: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
  ElMessage.success('重启成功')
  fetchDeployments()
} catch (error) {
  if (error !== 'cancel') {
    ElMessage.error('重启失败')
  }
}
}

const deleteDeployment = async (deploy) => {
try {
  await ElMessageBox.confirm(`确定要删除 ${deploy.name} 吗?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
  
  await request.delete(`/k8s/deployment/${deploy.name}`, {
    params: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
  ElMessage.success('删除成功')
  fetchDeployments()
} catch (error) {
  if (error !== 'cancel') {
    ElMessage.error('删除失败')
  }
}
}
</script>

<style scoped>
.header-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}
</style>