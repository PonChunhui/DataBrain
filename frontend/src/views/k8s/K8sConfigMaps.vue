<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">ConfigMap管理</h1>
      <div class="header-right">
        <K8sResourceFilter 
          ref="resourceFilter"
          @query="handleQuery"
          @cluster-change="handleClusterChange"
          @ready="handleReady"
        />
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建ConfigMap
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="configMaps" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="180"></el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120"></el-table-column>
        <el-table-column label="数据键" width="120" align="center">
          <template #default="scope">
            <el-tag size="small" type="info">{{ scope.row.data_keys?.length || 0 }}个</el-tag>
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
            <el-button type="danger" size="small" link @click="deleteConfigMap(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建ConfigMap弹窗 -->
    <el-dialog 
      v-model="createDialogVisible" 
      title="创建ConfigMap" 
      width="900px"
      top="5vh"
      :close-on-click-modal="false"
      class="configmap-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><Files /></el-icon>
          <span class="dialog-title">创建ConfigMap</span>
          <el-tag size="small" type="info" effect="plain">{{ searchParams.namespace }}</el-tag>
        </div>
      </template>
      
      <el-form :model="createForm" label-position="top" class="configmap-form">
        <el-form-item label="ConfigMap名称" required>
          <el-input 
            v-model="createForm.name" 
            placeholder="请输入ConfigMap名称（小写字母、数字和减号）"
            clearable
          >
            <template #prefix>
              <el-icon><Document /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="数据内容">
          <div class="data-editor-container">
            <div class="data-header">
              <span class="data-label">
                <el-icon><Edit /></el-icon>
                键值对配置
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
                        placeholder="配置键名（如：config.yaml）"
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
                        :rows="3"
                        placeholder="配置值内容（支持多行文本）"
                        resize="vertical"
                      />
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
          <el-button type="primary" @click="createConfigMap">
            <el-icon><Check /></el-icon>
            创建
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑ConfigMap弹窗 -->
    <el-dialog 
      v-model="editDialogVisible" 
      title="编辑ConfigMap" 
      width="900px"
      top="5vh"
      :close-on-click-modal="false"
      class="configmap-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><Edit /></el-icon>
          <span class="dialog-title">编辑ConfigMap</span>
          <el-tag size="small">{{ editForm.name }}</el-tag>
          <el-tag size="small" type="info" effect="plain">{{ searchParams.namespace }}</el-tag>
        </div>
      </template>
      
      <el-form :model="editForm" label-position="top" class="configmap-form">
        <el-form-item>
          <div class="configmap-info">
            <div class="info-item">
              <span class="info-label">名称：</span>
              <span class="info-value">{{ editForm.name }}</span>
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
                键值对配置
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
                        :rows="3"
                        placeholder="配置值内容"
                        resize="vertical"
                      />
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
          <el-button type="primary" @click="updateConfigMap">
            <el-icon><Check /></el-icon>
            保存修改
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      title="ConfigMap详情" 
      width="900px"
      top="5vh"
      class="configmap-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <el-icon class="dialog-icon"><View /></el-icon>
          <span class="dialog-title">ConfigMap详情</span>
          <el-tag size="small">{{ currentConfigMap.name }}</el-tag>
        </div>
      </template>
      
      <div class="detail-info-card">
        <div class="info-row">
          <div class="info-item">
            <el-icon><Document /></el-icon>
            <span class="info-label">名称：</span>
            <span class="info-value">{{ currentConfigMap.name }}</span>
          </div>
          <div class="info-item">
            <el-icon><Folder /></el-icon>
            <span class="info-label">命名空间：</span>
            <span class="info-value">{{ currentConfigMap.namespace }}</span>
          </div>
          <div class="info-item">
            <el-icon><Clock /></el-icon>
            <span class="info-label">创建时间：</span>
            <span class="info-value">{{ formatTime(currentConfigMap.created_at) }}</span>
          </div>
        </div>
      </div>
      
      <div class="detail-data-section" v-if="currentConfigMapDataList.length > 0">
        <div class="section-header">
          <el-icon><Files /></el-icon>
          <span>数据内容</span>
          <el-tag size="small" type="info">{{ currentConfigMapDataList.length }}项</el-tag>
        </div>
        
        <div class="detail-items-container">
          <div 
            v-for="(item, index) in currentConfigMapDataList" 
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
          </div>
        </div>
      </div>
      
      <el-empty v-else description="该ConfigMap暂无数据内容" :image-size="100" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Files, Document, Edit, Key, Check, View, Folder, Clock } from '@element-plus/icons-vue'
