<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Pod管理</h1>
      <div class="header-right">
        <K8sResourceFilter 
          ref="resourceFilter"
          @query="handleQuery"
          @cluster-change="handleClusterChange"
          @ready="handleReady"
        />
      </div>
    </div>

    <el-card class="content-card section-gap" shadow="hover">
      <el-table :data="pods" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="240">
          <template #default="scope">
            <el-button type="primary" link @click="goToDetail(scope.row)">{{ scope.row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="重启" width="70" align="center">
          <template #default="scope">
            <span :style="{ color: scope.row.restarts > 0 ? '#EF4444' : '' }">{{ scope.row.restarts }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="pod_ip" label="IP地址" width="140"></el-table-column>
        <el-table-column prop="node_name" label="节点" width="160"></el-table-column>
        <el-table-column label="容器" width="100" align="center">
          <template #default="scope">
            <span>{{ scope.row.containers?.length || 0 }}个</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="showLogs(scope.row)">日志</el-button>
            <el-button type="success" size="small" link @click="showEvents(scope.row)">事件</el-button>
            <el-button type="warning" size="small" link @click="showTerminal(scope.row)">终端</el-button>
            <el-button type="danger" size="small" link @click="deletePod(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog 
      v-model="logsDialogVisible" 
      width="1000px"
      :fullscreen="logsFullscreen"
      destroy-on-close
      top="2vh"
      class="logs-dialog"
      :show-close="false"
    >
      <PodLogTerminal
        v-if="logsDialogVisible && selectedContainer"
        :clusterAlias="searchParams.cluster"
        :namespace="searchParams.namespace"
        :podName="currentPod.name"
        :container="selectedContainer"
        :tailLines="tailLines"
        @fullscreen-change="handleLogFullscreenChange"
        @close="logsDialogVisible = false"
      />
    </el-dialog>

    <el-dialog v-model="eventsDialogVisible" title="Pod事件" width="900px">
      <el-table :data="podEvents" style="width: 100%">
        <el-table-column prop="type" label="类型" width="80" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'danger'" size="small">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" width="120"></el-table-column>
        <el-table-column prop="message" label="消息" min-width="300"></el-table-column>
        <el-table-column prop="count" label="次数" width="60" align="center"></el-table-column>
        <el-table-column label="时间" width="140" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.eventTime || scope.row.lastTimestamp) }}
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog 
      v-model="terminalDialogVisible" 
      width="1000px"
      :fullscreen="terminalFullscreen"
      destroy-on-close
      top="2vh"
      class="terminal-dialog"
      :show-close="false"
    >
      <PodExecTerminal
        v-if="terminalDialogVisible && terminalContainer"
        :clusterAlias="searchParams.cluster"
        :namespace="searchParams.namespace"
        :podName="currentPod.name"
        :container="terminalContainer"
        :shell="terminalShell"
        :containers="currentPod.containers || []"
        @fullscreen-change="handleTerminalFullscreenChange"
        @close="terminalDialogVisible = false"
        @container-change="terminalContainer = $event"
        @shell-change="terminalShell = $event"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, FullScreen, Close, Monitor } from '@element-plus/icons-vue'
import request from '../../utils/request'
import PodLogTerminal from '../../components/k8s/PodLogTerminal.vue'
import PodExecTerminal from '../../components/k8s/PodExecTerminal.vue'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const router = useRouter()
const resourceFilter = ref(null)
const searchParams = ref({ cluster: null, namespace: 'default' })
const pods = ref([])
const loading = ref(false)

const logsDialogVisible = ref(false)
const eventsDialogVisible = ref(false)
const terminalDialogVisible = ref(false)
const logsFullscreen = ref(false)
const terminalFullscreen = ref(false)
const currentPod = ref({})
const selectedContainer = ref('')
const terminalContainer = ref('')
const terminalShell = ref('/bin/sh')
const tailLines = ref(500)
const podEvents = ref([])

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getStatusType = (status) => {
  const types = {
    'Running': 'success',
    'Pending': 'warning',
    'Succeeded': 'info',
    'Failed': 'danger',
    'Unknown': 'info'
  }
  return types[status] || 'info'
}

const handleQuery = (params) => {
  searchParams.value = params
  fetchPods()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchPods()
}

const handleClusterChange = () => {
  pods.value = []
}

const fetchPods = async () => {
  if (!searchParams.value.cluster) return
  loading.value = true
  try {
    const res = await request.get('/k8s/pod', {
      params: { cluster: searchParams.value.cluster, namespace: searchParams.value.namespace }
    })
    pods.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Pod列表失败')
  } finally {
    loading.value = false
  }
}

const showLogs = (pod) => {
  currentPod.value = pod
  logsFullscreen.value = false
  if (pod.containers && pod.containers.length > 0) {
    selectedContainer.value = pod.containers[0].name
  }
  logsDialogVisible.value = true
}

const handleLogFullscreenChange = (fullscreen) => {
  logsFullscreen.value = fullscreen
}

const onContainerChange = () => {
}

const showEvents = async (pod) => {
  currentPod.value = pod
  eventsDialogVisible.value = true
  
  try {
    const res = await request.get(`/k8s/pod/${pod.name}/events`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    podEvents.value = res.data || []
  } catch (error) {
    ElMessage.error('获取事件失败')
    podEvents.value = []
  }
}

const showTerminal = (pod) => {
  currentPod.value = pod
  terminalFullscreen.value = false
  if (pod.containers && pod.containers.length > 0) {
    terminalContainer.value = pod.containers[0].name
  }
  terminalDialogVisible.value = true
}

const handleTerminalFullscreenChange = (fullscreen) => {
  terminalFullscreen.value = fullscreen
}

const openTerminal = () => {
  ElMessage.info('终端功能需要WebSocket支持，请使用kubectl exec命令')
}

const deletePod = async (pod) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod ${pod.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/pod/${pod.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('删除成功')
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const goToDetail = (pod) => {
  router.push({
    path: `/k8s/pod/${pod.name}`,
    query: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.terminal-placeholder {
  background: #1e1e1e;
  color: #cccccc;
  padding: 20px;
  border-radius: 8px;
  font-family: monospace;
  min-height: 300px;
}

.terminal-placeholder p {
  margin: 8px 0;
}
</style>