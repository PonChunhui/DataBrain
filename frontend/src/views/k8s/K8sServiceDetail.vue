<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ serviceName }}</h1>
        <el-tag :type="getServiceTypeTag(service?.type)" size="small" style="margin-left: 12px;">
          {{ service?.type }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showYamlDialog">
          <el-icon><Edit /></el-icon>
          编辑YAML
        </el-button>
        <el-button type="danger" @click="deleteService">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="资源状态" name="status">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="12">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Connection /></el-icon>
                    <span class="card-title">基本信息</span>
                  </div>
                </template>
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="服务类型">
                    <el-tag :type="getServiceTypeTag(service?.type)" size="small">{{ service?.type }}</el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="ClusterIP">
                    <span class="mono-text">{{ service?.cluster_ip || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="LoadBalancer IP" v-if="service?.type === 'LoadBalancer'">
                    <span class="mono-text" v-if="service?.external_ip && service?.external_ip.length > 0">{{ service?.external_ip.join(', ') }}</span>
                    <span v-else style="color: #909399;">-</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="Session Affinity">
                    {{ service?.session_affinity || 'None' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="创建时间">
                    {{ formatTime(service?.created_at) }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>

            <el-col :span="12">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Share /></el-icon>
                    <span class="card-title">端口映射</span>
                  </div>
                </template>
                <el-table :data="service?.ports || []" style="width: 100%">
                  <el-table-column prop="name" label="名称" min-width="100">
                    <template #default="scope">{{ scope.row.name || '-' }}</template>
                  </el-table-column>
                  <el-table-column prop="protocol" label="协议" width="80" align="center">
                    <template #default="scope">
                      <el-tag type="info" size="small">{{ scope.row.protocol }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="端口映射" min-width="150">
                    <template #default="scope">
                      <span class="port-mapping">
                        <span class="port-label">服务端口:</span>
                        <span class="port-value">{{ scope.row.port }}</span>
                        <el-icon class="arrow-icon"><Right /></el-icon>
                        <span class="port-label">目标端口:</span>
                        <span class="port-value target">{{ scope.row.target_port }}</span>
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="node_port" label="NodePort" width="100" align="right" v-if="service?.type === 'NodePort'">
                    <template #default="scope">
                      <span class="node-port" v-if="scope.row.node_port">{{ scope.row.node_port }}</span>
                      <span v-else style="color: #909399;">-</span>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="!service?.ports || service?.ports.length === 0" description="暂无端口配置" :image-size="60" />
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="24" style="margin-top: 24px;" v-if="deployments.length > 0">
            <el-col :span="24">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Collection /></el-icon>
                    <span class="card-title">关联的工作负载（Deployment）</span>
                    <el-badge :value="deployments.length" type="primary" style="margin-left: 12px;" />
                  </div>
                </template>
                <div class="selector-info" v-if="service?.selector && Object.keys(service?.selector).length > 0" style="margin-bottom: 16px;">
                  <span class="selector-label">Selector匹配规则：</span>
                  <el-tag v-for="(value, key) in service?.selector" :key="key" type="success" size="default" class="selector-tag">
                    <span class="selector-key">{{ key }}</span>=<span class="selector-value">{{ value }}</span>
                  </el-tag>
                </div>
                <el-table :data="deployments" style="width: 100%" v-loading="deploymentsLoading">
                  <el-table-column prop="name" label="名称" min-width="180">
                    <template #default="scope">
                      <el-button type="primary" link @click="goToDeploymentDetail(scope.row)">{{ scope.row.name }}</el-button>
                    </template>
                  </el-table-column>
                  <el-table-column prop="status" label="状态" width="100" align="center">
                    <template #default="scope">
                      <el-tag :type="getDeploymentStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="副本数" width="120" align="center">
                    <template #default="scope">
                      <span class="replicas-text">{{ scope.row.ready_replicas }}/{{ scope.row.replicas }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="images" label="镜像" min-width="200">
                    <template #default="scope">
                      <el-tag v-for="(img, idx) in scope.row.images" :key="idx" type="info" size="small" style="margin-right: 4px;">
                        {{ img }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="created_at" label="创建时间" width="180" align="center">
                    <template #default="scope">
                      {{ formatTime(scope.row.created_at) }}
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>

          <el-card class="content-card section-gap" shadow="hover" v-if="pods.length > 0">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Box /></el-icon>
                <span class="card-title">容器组（Pods）</span>
                <el-badge :value="pods.length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="pods" style="width: 100%" v-loading="podsLoading">
              <el-table-column prop="name" label="名称" min-width="180">
                <template #default="scope">
                  <el-button type="primary" link @click="goToPodDetail(scope.row)">{{ scope.row.name }}</el-button>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100" align="center">
                <template #default="scope">
                  <el-tag :type="getPodStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="pod_ip" label="Pod IP" width="140">
                <template #default="scope">
                  <span class="mono-text">{{ scope.row.pod_ip }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="node_name" label="节点" min-width="140"></el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="180" align="center">
                <template #default="scope">
                  {{ formatTime(scope.row.created_at) }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="元数据" name="metadata">
        <div class="tab-content">
          <el-card class="metadata-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Document /></el-icon>
                <span class="card-title">基本信息</span>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ service?.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ service?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="服务类型">{{ service?.type }}</el-descriptions-item>
              <el-descriptions-item label="ClusterIP">{{ service?.cluster_ip }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(service?.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="Session Affinity">{{ service?.session_affinity || 'None' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="metadata-card" shadow="hover" style="margin-top: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><PriceTag /></el-icon>
                <span class="card-title">标签（Labels）</span>
                <el-badge :value="Object.keys(service?.labels || {}).length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <div class="labels-container" v-if="service?.labels && Object.keys(service?.labels).length > 0">
              <el-tag v-for="(value, key) in service?.labels" :key="key" type="primary" size="default" class="label-tag">
                <span class="label-key">{{ key }}</span>=<span class="label-value">{{ value }}</span>
              </el-tag>
            </div>
            <el-empty v-else description="暂无标签" :image-size="60" />
          </el-card>

          <el-card class="metadata-card" shadow="hover" style="margin-top: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><EditPen /></el-icon>
                <span class="card-title">注解（Annotations）</span>
                <el-badge :value="Object.keys(service?.annotations || {}).length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <div class="annotations-container" v-if="service?.annotations && Object.keys(service?.annotations).length > 0">
              <div v-for="(value, key) in service?.annotations" :key="key" class="annotation-item">
                <div class="annotation-key">{{ key }}</div>
                <div class="annotation-value">{{ value || '-' }}</div>
              </div>
            </div>
            <el-empty v-else description="暂无注解" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="事件" name="events">
        <div class="tab-content">
          <el-card class="events-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Bell /></el-icon>
                <span class="card-title">最近事件</span>
                <el-badge :value="events.length" type="warning" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="events" style="width: 100%" v-loading="eventsLoading">
              <el-table-column prop="type" label="类型" width="100" align="center">
                <template #default="scope">
                  <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'danger'" size="small">
                    {{ scope.row.type }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" width="140">
                <template #default="scope">
                  <el-tag type="info" size="small">{{ scope.row.reason }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="message" label="消息" min-width="300"></el-table-column>
              <el-table-column prop="count" label="次数" width="80" align="center">
                <template #default="scope">
                  <el-badge :value="scope.row.count" type="primary" />
                </template>
              </el-table-column>
              <el-table-column prop="timestamp" label="时间" width="180" align="center">
                <template #default="scope">
                  {{ formatTime(scope.row.timestamp) }}
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="events.length === 0 && !eventsLoading" description="暂无事件记录" :image-size="80" />
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="yamlDialogVisible" title="编辑 Service YAML" direction="rtl" size="65%" destroy-on-close>
      <el-card class="yaml-card" shadow="hover" v-loading="yamlLoading">
        <template #header>
          <div class="yaml-header">
            <div class="yaml-title-info">
              <el-icon class="yaml-icon"><Document /></el-icon>
              <span class="yaml-title">YAML 配置</span>
              <el-tag size="small" type="info" style="margin-left: 8px">{{ serviceName }}</el-tag>
            </div>
            <div class="yaml-status">
              <el-tag v-if="yamlError" type="danger" size="small">
                <el-icon><Warning /></el-icon>
                {{ yamlError }}
              </el-tag>
              <el-tag v-else-if="yamlContent" type="success" size="small">
                <el-icon><SuccessFilled /></el-icon>
                YAML 格式正确
              </el-tag>
            </div>
          </div>
        </template>
        <div class="editor-wrapper">
          <codemirror
            v-model="yamlContent"
            :style="{ height: '100%' }"
            :extensions="yamlExtensions"
            @change="validateYAML"
            class="yaml-codemirror"
          />
        </div>
        <div class="editor-footer">
          <div class="footer-stats">
            <span class="stat-item">
              <el-icon><Document /></el-icon>
              行数: {{ lineCount }}
            </span>
            <span class="stat-item">
              <el-icon><EditPen /></el-icon>
              字符: {{ charCount }}
            </span>
          </div>
          <div class="footer-tips">
            <span class="tip-text">提示: 请勿修改 apiVersion 和 kind 字段</span>
          </div>
        </div>
      </el-card>
      <template #footer>
        <div class="yaml-footer-actions">
          <el-button @click="yamlDialogVisible = false">
            <el-icon><Close /></el-icon>
            取消
          </el-button>
          <el-button type="warning" @click="resetYaml" :loading="yamlLoading">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
          <el-button type="primary" @click="saveYaml" :loading="yamlSaving" :disabled="yamlError">
            <el-icon><Check /></el-icon>
            保存
          </el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted, shallowRef, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Codemirror } from 'vue-codemirror'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view'
import { ArrowLeft, Delete, Connection, Share, Right, Collection, Box, Document, PriceTag, EditPen, Bell, Edit, Warning, SuccessFilled, Refresh, Check, Close } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()
const route = useRoute()
const serviceName = ref(route.params.name)
const activeTab = ref('status')
const service = ref(null)
const pods = ref([])
const deployments = ref([])
const events = ref([])
const loading = ref(false)
const podsLoading = ref(false)
const deploymentsLoading = ref(false)
const eventsLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlSaving = ref(false)
const yamlLoading = ref(false)
const yamlError = ref('')
const yamlExtensions = [yaml(), oneDark, lineNumbers(), highlightActiveLine(), highlightActiveLineGutter()]

const lineCount = computed(() => {
  if (!yamlContent.value) return 0
  return yamlContent.value.split('\n').length
})

const charCount = computed(() => {
  return yamlContent.value?.length || 0
})

const validateYAML = () => {
  if (!yamlContent.value) {
    yamlError.value = ''
    return
  }

  try {
    const lines = yamlContent.value.split('\n')
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      if (line.includes('\t')) {
        yamlError.value = `第 ${i + 1} 行: YAML 不应使用 Tab，请使用空格`
        return
      }
    }

    const indentStack = []
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      if (line.trim() === '' || line.trim().startsWith('#')) continue

      const indent = line.search(/\S/)
      if (indent !== -1 && indentStack.length > 0) {
        const lastIndent = indentStack[indentStack.length - 1]
        if (indent % 2 !== 0 && indent !== lastIndent) {
          yamlError.value = `第 ${i + 1} 行: 缩进应为 2 的倍数`
          return
        }
      }
    }

    yamlError.value = ''
  } catch (e) {
    yamlError.value = e.message
  }
}

const cluster = ref(route.query.cluster)
const namespace = ref(route.query.namespace || 'default')

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getServiceTypeTag = (type) => {
  const types = {
    'ClusterIP': 'primary',
    'NodePort': 'success',
    'LoadBalancer': 'warning',
    'ExternalName': 'info'
  }
  return types[type] || 'info'
}

const getPodStatusType = (status) => {
  const statusTypes = {
    'Running': 'success',
    'Pending': 'warning',
    'Failed': 'danger',
    'Succeeded': 'info',
    'Unknown': 'info'
  }
  return statusTypes[status] || 'info'
}

const getDeploymentStatusType = (status) => {
  const statusTypes = {
    '运行中': 'success',
    '更新中': 'warning',
    '异常': 'danger',
    '未知': 'info'
  }
  return statusTypes[status] || 'info'
}

const fetchServiceDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/detail`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    service.value = res.data
  } catch (error) {
    ElMessage.error('获取Service详情失败')
  } finally {
    loading.value = false
  }
}

const fetchPods = async () => {
  podsLoading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/pods`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    pods.value = res.data || []
  } catch (error) {
    pods.value = []
  } finally {
    podsLoading.value = false
  }
}

const fetchEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/events`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    events.value = res.data || []
  } catch (error) {
    ElMessage.error('获取事件失败')
  } finally {
    eventsLoading.value = false
  }
}

const fetchDeployments = async () => {
  deploymentsLoading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/deployments`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    deployments.value = res.data || []
  } catch (error) {
    deployments.value = []
  } finally {
    deploymentsLoading.value = false
  }
}

const showYamlDialog = async () => {
  yamlLoading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/yaml`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    yamlContent.value = res.data || ''
    yamlDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取YAML失败')
  } finally {
    yamlLoading.value = false
  }
}

const resetYaml = async () => {
  yamlLoading.value = true
  try {
    const res = await request.get(`/k8s/service/${serviceName.value}/yaml`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    yamlContent.value = res.data || ''
    yamlError.value = ''
    ElMessage.success('已重置为原始YAML')
  } catch (error) {
    ElMessage.error('重置失败')
  } finally {
    yamlLoading.value = false
  }
}

const saveYaml = async () => {
  if (!yamlContent.value) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  
  if (yamlError.value) {
    ElMessage.warning('请修正 YAML 错误后再保存')
    return
  }
  
  yamlSaving.value = true
  try {
    await request.put(`/k8s/service/${serviceName.value}/yaml`, {
      yaml: yamlContent.value
    }, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    ElMessage.success('保存成功')
    yamlDialogVisible.value = false
    fetchServiceDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '保存失败')
  } finally {
    yamlSaving.value = false
  }
}

const deleteService = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除 Service ${serviceName.value} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await request.delete(`/k8s/service/${serviceName.value}`, {
      params: {
        cluster: cluster.value,
        namespace: namespace.value
      }
    })
    ElMessage.success('删除成功')
    router.back()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const goToPodDetail = (pod) => {
  router.push({
    path: '/k8s/pods',
    query: {
      cluster: cluster.value,
      namespace: namespace.value,
      search: pod.name
    }
  })
}

const goToDeploymentDetail = (deploy) => {
  router.push({
    path: `/k8s/deployment/${deploy.name}`,
    query: {
      cluster: cluster.value,
      namespace: namespace.value
    }
  })
}

onMounted(() => {
  fetchServiceDetail()
  fetchDeployments()
  fetchPods()
  fetchEvents()
})
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
  gap: 16px;
}

.header-right {
  display: flex;
  gap: 12px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.detail-tabs {
  margin-top: 0;
}

.tab-content {
  padding: 0;
}

.stat-card {
  height: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 18px;
  color: var(--el-color-primary);
}

.card-title {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.mono-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: #409EFF;
  font-weight: 500;
}

.port-mapping {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.port-label {
  color: #909399;
  font-size: 12px;
}

.port-value {
  color: #409EFF;
  font-weight: 600;
  font-family: 'Monaco', 'Menlo', monospace;
}

.port-value.target {
  color: #67C23A;
}

.arrow-icon {
  color: #C0C4CC;
  font-size: 12px;
}

.node-port {
  color: #E6A23C;
  font-weight: 600;
  font-family: 'Monaco', 'Menlo', monospace;
}

.selector-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f0f9eb;
  border-radius: 4px;
}

.selector-label {
  font-size: 13px;
  color: #67C23A;
  font-weight: 500;
}

.selector-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}

.selector-tag {
  font-size: 14px;
}

.selector-key {
  color: #67C23A;
  font-weight: 500;
}

.selector-value {
  color: #409EFF;
  font-weight: 500;
}

.replicas-text {
  color: #409EFF;
  font-weight: 600;
  font-family: 'Monaco', 'Menlo', monospace;
}

.section-gap {
  margin-top: 24px;
}

.metadata-card {
  margin-bottom: 0;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}

.label-tag {
  font-size: 14px;
}

.label-key {
  color: #606266;
  font-weight: 500;
}

.label-value {
  color: #409EFF;
  font-weight: 500;
}

.annotations-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 8px 0;
}

.annotation-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.annotation-key {
  font-size: 13px;
  color: #909399;
  font-weight: 500;
  font-family: 'Monaco', 'Menlo', monospace;
}

.annotation-value {
  font-size: 14px;
  color: #303133;
  word-break: break-all;
  white-space: pre-wrap;
}

.events-card {
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

.el-descriptions {
  --el-descriptions-item-bordered-label-background: #f5f7fa;
}

.yaml-card {
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
}

.yaml-card .el-card__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.yaml-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.yaml-title-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.yaml-icon {
  font-size: 18px;
  color: #409EFF;
}

.yaml-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.yaml-status {
  display: flex;
  align-items: center;
}

.yaml-status .el-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.editor-wrapper {
  flex: 1;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;
  background: #1e1e1e;
}

.yaml-codemirror {
  height: 100%;
}

.yaml-codemirror .cm-editor {
  height: 100%;
}

.yaml-codemirror .cm-scroller {
  overflow: auto;
}

.editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.footer-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #606266;
}

.stat-item .el-icon {
  font-size: 14px;
}

.footer-tips {
  font-size: 12px;
  color: #909399;
}

.yaml-footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.yaml-footer-actions .el-button {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>