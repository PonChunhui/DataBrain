<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">集群管理</h1>
      <div class="header-right">
        <el-button type="primary" @click="showDialog">
          <el-icon><Plus /></el-icon>
          新增集群
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="clusters" style="width: 100%">
        <el-table-column prop="id" label="ID" min-width="50" align="center"></el-table-column>
        <el-table-column prop="alias" label="标识" min-width="80" align="center">
          <template #default="scope">
            <el-link type="primary" v-if="scope.row.alias" @click="goToClusterDetail(scope.row)">{{ scope.row.alias }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="集群名称" min-width="120"></el-table-column>
        <el-table-column prop="namespace" label="默认命名空间" min-width="100"></el-table-column>
        <el-table-column prop="description" label="描述" min-width="160"></el-table-column>
        <el-table-column label="状态" min-width="70" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status ? 'success' : 'danger'" size="small">
              {{ scope.row.status ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="120" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="editCluster(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteCluster(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="drawerVisible" :title="clusterForm.id ? '编辑集群' : '新增集群'" direction="rtl" size="50%" destroy-on-close>
      <el-form :model="clusterForm" label-width="110px" :rules="rules" ref="formRef">
        <el-form-item label="集群名称" prop="name">
          <el-input v-model="clusterForm.name" placeholder="生产环境集群"></el-input>
        </el-form-item>
        <el-form-item label="集群标识" prop="alias">
          <el-input v-model="clusterForm.alias" placeholder="prod" maxlength="50" :disabled="clusterForm.id"></el-input>
          <div class="form-tip">简短标识，如：prod、dev、test（创建后不可修改）</div>
        </el-form-item>
        <el-form-item label="Kubeconfig" prop="kubeconfig">
          <el-input 
            v-model="clusterForm.kubeconfig" 
            type="textarea" 
            :rows="10" 
            placeholder="请粘贴kubeconfig文件内容..."
          ></el-input>
          <div class="kubeconfig-tip">
            <el-link type="primary" @click="testKubeconfig">测试连接</el-link>
            <span class="tip-text">粘贴 ~/.kube/config 文件内容</span>
          </div>
        </el-form-item>
        <el-form-item label="默认命名空间">
          <el-input v-model="clusterForm.namespace" placeholder="default"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="clusterForm.description" placeholder="集群描述"></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="clusterForm.status"></el-switch>
        </el-form-item>

        <el-divider content-position="left">Prometheus配置</el-divider>
        
        <el-form-item label="Prometheus地址">
          <el-input v-model="clusterForm.prometheus_url" placeholder="http://prometheus.example.com:9090"></el-input>
        </el-form-item>
        
        <el-form-item label="启用认证">
          <el-switch v-model="clusterForm.prometheus_auth_enabled"></el-switch>
        </el-form-item>
        
        <el-form-item label="认证类型" v-if="clusterForm.prometheus_auth_enabled">
          <el-radio-group v-model="clusterForm.prometheus_auth_type">
            <el-radio value="basic_auth">Basic Auth</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="用户名" v-if="clusterForm.prometheus_auth_enabled && clusterForm.prometheus_auth_type === 'basic_auth'">
          <el-input v-model="clusterForm.prometheus_basic_auth_user" placeholder="admin"></el-input>
        </el-form-item>
        
        <el-form-item label="密码" v-if="clusterForm.prometheus_auth_enabled && clusterForm.prometheus_auth_type === 'basic_auth'">
          <el-input v-model="clusterForm.prometheus_basic_auth_pass" type="password" placeholder="请输入密码" show-password></el-input>
        </el-form-item>
        
        <el-alert v-if="testResult" :type="testResult.success ? 'success' : 'error'" :closable="false" style="margin-top: 16px">
          <template #title>
            {{ testResult.success ? '连接测试成功' : '连接测试失败' }}
          </template>
          <div v-if="testResult.success">
            <p>API Server: {{ testResult.server }}</p>
            <p>Context: {{ testResult.context }}</p>
          </div>
          <div v-else>
            {{ testResult.message }}
          </div>
        </el-alert>
      </el-form>
      <template #footer>
        <div style="text-align: right">
          <el-button @click="drawerVisible = false">取消</el-button>
          <el-button type="primary" @click="saveCluster" :loading="saving">保存</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()

const clusters = ref([])
const drawerVisible = ref(false)
const saving = ref(false)
const formRef = ref(null)
const testResult = ref(null)
const clusterForm = ref({
  name: '',
  alias: '',
  kubeconfig: '',
  namespace: 'default',
  description: '',
  status: true,
  prometheus_url: '',
  prometheus_auth_enabled: false,
  prometheus_auth_type: 'basic_auth',
  prometheus_basic_auth_user: '',
  prometheus_basic_auth_pass: ''
})

const rules = {
  name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  kubeconfig: [{ required: true, message: '请输入kubeconfig内容', trigger: 'blur' }]
}

const fetchClusters = async () => {
  try {
    const res = await request.get('/k8s/cluster')
    clusters.value = res.data || []
  } catch (error) {
    ElMessage.error('获取集群列表失败')
  }
}

const showDialog = () => {
  clusterForm.value = {
    name: '',
    alias: '',
    kubeconfig: '',
    namespace: 'default',
    description: '',
    status: true,
    prometheus_url: '',
    prometheus_auth_enabled: false,
    prometheus_auth_type: 'basic_auth',
    prometheus_basic_auth_user: '',
    prometheus_basic_auth_pass: ''
  }
  testResult.value = null
  drawerVisible.value = true
}

const goToClusterDetail = (cluster) => {
  localStorage.setItem('k8s_selected_cluster', cluster.alias)
  router.push({
    path: `/k8s/cluster/${cluster.alias}`
  })
}

const editCluster = async (cluster) => {
  try {
    const res = await request.get(`/k8s/cluster/${cluster.id}`)
    clusterForm.value = {
      ...res.data,
      prometheus_auth_type: 'basic_auth'
    }
    
    const infoRes = await request.get(`/k8s/cluster/${cluster.id}/info`)
    testResult.value = {
      success: true,
      server: infoRes.data?.server || '未知',
      context: infoRes.data?.context || '未知'
    }
    
    drawerVisible.value = true
  } catch (error) {
    ElMessage.error('获取集群信息失败')
  }
}

const testKubeconfig = async () => {
  if (!clusterForm.value.kubeconfig) {
    ElMessage.warning('请先输入kubeconfig内容')
    return
  }
  
  try {
    const res = await request.post('/k8s/cluster/test', {
      kubeconfig: clusterForm.value.kubeconfig
    })
    testResult.value = {
      success: true,
      server: res.data?.server,
      context: res.data?.context,
      message: res.data?.message
    }
    ElMessage.success('kubeconfig连接测试成功')
  } catch (error) {
    testResult.value = {
      success: false,
      message: error.response?.data?.msg || '连接测试失败'
    }
    ElMessage.error('kubeconfig连接测试失败')
  }
}

const saveCluster = async () => {
  try {
    await formRef.value.validate()
    saving.value = true
    
    if (clusterForm.value.id) {
      await request.put(`/k8s/cluster/${clusterForm.value.id}`, clusterForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/k8s/cluster', clusterForm.value)
      ElMessage.success('创建成功')
    }
    drawerVisible.value = false
    fetchClusters()
  } catch (error) {
    if (error.response?.data?.msg) {
      ElMessage.error(error.response.data.msg)
    } else if (error !== false) {
      ElMessage.error('操作失败')
    }
  } finally {
    saving.value = false
  }
}

const deleteCluster = async (cluster) => {
  try {
    await ElMessageBox.confirm(`确定要删除集群 "${cluster.name}" 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/k8s/cluster/${cluster.id}`)
    ElMessage.success('删除成功')
    fetchClusters()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(fetchClusters)
</script>

<style scoped>
.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.el-dialog .el-form-item {
  margin-bottom: 20px;
}

.kubeconfig-tip {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.tip-text {
  color: #909399;
  font-size: 12px;
}

.form-tip {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
}
</style>