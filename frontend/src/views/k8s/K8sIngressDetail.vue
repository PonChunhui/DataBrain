<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ ingressName }}</h1>
        <el-tag v-if="ingress?.spec?.ingressClassName" type="info" size="small" style="margin-left: 12px;">
          {{ ingress?.spec?.ingressClassName }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button @click="refreshAll" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="editYAML">
          <el-icon><Edit /></el-icon>
          编辑YAML
        </el-button>
        <el-button type="warning" @click="showEditDialog">
          <el-icon><EditPen /></el-icon>
          编辑
        </el-button>
        <el-button type="danger" @click="deleteIngress">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs" v-loading="loading">
      <el-tab-pane label="规则" name="rules">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="8">
              <el-card class="info-card" shadow="hover" v-if="ingress">
                <template #header>
                  <span class="card-title">基本信息</span>
                </template>
                <div class="info-list">
                  <div class="info-item">
                    <span class="info-label">名称</span>
                    <span class="info-value">{{ ingress.metadata?.name }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">命名空间</span>
                    <span class="info-value">{{ ingress.metadata?.namespace }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">IngressClass</span>
                    <span class="info-value">{{ ingress?.spec?.ingressClassName || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">规则数</span>
                    <span class="info-value">{{ ingress?.spec?.rules?.length || 0 }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">创建时间</span>
                    <span class="info-value">{{ formatTime(ingress.metadata?.creationTimestamp) }}</span>
                  </div>
                  <div class="info-item" v-if="gatewayAddress">
                    <span class="info-label">网关地址</span>
                    <span class="info-value" style="color: #409EFF;">{{ gatewayAddress }}</span>
                  </div>
                </div>
              </el-card>
            </el-col>
            
            <el-col :span="16">
              <el-card class="content-card" shadow="hover">
                <template #header>
                  <div class="card-header-row">
                    <span class="card-title">路由规则</span>
                    <el-tag type="primary" size="small">共 {{ ingress?.spec?.rules?.length || 0 }} 条</el-tag>
                  </div>
                </template>
                
                <el-collapse v-if="ingress?.spec?.rules && ingress.spec.rules.length > 0">
                  <el-collapse-item v-for="(rule, idx) in ingress.spec.rules" :key="idx" :name="idx">
                    <template #title>
                      <div class="rule-title">
                        <el-icon><Link /></el-icon>
                        <span class="host-label">{{ rule.host || '(默认)' }}</span>
                        <el-tag type="info" size="small">{{ rule.http?.paths?.length || 0 }} 个路径</el-tag>
                      </div>
                    </template>
                    <el-table :data="rule.http?.paths || []" size="small">
                      <el-table-column prop="path" label="路径" min-width="150">
                        <template #default="scope">
                          <span style="color: #409EFF; font-weight: 500;">{{ scope.row.path }}</span>
                        </template>
                      </el-table-column>
                      <el-table-column label="类型" width="180">
                        <template #default="scope">
                          <el-tag size="small">{{ scope.row.pathType }}</el-tag>
                        </template>
                      </el-table-column>
                      <el-table-column label="后端服务" min-width="200">
                        <template #default="scope">
                          <span style="font-weight: 500;">{{ scope.row.backend?.service?.name }}</span>
                        </template>
                      </el-table-column>
                      <el-table-column label="端口" width="80" align="right">
                        <template #default="scope">
                          <span style="color: #67C23A; font-weight: 600;">{{ scope.row.backend?.service?.port?.number }}</span>
                        </template>
                      </el-table-column>
                    </el-table>
                  </el-collapse-item>
                </el-collapse>
                <el-empty v-else description="暂无路由规则" />
              </el-card>
              
              <el-card class="content-card section-gap" shadow="hover" v-if="ingress?.spec?.tls && ingress.spec.tls.length > 0">
                <template #header>
                  <div class="card-header-row">
                    <span class="card-title">TLS配置</span>
                    <el-tag type="success" size="small">共 {{ ingress?.spec?.tls?.length || 0 }} 条</el-tag>
                  </div>
                </template>
                <el-table :data="ingress?.spec?.tls || []" size="small">
                  <el-table-column label="Secret名称" min-width="200">
                    <template #default="scope">
                      <el-tag type="warning" size="small">{{ scope.row.secretName }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="主机名" min-width="300">
                    <template #default="scope">
                      <el-tag v-for="(host, idx) in scope.row.hosts" :key="idx" size="small" style="margin-right: 4px;">
                        {{ host }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="元数据" name="metadata">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="8">
              <el-card class="content-card metadata-card" shadow="hover">
                <template #header>
                  <div class="metadata-header">
                    <el-icon><PriceTag /></el-icon>
                    <span>标签</span>
                    <span class="metadata-count">{{ Object.keys(ingress?.metadata?.labels || {}).length }}</span>
                  </div>
                </template>
                <div class="labels-list">
                  <div v-for="(value, key) in ingress?.metadata?.labels" :key="key" class="label-item">
                    <span class="label-key">{{ key }}</span>
                    <span class="label-separator">:</span>
                    <span class="label-value">{{ value }}</span>
                  </div>
                  <el-empty v-if="!ingress?.metadata?.labels || Object.keys(ingress?.metadata?.labels).length === 0" description="暂无标签" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :span="16">
              <el-card class="content-card metadata-card" shadow="hover">
                <template #header>
                  <div class="metadata-header">
                    <el-icon><Document /></el-icon>
                    <span>注解</span>
                    <span class="metadata-count">{{ Object.keys(ingress?.metadata?.annotations || {}).length }}</span>
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
                <el-empty v-if="!ingress?.metadata?.annotations || Object.keys(ingress?.metadata?.annotations).length === 0" description="暂无注解" :image-size="60" />
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="事件" name="events">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
            <el-table :data="events" style="width: 100%" v-loading="loadingEvents">
              <el-table-column label="类型" width="100">
                <template #default="scope">
                  <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'warning'" size="small">
                    {{ scope.row.type }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" width="150"></el-table-column>
              <el-table-column prop="time" label="发生时间" width="180" align="center">
                <template #default="scope">{{ formatTime(scope.row.time) }}</template>
              </el-table-column>
              <el-table-column prop="source" label="来源" width="150"></el-table-column>
              <el-table-column prop="message" label="消息" min-width="300"></el-table-column>
            </el-table>
            <el-empty v-if="events.length === 0 && !loadingEvents" description="暂无事件" />
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="editDialogVisible" title="编辑Ingress" direction="rtl" size="60%" destroy-on-close>
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="editForm.name" disabled></el-input>
        </el-form-item>
        
        <el-form-item label="注解">
          <div class="annotations-container">
            <div v-for="(annotation, idx) in editForm.annotations" :key="idx" class="annotation-item">
              <el-row :gutter="10">
                <el-col :span="10">
                  <el-input v-model="annotation.key" placeholder="注解键名"></el-input>
                </el-col>
                <el-col :span="12">
                  <el-input v-model="annotation.value" placeholder="注解值"></el-input>
                </el-col>
                <el-col :span="2">
                  <el-button type="danger" size="small" link @click="removeAnnotation(idx)">
                    删除
                  </el-button>
                </el-col>
              </el-row>
            </div>
            <el-button type="primary" size="small" @click="addAnnotation" style="margin-top: 10px">
              添加注解
            </el-button>
          </div>
        </el-form-item>
        
        <el-form-item label="规则配置">
          <div v-for="(rule, ruleIdx) in editForm.rules" :key="ruleIdx" class="rule-item">
            <el-card shadow="hover" class="rule-card">
              <el-form-item label="主机名">
                <el-input v-model="rule.host" placeholder="例如: example.com"></el-input>
              </el-form-item>
              <el-form-item label="路径配置">
                <div class="paths-wrapper">
                  <div v-for="(path, pathIdx) in rule.paths" :key="pathIdx" class="path-item">
                    <el-row :gutter="12">
                      <el-col :span="7">
                        <el-input v-model="path.path" placeholder="路径"></el-input>
                      </el-col>
                      <el-col :span="4">
                        <el-select v-model="path.path_type" placeholder="路径类型">
                          <el-option label="Prefix" value="Prefix"></el-option>
                          <el-option label="Exact" value="Exact"></el-option>
                          <el-option label="ImplementationSpecific" value="ImplementationSpecific"></el-option>
                        </el-select>
                      </el-col>
                      <el-col :span="7">
                        <el-input v-model="path.service_name" placeholder="服务名称"></el-input>
                      </el-col>
                      <el-col :span="4">
                        <el-input-number v-model="path.service_port" :min="1" :max="65535" placeholder="端口" style="width: 100%"></el-input-number>
                      </el-col>
                      <el-col :span="2">
                        <el-button type="danger" size="small" link @click="removeEditPath(ruleIdx, pathIdx)" v-if="rule.paths.length > 1">
                          删除
                        </el-button>
                      </el-col>
                    </el-row>
                  </div>
                  <div class="add-path-btn">
                    <el-button type="primary" size="small" @click="addEditPath(ruleIdx)">
                      添加路径
                    </el-button>
                  </div>
                </div>
              </el-form-item>
            </el-card>
          </div>
        </el-form-item>
        
        <el-form-item label="TLS配置">
          <div class="tls-container">
            <div v-for="(tls, tlsIdx) in editForm.tls" :key="tlsIdx" class="tls-item">
              <el-card shadow="hover" class="tls-card">
                <div class="tls-header">
                  <span class="tls-title">TLS配置 {{ tlsIdx + 1 }}</span>
                  <el-button type="danger" size="small" link @click="removeEditTls(tlsIdx)">删除</el-button>
                </div>
                <el-row :gutter="16">
                  <el-col :span="12">
                    <el-form-item label="Secret名称" label-width="90px">
                      <el-input v-model="tls.secret_name" placeholder="Secret名称"></el-input>
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="主机名列表" label-width="90px">
                      <div class="tls-hosts-wrapper">
                        <div v-for="(host, hostIdx) in tls.hosts" :key="hostIdx" class="tls-host-item">
                          <el-input v-model="tls.hosts[hostIdx]" placeholder="主机名" style="flex: 1; margin-right: 8px"></el-input>
                          <el-button type="danger" size="small" link @click="removeEditTlsHost(tlsIdx, hostIdx)">删除</el-button>
                        </div>
                        <el-button type="primary" size="small" @click="addEditTlsHost(tlsIdx)" style="margin-top: 8px">
                          添加主机名
                        </el-button>
                      </div>
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-card>
            </div>
            <el-button type="primary" size="small" @click="addEditTls" style="margin-top: 12px">
              添加TLS配置
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="text-align: right">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="updateIngress">保存</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Refresh, Edit, EditPen, Delete, Link, PriceTag, Document } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()
const route = useRoute()

const clusterAlias = computed(() => localStorage.getItem('k8s_selected_cluster') || route.query.cluster)
const namespace = computed(() => localStorage.getItem('k8s_selected_namespace') || route.query.namespace || 'default')
const ingressName = computed(() => route.params.name)

const activeTab = ref('rules')
const loading = ref(false)
const loadingEvents = ref(false)
const editDialogVisible = ref(false)

const ingress = ref(null)
const events = ref([])
const gatewayAddress = computed(() => {
  if (!ingress.value?.status?.loadBalancer?.ingress) return ''
  const lbIngress = ingress.value.status.loadBalancer.ingress
  if (lbIngress.length > 0) {
    return lbIngress[0].ip || lbIngress[0].hostname || ''
  }
  return ''
})

const editForm = ref({
  name: '',
  annotations: [],
  rules: [],
  tls: []
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const truncateText = (text, maxLen) => {
  if (!text) return ''
  if (text.length > maxLen) {
    return text.substring(0, maxLen) + '...'
  }
  return text
}

const getAnnotationsList = () => {
  const annotations = ingress.value?.metadata?.annotations || {}
  return Object.entries(annotations).map(([key, value]) => ({ key, value }))
}

const fetchIngress = async () => {
  loading.value = true
  try {
    const res = await request.get(`/k8s/ingress/${ingressName.value}`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    ingress.value = res.data
  } catch (error) {
    ElMessage.error('获取Ingress详情失败')
  } finally {
    loading.value = false
  }
}

const fetchEvents = async () => {
  loadingEvents.value = true
  try {
    const res = await request.get(`/k8s/ingress/${ingressName.value}/events`, {
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
  await Promise.all([fetchIngress(), fetchEvents()])
}

const editYAML = () => {
  router.push({
    path: `/k8s/ingress/${ingressName.value}/edit`,
    query: { cluster: clusterAlias.value, namespace: namespace.value }
  })
}

const showEditDialog = async () => {
  try {
    const res = await request.get(`/k8s/ingress/${ingressName.value}`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    const ingressData = res.data
    
    editForm.value.name = ingressData.metadata.name
    editForm.value.annotations = Object.entries(ingressData.metadata.annotations || {}).map(([key, value]) => ({
      key,
      value
    }))
    
    if (ingressData.spec && ingressData.spec.rules) {
      editForm.value.rules = ingressData.spec.rules.map(rule => ({
        host: rule.host || '',
        paths: rule.http?.paths?.map(path => ({
          path: path.path || '',
          path_type: path.pathType || 'Prefix',
          service_name: path.backend?.service?.name || '',
          service_port: path.backend?.service?.port?.number || 80
        })) || []
      }))
    }
    
    if (ingressData.spec && ingressData.spec.tls && ingressData.spec.tls.length > 0) {
      editForm.value.tls = ingressData.spec.tls.map(tls => ({
        hosts: tls.hosts || [],
        secret_name: tls.secretName || ''
      }))
    } else {
      editForm.value.tls = []
    }
    
    editDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取Ingress详情失败')
  }
}

const addAnnotation = () => {
  editForm.value.annotations.push({ key: '', value: '' })
}

const removeAnnotation = (idx) => {
  editForm.value.annotations.splice(idx, 1)
}

const addEditPath = (ruleIdx) => {
  editForm.value.rules[ruleIdx].paths.push({
    path: '/',
    path_type: 'Prefix',
    service_name: '',
    service_port: 80
  })
}

const removeEditPath = (ruleIdx, pathIdx) => {
  editForm.value.rules[ruleIdx].paths.splice(pathIdx, 1)
}

const addEditTls = () => {
  editForm.value.tls.push({
    hosts: [''],
    secret_name: ''
  })
}

const removeEditTls = (tlsIdx) => {
  editForm.value.tls.splice(tlsIdx, 1)
}

const addEditTlsHost = (tlsIdx) => {
  if (editForm.value.tls && editForm.value.tls[tlsIdx]) {
    editForm.value.tls[tlsIdx].hosts.push('')
  }
}

const removeEditTlsHost = (tlsIdx, hostIdx) => {
  if (editForm.value.tls && editForm.value.tls[tlsIdx]) {
    editForm.value.tls[tlsIdx].hosts.splice(hostIdx, 1)
  }
}

const buildIngressSpec = (rules, tls) => {
  const ingressRules = rules.map(rule => {
    const paths = rule.paths.map(path => ({
      path: path.path,
      pathType: path.path_type,
      backend: {
        service: {
          name: path.service_name,
          port: {
            number: path.service_port
          }
        }
      }
    }))
    
    return {
      host: rule.host,
      http: {
        paths: paths
      }
    }
  })
  
  const spec = {
    rules: ingressRules
  }
  
  if (tls && tls.length > 0) {
    spec.tls = tls.map(t => ({
      hosts: t.hosts || [],
      secretName: t.secret_name || ''
    }))
  }
  
  return spec
}

const updateIngress = async () => {
  const validPaths = editForm.value.rules.every(rule => 
    rule.paths.every(path => path.service_name && path.service_port)
  )
  
  if (!validPaths) {
    ElMessage.warning('请填写完整的服务名称和端口')
    return
  }
  
  try {
    const ingressSpec = buildIngressSpec(editForm.value.rules, editForm.value.tls)
    
    const annotations = {}
    editForm.value.annotations.forEach(item => {
      if (item.key) {
        annotations[item.key] = item.value
      }
    })
    
    await request.put(`/k8s/ingress/${editForm.value.name}`, {
      metadata: {
        name: editForm.value.name,
        namespace: namespace.value,
        annotations: annotations
      },
      spec: ingressSpec
    }, {
      params: {
        cluster: clusterAlias.value,
        namespace: namespace.value
      }
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchIngress()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '更新失败')
  }
}

const deleteIngress = async () => {
  try {
    await ElMessageBox.confirm(`确认删除 Ingress "${ingressName.value}"? 此操作不可恢复。`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger'
    })
    
    await request.delete(`/k8s/ingress/${ingressName.value}`, {
      params: { cluster: clusterAlias.value, namespace: namespace.value }
    })
    
    ElMessage.success('删除成功')
    router.push({
      path: '/k8s/ingress',
      query: { cluster: clusterAlias.value, namespace: namespace.value }
    })
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '删除失败')
    }
  }
}

onMounted(() => {
  fetchIngress()
  fetchEvents()
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

.detail-tabs {
  margin-top: 16px;
}

.tab-content {
  padding: 16px 0;
}

.info-card {
  height: 100%;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 12px;
  color: #909399;
}

.info-value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.rule-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.host-label {
  font-weight: 500;
  color: #303133;
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
  color: #409EFF;
}

.metadata-count {
  margin-left: auto;
  padding: 2px 8px;
  background: #ecf5ff;
  border-radius: 10px;
  font-size: 12px;
  color: #409EFF;
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

.section-gap {
  margin-top: 16px;
}

.paths-wrapper {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.add-path-btn {
  margin-top: 8px;
}

.tls-card {
  background: #fff9e6;
  width: 100%;
}

.tls-container {
  width: 100%;
}

.tls-item {
  margin-bottom: 12px;
}

.tls-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.tls-title {
  font-weight: 500;
  color: #303133;
}

.tls-hosts-wrapper {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.tls-host-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.el-drawer__body {
  padding: 20px;
}

.el-drawer__body .el-form {
  width: 100%;
}

.el-drawer__body .el-form-item__content {
  width: 100%;
}

.el-drawer__body .el-card {
  width: 100%;
}

.el-drawer__body .el-row {
  width: 100%;
}

.rule-card {
  background: #f5f7fa;
  width: 100%;
}

.rule-item {
  margin-bottom: 16px;
  width: 100%;
}

.annotation-item {
  margin-bottom: 10px;
}

.annotations-container {
  width: 100%;
}

.path-item {
  margin-bottom: 10px;
  width: 100%;
}
</style>