<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ node?.name }}</h1>
        <el-tag :type="node?.status === 'Ready' ? 'success' : 'danger'" size="small" style="margin-left: 12px;">
          {{ node?.status }}
        </el-tag>
        <el-tag :type="getRoleType(node?.role)" size="small" effect="plain" style="margin-left: 8px;">
          {{ node?.role }}
        </el-tag>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="运行状态" name="status">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><InfoFilled /></el-icon>
                <span class="card-title">系统信息</span>
              </div>
            </template>
            <el-descriptions :column="4" border>
              <el-descriptions-item label="操作系统">{{ node?.operating_system }}</el-descriptions-item>
              <el-descriptions-item label="架构">{{ node?.architecture }}</el-descriptions-item>
              <el-descriptions-item label="内核版本">{{ node?.kernel_version }}</el-descriptions-item>
              <el-descriptions-item label="OS镜像">{{ node?.os_image }}</el-descriptions-item>
              <el-descriptions-item label="Kubelet">{{ node?.kubelet_version }}</el-descriptions-item>
              <el-descriptions-item label="Kube-proxy">{{ node?.kube_proxy_version }}</el-descriptions-item>
              <el-descriptions-item label="容器运行时">{{ node?.container_runtime }}</el-descriptions-item>
              <el-descriptions-item label="内部IP">{{ node?.internal_ip }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-row :gutter="24" style="margin-top: 24px;">
            <el-col :span="6">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-header">
                  <el-icon class="stat-icon"><Cpu /></el-icon>
                  <span class="stat-title">CPU容量</span>
                </div>
                <div class="stat-value">{{ node?.cpu_capacity }} 核</div>
                <div class="stat-sub">可分配: {{ node?.cpu_allocatable }} 核</div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-header">
                  <el-icon class="stat-icon"><Coin /></el-icon>
                  <span class="stat-title">内存容量</span>
                </div>
                <div class="stat-value">{{ formatMemory(node?.memory_capacity) }}</div>
                <div class="stat-sub">可分配: {{ formatMemory(node?.memory_allocatable) }}</div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-header">
                  <el-icon class="stat-icon"><Grid /></el-icon>
                  <span class="stat-title">Pod容量</span>
                </div>
                <div class="stat-value">{{ node?.pod_capacity }}</div>
                <div class="stat-sub">运行中: {{ runningPods }}</div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card" shadow="hover">
                <div class="stat-header">
                  <el-icon class="stat-icon"><Timer /></el-icon>
                  <span class="stat-title">运行时间</span>
                </div>
                <div class="stat-value">{{ formatDuration(node?.created_at) }}</div>
                <div class="stat-sub">创建于 {{ formatTime(node?.created_at) }}</div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="24" style="margin-top: 24px;">
            <el-col :span="8">
              <el-card class="health-card" shadow="hover" :class="{ 'health-success': isHealthy('Ready'), 'health-danger': !isHealthy('Ready') }">
                <div class="health-title-row">
                  <span class="health-title">就绪</span>
                  <svg v-if="isHealthy('Ready')" class="health-svg success" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm193.5 301.7l-210.6 292a31.8 31.8 0 0 1-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"/>
                  </svg>
                  <svg v-else class="health-svg danger" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 700c-35.3 0-64-28.7-64-64s28.7-64 64-64 64 28.7 64 64-28.7 64-64 64zm0-224c-35.3 0-64-28.7-64-64V288c0-35.3 28.7-64 64-64s64 28.7 64 64v288c0 35.3-28.7 64-64 64z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="health-desc-row">
                  <span class="health-desc-text">节点是否可以接收容器组</span>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="health-card" shadow="hover" :class="{ 'health-success': isHealthy('MemoryPressure'), 'health-danger': !isHealthy('MemoryPressure') }">
                <div class="health-title-row">
                  <span class="health-title">内存压力</span>
                  <svg v-if="isHealthy('MemoryPressure')" class="health-svg success" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm193.5 301.7l-210.6 292a31.8 31.8 0 0 1-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"/>
                  </svg>
                  <svg v-else class="health-svg danger" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 700c-35.3 0-64-28.7-64-64s28.7-64 64-64 64 28.7 64 64-28.7 64-64 64zm0-224c-35.3 0-64-28.7-64-64V288c0-35.3 28.7-64 64-64s64 28.7 64 64v288c0 35.3-28.7 64-64 64z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="health-desc-row">
                  <span class="health-desc-text">节点剩余内存资源状态</span>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="health-card" shadow="hover" :class="{ 'health-success': isHealthy('DiskPressure'), 'health-danger': !isHealthy('DiskPressure') }">
                <div class="health-title-row">
                  <span class="health-title">磁盘压力</span>
                  <svg v-if="isHealthy('DiskPressure')" class="health-svg success" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm193.5 301.7l-210.6 292a31.8 31.8 0 0 1-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"/>
                  </svg>
                  <svg v-else class="health-svg danger" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 700c-35.3 0-64-28.7-64-64s28.7-64 64-64 64 28.7 64 64-28.7 64-64 64zm0-224c-35.3 0-64-28.7-64-64V288c0-35.3 28.7-64 64-64s64 28.7 64 64v288c0 35.3-28.7 64-64 64z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="health-desc-row">
                  <span class="health-desc-text">节点剩余磁盘资源状态</span>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="24" style="margin-top: 24px;">
            <el-col :span="8">
              <el-card class="health-card" shadow="hover" :class="{ 'health-success': isHealthy('PIDPressure'), 'health-danger': !isHealthy('PIDPressure') }">
                <div class="health-title-row">
                  <span class="health-title">进程压力</span>
                  <svg v-if="isHealthy('PIDPressure')" class="health-svg success" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm193.5 301.7l-210.6 292a31.8 31.8 0 0 1-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"/>
                  </svg>
                  <svg v-else class="health-svg danger" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 700c-35.3 0-64-28.7-64-64s28.7-64 64-64 64 28.7 64 64-28.7 64-64 64zm0-224c-35.3 0-64-28.7-64-64V288c0-35.3 28.7-64 64-64s64 28.7 64 64v288c0 35.3-28.7 64-64 64z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="health-desc-row">
                  <span class="health-desc-text">节点剩余进程资源状态</span>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="health-card" shadow="hover" :class="{ 'health-success': isHealthy('NetworkUnavailable'), 'health-danger': !isHealthy('NetworkUnavailable') }">
                <div class="health-title-row">
                  <span class="health-title">网络可用性</span>
                  <svg v-if="isHealthy('NetworkUnavailable')" class="health-svg success" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm193.5 301.7l-210.6 292a31.8 31.8 0 0 1-51.7 0L318.5 484.9c-3.8-5.3 0-12.7 6.5-12.7h46.9c10.2 0 19.9 4.9 25.9 13.3l71.2 98.8 157.2-218c6-8.3 15.6-13.3 25.9-13.3H699c6.5 0 10.3 7.4 6.5 12.7z" fill="currentColor"/>
                  </svg>
                  <svg v-else class="health-svg danger" viewBox="0 0 1024 1024" width="14" height="14">
                    <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 700c-35.3 0-64-28.7-64-64s28.7-64 64-64 64 28.7 64 64-28.7 64-64 64zm0-224c-35.3 0-64-28.7-64-64V288c0-35.3 28.7-64 64-64s64 28.7 64 64v288c0 35.3-28.7 64-64 64z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="health-desc-row">
                  <span class="health-desc-text">节点网络连接状态</span>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-card class="content-card" shadow="hover" style="margin-top: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Warning /></el-icon>
                <span class="card-title">污点</span>
                <el-badge :value="node?.taints?.length || 0" type="warning" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="node?.taints || []" style="width: 100%">
              <el-table-column prop="key" label="键" min-width="200"></el-table-column>
              <el-table-column prop="value" label="值" min-width="100">
                <template #default="scope">{{ scope.row.value || '-' }}</template>
              </el-table-column>
              <el-table-column prop="effect" label="策略" min-width="100" align="center">
                <template #default="scope">
                  <el-tag :type="getTaintEffectType(scope.row.effect)" size="small">{{ scope.row.effect }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!node?.taints || node?.taints.length === 0" description="暂无污点" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="容器组" name="pods">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <span class="card-title">容器组列表</span>
                <el-tag type="info" size="small">共 {{ podsTotal }} 个</el-tag>
              </div>
            </template>
            <el-table :data="pods" style="width: 100%" v-loading="podsLoading">
              <el-table-column prop="name" label="名称" min-width="180">
                <template #default="scope">
                  <el-link type="primary" @click="goToPodDetail(scope.row)">{{ scope.row.name }}</el-link>
                </template>
              </el-table-column>
              <el-table-column prop="namespace" label="命名空间" min-width="100"></el-table-column>
              <el-table-column label="状态" min-width="120" align="center">
                <template #default="scope">
                  <el-tag :type="getPodStatusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="pod_ip" label="IP" min-width="100"></el-table-column>
              <el-table-column label="容器" min-width="60" align="center">
                <template #default="scope">{{ scope.row.containers }}</template>
              </el-table-column>
              <el-table-column label="重启次数" min-width="60" align="center">
                <template #default="scope">
                  <span :style="{ color: scope.row.restarts > 0 ? '#EF4444' : '' }">{{ scope.row.restarts }}</span>
                </template>
              </el-table-column>
              <el-table-column label="CPU请求" min-width="80" align="right">
                <template #default="scope">{{ formatCpu(scope.row.cpu_request) }}</template>
              </el-table-column>
              <el-table-column label="内存请求" min-width="80" align="right">
                <template #default="scope">{{ formatMemory(scope.row.mem_request) }}</template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="140" align="center">
                <template #default="scope">{{ formatTime(scope.row.created_at) }}</template>
              </el-table-column>
            </el-table>
            <el-pagination
              v-model:current-page="podsPage"
              v-model:page-size="podsPageSize"
              :total="podsTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchPods"
              @current-change="fetchPods"
              style="margin-top: 16px; justify-content: flex-end;"
            />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="元数据" name="metadata">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="12">
              <el-card class="content-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><PriceTag /></el-icon>
                    <span class="card-title">标签</span>
                    <el-badge :value="Object.keys(node?.labels || {}).length" type="primary" style="margin-left: 12px;" />
                  </div>
                </template>
                <div class="labels-container">
                  <div v-for="(value, key) in node?.labels" :key="key" class="label-item">
                    <el-tag type="success" size="small">{{ key }}</el-tag>
                    <span class="label-value">{{ value }}</span>
                  </div>
                  <el-empty v-if="!node?.labels || Object.keys(node?.labels).length === 0" description="暂无标签" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card class="content-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Document /></el-icon>
                    <span class="card-title">注解</span>
                    <el-badge :value="Object.keys(node?.annotations || {}).length" type="info" style="margin-left: 12px;" />
                  </div>
                </template>
                <div class="annotations-container">
                  <div v-for="(value, key) in node?.annotations" :key="key" class="annotation-item">
                    <div class="annotation-key">{{ key }}</div>
                    <div class="annotation-value">{{ truncateText(value, 100) }}</div>
                  </div>
                  <el-empty v-if="!node?.annotations || Object.keys(node?.annotations).length === 0" description="暂无注解" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="事件" name="events">
        <div class="tab-content">
          <el-card class="content-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Bell /></el-icon>
                <span class="card-title">最近事件</span>
                <el-badge :value="events.length" type="warning" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="events" style="width: 100%" v-loading="eventsLoading">
              <el-table-column prop="type" label="类型" min-width="80" align="center">
                <template #default="scope">
                  <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'warning'" size="small">{{ scope.row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" min-width="100">
                <template #default="scope">
                  <el-tag type="info" size="small">{{ scope.row.reason }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="message" label="消息" min-width="200"></el-table-column>
              <el-table-column prop="count" label="次数" min-width="60" align="center">
                <template #default="scope">
                  <el-badge :value="scope.row.count" type="primary" />
                </template>
              </el-table-column>
              <el-table-column prop="source" label="来源" min-width="100"></el-table-column>
              <el-table-column prop="last_timestamp" label="发生时间" min-width="140" align="center">
                <template #default="scope">{{ formatTime(scope.row.last_timestamp) }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-if="events.length === 0 && !eventsLoading" description="暂无事件记录" :image-size="80" />
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Cpu, Coin, Grid, Timer, PriceTag, Document, InfoFilled, Warning, Bell } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()
const route = useRoute()

const activeTab = ref('status')
const node = ref(null)
const pods = ref([])
const events = ref([])

const podsPage = ref(1)
const podsPageSize = ref(10)
const podsTotal = ref(0)

const podsLoading = ref(false)
const eventsLoading = ref(false)

const nodeName = computed(() => route.params.name)
const cluster = computed(() => localStorage.getItem('k8s_selected_cluster') || route.query.cluster)

const runningPods = computed(() => {
  return pods.value.filter(p => p.status === 'Running').length
})

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const formatMemory = (bytes) => {
  if (!bytes) return '0'
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) return gb.toFixed(2) + ' Gi'
  const mb = bytes / (1024 * 1024)
  if (mb >= 1) return mb.toFixed(2) + ' Mi'
  return bytes + ' B'
}

const formatCpu = (millicores) => {
  if (!millicores) return '0'
  if (millicores >= 1000) return (millicores / 1000).toFixed(2) + ' 核'
  return millicores + ' m'
}

const formatDuration = (createdAt) => {
  if (!createdAt) return ''
  const now = new Date()
  const created = new Date(createdAt)
  const diff = now - created
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days > 0) return days + ' 天'
  const hours = Math.floor(diff / (1000 * 60 * 60))
  if (hours > 0) return hours + ' 小时'
  const minutes = Math.floor(diff / (1000 * 60))
  return minutes + ' 分钟'
}

