<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">编辑 Ingress YAML</h1>
      <div class="page-actions">
        <el-button @click="goBack">返回</el-button>
        <el-button type="warning" @click="resetYAML">重置</el-button>
        <el-button type="primary" @click="saveYAML" :loading="saving">保存</el-button>
      </div>
    </div>

    <el-row :gutter="24" class="content-row">
      <el-col :span="6">
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
              <span class="info-label">规则数</span>
              <span class="info-value">{{ ingress.spec?.rules?.length || 0 }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ formatTime(ingress.metadata?.creationTimestamp) }}</span>
            </div>
            <div class="info-item" v-if="ingress.spec?.ingressClassName">
              <span class="info-label">IngressClass</span>
              <span class="info-value">{{ ingress.spec?.ingressClassName }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="18">
        <el-card class="editor-card" shadow="hover" v-loading="loading">
          <template #header>
            <div class="editor-header">
              <span class="card-title">YAML 配置</span>
              <div class="editor-status">
                <el-tag v-if="yamlError" type="danger" size="small">{{ yamlError }}</el-tag>
                <el-tag v-else-if="yamlContent" type="success" size="small">YAML 格式正确</el-tag>
              </div>
            </div>
          </template>
          <div class="editor-container">
            <codemirror
              v-model="yamlContent"
              :style="{ height: '500px' }"
              :extensions="extensions"
              @change="validateYAML"
            />
          </div>
          <div class="editor-footer">
            <span class="line-count">行数: {{ lineCount }}</span>
            <span class="char-count">字符: {{ yamlContent?.length || 0 }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Codemirror } from 'vue-codemirror'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view'
import request from '../../utils/request'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const ingress = ref(null)
const yamlContent = ref('')
const originalYAML = ref('')
const yamlError = ref('')

const extensions = [yaml(), oneDark, lineNumbers(), highlightActiveLine(), highlightActiveLineGutter()]

const lineCount = computed(() => {
  if (!yamlContent.value) return 0
  return yamlContent.value.split('\n').length
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const fetchIngress = async () => {
  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  const name = route.params.name
  if (!cluster || !namespace || !name) {
    ElMessage.error('缺少必要参数')
    router.push('/k8s/ingress')
    return
  }

  loading.value = true
  try {
    const res = await request.get(`/k8s/ingress/${name}`, {
      params: { cluster, namespace }
    })
    ingress.value = res.data
  } catch (error) {
    ElMessage.error('获取Ingress信息失败')
  }
}

const fetchYAML = async () => {
  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  const name = route.params.name
  loading.value = true
  try {
    const res = await request.get(`/k8s/ingress/${name}/yaml`, {
      params: { cluster, namespace }
    })
    yamlContent.value = res.data || ''
    originalYAML.value = res.data || ''
    yamlError.value = ''
  } catch (error) {
    ElMessage.error('获取YAML失败')
  } finally {
    loading.value = false
  }
}

const validateYAML = () => {
  if (!yamlContent.value) {
    yamlError.value = ''
    return
  }

  try {
    const parsed = yamlContent.value
    if (typeof parsed === 'string' && yamlContent.value.trim()) {
      yamlError.value = ''
    }
  } catch (error) {
    yamlError.value = 'YAML格式错误'
  }
}

const resetYAML = () => {
  yamlContent.value = originalYAML.value
  yamlError.value = ''
  ElMessage.success('已重置为原始YAML')
}

const saveYAML = async () => {
  if (!yamlContent.value) {
    ElMessage.warning('YAML内容不能为空')
    return
  }

  if (yamlError.value) {
    ElMessage.warning('请先修正YAML格式错误')
    return
  }

  try {
    await ElMessageBox.confirm('确认保存YAML配置?', '保存确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  const name = route.params.name
  saving.value = true
  try {
    await request.put(`/k8s/ingress/${name}/yaml`, {
      yaml: yamlContent.value
    }, {
      params: { cluster, namespace }
    })
    
    ElMessage.success('保存成功')
    originalYAML.value = yamlContent.value
    router.push({
      path: '/k8s/ingress',
      query: { cluster, namespace }
    })
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '保存失败')
  } finally {
    saving.value = false
  }
}

const goBack = () => {
  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  router.push({
    path: '/k8s/ingress',
    query: { cluster, namespace }
  })
}

onMounted(() => {
  fetchIngress()
  fetchYAML()
})
</script>

<style scoped>
.content-row {
  margin-top: 16px;
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

.info-value.image {
  font-size: 12px;
  word-break: break-all;
  color: #606266;
}

.editor-card {
  height: 100%;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.editor-status {
  display: flex;
  align-items: center;
}

.editor-container {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;
  font-size: 14px;
}

.editor-container .cm-editor {
  height: 500px;
}

.editor-container .cm-scroller {
  overflow: auto;
}

.editor-footer {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  font-size: 12px;
  color: #909399;
}

.line-count, .char-count {
  padding: 4px 8px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>