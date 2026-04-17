<template>
  <div class="k8s-filter-container">
    <el-select 
      v-model="selectedClusterAlias" 
      placeholder="选择集群" 
      @change="onClusterChange"
      style="width: 160px"
    >
      <el-option 
        v-for="cluster in clusters" 
        :key="cluster.id" 
        :label="cluster.name" 
        :value="cluster.alias || String(cluster.id)"
      >
        <span>{{ cluster.name }}</span>
        <el-tag v-if="cluster.alias" type="info" size="small" style="margin-left: 8px;">{{ cluster.alias }}</el-tag>
      </el-option>
    </el-select>
    
    <el-select 
      v-model="selectedNamespace" 
      placeholder="选择命名空间" 
      @change="onNamespaceChange"
      style="width: 140px"
    >
      <el-option 
        v-for="ns in namespaces" 
        :key="ns" 
        :label="ns" 
        :value="ns"
      />
    </el-select>
    
    <el-button type="primary" @click="handleQuery">
      <el-icon><Search /></el-icon>
      查询
    </el-button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import request from '../../utils/request'

const STORAGE_KEY_CLUSTER = 'k8s_selected_cluster'
const STORAGE_KEY_NAMESPACE = 'k8s_selected_namespace'

const props = defineProps({
  defaultNamespace: {
    type: String,
    default: 'default'
  }
})

const emit = defineEmits(['query', 'cluster-change', 'namespace-change', 'ready'])

const clusters = ref([])
const namespaces = ref([])
const selectedClusterAlias = ref(null)
const selectedClusterId = ref(null)
const selectedNamespace = ref('default')
const isInitialized = ref(false)

const getStoredCluster = () => {
  return localStorage.getItem(STORAGE_KEY_CLUSTER) || null
}

const getStoredNamespace = () => {
  return localStorage.getItem(STORAGE_KEY_NAMESPACE) || props.defaultNamespace
}

const saveClusterSelection = (alias) => {
  if (alias) {
    localStorage.setItem(STORAGE_KEY_CLUSTER, alias)
  }
}

const saveNamespaceSelection = (namespace) => {
  if (namespace) {
    localStorage.setItem(STORAGE_KEY_NAMESPACE, namespace)
  }
}

const fetchClusters = async () => {
  try {
    const res = await request.get('/k8s/cluster')
    clusters.value = res.data || []
    
    if (clusters.value.length > 0) {
      const storedAlias = getStoredCluster()
      const found = clusters.value.find(c => c.alias === storedAlias || String(c.id) === storedAlias)
      
      if (found) {
        selectedClusterAlias.value = found.alias || String(found.id)
        selectedClusterId.value = found.id
      } else {
        const first = clusters.value[0]
        selectedClusterAlias.value = first.alias || String(first.id)
        selectedClusterId.value = first.id
        saveClusterSelection(selectedClusterAlias.value)
      }
      
      await fetchNamespaces()
    }
  } catch (error) {
    ElMessage.error('获取集群列表失败')
  }
}

const fetchNamespaces = async () => {
  if (!selectedClusterAlias.value) return
  try {
    const res = await request.get(`/k8s/cluster/${selectedClusterAlias.value}/namespaces`)
    namespaces.value = res.data || []
    
    const storedNamespace = getStoredNamespace()
    const namespaceExists = namespaces.value.includes(storedNamespace)
    
    if (namespaceExists) {
      selectedNamespace.value = storedNamespace
    } else if (namespaces.value.length > 0) {
      selectedNamespace.value = namespaces.value[0]
      saveNamespaceSelection(selectedNamespace.value)
    }
  } catch (error) {
    ElMessage.error('获取命名空间失败')
  }
}

const onClusterChange = async () => {
  saveClusterSelection(selectedClusterAlias.value)
  const found = clusters.value.find(c => c.alias === selectedClusterAlias.value || String(c.id) === selectedClusterAlias.value)
  if (found) {
    selectedClusterId.value = found.id
  }
  selectedNamespace.value = 'default'
  await fetchNamespaces()
  saveNamespaceSelection(selectedNamespace.value)
  emit('cluster-change', selectedClusterAlias.value)
}

const onNamespaceChange = () => {
  saveNamespaceSelection(selectedNamespace.value)
  emit('namespace-change', selectedNamespace.value)
}

const handleQuery = () => {
  if (!selectedClusterAlias.value) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  saveClusterSelection(selectedClusterAlias.value)
  saveNamespaceSelection(selectedNamespace.value)
  
  emit('query', {
    cluster: selectedClusterAlias.value,
    namespace: selectedNamespace.value
  })
}

onMounted(async () => {
  await fetchClusters()
  isInitialized.value = true
  
  if (selectedClusterAlias.value && selectedNamespace.value) {
    emit('ready', {
      cluster: selectedClusterAlias.value,
      namespace: selectedNamespace.value
    })
  }
})

defineExpose({
  selectedClusterAlias,
  selectedClusterId,
  selectedNamespace,
  fetchClusters,
  fetchNamespaces,
  isInitialized
})
</script>

<style scoped>
.k8s-filter-container {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>