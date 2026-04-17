<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Ingress管理</h1>
      <div class="header-right">
        <K8sResourceFilter 
          ref="resourceFilter"
          @query="handleQuery"
          @cluster-change="handleClusterChange"
          @ready="handleReady"
        />
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建Ingress
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="ingresses" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="160">
          <template #default="scope">
            <el-link type="primary" @click="goToDetail(scope.row)">{{ scope.row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120"></el-table-column>
        <el-table-column prop="ingress_class" label="Ingress Class" width="140">
          <template #default="scope">
            <el-tag v-if="scope.row.ingress_class" type="info" size="small">{{ scope.row.ingress_class }}</el-tag>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="gateway_address" label="网关地址" width="140">
          <template #default="scope">
            <span v-if="scope.row.gateway_address !== '-'" style="color: #409EFF; font-family: monospace;">{{ scope.row.gateway_address }}</span>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="规则数" width="80" align="center">
          <template #default="scope">
            <el-tag type="primary" size="small">{{ scope.row.rule_count }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="主机" min-width="200">
          <template #default="scope">
            <el-tag v-for="(host, idx) in scope.row.hosts" :key="idx" size="small" style="margin-right: 4px">
              {{ host }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="140" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="scope">
            <el-button type="info" size="small" link @click="editYAML(scope.row)">编辑YAML</el-button>
            <el-button type="warning" size="small" link @click="showEditDialog(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteIngress(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="createDialogVisible" title="创建Ingress" direction="rtl" size="60%" destroy-on-close>
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="请输入Ingress名称"></el-input>
        </el-form-item>
        
        <el-form-item label="注解">
          <div class="annotations-container">
            <div v-for="(annotation, idx) in createForm.annotations" :key="idx" class="annotation-item">
              <el-row :gutter="10">
                <el-col :span="10">
                  <el-input v-model="annotation.key" placeholder="注解键名（如：nginx.ingress.kubernetes.io/rewrite-target）"></el-input>
                </el-col>
                <el-col :span="12">
                  <el-input v-model="annotation.value" placeholder="注解值"></el-input>
                </el-col>
                <el-col :span="2">
                  <el-button type="danger" size="small" link @click="removeCreateAnnotation(idx)">
                    删除
                  </el-button>
                </el-col>
              </el-row>
            </div>
            <el-button type="primary" size="small" @click="addCreateAnnotation" style="margin-top: 10px">
              添加注解
            </el-button>
          </div>
        </el-form-item>
        
        <el-form-item label="规则配置">
          <div v-for="(rule, ruleIdx) in createForm.rules" :key="ruleIdx" class="rule-item">
            <el-card shadow="hover" class="rule-card">
              <el-form-item label="主机名">
                <el-input v-model="rule.host" placeholder="例如: example.com"></el-input>
              </el-form-item>
              <el-form-item label="路径配置">
                <div class="paths-wrapper">
                  <div v-for="(path, pathIdx) in rule.paths" :key="pathIdx" class="path-item">
                    <el-row :gutter="12">
                      <el-col :span="7">
                        <el-input v-model="path.path" placeholder="路径，如 /api"></el-input>
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
                        <el-button type="danger" size="small" link @click="removePath(ruleIdx, pathIdx)" v-if="rule.paths.length > 1">
                          删除
                        </el-button>
                      </el-col>
                    </el-row>
                  </div>
                  <div class="add-path-btn">
                    <el-button type="primary" size="small" @click="addPath(ruleIdx)">
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
            <div v-for="(tls, tlsIdx) in createForm.tls" :key="tlsIdx" class="tls-item">
              <el-card shadow="hover" class="tls-card">
                <div class="tls-header">
                  <span class="tls-title">TLS配置 {{ tlsIdx + 1 }}</span>
                  <el-button type="danger" size="small" link @click="removeCreateTls(tlsIdx)">删除</el-button>
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
                          <el-input v-model="tls.hosts[hostIdx]" placeholder="主机名（如：example.com）" style="flex: 1; margin-right: 8px"></el-input>
                          <el-button type="danger" size="small" link @click="removeCreateTlsHost(tlsIdx, hostIdx)">删除</el-button>
                        </div>
                        <el-button type="primary" size="small" @click="addCreateTlsHost(tlsIdx)" style="margin-top: 8px">
                          添加主机名
                        </el-button>
                      </div>
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-card>
            </div>
            <el-button type="primary" size="small" @click="addCreateTls" style="margin-top: 12px">
              添加TLS配置
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="text-align: right">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="createIngress">创建</el-button>
        </div>
      </template>
    </el-drawer>

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
                  <el-input v-model="annotation.key" placeholder="注解键名（如：nginx.ingress.kubernetes.io/rewrite-target）"></el-input>
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
                        <el-input v-model="path.path" placeholder="路径（如 /api）"></el-input>
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
                          <el-input v-model="tls.hosts[hostIdx]" placeholder="主机名（如：example.com）" style="flex: 1; margin-right: 8px"></el-input>
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit } from '@element-plus/icons-vue'
import request from '../../utils/request'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const router = useRouter()
const resourceFilter = ref(null)
const ingresses = ref([])
const loading = ref(false)
const searchParams = ref({
  cluster: null,
  namespace: 'default'
})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)

const createForm = ref({
  name: '',
  annotations: [],
  rules: [
    {
      host: '',
      paths: [
        { path: '/', path_type: 'Prefix', service_name: '', service_port: 80 }
      ]
    }
  ],
  tls: []
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

const handleQuery = (params) => {
  searchParams.value = params
  fetchIngresses()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchIngresses()
}

const handleClusterChange = () => {
  ingresses.value = []
}

const fetchIngresses = async () => {
  if (!searchParams.value.cluster) {
    return
  }
  
  loading.value = true
  try {
    const res = await request.get('/k8s/ingress', {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ingresses.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Ingress列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  createDialogVisible.value = true
}

const addCreateAnnotation = () => {
  createForm.value.annotations.push({ key: '', value: '' })
}

const removeCreateAnnotation = (idx) => {
  createForm.value.annotations.splice(idx, 1)
}

const addPath = (ruleIdx) => {
  createForm.value.rules[ruleIdx].paths.push({
    path: '/',
    path_type: 'Prefix',
    service_name: '',
    service_port: 80
  })
}

const removePath = (ruleIdx, pathIdx) => {
  createForm.value.rules[ruleIdx].paths.splice(pathIdx, 1)
}

const addCreateTls = () => {
  createForm.value.tls.push({
    hosts: [''],
    secret_name: ''
  })
}

const removeCreateTls = (tlsIdx) => {
  createForm.value.tls.splice(tlsIdx, 1)
}

const addCreateTlsHost = (tlsIdx) => {
  if (createForm.value.tls && createForm.value.tls[tlsIdx]) {
    createForm.value.tls[tlsIdx].hosts.push('')
  }
}

const removeCreateTlsHost = (tlsIdx, hostIdx) => {
  if (createForm.value.tls && createForm.value.tls[tlsIdx]) {
    createForm.value.tls[tlsIdx].hosts.splice(hostIdx, 1)
  }
}

const createIngress = async () => {
  if (!createForm.value.name) {
    ElMessage.warning('请输入Ingress名称')
    return
  }
  
  const validPaths = createForm.value.rules.every(rule => 
    rule.paths.every(path => path.service_name && path.service_port)
  )
  
  if (!validPaths) {
    ElMessage.warning('请填写完整的服务名称和端口')
    return
  }
  
  try {
    const ingressSpec = buildIngressSpec(createForm.value.rules, createForm.value.tls)
    
    const annotations = {}
    createForm.value.annotations.forEach(item => {
      if (item.key) {
        annotations[item.key] = item.value
      }
    })
    
    await request.post('/k8s/ingress', {
      metadata: {
        name: createForm.value.name,
        namespace: searchParams.value.namespace,
        annotations: annotations
      },
      spec: ingressSpec
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    createForm.value = {
      name: '',
      annotations: [],
      rules: [
        {
          host: '',
          paths: [
            { path: '/', path_type: 'Prefix', service_name: '', service_port: 80 }
          ]
        }
      ],
      tls: []
    }
    fetchIngresses()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '创建失败')
  }
}

const goToDetail = (ingress) => {
  router.push({
    path: `/k8s/ingress/${ingress.name}`,
    query: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
}

const showEditDialog = async (ingress) => {
  try {
    const res = await request.get(`/k8s/ingress/${ingress.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    const ingressData = res.data
    
    editForm.value.name = ingressData.name
    editForm.value.annotations = Object.entries(ingressData.annotations || {}).map(([key, value]) => ({
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
        namespace: searchParams.value.namespace,
        annotations: annotations
      },
      spec: ingressSpec
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchIngresses()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '更新失败')
  }
}

const editYAML = (ingress) => {
  router.push({
    path: `/k8s/ingress/${ingress.name}/edit`,
    query: {
      cluster: searchParams.value.cluster,
      namespace: searchParams.value.namespace
    }
  })
}

const deleteIngress = async (ingress) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Ingress ${ingress.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/ingress/${ingress.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('删除成功')
    fetchIngresses()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
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

.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
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