const truncateText = (text, maxLen) => {
  if (!text) return ''
  if (text.length > maxLen) return text.substring(0, maxLen) + '...'
  return text
}

const getRoleType = (role) => {
  const types = { 'Master': 'danger', 'Control-Plane': 'warning', 'Worker': 'info' }
  return types[role] || 'info'
}

const getPodStatusType = (status) => {
  const map = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info',
    Terminating: 'warning',
    ContainersNotReady: 'warning',
    CrashLoopBackOff: 'danger'
  }
  return map[status] || 'info'
}

const getTaintEffectType = (effect) => {
  const map = { 'NoSchedule': 'danger', 'NoExecute': 'warning', 'PreferNoSchedule': 'info' }
  return map[effect] || 'info'
}

const isHealthy = (conditionType) => {
  const condition = node.value?.conditions?.find(c => c.type === conditionType)
  if (!condition) return conditionType === 'NetworkUnavailable' ? true : false
  if (conditionType === 'Ready') return condition.status === 'True'
  return condition.status === 'False'
}

const getConditionStatus = (conditionType) => {
  const condition = node.value?.conditions?.find(c => c.type === conditionType)
  if (!condition) return '-'
  return condition.status === 'True' ? 'True' : 'False'
}

const fetchNodeDetail = async () => {
  try {
    const res = await request.get(`/k8s/node/${nodeName.value}/detail`, {
      params: { cluster: cluster.value }
    })
    node.value = res.data
  } catch (error) {
    ElMessage.error('获取节点详情失败')
  }
}

