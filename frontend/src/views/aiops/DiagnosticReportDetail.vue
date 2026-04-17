<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <span class="resource-info">
          {{ record?.resource_type }}: {{ record?.resource_name }}
        </span>
        <span class="namespace-info">Namespace: {{ record?.namespace }}</span>
        <span class="cluster-info">Cluster: {{ record?.cluster_name }}</span>
      </div>
      <div class="header-right">
        <el-tag :type="getStatusColor(record?.status)" size="small">
          {{ getStatusLabel(record?.status) }}
        </el-tag>
        <el-button type="primary" @click="retryDiagnosis" v-if="record?.status === 'completed'">
          <el-icon><Refresh /></el-icon>
          重新诊断
        </el-button>
      </div>
    </div>

    <el-card class="overview-card" shadow="hover" v-if="record?.status === 'completed'">
      <el-row :gutter="16">
        <el-col :span="4">
          <div class="overview-item">
            <div class="overview-label">严重程度</div>
            <div class="overview-value">
              <el-tag :type="getSeverityColor(record?.severity)" effect="dark" size="large">
                {{ record?.severity }}
              </el-tag>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="overview-item">
            <div class="overview-label">置信度</div>
            <div class="overview-value">
              <el-progress :percentage="Math.round(record?.confidence * 100)" :stroke-width="16" />
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="overview-item">
            <div class="overview-label">诊断耗时</div>
            <div class="overview-value">{{ record?.duration }}ms</div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="overview-item">
            <div class="overview-label">使用模型</div>
            <div class="overview-value">
              <el-tag size="small">{{ record?.llm_provider }}</el-tag>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="overview-item">
            <div class="overview-label">诊断时间</div>
            <div class="overview-value">{{ formatTime(record?.created_at) }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <div class="report-sections" v-if="record?.status === 'completed'">
      <el-card class="section-card" shadow="hover">
        <template #header>
          <div class="section-header">
            <el-icon class="section-icon problem"><WarningFilled /></el-icon>
            <span class="section-title">问题描述</span>
          </div>
        </template>
        <div class="section-content">
          {{ record?.problem_desc }}
        </div>
      </el-card>

      <el-card class="section-card" shadow="hover">
        <template #header>
          <div class="section-header">
            <el-icon class="section-icon cause"><Search /></el-icon>
            <span class="section-title">根本原因分析</span>
          </div>
        </template>
        <div class="section-content">
          <div class="root-cause">{{ record?.root_cause }}</div>
          <div class="possible-causes" v-if="possibleCauses.length > 0">
            <div class="causes-label">可能原因：</div>
            <ul class="causes-list">
              <li v-for="(cause, idx) in possibleCauses" :key="idx">{{ cause }}</li>
            </ul>
          </div>
        </div>
      </el-card>

      <el-card class="section-card" shadow="hover">
        <template #header>
          <div class="section-header">
            <el-icon class="section-icon solution"><Tools /></el-icon>
            <span class="section-title">解决方案</span>
          </div>
        </template>
        <div class="section-content">
          <div class="solution-summary">{{ record?.solution }}</div>
          <div class="solution-steps" v-if="solutionSteps.length > 0">
            <div class="steps-label">详细步骤：</div>
            <div class="steps-list">
              <div v-for="(step, idx) in solutionSteps" :key="idx" class="step-item">
                <span class="step-number">Step {{ idx + 1 }}</span>
                <span class="step-content">{{ step }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="section-card" shadow="hover">
        <template #header>
          <div class="section-header">
            <el-icon class="section-icon impact"><Share /></el-icon>
            <span class="section-title">影响范围</span>
          </div>
        </template>
        <div class="section-content">
          {{ record?.impact_scope }}
        </div>
      </el-card>

      <el-card class="section-card" shadow="hover" v-if="relatedResources.length > 0">
        <template #header>
          <div class="section-header">
            <el-icon class="section-icon related"><Link /></el-icon>
            <span class="section-title">相关资源</span>
          </div>
        </template>
        <div class="section-content">
          <el-row :gutter="16">
            <el-col :span="8" v-for="(res, idx) in relatedResources" :key="idx">
              <el-card class="related-card" shadow="hover">
                <div class="related-type">{{ res.type }}</div>
                <div class="related-name">{{ res.name }}</div>
                <div class="related-reason">{{ res.reason }}</div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-card>
    </div>

    <el-card class="error-card" shadow="hover" v-if="record?.status === 'failed'">
      <div class="error-content">
        <el-icon class="error-icon"><CircleCloseFilled /></el-icon>
        <div class="error-title">诊断失败</div>
        <div class="error-message">{{ record?.error_message }}</div>
        <el-button type="primary" @click="retryDiagnosis">重新诊断</el-button>
      </div>
    </el-card>

    <el-card class="running-card" shadow="hover" v-if="record?.status === 'running'">
      <div class="running-content">
        <el-icon class="running-icon is-loading"><Loading /></el-icon>
        <div class="running-title">诊断进行中...</div>
        <div class="running-desc">正在采集数据并调用 LLM 分析</div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, WarningFilled, Search, Tools, Share, Link, CircleCloseFilled, Loading } from '@element-plus/icons-vue'
import request from '../../utils/request'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const record = ref(null)

const possibleCauses = computed(() => {
  if (!record.value?.possible_causes) return []
  try {
    return JSON.parse(record.value.possible_causes)
  } catch {
    return []
  }
})

const solutionSteps = computed(() => {
  if (!record.value?.solution_steps) return []
  try {
    return JSON.parse(record.value.solution_steps)
  } catch {
    return []
  }
})

const relatedResources = computed(() => {
  if (!record.value?.related_resources) return []
  try {
    return JSON.parse(record.value.related_resources)
  } catch {
    return []
  }
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getStatusColor = (status) => {
  const colors = { completed: 'success', running: 'warning', failed: 'danger', pending: 'info' }
  return colors[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = { completed: '已完成', running: '进行中', failed: '失败', pending: '等待' }
  return labels[status] || status
}

const getSeverityColor = (severity) => {
  const colors = { high: 'danger', medium: 'warning', low: 'success' }
  return colors[severity] || 'info'
}

const fetchRecord = async () => {
  loading.value = true
  try {
    const id = route.params.id
    const res = await request.get(`/aiops/diagnostic/${id}`)
    record.value = res.data
  } catch (error) {
    ElMessage.error('获取诊断记录失败')
    router.back()
  } finally {
    loading.value = false
  }
}

const retryDiagnosis = async () => {
  ElMessage.info('重试功能待实现')
}

onMounted(() => {
  fetchRecord()
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
  gap: 16px;
}

.resource-info {
  font-size: 18px;
  font-weight: 600;
}

.namespace-info,
.cluster-info {
  font-size: 14px;
  color: #606266;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.overview-card {
  margin-bottom: 16px;
}

.overview-item {
  text-align: center;
}

.overview-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.overview-value {
  font-size: 16px;
  font-weight: 500;
}

.report-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-card {
  transition: all 0.3s;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  font-size: 20px;
}

.section-icon.problem {
  color: #E6A23C;
}

.section-icon.cause {
  color: #409EFF;
}

.section-icon.solution {
  color: #67C23A;
}

.section-icon.impact {
  color: #909399;
}

.section-icon.related {
  color: #F56C6C;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
}

.section-content {
  line-height: 1.8;
  color: #303133;
}

.root-cause,
.solution-summary {
  margin-bottom: 16px;
}

.causes-label,
.steps-label {
  font-weight: 500;
  margin-bottom: 8px;
}

.causes-list {
  padding-left: 20px;
  margin: 0;
}

.causes-list li {
  margin-bottom: 4px;
}

.steps-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.step-number {
  font-weight: 600;
  color: #409EFF;
  margin-right: 12px;
}

.step-content {
  white-space: pre-wrap;
}

.related-card {
  padding: 16px;
}

.related-type {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.related-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
}

.related-reason {
  font-size: 12px;
  color: #606266;
}

.error-card,
.running-card {
  text-align: center;
  padding: 60px 20px;
}

.error-icon,
.running-icon {
  font-size: 48px;
}

.error-icon {
  color: #F56C6C;
}

.running-icon {
  color: #409EFF;
}

.error-title,
.running-title {
  font-size: 18px;
  font-weight: 600;
  margin-top: 16px;
}

.error-message {
  color: #909399;
  margin-top: 12px;
}

.running-desc {
  color: #909399;
  margin-top: 12px;
}
</style>