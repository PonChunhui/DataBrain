<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ cluster?.name }}</h1>
        <el-tag :type="cluster?.status ? 'success' : 'danger'" size="small" style="margin-left: 12px;">
          {{ cluster?.status ? '启用' : '禁用' }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button type="success" @click="showInspection">
          <el-icon><Monitor /></el-icon>
          智能巡检
        </el-button>
        <el-button type="primary" @click="editCluster">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="基本信息" name="info">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="12">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><InfoFilled /></el-icon>
                    <span class="card-title">集群信息</span>
                  </div>
                </template>
                <el-descriptions :column="1" border size="small">
                  <el-descriptions-item label="集群标识">
                    <el-tag type="info" size="small">{{ cluster?.alias }}</el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="API Server">
                    <span class="mono-text">{{ clusterInfo?.server || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="Context">
                    <span class="mono-text">{{ clusterInfo?.context || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="默认命名空间">{{ cluster?.namespace || 'default' }}</el-descriptions-item>
                  <el-descriptions-item label="描述">{{ cluster?.description || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="创建时间">{{ formatTime(cluster?.created_at) }}</el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>

            <el-col :span="12">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Monitor /></el-icon>
                    <span class="card-title">Prometheus配置</span>
                  </div>
                </template>
                <el-descriptions :column="1" border size="small">
                  <el-descriptions-item label="Prometheus地址">
                    <span class="mono-text">{{ cluster?.prometheus_url || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="认证状态">
                    <el-tag :type="cluster?.prometheus_auth_enabled ? 'success' : 'info'" size="small">
                      {{ cluster?.prometheus_auth_enabled ? '已启用' : '未启用' }}
                    </el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="认证类型" v-if="cluster?.prometheus_auth_enabled">
                    {{ cluster?.prometheus_auth_type || '-' }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="资源统计" name="stats">
        <div class="tab-content">
          <el-row :gutter="24" class="section-gap">
            <el-col :span="4">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-value">{{ stats?.namespace_count || 0 }}</div>
                <div class="stat-label">命名空间</div>
              </el-card>
            </el-col>
            <el-col :span="4">
              <el-card class="stat-card stat-card-success" shadow="hover">
                <div class="stat-value">{{ stats?.running_pods || 0 }}</div>
                <div class="stat-label">运行中Pod</div>
              </el-card>
            </el-col>
            <el-col :span="4">
              <el-card class="stat-card stat-card-warning" shadow="hover">
                <div class="stat-value">{{ stats?.pending_pods || 0 }}</div>
                <div class="stat-label">等待中Pod</div>
              </el-card>
            </el-col>
            <el-col :span="4">
              <el-card class="stat-card stat-card-danger" shadow="hover">
                <div class="stat-value">{{ stats?.failed_pods || 0 }}</div>
                <div class="stat-label">失败Pod</div>
              </el-card>
            </el-col>
            <el-col :span="4">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-value">{{ stats?.pod_count || 0 }}</div>
                <div class="stat-label">Pod总数</div>
              </el-card>
            </el-col>
            <el-col :span="4">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-value">{{ stats?.nodes?.length || 0 }}</div>
                <div class="stat-label">节点数</div>
              </el-card>
            </el-col>
          </el-row>

          <el-card class="content-card section-gap" shadow="hover">
            <template #header>
              <div class="card-header-title">节点列表</div>
            </template>
            <el-table :data="stats?.nodes || []" style="width: 100%">
              <el-table-column prop="name" label="节点名称" min-width="180"></el-table-column>
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
      </el-tab-pane>
    </el-tabs>

    <AIOPSStreamDialog
      v-model="inspectionVisible"
      title="集群智能巡检"
      assistant-name="巡检助手"
      config-title="巡检配置"
      loading-text="正在巡检分析中..."
      :start-message="inspectionStartMessage"
      stream-endpoint="/aiops/inspection/stream"
      chat-endpoint="/aiops/inspection/chat"
      :stream-params="inspectionStreamParams"
      :config-items="inspectionConfigItems"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Edit, InfoFilled, Monitor } from '@element-plus/icons-vue'
import request from '../../utils/request'
import AIOPSStreamDialog from '../../components/aiops/AIOPSStreamDialog.vue'

const router = useRouter()
const route = useRoute()

const activeTab = ref('info')
const cluster = ref(null)
const clusterInfo = ref(null)
const stats = ref(null)
const loading = ref(false)
const inspectionVisible = ref(false)

const clusterAlias = computed(() => route.params.alias)

const inspectionConfigItems = computed(() => [
  { label: '集群', value: cluster?.value?.alias || cluster?.value?.name || '未知', type: 'success' },
  { label: '节点数', value: stats?.value?.nodes?.length || 0, type: 'info' },
  { label: '命名空间', value: stats?.value?.namespace_count || 0, type: 'warning' },
  { label: 'Pod总数', value: stats?.value?.pod_count || 0, type: '' },
])

const inspectionStreamParams = computed(() => ({
  cluster_id: cluster?.value?.id || 0
}))

const inspectionStartMessage = computed(() => 
  `请对集群 ${cluster?.value?.alias || cluster?.value?.name || '未知集群'} 进行全面巡检分析`
)

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getRoleType = (role) => {
  const types = {
    'Master': 'danger',
    'Control-Plane': 'warning',
    'Worker': 'info'
  }
  return types[role] || 'info'
}

const fetchClusterDetail = async () => {
  loading.value = true
  try {
    const res = await request.get('/k8s/cluster', {
      params: { alias: clusterAlias.value }
    })
    if (res.data && res.data.length > 0) {
      cluster.value = res.data[0]
      if (cluster.value?.id) {
        fetchClusterInfo()
        fetchStats()
      }
    }
  } catch (error) {
    ElMessage.error('获取集群详情失败')
  } finally {
    loading.value = false
  }
}

const fetchClusterInfo = async () => {
  try {
    const res = await request.get(`/k8s/cluster/${cluster.value.id}/info`)
    clusterInfo.value = res.data
  } catch (error) {
    console.error('获取集群信息失败', error)
  }
}

const fetchStats = async () => {
  try {
    const res = await request.get(`/k8s/stats/${cluster.value.id}`)
    stats.value = res.data
  } catch (error) {
    console.error('获取集群统计失败', error)
  }
}

const editCluster = () => {
  router.push({
    path: '/k8s/clusters',
    query: { edit: cluster.value.id }
  })
}

const showInspection = () => {
  if (!cluster.value?.id) {
    ElMessage.warning('请先获取集群信息')
    return
  }
  inspectionVisible.value = true
}

onMounted(fetchClusterDetail)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  gap: 12px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.detail-tabs {
  margin-top: 16px;
}

.tab-content {
  padding: 16px 0;
}

.stat-card {
  border-radius: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 18px;
  color: #409EFF;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.mono-text {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  color: #606266;
}

.stat-card .stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
  text-align: center;
}

.stat-card .stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
  text-align: center;
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

.section-gap {
  margin-top: 24px;
}
</style>