const fetchPods = async () => {
  podsLoading.value = true
  try {
    const res = await request.get(`/k8s/node/${nodeName.value}/pods`, {
      params: {
        cluster: cluster.value,
        page: podsPage.value,
        pageSize: podsPageSize.value
      }
    })
    pods.value = res.data?.list || []
    podsTotal.value = res.data?.total || 0
  } catch (error) {
    ElMessage.error('获取容器组列表失败')
  } finally {
    podsLoading.value = false
  }
}

const fetchEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await request.get(`/k8s/node/${nodeName.value}/events`, {
      params: { cluster: cluster.value }
    })
    events.value = res.data || []
  } catch (error) {
    ElMessage.error('获取事件列表失败')
  } finally {
    eventsLoading.value = false
  }
}

const goToPodDetail = (pod) => {
  router.push({
    path: `/k8s/pod/${pod.name}`,
    query: { cluster: cluster.value, namespace: pod.namespace }
  })
}

const goBack = () => {
  router.push('/k8s/nodes')
}

onMounted(() => {
  fetchNodeDetail()
  fetchPods()
  fetchEvents()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.detail-tabs {
  margin-top: 16px;
}

.tab-content {
  padding: 16px 0;
}

.stat-card {
  text-align: center;
  padding: 20px;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.stat-icon {
  font-size: 20px;
  color: #409EFF;
}

.stat-title {
  font-size: 14px;
  color: #606266;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-sub {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 18px;
  color: #409EFF;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.content-card {
  margin-bottom: 24px;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 12px 0;
}

.label-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-value {
  font-size: 13px;
  color: #606266;
}

.annotations-container {
  padding: 12px 0;
}

.annotation-item {
  margin-bottom: 12px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

.annotation-key {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.annotation-value {
  font-size: 13px;
  color: #606266;
  word-break: break-all;
}

.health-card {
  padding: 20px;
  transition: all 0.3s;
}

.health-card.health-success {
  background: rgba(103, 194, 58, 0.05);
}

.health-card.health-danger {
  background: rgba(245, 108, 108, 0.05);
}

.health-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.health-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.health-desc-row {
  display: flex;
  align-items: center;
}

.health-desc-text {
  font-size: 12px;
  color: #909399;
}

.health-status-icon {
  font-size: 16px;
}

.health-svg {
  margin-left: auto;
}

.health-svg.success {
  color: #67C23A;
}

.health-svg.danger {
  color: #F56C6C;
}
</style>