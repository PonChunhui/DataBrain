<template>
  <div class="deployment-create-container">
    <div class="page-header">
      <h1 class="page-title">创建Deployment</h1>
      <el-button @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-form :model="deploymentForm" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="集群" prop="cluster">
          <el-select v-model="deploymentForm.cluster" @change="fetchNamespaces" style="width: 200px">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.alias || String(cluster.id)">
              <span>{{ cluster.name }}</span>
              <el-tag v-if="cluster.alias" type="info" size="small" style="margin-left: 8px;">{{ cluster.alias }}</el-tag>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="命名空间" prop="namespace">
          <el-select v-model="deploymentForm.namespace" style="width: 200px">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="Deployment名称" prop="name">
          <el-input v-model="deploymentForm.name" placeholder="my-deployment" style="width: 300px"></el-input>
        </el-form-item>

        <el-form-item label="副本数" prop="replicas">
          <el-input-number v-model="deploymentForm.replicas" :min="0" :max="100" style="width: 150px"></el-input-number>
        </el-form-item>

        <el-form-item label="镜像" prop="image">
          <el-input v-model="deploymentForm.image" placeholder="nginx:latest" style="width: 400px"></el-input>
        </el-form-item>

        <el-form-item label="容器名称">
          <el-input v-model="deploymentForm.container_name" placeholder="nginx" style="width: 200px"></el-input>
        </el-form-item>

        <el-form-item label="容器端口">
          <el-input-number v-model="deploymentForm.container_port" :min="1" :max="65535" style="width: 150px"></el-input-number>
        </el-form-item>

        <el-form-item label="环境变量">
          <div v-for="(env, index) in deploymentForm.envs" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px">
            <el-input v-model="env.name" placeholder="变量名" style="width: 150px"></el-input>
            <el-input v-model="env.value" placeholder="变量值" style="width: 200px"></el-input>
            <el-button type="danger" size="small" @click="removeEnv(index)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addEnv">添加环境变量</el-button>
        </el-form-item>

        <el-form-item label="资源限制">
          <el-row :gutter="16">
            <el-col :span="6">
              <el-input v-model="deploymentForm.resources.requests.cpu" placeholder="CPU请求(如100m)">
                <template #prepend>CPU请求</template>
              </el-input>
            </el-col>
            <el-col :span="6">
              <el-input v-model="deploymentForm.resources.requests.memory" placeholder="内存请求(如128Mi)">
                <template #prepend>内存请求</template>
              </el-input>
            </el-col>
            <el-col :span="6">
              <el-input v-model="deploymentForm.resources.limits.cpu" placeholder="CPU限制(如500m)">
                <template #prepend>CPU限制</template>
              </el-input>
            </el-col>
            <el-col :span="6">
              <el-input v-model="deploymentForm.resources.limits.memory" placeholder="内存限制(如512Mi)">
                <template #prepend>内存限制</template>
              </el-input>
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="标签">
          <div v-for="(label, index) in deploymentForm.labels" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px">
            <el-input v-model="label.key" placeholder="键" style="width: 150px"></el-input>
            <el-input v-model="label.value" placeholder="值" style="width: 150px"></el-input>
            <el-button type="danger" size="small" @click="removeLabel(index)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addLabel">添加标签</el-button>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="createDeployment" :loading="creating">创建</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()
const formRef = ref(null)
const creating = ref(false)
const clusters = ref([])
const namespaces = ref([])

const deploymentForm = ref({
  cluster: '',
  namespace: 'default',
  name: '',
  replicas: 1,
  image: '',
  container_name: '',
  container_port: 80,
  envs: [],
  resources: {
    requests: { cpu: '', memory: '' },
    limits: { cpu: '', memory: '' }
  },
  labels: []
})

