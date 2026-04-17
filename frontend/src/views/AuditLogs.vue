<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">审计日志</h1>
      <div class="page-actions">
        <el-button type="primary" @click="fetchAuditLogs">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <div class="filter-section">
        <el-form :inline="true">
          <el-form-item label="用户名">
            <el-input v-model="filters.username" placeholder="请输入用户名" clearable style="width: 180px" />
          </el-form-item>
          <el-form-item label="操作类型">
            <el-select v-model="filters.action" placeholder="请选择" clearable style="width: 140px">
              <el-option label="创建" value="create" />
              <el-option label="更新" value="update" />
              <el-option label="删除" value="delete" />
              <el-option label="登录" value="login" />
              <el-option label="登出" value="logout" />
            </el-select>
          </el-form-item>
          <el-form-item label="资源类型">
            <el-select v-model="filters.resource" placeholder="请选择" clearable style="width: 140px">
              <el-option label="用户" value="user" />
              <el-option label="角色" value="role" />
              <el-option label="菜单" value="menu" />
              <el-option label="API" value="api" />
              <el-option label="K8s集群" value="cluster" />
              <el-option label="Deployment" value="deployment" />
              <el-option label="Pod" value="pod" />
              <el-option label="Service" value="service" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filters.status" placeholder="请选择" clearable style="width: 100px">
              <el-option label="成功" value="success" />
              <el-option label="失败" value="failed" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 360px"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchAuditLogs">查询</el-button>
            <el-button @click="resetFilters">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="auditLogs" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" align="center" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="action" label="操作类型" width="100" align="center">
          <template #default="scope">
            <el-tag :type="getActionTagType(scope.row.action)" size="small">
              {{ getActionLabel(scope.row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="资源类型" width="120" />
        <el-table-column prop="resource_id" label="资源ID" width="100" />
        <el-table-column label="详情" min-width="200">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="showDetail(scope.row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="160" align="center">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchAuditLogs"
          @current-change="fetchAuditLogs"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="操作详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户名">{{ currentLog.username }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">{{ getActionLabel(currentLog.action) }}</el-descriptions-item>
        <el-descriptions-item label="资源类型">{{ currentLog.resource }}</el-descriptions-item>
        <el-descriptions-item label="资源ID">{{ currentLog.resource_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog.ip }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'" size="small">
            {{ currentLog.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">{{ formatTime(currentLog.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <div class="detail-json" style="margin-top: 20px">
        <div style="font-weight: 600; margin-bottom: 10px">请求详情：</div>
        <el-input
          v-model="detailJson"
          type="textarea"
          :rows="10"
          readonly
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import request from '../utils/request'

const auditLogs = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const currentLog = ref({})
const detailJson = ref('')

const filters = reactive({
  username: '',
  action: '',
  resource: '',
  status: '',
  start_time: '',
  end_time: ''
})

const timeRange = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getActionTagType = (action) => {
  const typeMap = {
    'create': 'success',
    'update': 'warning',
    'delete': 'danger',
    'login': 'primary',
    'logout': 'info'
  }
  return typeMap[action] || ''
}

const getActionLabel = (action) => {
  const labelMap = {
    'create': '创建',
    'update': '更新',
    'delete': '删除',
    'login': '登录',
    'logout': '登出'
  }
  return labelMap[action] || action
}

const fetchAuditLogs = async () => {
  loading.value = true
  try {
    const params = {
      username: filters.username,
      action: filters.action,
      resource: filters.resource,
      status: filters.status,
      page: pagination.page,
      pageSize: pagination.pageSize
    }

    if (timeRange.value && timeRange.value.length === 2) {
      params.start_time = timeRange.value[0]
      params.end_time = timeRange.value[1]
    }

    const res = await request.get('/audit', { params })
    auditLogs.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    ElMessage.error('获取审计日志失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.username = ''
  filters.action = ''
  filters.resource = ''
  filters.status = ''
  timeRange.value = []
  pagination.page = 1
  fetchAuditLogs()
}

const showDetail = (log) => {
  currentLog.value = log
  try {
    const detailObj = JSON.parse(log.detail)
    detailJson.value = JSON.stringify(detailObj, null, 2)
  } catch (e) {
    detailJson.value = log.detail || ''
  }
  detailVisible.value = true
}

onMounted(fetchAuditLogs)
</script>

<style scoped>
.filter-section {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.detail-json {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
}

.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}
</style>