<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ deploymentName }}</h1>
        <el-tag :type="getStatusType(deployment?.status)" size="small" style="margin-left: 12px;">
          {{ deployment?.status }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button @click="refreshAll" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <div class="auto-refresh-wrapper">
          <span class="auto-refresh-label">自动刷新:</span>
          <el-select v-model="autoRefreshInterval" @change="handleAutoRefreshChange" style="width: 100px;" size="default">
            <el-option label="关闭" :value="0"></el-option>
            <el-option label="15秒" :value="15"></el-option>
            <el-option label="30秒" :value="30"></el-option>
            <el-option label="1分钟" :value="60"></el-option>
            <el-option label="5分钟" :value="300"></el-option>
          </el-select>
        </div>
        <el-button type="primary" @click="editYAML">
          <el-icon><Edit /></el-icon>
          编辑YAML
        </el-button>
        <el-button type="warning" @click="restartDeployment">
          <el-icon><RefreshRight /></el-icon>
          重启
        </el-button>
        <el-button type="danger" @click="deleteDeployment">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

<el-tabs v-model="activeTab" class="detail-tabs">
      <!-- 资源状态 -->
      <el-tab-pane label="资源状态" name="status">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="12">
              <el-card class="stat-card replicas-card" shadow="hover">
                <div class="stat-header">
                  <el-icon class="stat-icon"><Monitor /></el-icon>
                  <span class="stat-title">副本数</span>
                </div>
                
                <div class="replicas-container">
                  <div class="replicas-pie">
                    <div class="pie-chart" :style="{ background: getPieGradient() }">
                      <div class="pie-center">
                        <span class="pie-text">{{ deployment?.ready_replicas || 0 }}/{{ deployment?.replicas || 0 }}</span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="replicas-info">
                    <div class="replica-item">
                      <div class="replica-left">
                        <div class="replica-dot expected"></div>
                        <span class="replica-label">期望副本数</span>
                        <span class="replica-value">{{ deployment?.replicas || 0 }}</span>
                      </div>
                      <el-button size="default" class="adjust-btn decrease" @click="adjustReplicas(-1)" :disabled="deployment?.replicas <= 0">
                        <el-icon><Remove /></el-icon>
                      </el-button>
                    </div>
                    <div class="replica-item">
                      <div class="replica-left">
                        <div class="replica-dot current"></div>
                        <span class="replica-label">当前副本数</span>
                        <span class="replica-value ready">{{ deployment?.ready_replicas || 0 }}</span>
                      </div>
                      <el-button size="default" class="adjust-btn increase" @click="adjustReplicas(1)">
                        <el-icon><CirclePlusFilled /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </div>
              </el-card>
            </el-col>
            
            <el-col :span="12">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <span class="card-title">端口</span>
                </template>
                <el-table :data="deployment?.ports || []" style="width: 100%">
                  <el-table-column prop="name" label="名称" min-width="100">
                    <template #default="scope">{{ scope.row.name || '-' }}</template>
                  </el-table-column>
                  <el-table-column prop="protocol" label="协议" width="80" align="center"></el-table-column>
                  <el-table-column prop="port" label="端口" width="100" align="right">
                    <template #default="scope">
                      <span style="color: var(--primary-color); font-weight: 600;">{{ scope.row.port }}</span>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="!deployment?.ports || deployment?.ports.length === 0" description="暂无端口配置" :image-size="60" />
              </el-card>
            </el-col>
          </el-row>

          <el-card class="content-card section-gap" shadow="hover">
            <template #header>
              <span class="card-title">容器组</span>
            </template>
            <el-table :data="pods" style="width: 100%" v-loading="loadingPods">
              <el-table-column prop="name" label="名称" min-width="180">
                <template #default="scope">
                  <el-link type="primary" @click="showPodLogs(scope.row)">{{ scope.row.name }}</el-link>
                </template>
              </el-table-column>
              <el-table-column label="状态" min-width="120" align="center">
                <template #default="scope">
                  <el-tag :type="getPodStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="node_name" label="节点" min-width="100"></el-table-column>
              <el-table-column prop="node_ip" label="节点IP" min-width="100"></el-table-column>
              <el-table-column prop="pod_ip" label="容器组IP" min-width="100"></el-table-column>
              <el-table-column label="重启次数" min-width="80" align="center">
                <template #default="scope">
                  <span :style="{ color: scope.row.restarts > 0 ? '#EF4444' : '' }">{{ scope.row.restarts }}</span>
                </template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="140" align="center">
                <template #default="scope">{{ formatTime(scope.row.created_at) }}</template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>

      <!-- 修改记录 -->
      <el-tab-pane label="修改记录" name="history">
        <div class="history-container">
          <div class="revision-list">
            <div class="revision-header">
              <span class="header-title">版本历史</span>
              <span class="header-count">共 {{ revisions.length }} 个版本</span>
            </div>
            
            <div class="revision-items">
              <div 
                v-for="rev in revisions" 
                :key="rev.revision"
                class="revision-card"
                :class="{ 
                  current: rev.is_current,
                  selected: selectedRev1 === rev.revision || selectedRev2 === rev.revision
                }"
                @click="selectRevision(rev)"
              >
                <div class="revision-main">
                  <div class="revision-badge">
                    <span class="revision-num">v{{ rev.revision }}</span>
                    <el-tag v-if="rev.is_current" type="success" size="small" effect="dark">当前</el-tag>
                  </div>
                  <div class="revision-info">
                    <div class="revision-time">{{ formatTime(rev.created_at) }}</div>
                    <div class="revision-meta">
                      <span class="meta-item">
                        <el-icon><Document /></el-icon>
                        {{ rev.change_reason }}
                      </span>
                      <span class="meta-item">
                        <el-icon><Monitor /></el-icon>
                        {{ rev.replicas }}副本
                      </span>
                    </div>
                  </div>
                </div>
                <div class="revision-image">
                  <el-tag type="info" size="small" effect="plain">{{ truncateImage(rev.image) }}</el-tag>
                </div>
              </div>
              
              <el-empty v-if="revisions.length === 0" description="暂无版本历史" />
            </div>
          </div>
          
          <div class="diff-panel" v-if="selectedRev1 && selectedRev2">
            <div class="diff-header">
              <div class="diff-title">
                <el-tag type="primary" size="small">v{{ selectedRev1 }}</el-tag>
                <span class="diff-arrow">对比</span>
                <el-tag type="primary" size="small">v{{ selectedRev2 }}</el-tag>
              </div>
              <div class="diff-stats">
                <span class="stat-added">+{{ diffResult?.added_count || 0 }}</span>
                <span class="stat-removed">-{{ diffResult?.removed_count || 0 }}</span>
              </div>
            </div>
            
            <div class="diff-content" v-loading="loadingDiff">
              <div 
                v-for="(line, idx) in diffResult?.lines" 
                :key="idx" 
                class="diff-line"
                :class="line.type"
              >
                <span class="line-num">{{ line.new_num || line.old_num || '' }}</span>
                <span class="line-marker">{{ line.marker }}</span>
                <span class="line-text">{{ line.content }}</span>
              </div>
            </div>
            
            <div class="diff-actions">
              <el-button type="warning" @click="rollbackRevision(selectedRev2)" :disabled="isCurrentRevision(selectedRev2)">
                回退到 v{{ selectedRev2 }}
              </el-button>
            </div>
          </div>
          
          <div class="diff-empty" v-else>
            <el-empty description="请选择两个版本进行对比">
              <template #description>
                <p>点击上方版本卡片选择对比版本</p>
              </template>
            </el-empty>
          </div>
        </div>
      </el-tab-pane>

      <!-- 元数据 -->
      <el-tab-pane label="元数据" name="metadata">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="8">
              <el-card class="content-card metadata-card" shadow="hover">
                <template #header>
                  <div class="metadata-header">
                    <el-icon><PriceTag /></el-icon>
                    <span>标签</span>
                    <span class="metadata-count">{{ Object.keys(deployment?.labels || {}).length }}</span>
                  </div>
                </template>
                <div class="labels-list">
                  <div v-for="(value, key) in deployment?.labels" :key="key" class="label-item">
                    <span class="label-key">{{ key }}</span>
                    <span class="label-separator">:</span>
                    <span class="label-value">{{ value }}</span>
                  </div>
                  <el-empty v-if="!deployment?.labels || Object.keys(deployment?.labels).length === 0" description="暂无标签" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :span="16">
              <el-card class="content-card metadata-card" shadow="hover">
                <template #header>
                  <div class="metadata-header">
                    <el-icon><Document /></el-icon>
                    <span>注解</span>
                    <span class="metadata-count">{{ Object.keys(deployment?.annotations || {}).length }}</span>
                  </div>
                </template>
                <el-table :data="getAnnotationsList()" style="width: 100%" size="small">
                  <el-table-column prop="key" label="Key" min-width="280">
                    <template #default="scope">
                      <span class="annotation-key-text">{{ scope.row.key }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="value" label="Value" min-width="400">
                    <template #default="scope">
                      <el-tooltip :content="scope.row.value" placement="top" :disabled="scope.row.value.length < 50">
                        <span class="annotation-value-text">{{ truncateText(scope.row.value, 50) }}</span>
                      </el-tooltip>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="!deployment?.annotations || Object.keys(deployment?.annotations).length === 0" description="暂无注解" :image-size="60" />
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <!-- 环境变量 -->
      <el-tab-pane label="环境变量" name="env">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
           
            <div v-for="container in deployment?.containers" :key="container.name" class="container-env">
              
              <el-table :data="container.env" style="width: 100%" v-if="container.env?.length > 0">
                <el-table-column prop="name" label="名称" min-width="200"></el-table-column>
                <el-table-column prop="value" label="值" min-width="300">
                  <template #default="scope">
                    <span v-if="scope.row.value">{{ scope.row.value }}</span>
                    <el-tag v-else-if="scope.row.valueFrom" size="small" type="info">
                      {{ getValueFromType(scope.row.valueFrom) }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="无环境变量" />
            </div>
          </el-card>
        </div>
      </el-tab-pane>

      <!-- 事件 -->
      <el-tab-pane label="事件" name="events">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
            <el-table :data="events" style="width: 100%" v-loading="loadingEvents">
              <el-table-column label="类型" min-width="80">
                <template #default="scope">
                  <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'warning'" size="small">
                    {{ scope.row.type }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" min-width="120"></el-table-column>
              <el-table-column prop="time" label="发生时间" min-width="140" align="center">
                <template #default="scope">{{ formatTime(scope.row.time) }}</template>
              </el-table-column>
              <el-table-column prop="source" label="来源" min-width="120"></el-table-column>
              <el-table-column prop="message" label="消息" min-width="200"></el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'

import { ArrowLeft, Monitor, Remove, CirclePlusFilled, Document, PriceTag, Edit, RefreshRight, Delete, Refresh } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const clusterAlias = computed(() => localStorage.getItem('k8s_selected_cluster') || route.query.cluster)
const namespace = computed(() => localStorage.getItem('k8s_selected_namespace') || route.query.namespace || 'default')
const deploymentName = computed(() => route.params.name)

const activeTab = ref('status')
const loading = ref(false)
const loadingPods = ref(false)
const loadingHistories = ref(false)
const loadingEvents = ref(false)
const autoRefreshInterval = ref(15)
let autoRefreshTimer = null

const deployment = ref({})
const pods = ref([])
const events = ref([])

const revisions = ref([])
const selectedRev1 = ref(null)
const selectedRev2 = ref(null)
const diffResult = ref(null)
const loadingDiff = ref(false)

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const truncateImage = (image) => {
  if (!image) return ''
  if (image.length > 30) {
    return image.substring(0, 30) + '...'
  }
  return image
}

const truncateText = (text, maxLen) => {
  if (!text) return ''
  if (text.length > maxLen) {
    return text.substring(0, maxLen) + '...'
  }
  return text
}

const getAnnotationsList = () => {
  const annotations = deployment.value?.annotations || {}
  return Object.entries(annotations).map(([key, value]) => ({ key, value }))
}

const getStatusType = (status) => {
  const map = { Running: 'success', Updating: 'warning', Starting: 'info', Failed: 'danger' }
  return map[status] || 'info'
}

const getPodStatusType = (status) => {
  const map = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info',
    Terminating: 'warning',
    ContainersNotReady: 'warning',
    CrashLoopBackOff: 'danger',
    ImagePullBackOff: 'danger',
    ErrImagePull: 'danger',
    ContainerCreating: 'info',
    Unschedulable: 'danger',
    NotReady: 'warning'
  }
  return map[status] || 'info'
}

const getChangeType = (type) => {
  const map = { create: 'success', update: 'primary', scale: 'warning', rollback: 'info' }
  return map[type] || 'info'
}

const getVersionTagType = (row) => {
  const rev = revisions.value.find(r => r.revision === row.revision)
  if (rev?.is_current) return 'success'
  return 'info'
}

const getValueFromType = (valueFrom) => {
  if (valueFrom.fieldRef) return 'FieldRef: ' + valueFrom.fieldRef.fieldPath
  if (valueFrom.secretKeyRef) return 'Secret: ' + valueFrom.secretKeyRef.name
  if (valueFrom.configMapKeyRef) return 'ConfigMap: ' + valueFrom.configMapKeyRef.name
  return 'Unknown'
}

const getReplicaProgress = () => {
  if (!deployment.value?.replicas || deployment.value.replicas === 0) return 0
  const ready = deployment.value.ready_replicas || 0
  const total = deployment.value.replicas
  return Math.round((ready / total) * 100)
}

const getProgressClass = () => {
  const progress = getReplicaProgress()
  if (progress >= 100) return 'success'
  if (progress >= 50) return 'warning'
  return 'danger'
}

const getPieGradient = () => {
  const progress = getReplicaProgress()
  const readyColor = '#10B981'
  const pendingColor = '#E2E8F0'
  return `conic-gradient(${readyColor} 0% ${progress}%, ${pendingColor} ${progress}% 100%)`
}

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/k8s/deployment/${deploymentName.value}/detail`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    deployment.value = res.data || {}
  } catch (error) {
    ElMessage.error('获取详情失败')
  } finally {
    loading.value = false
  }
}

const fetchPods = async () => {
  loadingPods.value = true
  try {
    const res = await request.get(`/k8s/deployment/${deploymentName.value}/pods`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    pods.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Pod列表失败')
  } finally {
    loadingPods.value = false
  }
}

const fetchRevisions = async () => {
  loadingHistories.value = true
  try {
    const res = await request.get(`/k8s/deployment/${deploymentName.value}/revisions`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    revisions.value = res.data || []
    
    if (revisions.value.length >= 2) {
      selectedRev1.value = revisions.value[0].revision
      selectedRev2.value = revisions.value[1].revision
      loadDiff()
    }
  } catch (error) {
    ElMessage.error('获取版本历史失败')
  } finally {
    loadingHistories.value = false
  }
}

const selectRevision = (rev) => {
  if (!selectedRev1.value) {
    selectedRev1.value = rev.revision
  } else if (!selectedRev2.value && rev.revision !== selectedRev1.value) {
    selectedRev2.value = rev.revision
    loadDiff()
  } else {
    selectedRev1.value = rev.revision
    selectedRev2.value = null
    diffResult.value = null
  }
}

const loadDiff = async () => {
  if (!selectedRev1.value || !selectedRev2.value) return
  
  loadingDiff.value = true
  try {
    const res = await request.get(`/k8s/deployment/${deploymentName.value}/diff`, {
      params: {
        cluster: clusterAlias.value,
        namespace: namespace.value,
        revision1: selectedRev1.value,
        revision2: selectedRev2.value
      }
    })
    diffResult.value = res.data
  } catch (error) {
    ElMessage.error('加载对比失败')
  } finally {
    loadingDiff.value = false
  }
}

const isCurrentRevision = (revision) => {
  const rev = revisions.value.find(r => r.revision === revision)
  return rev?.is_current
}

const rollbackRevision = async (revision) => {
  try {
    await ElMessageBox.confirm(`确认回退到版本 v${revision}?`, '回退确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.post(`/k8s/deployment/${deploymentName.value}/rollback/${revision}`, null, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    
    ElMessage.success('回退成功')
    fetchDetail()
    fetchPods()
    fetchRevisions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '回退失败')
    }
  }
}

const fetchEvents = async () => {
  loadingEvents.value = true
  try {
    const res = await request.get(`/k8s/deployment/${deploymentName.value}/events`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    events.value = res.data || []
  } catch (error) {
    ElMessage.error('获取事件失败')
  } finally {
    loadingEvents.value = false
  }
}

const refreshAll = async () => {
  await Promise.all([
    fetchDetail(),
    fetchPods(),
    fetchRevisions(),
    fetchEvents()
  ])
}

const startAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
  if (autoRefreshInterval.value > 0) {
    autoRefreshTimer = setInterval(() => {
      refreshAll()
    }, autoRefreshInterval.value * 1000)
  }
}

const handleAutoRefreshChange = () => {
  startAutoRefresh()
  if (autoRefreshInterval.value === 0) {
    ElMessage.info('自动刷新已关闭')
  } else {
    ElMessage.success(`自动刷新已设置为 ${autoRefreshInterval.value}秒`)
  }
}

const adjustReplicas = async (delta) => {
  const newReplicas = (deployment.value.replicas || 0) + delta
  if (newReplicas < 0) return
  
  try {
    await ElMessageBox.confirm(`确认将副本数调整为 ${newReplicas}?`, '调整副本数', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.post(`/k8s/deployment/${deploymentName.value}/scale-with-history`, { replicas: newReplicas }, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    
    ElMessage.success('副本数调整成功')
    fetchDetail()
    fetchRevisions()
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '调整失败')
    }
  }
}

const editYAML = () => {
  router.push({
    path: `/k8s/deployment/${deploymentName.value}/edit`,
    query: { cluster: clusterAlias.value, namespace: namespace.value }
  })
}

const restartDeployment = async () => {
  try {
    await ElMessageBox.confirm('确认重启Deployment? 这将重启所有Pod。', '重启确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.post(`/k8s/deployment/${deploymentName.value}/restart`, null, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    
    ElMessage.success('重启成功')
    fetchDetail()
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '重启失败')
    }
  }
}

const deleteDeployment = async () => {
  try {
    await ElMessageBox.confirm(`确认删除Deployment "${deploymentName.value}"? 此操作不可恢复。`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger'
    })
    
    await request.delete(`/k8s/deployment/${deploymentName.value}`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    
    ElMessage.success('删除成功')
    router.push({
      path: '/k8s/deployments',
      query: { cluster: clusterAlias.value, namespace: namespace.value }
    })
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '删除失败')
    }
  }
}

const showPodLogs = (pod) => {
  router.push({
    path: `/k8s/pod/${pod.name}`,
    query: { cluster: clusterAlias.value, namespace: namespace.value }
  })
}

watch([selectedRev1, selectedRev2], () => {
  if (selectedRev1.value && selectedRev2.value) {
    loadDiff()
  }
})

onMounted(() => {
  fetchDetail()
  fetchPods()
  fetchRevisions()
  fetchEvents()
  startAutoRefresh()
})

onBeforeUnmount(() => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auto-refresh-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.auto-refresh-label {
  font-size: 14px;
  color: #606266;
}

.detail-tabs {
  margin-top: 16px;
}

.tab-content {
  padding: 16px 0;
}

.stat-card {
  text-align: center;
  padding: 20px;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.stat-icon {
  font-size: 20px;
  color: var(--primary-color);
}

.stat-title {
  font-size: 14px;
  color: var(--text-secondary);
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.stat-value .ready {
  color: var(--success-color);
}

.stat-value .separator {
  color: var(--text-muted);
  margin: 0 4px;
}

.stat-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.replicas-card {
  text-align: left;
}

.replicas-container {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 16px;
}

.replicas-pie {
  flex-shrink: 0;
}

.pie-chart {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.pie-center {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.pie-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.replicas-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.replica-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.replica-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.replica-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.replica-dot.expected {
  background: #E2E8F0;
}

.replica-dot.current {
  background: #10B981;
}

.replica-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.replica-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.replica-value.ready {
  color: var(--success-color);
}

.adjust-btn {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-md);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s ease;
}

.adjust-btn.decrease {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.adjust-btn.decrease:hover:not(:disabled) {
  background: var(--border-light);
  border-color: #EF4444;
  color: #EF4444;
}

.adjust-btn.decrease:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.adjust-btn.increase {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.adjust-btn.increase:hover {
  background: var(--border-light);
  border-color: #10B981;
  color: #10B981;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
}

.metadata-card {
  height: 100%;
}

.metadata-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
}

.metadata-header .el-icon {
  color: var(--primary-color);
}

.metadata-count {
  margin-left: auto;
  padding: 2px 8px;
  background: var(--primary-lighter);
  border-radius: 10px;
  font-size: 12px;
  color: var(--primary-color);
}

.labels-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.label-item {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
  transition: all 0.2s ease;
}

.label-item:hover {
  background: var(--primary-lighter);
}

.label-key {
  padding: 2px 6px;
  background: #e4e7ed;
  border-radius: 3px;
  color: #606266;
  font-weight: 500;
}

.label-separator {
  margin: 0 4px;
  color: #909399;
}

.label-value {
  color: #303133;
}

.annotation-key-text {
  color: #606266;
  font-weight: 500;
}

.annotation-value-text {
  color: #303133;
  word-break: break-all;
}

.container-env {
  margin-bottom: 24px;
}

.container-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--border-light);
  border-radius: var(--radius-md);
}

.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.version-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}

.vs-text {
  color: var(--text-secondary);
  font-size: 13px;
}

.history-table {
  margin-bottom: 0;
}

.diff-stats {
  display: flex;
  gap: 16px;
}

.stat-added {
  color: #10B981;
  font-size: 13px;
  font-weight: 500;
}

.stat-removed {
  color: #EF4444;
  font-size: 13px;
  font-weight: 500;
}

.diff-view {
  font-family: 'JetBrains Mono', 'SF Mono', Monaco, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1a1b26;
  border-radius: var(--radius-md);
  overflow: auto;
  max-height: 500px;
}

.diff-line {
  display: flex;
  align-items: flex-start;
  padding: 2px 8px;
  white-space: pre;
}

.diff-line.normal {
  background: transparent;
}

.diff-line.added {
  background: rgba(16, 185, 129, 0.15);
}

.diff-line.removed {
  background: rgba(239, 68, 68, 0.15);
}

.diff-num {
  min-width: 40px;
  padding: 0 8px;
  text-align: right;
  color: #6366F1;
  font-size: 12px;
  user-select: none;
}

.diff-num.old {
  border-right: 1px solid #3B4872;
}

.diff-num.new {
  border-right: 1px solid #3B4872;
}

.diff-marker {
  min-width: 20px;
  text-align: center;
  font-weight: bold;
  user-select: none;
}

.diff-line.added .diff-marker {
  color: #10B981;
}

.diff-line.removed .diff-marker {
  color: #EF4444;
}

.diff-line.normal .diff-marker {
  color: #565F89;
}

.diff-content {
  flex: 1;
  padding-left: 8px;
  color: #A9B1D6;
}

.diff-line.added .diff-content {
  color: #73DACA;
}

.diff-line.removed .diff-content {
  color: #F7768E;
}

.history-container {
  display: flex;
  gap: 24px;
  padding: 16px 0;
}

.revision-list {
  flex: 0 0 320px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.revision-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--primary-lighter);
  border-radius: var(--radius-md);
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-count {
  font-size: 12px;
  color: var(--text-secondary);
}

.revision-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 600px;
  overflow-y: auto;
}

.revision-card {
  padding: 16px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.revision-card:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.revision-card.current {
  background: rgba(16, 185, 129, 0.05);
  border-color: #10B981;
}

.revision-card.selected {
  background: rgba(59, 130, 246, 0.05);
  border-color: #3B82F6;
}

.revision-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.revision-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.revision-num {
  font-size: 16px;
  font-weight: 600;
  color: var(--primary-color);
}

.revision-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.revision-time {
  font-size: 12px;
  color: var(--text-secondary);
}

.revision-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-muted);
}

.revision-image {
  margin-top: 8px;
}

.diff-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--border-light);
  border-bottom: 1px solid var(--border-color);
}

.diff-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-arrow {
  font-size: 12px;
  color: var(--text-secondary);
}

.diff-stats {
  display: flex;
  gap: 12px;
}

.diff-panel .diff-content {
  flex: 1;
  padding: 16px;
  font-family: 'JetBrains Mono', 'SF Mono', Monaco, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1a1b26;
  overflow: auto;
  max-height: 500px;
}

.diff-panel .diff-line {
  display: flex;
  align-items: flex-start;
  padding: 2px 8px;
  white-space: pre;
}

.diff-panel .line-num {
  min-width: 40px;
  padding: 0 8px;
  text-align: right;
  color: #6366F1;
  font-size: 12px;
  user-select: none;
  border-right: 1px solid #3B4872;
}

.diff-panel .line-marker {
  min-width: 20px;
  text-align: center;
  font-weight: bold;
  user-select: none;
}

.diff-panel .line-text {
  flex: 1;
  padding-left: 8px;
  color: #A9B1D6;
}

.diff-panel .diff-line.added {
  background: rgba(16, 185, 129, 0.15);
}

.diff-panel .diff-line.removed {
  background: rgba(239, 68, 68, 0.15);
}

.diff-panel .diff-line.added .line-marker {
  color: #10B981;
}

.diff-panel .diff-line.removed .line-marker {
  color: #EF4444;
}

.diff-panel .diff-line.added .line-text {
  color: #73DACA;
}

.diff-panel .diff-line.removed .line-text {
  color: #F7768E;
}

.diff-actions {
  padding: 12px 16px;
  background: var(--border-light);
  border-top: 1px solid var(--border-color);
}

.diff-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}
</style>