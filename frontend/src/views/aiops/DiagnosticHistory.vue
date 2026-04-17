<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">诊断历史记录</h1>
      <div class="header-actions">
        <el-button @click="fetchHistory">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="stats-overview">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon"><el-icon><Search /></el-icon></div>
            <div class="stat-value">{{ stats.total_count || 0 }}</div>
            <div class="stat-label">总诊断次数</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-icon"><el-icon><Calendar /></el-icon></div>
            <div class="stat-value">{{ stats.week_count || 0 }}</div>
            <div class="stat-label">本周诊断</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stat-card warning">
            <div class="stat-icon"><el-icon><Warning /></el-icon></div>
            <div class="stat-value">{{ stats.high_count || 0 }}</div>
            <div class="stat-label">高严重度</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stat-card success">
            <div class="stat-icon"><el-icon><TrendCharts /></el-icon></div>
            <div class="stat-value">{{ stats.avg_confidence || '0%' }}</div>
            <div class="stat-label">平均置信度</div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-card class="filter-card" shadow="hover">
      <el-row :gutter="16">
        <el-col :span="4">
          <el-select v-model="filters.resource_type" placeholder="资源类型" clearable>
            <el-option label="全部" value="" />
            <el-option label="Pod" value="pod" />
            <el-option label="Deployment" value="deployment" />
            <el-option label="Service" value="service" />
            <el-option label="Ingress" value="ingress" />
            <el-option label="Node" value="node" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.severity" placeholder="严重程度" clearable>
            <el-option label="全部" value="" />
            <el-option label="High" value="high" />
            <el-option label="Medium" value="medium" />
            <el-option label="Low" value="low" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.llm_provider" placeholder="模型" clearable>
            <el-option label="全部" value="" />
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="Claude" value="claude" />
            <el-option label="通义千问" value="qwen" />
            <el-option label="Gemini" value="gemini" />
            <el-option label="OpenAI" value="openai" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-col>
        <el-col :span="4">
          <el-input v-model="filters.keyword" placeholder="搜索资源名称" clearable />
        </el-col>
      </el-row>
      <el-row :gutter="16" style="margin-top: 12px">
        <el-col :span="24">
          <el-button type="primary" @click="applyFilters">筛选</el-button>
          <el-button @click="clearFilters">清除筛选</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="content-card" shadow="hover">
      <el-table :data="records" style="width: 100%" v-loading="loading" @row-click="showReport">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="scope">{{ formatTime(scope.row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="resource_type" label="资源类型" width="120">
          <template #default="scope">
            <el-tag size="small" :type="getResourceTypeColor(scope.row.resource_type)">
              {{ scope.row.resource_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource_name" label="资源名称" min-width="200">
          <template #default="scope">
            <el-link type="primary">{{ scope.row.resource_name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="cluster_name" label="集群" width="120" />
        <el-table-column prop="severity" label="严重程度" width="100" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.severity" :type="getSeverityColor(scope.row.severity)" size="small" effect="dark">
              {{ scope.row.severity }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="llm_provider" label="模型" width="120">
          <template #default="scope">
            <el-tag size="small">{{ scope.row.llm_provider }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="scope">
            <el-tag :type="getStatusColor(scope.row.status)" size="small">
              {{ getStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button size="small" link type="primary" @click.stop="goToReport(scope.row)">查看详情</el-button>
            <el-button size="small" link type="warning" @click.stop="retryDiagnosis(scope.row)" v-if="scope.row.status === 'failed'">
              重试
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchHistory"
          @current-change="fetchHistory"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search, Calendar, Warning, TrendCharts } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()

const loading = ref(false)
const records = ref([])
const stats = ref({})
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dateRange = ref([])

const filters = ref({
  resource_type: '',
  severity: '',
  llm_provider: '',
  keyword: ''
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getResourceTypeColor = (type) => {
  const colors = { pod: 'primary', deployment: 'success', service: 'warning', ingress: 'info', node: 'danger' }
  return colors[type] || 'info'
}

const getSeverityColor = (severity) => {
  const colors = { high: 'danger', medium: 'warning', low: 'success' }
  return colors[severity] || 'info'
}

const getStatusColor = (status) => {
  const colors = { completed: 'success', running: 'warning', failed: 'danger', pending: 'info' }
  return colors[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = { completed: '完成', running: '进行中', failed: '失败', pending: '等待' }
  return labels[status] || status
}

const fetchStats = async () => {
  try {
    const res = await request.get('/aiops/history/stats')
    stats.value = res.data || {}
  } catch (error) {
    console.error('获取统计失败')
  }
}

const fetchHistory = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      ...filters.value
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    const res = await request.get('/aiops/history', { params })
    records.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    ElMessage.error('获取历史记录失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  page.value = 1
  fetchHistory()
}

const clearFilters = () => {
  filters.value = { resource_type: '', severity: '', llm_provider: '', keyword: '' }
  dateRange.value = []
  page.value = 1
  fetchHistory()
}

const showReport = (row) => {
  goToReport(row)
}

const goToReport = (row) => {
  router.push(`/aiops/report/${row.id}`)
}

const retryDiagnosis = async (row) => {
  ElMessage.info('重试功能待实现')
}

onMounted(() => {
  fetchStats()
  fetchHistory()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.stats-overview {
  margin: 16px 0;
}

.stat-card {
  text-align: center;
  padding: 20px;
}

.stat-card .stat-icon {
  font-size: 24px;
  margin-bottom: 12px;
}

.stat-card .stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-card .stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.stat-card.warning .stat-value {
  color: #E6A23C;
}

.stat-card.warning .stat-icon {
  color: #E6A23C;
}

.stat-card.success .stat-value {
  color: #67C23A;
}

.stat-card.success .stat-icon {
  color: #67C23A;
}

.filter-card {
  margin-bottom: 16px;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>