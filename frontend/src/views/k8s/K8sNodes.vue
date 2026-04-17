<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">集群节点</h1>
      <el-select v-model="selectedCluster" @change="fetchStats" style="width: 200px">
        <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id"></el-option>
      </el-select>
    </div>

    <div v-if="stats">
      <el-row :gutter="24" class="section-gap">
        <el-col :span="4">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-value">{{ stats.namespace_count }}</div>
            <div class="stat-label">命名空间</div>
          </el-card>
        </el-col>
        <el-col :span="4">
          <el-card class="stat-card stat-card-success" shadow="hover">
            <div class="stat-value">{{ stats.running_pods }}</div>
            <div class="stat-label">运行中Pod</div>
          </el-card>
        </el-col>
        <el-col :span="4">
          <el-card class="stat-card stat-card-warning" shadow="hover">
            <div class="stat-value">{{ stats.pending_pods }}</div>
            <div class="stat-label">等待中Pod</div>
          </el-card>
        </el-col>
        <el-col :span="4">
          <el-card class="stat-card stat-card-danger" shadow="hover">
            <div class="stat-value">{{ stats.failed_pods }}</div>
            <div class="stat-label">失败Pod</div>
          </el-card>
        </el-col>
        <el-col :span="4">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-value">{{ stats.pod_count }}</div>
            <div class="stat-label">Pod总数</div>
          </el-card>
        </el-col>
        <el-col :span="4">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-value">{{ stats.nodes?.length || 0 }}</div>
            <div class="stat-label">节点数</div>
          </el-card>
        </el-col>
      </el-row>

      <el-card class="content-card section-gap" shadow="hover">
        <template #header>
          <div class="card-header-title">节点列表</div>
        </template>
        <el-table :data="stats.nodes" style="width: 100%">
          <el-table-column prop="name" label="节点名称" min-width="180">
            <template #default="scope">
              <el-link type="primary" @click="goToNodeDetail(scope.row)">{{ scope.row.name }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="节点IP" min-width="140"></el-table-column>
          <el-table-column prop="role" label="角色" min-width="100" align="center">
            <template #default="scope">
              <el-tag :type="getRoleType(scope.row.role)" size="small" effect="plain">
                {{ scope.row.role }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" min-width="80" align="center">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'Ready' ? 'success' : 'danger'" size="small">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="cpu_capacity" label="CPU容量" min-width="100" align="center"></el-table-column>
          <el-table-column prop="mem_capacity" label="内存容量" min-width="100" align="center"></el-table-column>
          <el-table-column prop="pod_count" label="Pod数量" min-width="80" align="center"></el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-empty v-else description="请选择集群查看监控信息" />
  </div>
</template>

<script setup>
import { ref, onMounted, onActivated } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '../../utils/request'

const router = useRouter()
const route = useRoute()
const clusters = ref([])
const selectedCluster = ref(null)
const stats = ref(null)

const fetchClusters = async () => {
  try {
    const res = await request.get('/k8s/cluster')
    clusters.value = res.data || []
    if (clusters.value.length > 0) {
      selectedCluster.value = clusters.value[0].id
      fetchStats()
    }
  } catch (error) {
    ElMessage.error('获取集群列表失败')
  }
}

const fetchStats = async () => {
  if (!selectedCluster.value) return
  
  try {
    const res = await request.get(`/k8s/stats/${selectedCluster.value}`)
    stats.value = res.data
  } catch (error) {
    ElMessage.error('获取集群节点数据失败')
  }
}

const getRoleType = (role) => {
  const types = {
    'Master': 'danger',
    'Control-Plane': 'warning',
    'Worker': 'info'
  }
  return types[role] || 'info'
}

const goToNodeDetail = (node) => {
  const clusterAlias = clusters.value.find(c => c.id === selectedCluster.value)?.alias
  localStorage.setItem('k8s_selected_cluster', clusterAlias)
  router.push({
    path: `/k8s/node/${node.name}`,
    query: { cluster: clusterAlias }
  })
}

onMounted(fetchClusters)

onActivated(() => {
  if (selectedCluster.value) {
    fetchStats()
  }
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.stat-card {
  border-radius: 12px;
  padding: 20px;
  text-align: center;
}

.stat-card .stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
}

.stat-card .stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.stat-card-success .stat-value {
  color: #67C23A;
}

.stat-card-warning .stat-value {
  color: #E6A23C;
}

.stat-card-danger .stat-value {
  color: #F56C6C;
}

.card-header-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
</style>