<template>
  <AIOPSStreamDialog
    v-model="visible"
    title="智能诊断"
    assistant-name="诊断助手"
    config-title="诊断配置"
    loading-text="正在分析中..."
    :start-message="`请诊断 ${resourceType}/${resourceName} 在 ${namespace} 命名空间的问题`"
    stream-endpoint="/aiops/diagnostic/stream"
    chat-endpoint="/aiops/diagnostic/chat"
    :stream-params="streamParams"
    :config-items="configItems"
    :ready="ready"
    @resolve-cluster="resolveClusterId"
  />
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import request from '../../utils/request'
import AIOPSStreamDialog from './AIOPSStreamDialog.vue'

const props = defineProps({
  modelValue: Boolean,
  clusterId: { type: Number, default: 0 },
  clusterAlias: { type: String, default: '' },
  namespace: String,
  resourceType: String,
  resourceName: String
})

const emit = defineEmits(['update:modelValue'])

const ready = ref(false)
const resolvedClusterId = ref(0)

const visible = computed({
  get: () => props.modelValue,
  set: (val) => {
    if (!val) {
      ready.value = false
    }
    emit('update:modelValue', val)
  }
})

const streamParams = computed(() => ({
  cluster_id: resolvedClusterId.value || props.clusterId,
  namespace: props.namespace,
  resource_type: props.resourceType,
  resource_name: props.resourceName,
}))

const configItems = computed(() => [
  { label: '资源类型', value: props.resourceType, type: '' },
  { label: '资源名称', value: props.resourceName, type: 'info' },
  { label: '命名空间', value: props.namespace, type: 'warning' },
  { label: '集群', value: props.clusterAlias, type: 'success' },
])

const resolveClusterId = async () => {
  if (ready.value) return
  
  if (props.clusterId && props.clusterId > 0) {
    resolvedClusterId.value = props.clusterId
    ready.value = true
    return
  }
  
  if (props.clusterAlias) {
    try {
      const res = await request.get('/k8s/cluster', {
        params: { alias: props.clusterAlias }
      })
      if (res.data && res.data.length > 0) {
        resolvedClusterId.value = res.data[0].id
        ready.value = true
      } else {
        ElMessage.error('无法找到集群信息')
        visible.value = false
      }
    } catch (error) {
      ElMessage.error('获取集群信息失败')
      visible.value = false
    }
  } else {
    ElMessage.error('缺少集群信息')
    visible.value = false
  }
}
</script>