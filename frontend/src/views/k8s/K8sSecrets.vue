<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Secret管理</h1>
      <div class="header-right">
        <K8sResourceFilter ref="resourceFilter" @query="handleQuery" @cluster-change="handleClusterChange" @ready="handleReady" />
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建Secret
        </el-button>
      </div>
    </div>

    <el-card class="content-card section-gap" shadow="hover">
      <el-table :data="secrets" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="180"></el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120"></el-table-column>
        <el-table-column label="类型" width="200">
          <template #default="scope">
            <el-tag :type="getSecretTypeColor(scope.row.type)" size="small" effect="plain">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="数据键" width="120" align="center">
          <template #default="scope">
            <el-tag size="small" type="warning">{{ scope.row.data_keys?.length || 0 }}个</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="showDetail(scope.row)">详情</el-button>
            <el-button type="warning" size="small" link @click="showEditDialog(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteSecret(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建Secret弹窗 -->
    <el-dialog 
      v-model="createDialogVisible" 
      title="创建Secret" 
      width="1000px"
      top="3vh"
      :close-on-click-modal="false"
      class="secret-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><Lock /></el-icon>
          <span class="dialog-title">创建Secret</span>
          <el-tag size="small" type="info" effect="plain">{{ searchParams.namespace }}</el-tag>
        </div>
      </template>
      
      <el-form :model="createForm" label-position="top" class="secret-form">
        <el-form-item label="Secret名称" required>
          <el-input 
            v-model="createForm.name" 
            placeholder="请输入Secret名称（小写字母、数字和减号）"
            clearable
          >
            <template #prefix>
              <el-icon><Document /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="Secret类型" required>
          <el-select v-model="createForm.type" style="width: 100%">
            <el-option label="Opaque (通用密钥)" value="Opaque">
              <div style="display: flex; justify-content: space-between;">
                <span>Opaque</span>
                <span style="color: #8b949e; font-size: 12px;">通用密钥</span>
              </div>
            </el-option>
            <el-option label="kubernetes.io/tls (TLS证书)" value="kubernetes.io/tls">
              <div style="display: flex; justify-content: space-between;">
                <span>kubernetes.io/tls</span>
                <span style="color: #8b949e; font-size: 12px;">TLS证书</span>
              </div>
            </el-option>
            <el-option label="kubernetes.io/basic-auth (基本认证)" value="kubernetes.io/basic-auth">
              <div style="display: flex; justify-content: space-between;">
                <span>kubernetes.io/basic-auth</span>
                <span style="color: #8b949e; font-size: 12px;">基本认证</span>
              </div>
            </el-option>
            <el-option label="kubernetes.io/dockerconfigjson (镜像仓库)" value="kubernetes.io/dockerconfigjson">
              <div style="display: flex; justify-content: space-between;">
                <span>kubernetes.io/dockerconfigjson</span>
                <span style="color: #8b949e; font-size: 12px;">镜像仓库认证</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="数据内容">
          <div class="data-editor-container">
            <div class="data-header">
              <span class="data-label">
                <el-icon><Edit /></el-icon>
                键值对配置（Base64编码）
              </span>
              <el-button type="primary" size="small" @click="addDataItem('create')">
                <el-icon><Plus /></el-icon>
                添加键值
              </el-button>
            </div>
            
            <div class="data-items-container" v-if="createForm.dataItems.length > 0">
              <div 
                v-for="(item, index) in createForm.dataItems" 
                :key="index" 
                class="data-item-card"
              >
                <div class="item-header">
                  <span class="item-index">#{{ index + 1 }}</span>
                  <el-button 
                    type="danger" 
                    size="small" 
                    circle
                    @click="removeDataItem('create', index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                
                <div class="item-content">
                  <div class="item-row">
                    <div class="key-input">
                      <el-input 
                        v-model="item.key" 
                        placeholder="配置键名（如：password）"
                        clearable
                      >
                        <template #prepend>
                          <el-icon><Key /></el-icon>
                        </template>
                      </el-input>
                    </div>
                    
                    <div class="value-input">
                      <el-input 
                        v-model="item.value" 
                        type="textarea"
                        :rows="4"
                        placeholder="配置值内容"
                        resize="vertical"
                      />
                    </div>
                    
                    <div class="action-buttons">
                      <el-button 
                        size="small" 
                        @click="encodeBase64('create', index)"
                        title="编码为Base64"
                        circle
                      >
                        <el-icon><Lock /></el-icon>
                      </el-button>
                      <el-button 
                        size="small" 
                        @click="decodeBase64('create', index)"
                        title="解码Base64"
                        circle
                      >
                        <el-icon><Unlock /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="empty-data" v-else>
              <el-empty description="暂无数据，请点击上方按钮添加键值对" :image-size="80" />
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="createSecret">
            <el-icon><Check /></el-icon>
            创建
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑Secret弹窗 -->
    <el-dialog 
      v-model="editDialogVisible" 
      title="编辑Secret" 
      width="1000px"
      top="3vh"
      :close-on-click-modal="false"
      class="secret-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><Edit /></el-icon>
          <span class="dialog-title">编辑Secret</span>
          <el-tag size="small">{{ editForm.name }}</el-tag>
          <el-tag size="small" type="info" effect="plain">{{ searchParams.namespace }}</el-tag>
        </div>
      </template>
      
      <el-form :model="editForm" label-position="top" class="secret-form">
        <el-form-item>
          <div class="secret-info">
            <div class="info-item">
              <span class="info-label">名称：</span>
              <span class="info-value">{{ editForm.name }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">类型：</span>
              <el-tag :type="getSecretTypeColor(editForm.type)" size="small" effect="plain">
                {{ editForm.type }}
              </el-tag>
            </div>
            <div class="info-item">
              <span class="info-label">命名空间：</span>
              <span class="info-value">{{ searchParams.namespace }}</span>
            </div>
          </div>
        </el-form-item>
        
        <el-form-item label="数据内容">
          <div class="data-editor-container">
            <div class="data-header">
              <span class="data-label">
                <el-icon><Edit /></el-icon>
                键值对配置（Base64编码）
              </span>
              <el-button type="primary" size="small" @click="addDataItem('edit')">
                <el-icon><Plus /></el-icon>
                添加键值
              </el-button>
            </div>
            
            <div class="data-items-container" v-if="editForm.dataItems.length > 0">
              <div 
                v-for="(item, index) in editForm.dataItems" 
                :key="index" 
                class="data-item-card"
              >
                <div class="item-header">
                  <span class="item-index">#{{ index + 1 }}</span>
                  <el-button 
                    type="danger" 
                    size="small" 
                    circle
                    @click="removeDataItem('edit', index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                
                <div class="item-content">
                  <div class="item-row">
                    <div class="key-input">
                      <el-input 
                        v-model="item.key" 
                        placeholder="配置键名"
                        clearable
                      >
                        <template #prepend>
                          <el-icon><Key /></el-icon>
                        </template>
                      </el-input>
                    </div>
                    
                    <div class="value-input">
                      <el-input 
                        v-model="item.value" 
                        type="textarea"
                        :rows="4"
                        placeholder="配置值内容"
                        resize="vertical"
                      />
                    </div>
                    
                    <div class="action-buttons">
                      <el-button 
                        size="small" 
                        @click="encodeBase64('edit', index)"
                        title="编码为Base64"
                        circle
                      >
                        <el-icon><Lock /></el-icon>
                      </el-button>
                      <el-button 
                        size="small" 
                        @click="decodeBase64('edit', index)"
                        title="解码Base64"
                        circle
                      >
                        <el-icon><Unlock /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="empty-data" v-else>
              <el-empty description="暂无数据，请点击上方按钮添加键值对" :image-size="80" />
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="updateSecret">
            <el-icon><Check /></el-icon>
            保存修改
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      title="Secret详情" 
      width="1000px"
      top="3vh"
      class="secret-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><View /></el-icon>
          <span class="dialog-title">Secret详情</span>
          <el-tag size="small">{{ currentSecret.name }}</el-tag>
        </div>
      </template>
      
      <div class="detail-info-card">
        <div class="info-row">
          <div class="info-item">
            <el-icon><Document /></el-icon>
            <span class="info-label">名称：</span>
            <span class="info-value">{{ currentSecret.name }}</span>
          </div>
          <div class="info-item">
            <el-icon><Folder /></el-icon>
            <span class="info-label">命名空间：</span>
            <span class="info-value">{{ currentSecret.namespace }}</span>
          </div>
          <div class="info-item">
            <el-icon><Clock /></el-icon>
            <span class="info-label">创建时间：</span>
            <span class="info-value">{{ formatTime(currentSecret.created_at) }}</span>
          </div>
        </div>
        <div class="info-row" style="margin-top: 12px;">
          <div class="info-item">
            <el-icon><Lock /></el-icon>
            <span class="info-label">类型：</span>
            <el-tag :type="getSecretTypeColor(currentSecret.type)" size="small" effect="plain">
              {{ currentSecret.type }}
            </el-tag>
          </div>
        </div>
      </div>
      
      <div class="detail-data-section" v-if="currentSecretData.length > 0">
        <div class="section-header">
          <el-icon><Files /></el-icon>
          <span>数据内容</span>
          <el-tag size="small" type="warning">{{ currentSecretData.length }}项</el-tag>
        </div>
        
        <div class="detail-items-container">
          <div 
            v-for="(item, index) in currentSecretData" 
            :key="index"
            class="detail-item-card"
          >
            <div class="detail-key">
              <el-icon><Key /></el-icon>
              <span>{{ item.key }}</span>
            </div>
            <div class="detail-value">
              <el-input 
                :model-value="item.value" 
                type="textarea"
                :rows="3"
                readonly
                resize="none"
              />
            </div>
            <div class="detail-actions">
              <el-button 
                size="small" 
                @click="decodeValue(item.value)"
                title="解码Base64"
              >
                <el-icon><Unlock /></el-icon>
                解码
              </el-button>
            </div>
          </div>
        </div>
      </div>
      
      <el-empty v-else description="该Secret暂无数据内容" :image-size="100" />
    </el-dialog>

    <!-- 解码结果弹窗 -->
    <el-dialog 
      v-model="decodedDialogVisible" 
      title="解码结果" 
      width="500px"
      class="decoded-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><Unlock /></el-icon>
          <span class="dialog-title">解码结果</span>
        </div>
      </template>
      
      <el-input 
        v-model="decodedValue" 
        type="textarea"
        :rows="10"
        readonly
        resize="vertical"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Lock, Unlock, Document, Edit, Key, Check, View, Folder, Clock, Files } from '@element-plus/icons-vue'