import request from '../../utils/request'
import K8sResourceFilter from '../../components/k8s/K8sResourceFilter.vue'

const resourceFilter = ref(null)
const configMaps = ref([])
const loading = ref(false)
const searchParams = ref({
  cluster: null,
  namespace: 'default'
})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const currentConfigMap = ref({})
const currentConfigMapData = ref({})
const currentConfigMapDataList = ref([])

const createForm = ref({
  name: '',
  dataItems: []
})

const editForm = ref({
  name: '',
  dataItems: []
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const handleQuery = (params) => {
  searchParams.value = params
  fetchConfigMaps()
}

const handleReady = (params) => {
  searchParams.value = params
  fetchConfigMaps()
}

const handleClusterChange = () => {
  configMaps.value = []
}

const fetchConfigMaps = async () => {
  if (!searchParams.value.cluster) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  loading.value = true
  try {
    const res = await request.get('/k8s/configmap', {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    configMaps.value = res.data || []
  } catch (error) {
    ElMessage.error('获取ConfigMap列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  createForm.value = { name: '', dataItems: [] }
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

const createConfigMap = async () => {
  if (!createForm.value.name) {
    ElMessage.warning('请输入ConfigMap名称')
    return
  }

  const data = {}
  createForm.value.dataItems.forEach(item => {
    if (item.key) {
      data[item.key] = item.value
    }
  })

  try {
    await request.post('/k8s/configmap', {
      name: createForm.value.name,
      data: data
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchConfigMaps()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const showDetail = async (cm) => {
  try {
    const res = await request.get(`/k8s/configmap/${cm.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    currentConfigMap.value = cm
    currentConfigMapData.value = res.data?.data || {}
    currentConfigMapDataList.value = Object.entries(currentConfigMapData.value).map(([key, value]) => ({ key, value }))
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

const showEditDialog = async (cm) => {
  try {
    const res = await request.get(`/k8s/configmap/${cm.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    editForm.value.name = cm.name
    const data = res.data?.data || {}
    editForm.value.dataItems = Object.entries(data).map(([key, value]) => ({ key, value }))
    editDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

const updateConfigMap = async () => {
  const data = {}
  editForm.value.dataItems.forEach(item => {
    if (item.key) {
      data[item.key] = item.value
    }
  })

  try {
    await request.put(`/k8s/configmap/${editForm.value.name}`, {
      data: data
    }, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchConfigMaps()
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const deleteConfigMap = async (cm) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ConfigMap ${cm.name} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/configmap/${cm.name}`, {
      params: {
        cluster: searchParams.value.cluster,
        namespace: searchParams.value.namespace
      }
    })
    ElMessage.success('删除成功')
    fetchConfigMaps()
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

.configmap-dialog .el-dialog__header {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecef 100%);
  padding: 20px 24px;
  margin: 0;
  border-bottom: 1px solid #e4e7ed;
}

.configmap-dialog .el-dialog__body {
  padding: 24px;
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon {
  color: #409eff;
  font-size: 22px;
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.configmap-form {
  width: 100%;
}

.configmap-info {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
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
  background: #fafbfc;
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  padding: 16px;
}

.data-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
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
  color: #409eff;
}

.data-items-container {
  max-height: 400px;
  overflow-y: auto;
  padding-right: 4px;
}

.data-item-card {
  background: #ffffff;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  transition: all 0.3s;
}

.data-item-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.1);
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
  color: #909399;
  background: #f5f7fa;
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
  flex: 0 0 200px;
}

.key-input .el-input {
  width: 100%;
}

.key-input .el-input-group__prepend {
  background: #f5f7fa;
  color: #606266;
}

.value-input {
  flex: 1;
}

.value-input .el-textarea {
  width: 100%;
}

.empty-data {
  padding: 20px;
  background: #ffffff;
  border-radius: 8px;
  border: 1px dashed #dcdfe6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.detail-info-card {
  background: #f5f7fa;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  border: 1px solid #e4e7ed;
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
  color: #409eff;
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
  color: #409eff;
  font-size: 20px;
}

.detail-items-container {
  max-height: 500px;
  overflow-y: auto;
}

.detail-item-card {
  background: #ffffff;
  border: 1px solid #e4e7ed;
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
  background: #f5f7fa;
  padding: 8px 12px;
  border-radius: 6px;
}

.detail-key .el-icon {
  color: #409eff;
}

.detail-value .el-textarea {
  background: #fafbfc;
}
</style>