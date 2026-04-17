<template>
  <el-dialog 
    v-model="visible" 
    :title="title" 
    width="70%" 
    destroy-on-close 
    @close="handleClose" 
    @open="handleOpen" 
    custom-class="aiops-dialog"
  >
    <div class="aiops-chat">
      <div class="chat-container">
        <div class="chat-messages" ref="messagesContainer">
          <div v-for="(msg, idx) in messages" :key="idx" :class="['message', msg.role]">
            <div class="message-avatar">
              <el-avatar v-if="msg.role === 'user'" :size="32">
                {{ username?.charAt(0) || 'U' }}
              </el-avatar>
              <el-avatar v-else :size="32" style="background: #409eff">
                AI
              </el-avatar>
            </div>
            <div class="message-content">
              <div class="message-header">
                <span class="role-name">{{ msg.role === 'user' ? '用户' : assistantName }}</span>
                <span class="message-time">{{ msg.time }}</span>
              </div>
              <div class="message-text" v-html="renderMarkdown(msg.content)"></div>
            </div>
          </div>

          <div v-if="loading" class="message assistant">
            <div class="message-avatar">
              <el-avatar :size="32" style="background: #409eff">AI</el-avatar>
            </div>
            <div class="message-content">
              <div class="message-header">
                <span class="role-name">{{ assistantName }}</span>
              </div>
              <div class="message-text typing">
                <span class="typing-indicator"></span>
                {{ loadingText }}
              </div>
            </div>
          </div>
        </div>

        <div class="chat-input">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="2"
            placeholder="输入问题继续对话..."
            :disabled="loading || !sessionId"
            @keyup.enter.ctrl="sendMessage"
          />
          <el-button type="primary" @click="sendMessage" :loading="loading" :disabled="!inputMessage.trim() || loading || !sessionId">
            发送
          </el-button>
        </div>
      </div>

      <div class="aiops-sidebar">
        <el-card shadow="hover">
          <template #header>
            <span>{{ configTitle }}</span>
          </template>
          <el-form label-width="80px" size="small">
            <el-form-item label="LLM模型">
              <el-select v-model="llmConfigId" placeholder="选择模型">
                <el-option label="默认配置" value="" />
                <el-option v-for="config in llmConfigs" :key="config.id" :label="config.name" :value="config.id" />
              </el-select>
            </el-form-item>
            <el-form-item v-for="item in configItems" :key="item.label" :label="item.label">
              <el-tag :type="item.type">{{ item.value }}</el-tag>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card v-if="sessionId" shadow="hover" style="margin-top: 16px">
          <template #header>
            <span>记录信息</span>
          </template>
          <el-descriptions :column="1" size="small">
            <el-descriptions-item label="记录ID">{{ sessionId }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="status === 'completed' ? 'success' : 'warning'">{{ status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="耗时">{{ duration }}ms</el-descriptions-item>
          </el-descriptions>
          <el-button type="primary" link @click="viewFullReport" style="margin-top: 8px">
            查看完整报告
          </el-button>
        </el-card>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { marked } from 'marked'
import request from '../../utils/request'

const props = defineProps({
  modelValue: Boolean,
  title: { type: String, default: 'AIOPS智能助手' },
  assistantName: { type: String, default: '智能助手' },
  configTitle: { type: String, default: '配置信息' },
  loadingText: { type: String, default: '正在分析中...' },
  configItems: { type: Array, default: () => [] },
  startMessage: { type: String, default: '' },
  streamEndpoint: { type: String, required: true },
  chatEndpoint: { type: String, required: true },
  streamParams: { type: Object, default: () => {} },
  ready: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue', 'resolve-cluster'])

const router = useRouter()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const llmConfigs = ref([])
const llmConfigId = ref('')
const messages = ref([])
const inputMessage = ref('')
const loading = ref(false)
const sessionId = ref(null)
const status = ref('')
const duration = ref(0)
const username = ref('')
const messagesContainer = ref(null)

marked.setOptions({
  breaks: true,
  gfm: true
})

const renderMarkdown = (content) => {
  if (!content) return ''
  return marked.parse(content)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const addMessage = (role, content) => {
  messages.value.push({
    role,
    content,
    time: new Date().toLocaleTimeString()
  })
  scrollToBottom()
}

const fetchLLMConfigs = async () => {
  try {
    const res = await request.get('/aiops/llm-config')
    llmConfigs.value = res.data || []
    if (llmConfigs.value.length > 0 && !llmConfigId.value) {
      const defaultConfig = llmConfigs.value.find(c => c.is_default)
      if (defaultConfig) {
        llmConfigId.value = defaultConfig.id
      } else {
        llmConfigId.value = llmConfigs.value[0].id
      }
    }
  } catch (error) {
    console.error('获取LLM配置失败')
  }
}

const handleOpen = async () => {
  messages.value = []
  sessionId.value = null
  status.value = ''
  duration.value = 0

  const userInfo = localStorage.getItem('userInfo')
  if (userInfo) {
    const user = JSON.parse(userInfo)
    username.value = user.username || ''
  }

  await fetchLLMConfigs()

  if (!props.ready) {
    emit('resolve-cluster')
    return
  }

  startStream()
}

const startStream = async () => {
  if (!props.ready) {
    return
  }

  loading.value = true

  if (props.startMessage) {
    addMessage('user', props.startMessage)
  }

  try {
    const token = localStorage.getItem('token')
    const params = {
      ...props.streamParams,
      llm_config_id: llmConfigId.value || 0
    }

    const response = await fetch(`/api${props.streamEndpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token,
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify(params)
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.msg || '请求失败')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let assistantContent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event:')) {
          continue
        }
        if (line.startsWith('data:')) {
          const data = line.slice(5).trim()
          if (data) {
            try {
              const event = JSON.parse(data)

              if (event.id) {
                sessionId.value = event.id
              }

              if (event.content) {
                assistantContent += event.content
                if (messages.value.length > 0 && messages.value[messages.value.length - 1].role === 'assistant') {
                  messages.value[messages.value.length - 1].content = assistantContent
                } else {
                  addMessage('assistant', assistantContent)
                }
                scrollToBottom()
              }

              if (event.duration) {
                duration.value = event.duration
              }

              if (event.message) {
                ElMessage.error(event.message)
              }
            } catch (e) {
              console.error('Parse SSE error:', e)
            }
          }
        }
      }
    }

    status.value = 'completed'
  } catch (error) {
    ElMessage.error(error.message || '处理失败')
    addMessage('assistant', '处理失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || loading.value || !sessionId.value) return

  const message = inputMessage.value.trim()
  inputMessage.value = ''
  addMessage('user', message)
  loading.value = true

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api${props.chatEndpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token,
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({
        record_id: sessionId.value,
        message: message,
        llm_config_id: llmConfigId.value || 0
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.msg || '请求失败')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let assistantContent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event:')) {
          continue
        }
        if (line.startsWith('data:')) {
          const data = line.slice(5).trim()
          if (data) {
            try {
              const event = JSON.parse(data)

              if (event.content) {
                assistantContent += event.content
                if (messages.value.length > 0 && messages.value[messages.value.length - 1].role === 'assistant') {
                  messages.value[messages.value.length - 1].content = assistantContent
                } else {
                  addMessage('assistant', assistantContent)
                }
                scrollToBottom()
              }
            } catch (e) {
              console.error('Parse SSE error:', e)
            }
          }
        }
      }
    }
  } catch (error) {
    ElMessage.error(error.message || '发送失败')
    addMessage('assistant', '发送失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const viewFullReport = () => {
  if (sessionId.value) {
    router.push(`/aiops/report/${sessionId.value}`)
    visible.value = false
  }
}

const handleClose = () => {
  visible.value = false
  messages.value = []
  sessionId.value = null
  inputMessage.value = ''
}

watch(() => props.ready, (newVal) => {
  if (newVal && visible.value && !loading.value && messages.value.length === 0) {
    startStream()
  }
})
</script>

<style scoped>
.aiops-chat {
  display: flex;
  gap: 16px;
  height: 600px;
  overflow: hidden;
}

.chat-container {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  overflow: hidden;
}

.chat-messages {
  flex: 1 1 0;
  padding: 16px;
  overflow-y: auto;
  overflow-x: hidden;
  background: #f5f7fa;
  min-height: 0;
  max-height: 100%;
}

.message {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  max-width: 80%;
  min-width: 0;
  overflow: hidden;
}

.message.user .message-content {
  text-align: right;
}

.message-header {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
}

.message.user .message-header {
  flex-direction: row-reverse;
}

.role-name {
  font-size: 12px;
  color: #909399;
}

.message-time {
  font-size: 12px;
  color: #c0c4cc;
}

.message-text {
  padding: 12px 16px;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  line-height: 1.6;
  word-wrap: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
  overflow-x: hidden;
}

.message-text :deep(pre) {
  background: #282c34;
  color: #abb2bf;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  max-width: 100%;
  margin: 8px 0;
}

.message-text :deep(code) {
  background: #282c34;
  color: #abb2bf;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.message-text :deep(pre code) {
  background: transparent;
  padding: 0;
}

.message-text :deep(p) {
  margin: 8px 0;
}

.message-text :deep(h1), .message-text :deep(h2), .message-text :deep(h3) {
  margin: 12px 0 8px 0;
}

.message-text :deep(ul), .message-text :deep(ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.message-text :deep(li) {
  margin: 4px 0;
}

.message-text :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
}

.message-text :deep(th), .message-text :deep(td) {
  border: 1px solid #dcdfe6;
  padding: 8px;
}

.message-text :deep(th) {
  background: #f5f7fa;
}

.message.user .message-text {
  background: #409eff;
  color: #fff;
}

.message.user .message-text :deep(pre) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.message.user .message-text :deep(code) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.message-text.typing {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.typing-indicator {
  width: 8px;
  height: 8px;
  background: #409eff;
  border-radius: 50%;
  animation: typing 1s infinite;
}

@keyframes typing {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

.chat-input {
  padding: 12px;
  border-top: 1px solid #e4e7ed;
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.chat-input .el-input {
  flex: 1;
}

.aiops-sidebar {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  min-height: 0;
}

.aiops-sidebar :deep(.el-card) {
  flex-shrink: 0;
}
</style>

<style>
.aiops-dialog {
  height: 80vh;
  max-height: 800px;
}

.aiops-dialog .el-dialog__body {
  height: calc(80vh - 120px);
  max-height: calc(800px - 120px);
  overflow: hidden;
  padding: 20px;
}
</style>