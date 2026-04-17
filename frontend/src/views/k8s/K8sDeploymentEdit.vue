<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">编辑 Deployment YAML</h1>
      <div class="page-actions">
        <el-button @click="goBack">返回</el-button>
        <el-button type="warning" @click="resetYAML">重置</el-button>
        <el-button type="primary" @click="saveYAML" :loading="saving">保存</el-button>
      </div>
    </div>

    <el-row :gutter="24" class="content-row">
      <el-col :span="6">
        <el-card class="info-card" shadow="hover" v-if="deployment">
          <template #header>
            <span class="card-title">基本信息</span>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">名称</span>
              <span class="info-value">{{ deployment.metadata?.name }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">命名空间</span>
              <span class="info-value">{{ deployment.metadata?.namespace }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">副本数</span>
              <span class="info-value">
                <span class="ready">{{ deployment.status?.readyReplicas || 0 }}</span>
                <span class="separator">/</span>
                <span>{{ deployment.spec?.replicas || 0 }}</span>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ formatTime(deployment.metadata?.creationTimestamp) }}</span>
            </div>
            <div class="info-item" v-if="deployment.spec?.template?.spec?.containers?.length">
              <span class="info-label">容器镜像</span>
              <span class="info-value image">{{ deployment.spec?.template?.spec?.containers[0]?.image }}</span>
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
const deployment = ref(null)
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

const fetchDeployment = async () => {
  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  const name = route.params.name
  if (!cluster || !namespace || !name) {
    ElMessage.error('缺少必要参数')
    router.push('/k8s/deployments')
    return
  }

  loading.value = true
  try {
    const res = await request.get(`/k8s/deployment/${name}`, {
      params: { cluster, namespace }
    })
    deployment.value = res.data
  } catch (error) {
    ElMessage.error('获取Deployment信息失败')
  }
}

const fetchYAML = async () => {
  const cluster = route.query.cluster || localStorage.getItem('k8s_selected_cluster')
  const namespace = route.query.namespace || localStorage.getItem('k8s_selected_namespace') || 'default'
  const name = route.params.name
  loading.value = true
  try {
    const res = await request.get(`/k8s/deployment/${name}/yaml`, {
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

const saveYAML = async () => {
  if (!yamlContent.value) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }

  if (yamlError.value) {
    ElMessage.warning('请修正 YAML 错误后再保存')
    return
  }

  try {
    await ElMessageBox.confirm('确定要保存修改吗？此操作将直接更新 Deployment。', '确认保存', {
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
    await request.put(`/k8s/deployment/${name}/yaml`, {
      yaml: yamlContent.value
    }, {
      params: { cluster, namespace }
    })
    ElMessage.success('保存成功')
    originalYAML.value = yamlContent.value
    await fetchDeployment()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '保存失败')
  } finally {
    saving.value = false
  }
}

const resetYAML = () => {
  if (yamlContent.value === originalYAML.value) {
    ElMessage.info('内容未修改')
    return
  }

  ElMessageBox.confirm('确定要重置为原始 YAML 吗？', '确认重置', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    yamlContent.value = originalYAML.value
    yamlError.value = ''
    ElMessage.success('已重置')
  }).catch(() => {})
}

const goBack = () => {
  router.push('/k8s/deployments')
}

onMounted(() => {
  fetchYAML()
  fetchDeployment()
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

.info-value .ready {
  color: #10B981;
  font-weight: 600;
}

.info-value .separator {
  color: #909399;
  margin: 0 2px;
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