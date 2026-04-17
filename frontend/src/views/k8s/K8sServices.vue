<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Service管理</h1>
      <div class="header-right">
        <K8sResourceFilter
          ref="resourceFilter"
          @query="handleQuery"
          @cluster-change="handleClusterChange"
          @ready="handleReady"
        />
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="services" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="160">
          <template #default="scope">
            <el-button type="primary" link @click="goToDetail(scope.row)">{{ scope.row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120"></el-table-column>
        <el-table-column prop="type" label="类型" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getServiceType(scope.row.type)" size="small">{{ scope.row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cluster_ip" label="ClusterIP" width="140"></el-table-column>
        <el-table-column label="端口映射" min-width="200">
          <template #default="scope">
            <el-tag v-for="(port, idx) in scope.row.ports" :key="idx" size="small" style="margin-right: 4px">
              {{ port.port }}:{{ port.target_port }}/{{ port.protocol }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="200" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center" fixed="right">
          <template #default="scope">
            <el-button type="danger" size="small" link @click="deleteService(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const router = useRouter()
const resourceFilter = ref(null)
const services = ref([])
const loading = ref(false)
const searchParams = ref({
  cluster: null,
  namespace: 'default'
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getServiceType = (type) => {
  const types = {
    'ClusterIP': 'primary',
    'NodePort': 'success',
    'LoadBalancer': 'warning',
    'ExternalName': 'info'
  }
  return types[type] || 'info'
}

const handleQuery = (params) => {
  searchParams.value = params
  fetchServices()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchServices()
}

const handleClusterChange = () => {
  services.value = []
}

const fetchServices = async () => {
  if (!searchParams.value.cluster) {
    return
  }
  
  loading.value = true
  try {
    const res = await request.get('/k8s/service', {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    services.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Service列表失败')
  } finally {
    loading.value = false
  }
}

const deleteService = async (svc) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Service ${svc.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/service/${svc.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('删除成功')
    fetchServices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const goToDetail = (svc) => {
  router.push({
    path: `/k8s/service/${svc.name}`,
    query: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
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