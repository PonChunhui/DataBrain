<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">LLM 模型配置</h1>
      <div class="header-actions">
        <el-button @click="fetchConfigs">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          添加配置
        </el-button>
      </div>
    </div>

    <div class="config-list" v-loading="loading">
      <el-card v-for="config in configs" :key="config.id" class="config-card" shadow="hover">
        <div class="card-header">
          <div class="header-left">
            <el-icon class="provider-icon"><Cpu /></el-icon>
            <span class="config-name">{{ config.name }}</span>
            <el-tag v-if="config.is_default" type="success" size="small" effect="dark">默认</el-tag>
            <el-tag :type="config.is_enabled ? 'success' : 'info'" size="small">
              {{ config.is_enabled ? '已启用' : '已禁用' }}
            </el-tag>
          </div>
          <div class="header-right">
            <el-button size="small" link @click="testConfig(config)" :loading="config.testing">
              测试连接
            </el-button>
            <el-button size="small" link type="primary" @click="showEditDialog(config)">编辑</el-button>
            <el-button size="small" link type="warning" @click="setDefault(config)" v-if="!config.is_default">
              设为默认
            </el-button>
            <el-button size="small" link :type="config.is_enabled ? 'warning' : 'success'" @click="toggleEnabled(config)">
              {{ config.is_enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button size="small" link type="danger" @click="deleteConfig(config)">删除</el-button>
          </div>
        </div>

        <div class="card-content">
          <el-row :gutter="16">
            <el-col :span="6">
              <div class="detail-item">
                <span class="detail-label">服务商</span>
                <span class="detail-value">{{ getProviderLabel(config.provider) }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="detail-item">
                <span class="detail-label">模型</span>
                <span class="detail-value">{{ config.model }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="detail-item">
                <span class="detail-label">Max Tokens</span>
                <span class="detail-value">{{ config.max_tokens }}</span>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="detail-item">
                <span class="detail-label">Temperature</span>
                <span class="detail-value">{{ config.temperature }}</span>
              </div>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <div class="detail-item">
                <span class="detail-label">API地址</span>
                <span class="detail-value">{{ config.base_url || getProviderDefaultURL(config.provider) }}</span>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="detail-item">
                <span class="detail-label">API Key</span>
                <span class="detail-value masked-key">{{ config.api_key }}</span>
              </div>
            </el-col>
          </el-row>
        </div>

        <div class="test-result" v-if="config.testResult">
          <el-divider />
          <div :class="['result-content', config.testResult.success ? 'success' : 'failed']">
            <el-icon v-if="config.testResult.success"><SuccessFilled /></el-icon>
            <el-icon v-else><CircleCloseFilled /></el-icon>
            <span>{{ config.testResult.success ? '连接成功' : '连接失败' }}</span>
          </div>
          <div class="result-details" v-if="config.testResult.success">
            <span>模型: {{ config.testResult.model }}</span>
            <span>响应时间: {{ config.testResult.response_time }}ms</span>
          </div>
          <div class="result-response" v-if="config.testResult.test_response">
            <span class="response-label">测试回复:</span>
            <span>{{ config.testResult.test_response }}</span>
          </div>
          <div class="result-error" v-if="config.testResult.error">
            <span class="error-label">错误信息:</span>
            <span>{{ config.testResult.error }}</span>
          </div>
        </div>
      </el-card>

      <el-empty v-if="configs.length === 0 && !loading" description="暂无LLM配置">
        <el-button type="primary" @click="showCreateDialog">添加配置</el-button>
      </el-empty>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑配置' : '添加配置'" width="600px" destroy-on-close>
      <el-form :model="formData" label-width="100px">
        <el-form-item label="配置名称" required>
          <el-input v-model="formData.name" placeholder="如: DeepSeek-Production" />
        </el-form-item>

        <el-form-item label="服务商" required>
          <el-select v-model="formData.provider" placeholder="选择服务商" @change="onProviderChange">
            <el-option label="DeepSeek - 高性价比国产模型" value="deepseek" />
            <el-option label="Claude - Anthropic强大推理" value="claude" />
            <el-option label="通义千问 - 阿里云百炼" value="qwen" />
            <el-option label="Gemini - Google多模态" value="gemini" />
            <el-option label="OpenAI - GPT系列" value="openai" />
          </el-select>
        </el-form-item>

        <el-form-item label="模型" required>
          <el-select v-model="formData.model" placeholder="选择模型">
            <el-option v-for="model in availableModels" :key="model" :label="model" :value="model" />
          </el-select>
        </el-form-item>

        <el-form-item label="API Key" required>
          <el-input v-model="formData.api_key" placeholder="输入API密钥" show-password />
        </el-form-item>

        <el-form-item label="API地址">
          <el-input v-model="formData.base_url" placeholder="可选，默认使用官方地址" />
        </el-form-item>

        <el-divider>高级参数</el-divider>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Max Tokens">
              <el-input-number v-model="formData.max_tokens" :min="1" :max="32000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Temperature">
              <el-input-number v-model="formData.temperature" :min="0" :max="2" :step="0.1" :precision="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="状态">
          <el-checkbox v-model="formData.is_enabled">启用此配置</el-checkbox>
          <el-checkbox v-model="formData.is_default">设为默认配置</el-checkbox>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button @click="testNewConfig" :loading="testing">测试连接</el-button>
        <el-button type="primary" @click="saveConfig" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Cpu, SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import request from '../../utils/request'

const loading = ref(false)
const configs = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const testing = ref(false)
const saving = ref(false)

const formData = ref({
  id: null,
  name: '',
  provider: '',
  model: '',
  api_key: '',
  base_url: '',
  max_tokens: 4000,
  temperature: 0.7,
  is_enabled: true,
  is_default: false
})

const providerModels = {
  deepseek: ['deepseek-chat', 'deepseek-coder'],
  claude: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
  qwen: ['qwen-max', 'qwen-plus', 'qwen-turbo'],
  gemini: ['gemini-pro', 'gemini-1.5-pro', 'gemini-1.5-flash'],
  openai: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo']
}

const availableModels = ref([])

const getProviderLabel = (provider) => {
  const labels = {
    deepseek: 'DeepSeek',
    claude: 'Claude',
    qwen: '通义千问',
    gemini: 'Gemini',
    openai: 'OpenAI'
  }
  return labels[provider] || provider
}

const getProviderDefaultURL = (provider) => {
  const urls = {
    deepseek: 'https://api.deepseek.com',
    claude: 'https://api.anthropic.com',
    qwen: 'https://dashscope.aliyuncs.com',
    gemini: 'https://generativelanguage.googleapis.com',
    openai: 'https://api.openai.com'
  }
  return urls[provider] || ''
}

const onProviderChange = () => {
  availableModels.value = providerModels[formData.value.provider] || []
  if (availableModels.value.length > 0 && !formData.value.model) {
    formData.value.model = availableModels.value[0]
  }
}

const fetchConfigs = async () => {
  loading.value = true
  try {
    const res = await request.get('/aiops/llm-config')
    configs.value = res.data || []
    configs.value.forEach(c => {
      c.testing = false
      c.testResult = null
    })
  } catch (error) {
    ElMessage.error('获取配置列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  formData.value = {
    id: null,
    name: '',
    provider: '',
    model: '',
    api_key: '',
    base_url: '',
    max_tokens: 4000,
    temperature: 0.7,
    is_enabled: true,
    is_default: false
  }
  availableModels.value = []
  dialogVisible.value = true
}

const showEditDialog = (config) => {
  isEdit.value = true
  formData.value = {
    id: config.id,
    name: config.name,
    provider: config.provider,
    model: config.model,
    api_key: '',
    base_url: config.base_url,
    max_tokens: config.max_tokens,
    temperature: config.temperature,
    is_enabled: config.is_enabled,
    is_default: config.is_default
  }
  availableModels.value = providerModels[config.provider] || []
  dialogVisible.value = true
}

const testConfig = async (config) => {
  config.testing = true
  config.testResult = null
  try {
    const res = await request.post(`/aiops/llm-config/${config.id}/test`)
    config.testResult = res.data
  } catch (error) {
    config.testResult = {
      success: false,
      error: error.response?.data?.msg || '测试失败'
    }
  } finally {
    config.testing = false
  }
}

const testNewConfig = async () => {
  if (!formData.value.api_key) {
    ElMessage.warning('请填写 API Key')
    return
  }
  
  if (!formData.value.provider || !formData.value.model) {
    ElMessage.warning('请选择服务商和模型')
    return
  }
  
  testing.value = true
  try {
    if (isEdit.value && formData.value.id) {
      const res = await request.post(`/aiops/llm-config/${formData.value.id}/test`)
      ElMessage.success('连接测试成功')
    } else {
      ElMessage.info('新配置需先保存后才能测试，请点击保存按钮')
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '测试失败')
  } finally {
    testing.value = false
  }
}

const saveConfig = async () => {
  if (!formData.value.name || !formData.value.provider || !formData.value.model) {
    ElMessage.warning('请填写必填字段')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await request.put(`/aiops/llm-config/${formData.value.id}`, formData.value)
      ElMessage.success('更新成功')
    } else {
      if (!formData.value.api_key) {
        ElMessage.warning('请填写API Key')
        saving.value = false
        return
      }
      await request.post('/aiops/llm-config', formData.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchConfigs()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '保存失败')
  } finally {
    saving.value = false
  }
}

const setDefault = async (config) => {
  try {
    await request.post(`/aiops/llm-config/${config.id}/set-default`)
    ElMessage.success('已设为默认')
    fetchConfigs()
  } catch (error) {
    ElMessage.error('设置失败')
  }
}

const toggleEnabled = async (config) => {
  try {
    await request.put(`/aiops/llm-config/${config.id}`, {
      is_enabled: !config.is_enabled
    })
    ElMessage.success(config.is_enabled ? '已禁用' : '已启用')
    fetchConfigs()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const deleteConfig = async (config) => {
  try {
    await ElMessageBox.confirm(`确定删除配置 "${config.name}"?`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await request.delete(`/aiops/llm-config/${config.id}`)
    ElMessage.success('删除成功')
    fetchConfigs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchConfigs()
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

.config-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 16px;
}

.config-card {
  transition: all 0.3s;
}

.config-card:hover {
  border-color: #409EFF;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.provider-icon {
  font-size: 20px;
  color: #409EFF;
}

.config-name {
  font-size: 16px;
  font-weight: 600;
}

.header-right {
  display: flex;
  gap: 8px;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: #909399;
}

.detail-value {
  font-size: 14px;
  color: #303133;
}

.masked-key {
  font-family: monospace;
  color: #606266;
}

.test-result {
  margin-top: 8px;
}

.result-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.result-content.success {
  color: #67C23A;
}

.result-content.failed {
  color: #F56C6C;
}

.result-details {
  display: flex;
  gap: 16px;
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}

.result-response,
.result-error {
  margin-top: 8px;
}

.response-label,
.error-label {
  font-weight: 500;
  color: #606266;
}

.result-error {
  color: #F56C6C;
}
</style>