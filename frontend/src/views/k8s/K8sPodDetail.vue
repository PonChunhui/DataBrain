<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">{{ podName }}</h1>
        <el-tag :type="getStatusType(pod?.status)" size="small" style="margin-left: 12px;">
          {{ pod?.status }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showDiagnosticDialog">
          <el-icon><Cpu /></el-icon>
          智能诊断
        </el-button>
        <el-button type="danger" @click="deletePod">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="资源状态" name="status">
        <div class="tab-content">
          <el-row :gutter="24">
            <el-col :span="8">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><InfoFilled /></el-icon>
                    <span class="card-title">基本信息</span>
                  </div>
                </template>
                <el-descriptions :column="1" border size="small">
                  <el-descriptions-item label="状态">
                    <el-tag :type="getStatusType(pod?.status)" size="small">{{ pod?.status }}</el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="Pod IP">
                    <span class="mono-text">{{ pod?.pod_ip || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="Host IP">
                    <span class="mono-text">{{ pod?.host_ip || '-' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="重启次数">
                    <span :style="{ color: pod?.restarts > 0 ? '#EF4444' : '#67C23A', fontWeight: 600 }">{{ pod?.restarts }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="服务账户">{{ pod?.service_account || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="QoS级别">{{ pod?.qos_class || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="创建时间">{{ formatTime(pod?.created_at) }}</el-descriptions-item>
                  <el-descriptions-item label="启动时间">{{ formatTime(pod?.start_time) }}</el-descriptions-item>
                </el-descriptions>
                <div v-if="pod?.message" style="margin-top: 12px; padding: 12px; background: #fff3cd; border-radius: 4px;">
                  <strong>消息:</strong> {{ pod?.message }}
                </div>
                <div v-if="pod?.reason" style="margin-top: 8px; padding: 12px; background: #f8d7da; border-radius: 4px;">
                  <strong>原因:</strong> {{ pod?.reason }}
                </div>
              </el-card>
            </el-col>

            <el-col :span="16">
              <el-card class="stat-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Box /></el-icon>
                    <span class="card-title">容器状态</span>
                    <el-badge :value="pod?.containers?.length || 0" type="primary" style="margin-left: 12px;" />
                  </div>
                </template>
                <el-table :data="pod?.containers || []" style="width: 100%">
                  <el-table-column prop="name" label="名称" min-width="120"></el-table-column>
                  <el-table-column prop="image" label="镜像" min-width="180">
                    <template #default="scope">
                      <el-tag type="info" size="small">{{ scope.row.image }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="状态" width="100" align="center">
                    <template #default="scope">
                      <el-tag :type="getContainerStatusType(scope.row.status?.status)" size="small">
                        {{ scope.row.status?.status }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="就绪" width="70" align="center">
                    <template #default="scope">
                      <el-icon v-if="scope.row.ready" color="#67C23A"><SuccessFilled /></el-icon>
                      <el-icon v-else color="#F56C6C"><CircleCloseFilled /></el-icon>
                    </template>
                  </el-table-column>
                  <el-table-column prop="restarts" label="重启" width="70" align="center">
                    <template #default="scope">
                      <span :style="{ color: scope.row.restarts > 0 ? '#EF4444' : '' }">{{ scope.row.restarts }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="140" align="center" fixed="right">
                    <template #default="scope">
                      <el-button type="primary" size="small" link @click="showContainerLogs(scope.row.name)">日志</el-button>
                      <el-button type="success" size="small" link @click="showContainerTerminal(scope.row.name)">终端</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>

          <el-card class="content-card section-gap" shadow="hover" v-if="pod?.volumes && pod?.volumes.length > 0">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Folder /></el-icon>
                <span class="card-title">存储卷</span>
                <el-badge :value="pod?.volumes?.length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="pod?.volumes || []" style="width: 100%">
              <el-table-column type="expand">
                <template #default="scope">
                  <div v-if="scope.row.mounts && scope.row.mounts.length > 0" class="mounts-expand">
                    <el-table :data="scope.row.mounts" size="small" style="width: 100%">
                      <el-table-column prop="container" label="容器" width="150"></el-table-column>
                      <el-table-column prop="mount_path" label="挂载路径" min-width="200">
                        <template #default="subScope">
                          <span class="mono-text">{{ subScope.row.mount_path }}</span>
                        </template>
                      </el-table-column>
                      <el-table-column prop="sub_path" label="子路径" width="150">
                        <template #default="subScope">
                          <span v-if="subScope.row.sub_path">{{ subScope.row.sub_path }}</span>
                          <span v-else style="color: #909399;">-</span>
                        </template>
                      </el-table-column>
                      <el-table-column label="只读" width="80" align="center">
                        <template #default="subScope">
                          <el-tag :type="subScope.row.read_only ? 'warning' : 'success'" size="small">
                            {{ subScope.row.read_only ? '是' : '否' }}
                          </el-tag>
                        </template>
                      </el-table-column>
                    </el-table>
                  </div>
                  <div v-else class="mounts-empty">暂无挂载信息</div>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="名称" min-width="150"></el-table-column>
              <el-table-column prop="type" label="类型" width="120" align="center">
                <template #default="scope">
                  <el-tag :type="getVolumeType(scope.row.type)" size="small">{{ scope.row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="source" label="来源" min-width="200"></el-table-column>
              <el-table-column label="挂载" width="80" align="center">
                <template #default="scope">
                  <el-badge :value="scope.row.mounts?.length || 0" type="primary" />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="调度信息" name="schedule">
        <div class="tab-content">
          <el-card class="stat-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Location /></el-icon>
                <span class="card-title">调度详情</span>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="节点">{{ pod?.node_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="调度器">{{ pod?.scheduler_name || 'default-scheduler' }}</el-descriptions-item>
              <el-descriptions-item label="优先级类">{{ pod?.priority_class || '-' }}</el-descriptions-item>
              <el-descriptions-item label="QoS级别">{{ pod?.qos_class || '-' }}</el-descriptions-item>
              <el-descriptions-item label="推荐节点">{{ pod?.nominated_node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="服务账户">{{ pod?.service_account || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="stat-card" shadow="hover" style="margin-top: 24px;" v-if="pod?.tolerations && pod?.tolerations.length > 0">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Warning /></el-icon>
                <span class="card-title">容忍（Tolerations）</span>
                <el-badge :value="pod?.tolerations?.length" type="warning" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="pod?.tolerations || []" style="width: 100%">
              <el-table-column prop="key" label="键" min-width="150">
                <template #default="scope">{{ scope.row.key || '*' }}</template>
              </el-table-column>
              <el-table-column prop="operator" label="操作符" width="120" align="center">
                <template #default="scope">
                  <el-tag type="info" size="small">{{ scope.row.operator }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="value" label="值" min-width="120">
                <template #default="scope">{{ scope.row.value || '-' }}</template>
              </el-table-column>
              <el-table-column prop="effect" label="效果" width="150" align="center">
                <template #default="scope">
                  <el-tag :type="getTolerationEffectType(scope.row.effect)" size="small">{{ scope.row.effect }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card class="stat-card" shadow="hover" style="margin-top: 24px;" v-if="pod?.affinity">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Connection /></el-icon>
                <span class="card-title">亲和性（Affinity）</span>
              </div>
            </template>
            <div v-if="pod?.affinity?.node_affinity" class="affinity-section">
              <h4>节点亲和性</h4>
              <div v-if="pod?.affinity?.node_affinity?.required" class="affinity-item">
                <span class="affinity-label">必需:</span>
                <el-tag type="danger" size="small">硬性要求</el-tag>
              </div>
              <div v-if="pod?.affinity?.node_affinity?.preferred" class="affinity-item">
                <span class="affinity-label">偏好:</span>
                <el-tag type="warning" size="small">软性偏好</el-tag>
              </div>
            </div>
            <div v-if="pod?.affinity?.pod_affinity" class="affinity-section">
              <h4>Pod亲和性</h4>
              <el-tag type="success" size="small">倾向于与其他Pod共存</el-tag>
            </div>
            <div v-if="pod?.affinity?.pod_antiaffinity" class="affinity-section">
              <h4>Pod反亲和性</h4>
              <el-tag type="warning" size="small">倾向于与其他Pod分离</el-tag>
            </div>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="元数据" name="metadata">
        <div class="tab-content">
          <el-card class="metadata-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Document /></el-icon>
                <span class="card-title">基本信息</span>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ pod?.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ pod?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ pod?.status }}</el-descriptions-item>
              <el-descriptions-item label="节点">{{ pod?.node_name }}</el-descriptions-item>
              <el-descriptions-item label="Pod IP">{{ pod?.pod_ip }}</el-descriptions-item>
              <el-descriptions-item label="Host IP">{{ pod?.host_ip }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(pod?.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="服务账户">{{ pod?.service_account }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="metadata-card" shadow="hover" style="margin-top: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><PriceTag /></el-icon>
                <span class="card-title">标签（Labels）</span>
                <el-badge :value="Object.keys(pod?.labels || {}).length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <div class="labels-container" v-if="pod?.labels && Object.keys(pod?.labels).length > 0">
              <el-tag v-for="(value, key) in pod?.labels" :key="key" type="primary" size="default" class="label-tag">
                <span class="label-key">{{ key }}</span>=<span class="label-value">{{ value }}</span>
              </el-tag>
            </div>
            <el-empty v-else description="暂无标签" :image-size="60" />
          </el-card>

          <el-card class="metadata-card" shadow="hover" style="margin-top: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><EditPen /></el-icon>
                <span class="card-title">注解（Annotations）</span>
                <el-badge :value="Object.keys(pod?.annotations || {}).length" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <div class="annotations-container" v-if="pod?.annotations && Object.keys(pod?.annotations).length > 0">
              <div v-for="(value, key) in pod?.annotations" :key="key" class="annotation-item">
                <div class="annotation-key">{{ key }}</div>
                <div class="annotation-value">{{ value || '-' }}</div>
              </div>
            </div>
            <el-empty v-else description="暂无注解" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="监控" name="monitor">
        <div class="tab-content">
          <div class="monitor-header">
            <div class="monitor-controls">
              <div class="time-range-section">
                <span class="control-label">时间范围：</span>
                <el-select v-model="monitorDuration" @change="handleDurationChange" style="width: 120px;">
                  <el-option value="15" label="15分钟"></el-option>
                  <el-option value="30" label="30分钟"></el-option>
                  <el-option value="60" label="1小时"></el-option>
                  <el-option value="120" label="2小时"></el-option>
                  <el-option value="180" label="3小时"></el-option>
<el-option value="300" label="5小时"></el-option>
                <el-option value="480" label="8小时"></el-option>
                <el-option value="720" label="12小时"></el-option>
                  <el-option value="1440" label="1天"></el-option>
                  <el-option value="2880" label="2天"></el-option>
                  <el-option value="4320" label="3天"></el-option>
                  <el-option value="10080" label="7天"></el-option>
                  <el-option value="custom" label="自定义"></el-option>
                </el-select>
                
                <el-popover 
                  placement="bottom" 
                  :width="360" 
                  trigger="click"
                  v-model:visible="customTimePopoverVisible"
                  v-if="monitorDuration === 'custom'"
                >
                  <template #reference>
                    <el-button size="small" style="margin-left: 8px;">
                      <el-icon><Calendar /></el-icon>
                      设置时间
                    </el-button>
                  </template>
                  <div class="custom-time-panel">
                    <div class="custom-time-row">
                      <span class="custom-label">开始时间：</span>
                      <el-date-picker 
                        v-model="customStartTime" 
                        type="datetime" 
                        placeholder="选择开始时间"
                        format="YYYY-MM-DD HH:mm"
                        value-format="YYYY-MM-DD HH:mm:ss"
                        style="width: 200px;"
                      ></el-date-picker>
                    </div>
                    <div class="custom-time-row">
                      <span class="custom-label">结束时间：</span>
                      <el-date-picker 
                        v-model="customEndTime" 
                        type="datetime" 
                        placeholder="选择结束时间"
                        format="YYYY-MM-DD HH:mm"
                        value-format="YYYY-MM-DD HH:mm:ss"
                        style="width: 200px;"
                      ></el-date-picker>
                    </div>
                    <div class="custom-time-actions">
                      <el-button size="small" @click="applyCustomTime" type="primary">应用</el-button>
                      <el-button size="small" @click="customTimePopoverVisible = false">取消</el-button>
                    </div>
                  </div>
                </el-popover>
              </div>
              
              <div class="step-section">
                <span class="control-label">采样时间：</span>
                <el-select v-model="monitorStep" @change="fetchPrometheusMetrics" style="width: 100px;">
                  <el-option value="60" label="1分钟"></el-option>
                  <el-option value="120" label="2分钟"></el-option>
                  <el-option value="300" label="5分钟"></el-option>
                  <el-option value="600" label="10分钟"></el-option>
                  <el-option value="900" label="15分钟"></el-option>
                  <el-option value="1800" label="30分钟"></el-option>
                  <el-option value="3600" label="1小时"></el-option>
                </el-select>
              </div>
            </div>
            
            <el-button type="primary" size="small" @click="fetchPrometheusMetrics" :loading="prometheusLoading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>

          <el-alert v-if="prometheusError" type="error" :closable="false" style="margin-bottom: 16px">
            {{ prometheusError }}
          </el-alert>

          <el-row :gutter="24">
            <el-col :span="12">
              <el-card class="monitor-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Cpu /></el-icon>
                    <span class="card-title">CPU用量（{{ cpuUnit }}）</span>
                  </div>
                </template>
                <div ref="cpuChartRef" class="chart-container" v-loading="prometheusLoading"></div>
              </el-card>
            </el-col>

            <el-col :span="12">
              <el-card class="monitor-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Coin /></el-icon>
                    <span class="card-title">内存用量（Gi）</span>
                  </div>
                </template>
                <div ref="memChartRef" class="chart-container" v-loading="prometheusLoading"></div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="24" style="margin-top: 24px;">
            <el-col :span="12">
              <el-card class="monitor-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Upload /></el-icon>
                    <span class="card-title">出站流量（Kbps）</span>
                  </div>
                </template>
                <div ref="txChartRef" class="chart-container" v-loading="prometheusLoading"></div>
              </el-card>
            </el-col>

            <el-col :span="12">
              <el-card class="monitor-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <el-icon class="header-icon"><Download /></el-icon>
                    <span class="card-title">入站流量（Kbps）</span>
                  </div>
                </template>
                <div ref="rxChartRef" class="chart-container" v-loading="prometheusLoading"></div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="环境变量" name="env">
        <div class="tab-content">
          <el-card class="stat-card" shadow="hover" v-for="(container, idx) in pod?.containers" :key="idx" style="margin-bottom: 24px;">
            <template #header>
              <div class="card-header">
                <el-icon class="header-icon"><Setting /></el-icon>
                <span class="card-title">容器: {{ container.name }}</span>
                <el-badge :value="container.env?.length || 0" type="primary" style="margin-left: 12px;" />
              </div>
            </template>
            <el-table :data="container.env || []" style="width: 100%">
              <el-table-column prop="name" label="变量名" min-width="200">
                <template #default="scope">
                  <span class="mono-text env-name">{{ scope.row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="value" label="值" min-width="300">
                <template #default="scope">
                  <span v-if="scope.row.value.startsWith('[from')" class="env-ref">{{ scope.row.value }}</span>
                  <span v-else class="mono-text">{{ scope.row.value }}</span>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!container.env || container.env.length === 0" description="暂无环境变量" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="事件" name="events">
        <div class="tab-content">
          <el-card class="events-card" shadow="hover">
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
                  <el-tag :type="scope.row.type === 'Normal' ? 'success' : 'danger'" size="small">
                    {{ scope.row.type }}
                  </el-tag>
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
              <el-table-column prop="lastTimestamp" label="时间" min-width="140" align="center">
                <template #default="scope">
                  {{ formatTime(scope.row.lastTimestamp) }}
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="events.length === 0 && !eventsLoading" description="暂无事件记录" :image-size="80" />
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="logsDialogVisible" width="1000px" :fullscreen="logsFullscreen" destroy-on-close top="2vh" class="logs-dialog" :show-close="false">
      <PodLogTerminal
        v-if="logsDialogVisible && selectedContainer"
        :clusterAlias="cluster"
        :namespace="namespace"
        :podName="podName"
        :container="selectedContainer"
        :tailLines="500"
        @fullscreen-change="handleLogFullscreenChange"
        @close="logsDialogVisible = false"
      />
    </el-dialog>

    <el-dialog v-model="terminalDialogVisible" width="1000px" :fullscreen="terminalFullscreen" destroy-on-close top="2vh" class="terminal-dialog" :show-close="false">
      <PodExecTerminal
        v-if="terminalDialogVisible && terminalContainer"
        :clusterAlias="cluster"
        :namespace="namespace"
        :podName="podName"
        :container="terminalContainer"
        :shell="terminalShell"
        @fullscreen-change="handleTerminalFullscreenChange"
        @close="terminalDialogVisible = false"
      />
    </el-dialog>

    <DiagnosticDialog
      v-model="diagnosticDialogVisible"
      :clusterAlias="cluster"
      :namespace="namespace"
      resourceType="pod"
      :resourceName="podName"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete, Document, Monitor, InfoFilled, Box, Folder, Location, Warning, Connection, PriceTag, EditPen, Bell, DataLine, Setting, SuccessFilled, CircleCloseFilled, Refresh, Cpu, Upload, Download, Coin, Calendar } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import request from '../../utils/request'
import PodLogTerminal from '../../components/k8s/PodLogTerminal.vue'
import PodExecTerminal from '../../components/k8s/PodExecTerminal.vue'
import DiagnosticDialog from '../../components/aiops/DiagnosticDialog.vue'

const router = useRouter()
const route = useRoute()
const podName = ref(route.params.name)
const activeTab = ref('status')
const pod = ref(null)
const metrics = ref([])
const events = ref([])
const loading = ref(false)
const metricsLoading = ref(false)
const eventsLoading = ref(false)

const cluster = ref(route.query.cluster)
const namespace = ref(route.query.namespace || 'default')

const logsDialogVisible = ref(false)
const terminalDialogVisible = ref(false)
const logsFullscreen = ref(false)
const terminalFullscreen = ref(false)
const selectedContainer = ref('')
const terminalContainer = ref('')
const terminalShell = ref('/bin/sh')

const diagnosticDialogVisible = ref(false)

const monitorDuration = ref('480')
const monitorStep = ref('600')
const prometheusLoading = ref(false)
const prometheusError = ref('')
const prometheusMetrics = ref({})
const cpuUnit = ref('cores')
const cpuChartRef = ref(null)
const memChartRef = ref(null)
const txChartRef = ref(null)
const rxChartRef = ref(null)
const customTimePopoverVisible = ref(false)
const customStartTime = ref(null)
const customEndTime = ref(null)
let cpuChart = null
let memChart = null
let txChart = null
let rxChart = null

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getStatusType = (status) => {
  const types = {
    'Running': 'success',
    'Pending': 'warning',
    'Succeeded': 'info',
    'Failed': 'danger',
    'Unknown': 'info'
  }
  return types[status] || 'info'
}

const getContainerStatusType = (status) => {
  const types = {
    'Running': 'success',
    'Waiting': 'warning',
    'Terminated': 'danger',
    'Unknown': 'info'
  }
  return types[status] || 'info'
}

const getVolumeType = (type) => {
  const types = {
    'PVC': 'primary',
    'ConfigMap': 'success',
    'Secret': 'warning',
    'EmptyDir': 'info',
    'HostPath': 'danger',
    'Other': 'info'
  }
  return types[type] || 'info'
}

const getTolerationEffectType = (effect) => {
  const types = {
    'NoSchedule': 'warning',
    'PreferNoSchedule': 'info',
    'NoExecute': 'danger'
  }
  return types[effect] || 'info'
}

const fetchPodDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/k8s/pod/${podName.value}/detail`, {
      params: { cluster: cluster.value, namespace: namespace.value }
    })
    pod.value = res.data
  } catch (error) {
    ElMessage.error('获取Pod详情失败')
  } finally {
    loading.value = false
  }
}

const fetchMetrics = async () => {
  metricsLoading.value = true
  try {
    const res = await request.get(`/k8s/pod/${podName.value}/metrics`, {
      params: { cluster: cluster.value, namespace: namespace.value }
    })
    metrics.value = res.data || []
  } catch (error) {
    metrics.value = []
  } finally {
    metricsLoading.value = false
  }
}

const fetchEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await request.get(`/k8s/pod/${podName.value}/events`, {
      params: { cluster: cluster.value, namespace: namespace.value }
    })
    events.value = res.data || []
  } catch (error) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

const fetchPrometheusMetrics = async () => {
  prometheusLoading.value = true
  prometheusError.value = ''
  try {
    let params = {
      cluster: cluster.value,
      namespace: namespace.value,
      step: monitorStep.value
    }
    
    if (monitorDuration.value === 'custom' && customStartTime.value && customEndTime.value) {
      params.start = Math.floor(new Date(customStartTime.value).getTime() / 1000)
      params.end = Math.floor(new Date(customEndTime.value).getTime() / 1000)
    } else {
      params.duration = monitorDuration.value
    }
    
    const res = await request.get(`/k8s/pod/${podName.value}/prometheus`, { params })
    prometheusMetrics.value = res.data || {}
    console.log('Prometheus metrics:', prometheusMetrics.value)
    console.log('CPU data:', prometheusMetrics.value.cpu)
    console.log('Memory data:', prometheusMetrics.value.memory)
    await nextTick()
    renderCharts()
  } catch (error) {
    console.error('Prometheus error:', error)
    prometheusError.value = error.response?.data?.msg || error.message || '获取Prometheus指标失败'
  } finally {
    prometheusLoading.value = false
  }
}

const handleDurationChange = () => {
  if (monitorDuration.value !== 'custom') {
    fetchPrometheusMetrics()
  }
}

const applyCustomTime = () => {
  if (!customStartTime.value || !customEndTime.value) {
    ElMessage.warning('请选择开始和结束时间')
    return
  }
  
  const start = new Date(customStartTime.value)
  const end = new Date(customEndTime.value)
  
  if (start >= end) {
    ElMessage.warning('开始时间必须早于结束时间')
    return
  }
  
  customTimePopoverVisible.value = false
  fetchPrometheusMetrics()
}

const renderCharts = () => {
  setTimeout(() => {
    initCpuChart()
    initMemChart()
    initTxChart()
    initRxChart()
  }, 100)
}

const formatChartTime = (timestamp) => {
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const getChartOption = (title, unit, data, color) => {
  const times = []
  const values = []
  
  console.log('getChartOption called:', title, 'data:', data)
  
  let actualUnit = unit
  
  if (data && Array.isArray(data) && data.length > 0) {
    const containerData = data[0]
    console.log('containerData:', containerData)
    if (containerData && containerData.values && Array.isArray(containerData.values) && containerData.values.length > 0) {
      const rawValues = containerData.values.map(v => v.value || 0)
      const maxValue = Math.max(...rawValues)
      
      containerData.values.forEach(v => {
        times.push(formatChartTime(v.time))
        let value = v.value || 0
        
        if (unit === 'cores') {
          if (maxValue < 1) {
            value = value * 1000
            actualUnit = 'm'
          }
        } else if (unit === 'Gi') {
          value = value / (1024 * 1024 * 1024)
        } else if (unit === 'Kbps') {
          value = value / 1000
        }
        values.push(parseFloat(value.toFixed(2)))
      })
      
      if (unit === 'cores' && maxValue < 1) {
        cpuUnit.value = 'm'
      } else if (unit === 'cores') {
        cpuUnit.value = 'cores'
      }
    }
  }

  console.log('Chart times:', times.length, 'values:', values.length)

  return {
    title: {
      show: values.length === 0,
      text: '暂无数据',
      left: 'center',
      top: 'center',
      textStyle: {
        color: '#909399',
        fontSize: 14
      }
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        if (!params || params.length === 0) return ''
        const val = params[0]?.value || 0
        return `${params[0]?.axisValue}<br/>${title}: ${val} ${actualUnit}`
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        fontSize: 11,
        color: '#909399'
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        fontSize: 11,
        color: '#909399',
        formatter: `{value} ${actualUnit}`
      }
    },
    series: [{
      name: title,
      type: 'line',
      smooth: true,
      symbol: 'none',
      lineStyle: {
        width: 2,
        color: color
      },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: color + '40' },
          { offset: 1, color: color + '10' }
        ])
      },
      data: values
    }]
  }
}

const initCpuChart = () => {
  if (!cpuChartRef.value) return
  if (cpuChart) cpuChart.dispose()
  cpuChart = echarts.init(cpuChartRef.value)
  const option = getChartOption('CPU', 'cores', prometheusMetrics.value.cpu, '#409EFF')
  cpuChart.setOption(option)
}

const initMemChart = () => {
  if (!memChartRef.value) return
  if (memChart) memChart.dispose()
  memChart = echarts.init(memChartRef.value)
  const option = getChartOption('内存', 'Gi', prometheusMetrics.value.memory, '#67C23A')
  memChart.setOption(option)
}

const initTxChart = () => {
  if (!txChartRef.value) return
  if (txChart) txChart.dispose()
  txChart = echarts.init(txChartRef.value)
  const option = getChartOption('出站流量', 'Kbps', prometheusMetrics.value.network_tx, '#E6A23C')
  txChart.setOption(option)
}

const initRxChart = () => {
  if (!rxChartRef.value) return
  if (rxChart) rxChart.dispose()
  rxChart = echarts.init(rxChartRef.value)
  const option = getChartOption('入站流量', 'Kbps', prometheusMetrics.value.network_rx, '#F56C6C')
  rxChart.setOption(option)
}

const resizeCharts = () => {
  if (cpuChart) cpuChart.resize()
  if (memChart) memChart.resize()
  if (txChart) txChart.resize()
  if (rxChart) rxChart.resize()
}

watch(activeTab, (newTab) => {
  if (newTab === 'monitor') {
    fetchPrometheusMetrics()
  }
})

const showContainerLogs = (containerName) => {
  selectedContainer.value = containerName
  logsFullscreen.value = false
  logsDialogVisible.value = true
}

const showContainerTerminal = (containerName) => {
  terminalContainer.value = containerName
  terminalFullscreen.value = false
  terminalDialogVisible.value = true
}

const handleLogFullscreenChange = (fullscreen) => {
  logsFullscreen.value = fullscreen
}

const handleTerminalFullscreenChange = (fullscreen) => {
  terminalFullscreen.value = fullscreen
}

const showDiagnosticDialog = () => {
  diagnosticDialogVisible.value = true
}

const deletePod = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod ${podName.value} 吗?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await request.delete(`/k8s/pod/${podName.value}`, {
      params: { cluster: cluster.value, namespace: namespace.value }
    })
    ElMessage.success('删除成功')
    router.back()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchPodDetail()
  fetchMetrics()
  fetchEvents()
  window.addEventListener('resize', resizeCharts)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  if (cpuChart) cpuChart.dispose()
  if (memChart) memChart.dispose()
  if (txChart) txChart.dispose()
  if (rxChart) rxChart.dispose()
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
  gap: 16px;
}

.header-right {
  display: flex;
  gap: 12px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.detail-tabs {
  margin-top: 0;
}

.tab-content {
  padding: 0;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.monitor-controls {
  display: flex;
  align-items: center;
  gap: 24px;
}

.time-range-section,
.step-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.custom-time-panel {
  padding: 12px;
}

.custom-time-row {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.custom-label {
  font-size: 14px;
  color: #606266;
  width: 70px;
}

.custom-time-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.monitor-card {
  height: 100%;
}

.chart-container {
  height: 280px;
  width: 100%;
}

.stat-card {
  height: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 18px;
  color: var(--el-color-primary);
}

.card-title {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.mono-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: #409EFF;
  font-weight: 500;
}

.mono-text.limit {
  color: #E6A23C;
}

.section-gap {
  margin-top: 24px;
}

.mounts-expand {
  padding: 12px 20px;
  background: #f5f7fa;
}

.mounts-empty {
  padding: 12px 20px;
  color: #909399;
  text-align: center;
}

.metadata-card {
  margin-bottom: 0;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}

.label-tag {
  font-size: 14px;
}

.label-key {
  color: #606266;
  font-weight: 500;
}

.label-value {
  color: #409EFF;
  font-weight: 500;
}

.annotations-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 8px 0;
}

.annotation-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.annotation-key {
  font-size: 13px;
  color: #909399;
  font-weight: 500;
  font-family: 'Monaco', 'Menlo', monospace;
}

.annotation-value {
  font-size: 14px;
  color: #303133;
  word-break: break-all;
  white-space: pre-wrap;
}

.affinity-section {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.affinity-section h4 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 14px;
}

.affinity-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.affinity-label {
  font-size: 13px;
  color: #606266;
}

.env-name {
  color: #67C23A;
  font-weight: 600;
}

.env-ref {
  color: #E6A23C;
  font-style: italic;
}

.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.el-descriptions {
  --el-descriptions-item-bordered-label-background: #f5f7fa;
}
</style>