import request from '../../utils/request'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const resourceFilter = ref(null)
const secrets = ref([])
const loading = ref(false)
const searchParams = ref({
  cluster: null,
  namespace: 'default'
})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const decodedDialogVisible = ref(false)
const currentSecret = ref({})
const currentSecretData = ref([])
const decodedValue = ref('')

const createForm = ref({
  name: '',
  type: 'Opaque',
  dataItems: []
})

const editForm = ref({
  name: '',
  type: '',
  dataItems: []
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getSecretTypeColor = (type) => {
  const colors = {
    'Opaque': '',
    'kubernetes.io/tls': 'success',
    'kubernetes.io/basic-auth': 'warning',
    'kubernetes.io/dockerconfigjson': 'info',
    'kubernetes.io/service-account-token': 'danger'
  }
  return colors[type] || 'info'
}

const handleQuery = (params) => {
  searchParams.value = params
  fetchSecrets()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchSecrets()
}

const handleClusterChange = () => {
  secrets.value = []
}

const fetchSecrets = async () => {
  if (!searchParams.value.cluster) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  loading.value = true
  try {
    const res = await request.get('/k8s/secret', {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    secrets.value = res.data || []
  } catch (error) {
    ElMessage.error('获取Secret列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  createForm.value = { name: '', type: 'Opaque', dataItems: [] }
  createDialogVisible.value = true
}

const addDataItem = (type) => {
  if (type === 'create') {
    createForm.value.dataItems.push({ key: '', value: '' })
  } else {
    editForm.value.dataItems.push({ key: '', value: '' })
  }
}

const removeDataItem = (type, index) => {
  if (type === 'create') {
    createForm.value.dataItems.splice(index, 1)
  } else {
    editForm.value.dataItems.splice(index, 1)
  }
}

const encodeBase64 = (type, index) => {
  const items = type === 'create' ? createForm.value.dataItems : editForm.value.dataItems
  if (items[index].value) {
    try {
      items[index].value = btoa(items[index].value)
      ElMessage.success('已编码为Base64')
    } catch (e) {
      ElMessage.error('编码失败')
    }
  }
}

const decodeBase64 = (type, index) => {
  const items = type === 'create' ? createForm.value.dataItems : editForm.value.dataItems
  if (items[index].value) {
    try {
      items[index].value = atob(items[index].value)
      ElMessage.success('已解码')
    } catch (e) {
      ElMessage.error('解码失败，可能不是有效的Base64')
    }
  }
}

const decodeValue = (value) => {
  try {
    decodedValue.value = atob(value)
    decodedDialogVisible.value = true
  } catch (e) {
    ElMessage.error('解码失败')
  }
}

const createSecret = async () => {
  if (!createForm.value.name) {
    ElMessage.warning('请输入Secret名称')
    return
  }

  const data = {}
  createForm.value.dataItems.forEach(item => {
    if (item.key) {
      data[item.key] = item.value
    }
  })

  try {
    await request.post('/k8s/secret', {
      name: createForm.value.name,
      type: createForm.value.type,
      data: data
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchSecrets()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const showDetail = async (secret) => {
  try {
    const res = await request.get(`/k8s/secret/${secret.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    currentSecret.value = secret
    const data = res.data?.data || {}
    currentSecretData.value = Object.entries(data).map(([key, value]) => ({
      key,
      value: typeof value === 'string' ? value : btoa(String.fromCharCode(...value))
    }))
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

const showEditDialog = async (secret) => {
  try {
    const res = await request.get(`/k8s/secret/${secret.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    editForm.value.name = secret.name
    editForm.value.type = secret.type
    const data = res.data?.data || {}
    editForm.value.dataItems = Object.entries(data).map(([key, value]) => ({
      key,
      value: typeof value === 'string' ? value : btoa(String.fromCharCode(...value))
    }))
    editDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

const updateSecret = async () => {
  const data = {}
  editForm.value.dataItems.forEach(item => {
    if (item.key) {
      data[item.key] = item.value
    }
  })

  try {
    await request.put(`/k8s/secret/${editForm.value.name}`, {
      data: data
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchSecrets()
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const deleteSecret = async (secret) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Secret ${secret.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/secret/${secret.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('删除成功')
    fetchSecrets()
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

.secret-dialog .el-dialog__header {
  background: linear-gradient(135deg, #fff9e6 0%, #ffe8cc 100%);
  padding: 20px 24px;
  margin: 0;
  border-bottom: 1px solid #ffd666;
}

.secret-dialog .el-dialog__body {
  padding: 24px;
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon {
  color: #ff9800;
  font-size: 22px;
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.secret-form {
  width: 100%;
}

.secret-info {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: #fff9e6;
  border-radius: 8px;
  border: 1px solid #ffd666;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.info-label {
  color: #606266;
  font-size: 14px;
}

.info-value {
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.data-editor-container {
  width: 100%;
  background: #fffbf0;
  border-radius: 12px;
  border: 1px solid #ffd666;
  padding: 16px;
}

.data-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ffd666;
}

.data-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.data-label .el-icon {
  color: #ff9800;
}

.data-items-container {
  max-height: 450px;
  overflow-y: auto;
  padding-right: 4px;
}

.data-item-card {
  background: #ffffff;
  border: 1px solid #ffd666;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  transition: all 0.3s;
}

.data-item-card:hover {
  border-color: #ff9800;
  box-shadow: 0 2px 12px rgba(255, 152, 0, 0.1);
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.item-index {
  font-size: 13px;
  font-weight: 600;
  color: #ff9800;
  background: #fff9e6;
  padding: 4px 12px;
  border-radius: 4px;
}

.item-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.item-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.key-input {
  flex: 0 0 220px;
}

.key-input .el-input {
  width: 100%;
}

.key-input .el-input-group__prepend {
  background: #fff9e6;
  color: #ff9800;
}

.value-input {
  flex: 1;
}

.value-input .el-textarea {
  width: 100%;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex: 0 0 auto;
}

.empty-data {
  padding: 20px;
  background: #ffffff;
  border-radius: 8px;
  border: 1px dashed #ffd666;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.detail-info-card {
  background: #fff9e6;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  border: 1px solid #ffd666;
}

.info-row {
  display: flex;
  gap: 32px;
}

.detail-info-card .info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-info-card .info-item .el-icon {
  color: #ff9800;
  font-size: 18px;
}

.detail-info-card .info-label {
  color: #606266;
  font-size: 14px;
}

.detail-info-card .info-value {
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.detail-data-section {
  margin-top: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.section-header .el-icon {
  color: #ff9800;
  font-size: 20px;
}

.detail-items-container {
  max-height: 550px;
  overflow-y: auto;
}

.detail-item-card {
  background: #ffffff;
  border: 1px solid #ffd666;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.detail-key {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  background: #fff9e6;
  padding: 8px 12px;
  border-radius: 6px;
}

.detail-key .el-icon {
  color: #ff9800;
}

.detail-value {
  margin-bottom: 12px;
}

.detail-value .el-textarea {
  background: #fffbf0;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
}

.decoded-dialog .el-dialog__header {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecef 100%);
  padding: 20px 24px;
  margin: 0;
  border-bottom: 1px solid #e4e7ed;
}

.decoded-dialog .el-dialog__body {
  padding: 24px;
}

.decoded-dialog .dialog-icon {
  color: #67c23a;
}
</style>