const rules = {
  cluster: [{ required: true, message: '请选择集群', trigger: 'change' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [{ required: true, message: '请输入Deployment名称', trigger: 'blur' }],
  replicas: [{ required: true, message: '请设置副本数', trigger: 'change' }],
  image: [{ required: true, message: '请输入镜像地址', trigger: 'blur' }]
}

const fetchClusters = async () => {
  try {
    const res = await request.get('/k8s/cluster')
    clusters.value = res.data || []
    if (clusters.value.length > 0) {
      const first = clusters.value[0]
      deploymentForm.value.cluster = first.alias || String(first.id)
      fetchNamespaces()
    }
  } catch (error) {
    ElMessage.error('获取集群列表失败')
  }
}

const fetchNamespaces = async () => {
  if (!deploymentForm.value.cluster) return
  try {
    const res = await request.get(`/k8s/cluster/${deploymentForm.value.cluster}/namespaces`)
    namespaces.value = res.data || []
    if (namespaces.value.includes('default')) {
      deploymentForm.value.namespace = 'default'
    } else if (namespaces.value.length > 0) {
      deploymentForm.value.namespace = namespaces.value[0]
    }
  } catch (error) {
    ElMessage.error('获取命名空间失败')
  }
}

const addEnv = () => {
  deploymentForm.value.envs.push({ name: '', value: '' })
}

const removeEnv = (index) => {
  deploymentForm.value.envs.splice(index, 1)
}

const addLabel = () => {
  deploymentForm.value.labels.push({ key: '', value: '' })
}

const removeLabel = (index) => {
  deploymentForm.value.labels.splice(index, 1)
}

const createDeployment = async () => {
  try {
    await formRef.value.validate()
    creating.value = true

    const labels = {}
    deploymentForm.value.labels.forEach(l => {
      if (l.key && l.value) {
        labels[l.key] = l.value
      }
    })
    labels['app'] = deploymentForm.value.name

    const envs = deploymentForm.value.envs
      .filter(e => e.name && e.value)
      .map(e => ({ name: e.name, value: e.value }))

    const containerName = deploymentForm.value.container_name || deploymentForm.value.name

    const resources = {}
    if (deploymentForm.value.resources.requests.cpu) {
      resources.requests = { cpu: deploymentForm.value.resources.requests.cpu }
    }
    if (deploymentForm.value.resources.requests.memory) {
      resources.requests = resources.requests || {}
      resources.requests.memory = deploymentForm.value.resources.requests.memory
    }
    if (deploymentForm.value.resources.limits.cpu) {
      resources.limits = { cpu: deploymentForm.value.resources.limits.cpu }
    }
    if (deploymentForm.value.resources.limits.memory) {
      resources.limits = resources.limits || {}
      resources.limits.memory = deploymentForm.value.resources.limits.memory
    }

    const payload = {
      name: deploymentForm.value.name,
      replicas: deploymentForm.value.replicas,
      selector: {
        matchLabels: { app: deploymentForm.value.name }
      },
      template: {
        metadata: {
          labels: labels
        },
        spec: {
          containers: [{
            name: containerName,
            image: deploymentForm.value.image,
            ports: deploymentForm.value.container_port ? [{ containerPort: deploymentForm.value.container_port }] : [],
            env: envs,
            resources: resources
          }]
        }
      }
    }

    await request.post('/k8s/deployment', payload, {
      params: {
        cluster: deploymentForm.value.cluster,
        namespace: deploymentForm.value.namespace
      }
    })

    ElMessage.success('Deployment创建成功')
    router.push('/k8s/deployments')
  } catch (error) {
    if (error !== false) {
      ElMessage.error(error.response?.data?.msg || '创建失败')
    }
  } finally {
    creating.value = false
  }
}

const resetForm = () => {
  deploymentForm.value = {
    cluster: deploymentForm.value.cluster,
    namespace: 'default',
    name: '',
    replicas: 1,
    image: '',
    container_name: '',
    container_port: 80,
    envs: [],
    resources: {
      requests: { cpu: '', memory: '' },
      limits: { cpu: '', memory: '' }
    },
    labels: []
  }
  formRef.value?.resetFields()
}

onMounted(fetchClusters)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.content-card {
  background: #ffffff;
  border-radius: 12px;
}

.el-form-item {
  margin-bottom: 22px;
}
</style>