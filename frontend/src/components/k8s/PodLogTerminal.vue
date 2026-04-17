<template>
  <div class="log-terminal-container" :class="{ fullscreen: isFullscreen }">
    <div class="terminal-header">
      <div class="header-left">
        <div class="terminal-badge">
          <el-icon><Document /></el-icon>
          <span>日志</span>
        </div>
        <div class="terminal-meta">
          <span class="meta-item pod">{{ podName }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item ns">{{ namespace }}</span>
          <span v-if="container" class="meta-sep">·</span>
          <span v-if="container" class="meta-item container">{{ container }}</span>
        </div>
      </div>
      
      <div class="header-right">
        <div class="tool-group">
          <el-tooltip content="搜索" placement="bottom">
            <button class="tool-btn" :class="{ active: showSearch }" @click="toggleSearch">
              <el-icon><Search /></el-icon>
            </button>
          </el-tooltip>
          
          <el-tooltip content="自动滚动" placement="bottom">
            <button class="tool-btn" :class="{ active: autoScroll }" @click="toggleAutoScroll">
              <el-icon><Bottom /></el-icon>
            </button>
          </el-tooltip>
          
          <el-tooltip content="下载日志" placement="bottom">
            <button class="tool-btn" @click="downloadLogs">
              <el-icon><Download /></el-icon>
            </button>
          </el-tooltip>
          
          <el-tooltip content="清屏" placement="bottom">
            <button class="tool-btn" @click="clearTerminal">
              <el-icon><Delete /></el-icon>
            </button>
          </el-tooltip>
          
          <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
            <button class="tool-btn" @click="toggleFullscreen">
              <el-icon><component :is="isFullscreen ? 'Close' : 'FullScreen'" /></el-icon>
            </button>
          </el-tooltip>
        </div>
        
        <div class="tool-divider"></div>
        
        <div class="tool-group">
          <el-tooltip content="重新连接" placement="bottom">
            <button class="tool-btn" @click="reconnect" :disabled="connected">
              <el-icon><Refresh /></el-icon>
            </button>
          </el-tooltip>
          
          <el-tooltip :content="connected ? '停止' : '已停止'" placement="bottom">
            <button class="tool-btn stop" @click="disconnect" :disabled="!connected">
              <el-icon><VideoPause /></el-icon>
            </button>
          </el-tooltip>
        </div>
        
        <div class="tool-divider"></div>
        
        <div class="tool-group">
          <el-tooltip content="关闭" placement="bottom">
            <button class="tool-btn close-btn" @click="handleClose">
              <el-icon><Close /></el-icon>
            </button>
          </el-tooltip>
        </div>
      </div>
    </div>
    
    <div v-if="showSearch" class="search-bar">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索关键词..."
        clearable
        size="small"
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <span class="search-hint" v-if="searchKeyword">{{ searchMatches }} 处匹配</span>
    </div>
    
    <div ref="terminalRef" class="terminal-body" :style="terminalBodyStyle"></div>
    
    <div class="terminal-footer">
      <div class="footer-left">
        <span class="status-dot" :class="connected ? 'connected' : 'disconnected'"></span>
        <span class="status-text">{{ connected ? '实时连接' : '已断开' }}</span>
        <span v-if="autoScroll" class="status-badge scroll">自动滚动</span>
      </div>
      
      <div class="footer-right">
        <span class="stat">{{ lineCount }} 行</span>
        <span class="stat" v-if="connected">{{ connectionDuration }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document, Search, Download, Delete, Refresh, VideoPause,
  Bottom, FullScreen, Close
} from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { SearchAddon } from '@xterm/addon-search'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({
  clusterAlias: { type: String, required: true },
  namespace: { type: String, required: true },
  podName: { type: String, required: true },
  container: { type: String, default: '' },
  tailLines: { type: Number, default: 100 }
})

const emit = defineEmits(['close', 'fullscreen-change'])

const terminalRef = ref(null)
const connected = ref(false)
const lineCount = ref(0)
const terminal = ref(null)
const fitAddon = ref(null)
const searchAddon = ref(null)
const webLinksAddon = ref(null)
const ws = ref(null)
const reconnectAttempts = ref(0)
const maxReconnectAttempts = 3
const isTerminalReady = ref(false)
const logBuffer = ref([])
const autoScroll = ref(true)
const showSearch = ref(false)
const searchKeyword = ref('')
const searchMatches = ref(0)
const isFullscreen = ref(false)
const connectionStartTime = ref(null)
const connectionDuration = ref('00:00')
let durationTimer = null

const terminalBodyStyle = computed(() => {
  if (isFullscreen.value) {
    const headerHeight = 48
    const searchHeight = showSearch.value ? 44 : 0
    const footerHeight = 36
    const totalHeight = headerHeight + searchHeight + footerHeight
    return { height: `calc(100vh - ${totalHeight}px)` }
  }
  return {}
})

const initTerminal = () => {
  if (terminal.value) {
    try { terminal.value.dispose() } catch (e) { console.warn(e) }
    terminal.value = null
    fitAddon.value = null
    searchAddon.value = null
    webLinksAddon.value = null
    isTerminalReady.value = false
  }

  terminal.value = new Terminal({
    theme: {
      background: '#1a1b26',
      foreground: '#a9b1d6',
      cursor: '#7aa2f7',
      cursorAccent: '#1a1b26',
      selectionBackground: '#364a82',
      black: '#3d4148',
      red: '#f7768e',
      green: '#9ece6a',
      yellow: '#e0af68',
      blue: '#7aa2f7',
      magenta: '#bb9af7',
      cyan: '#7dcfff',
      white: '#a9b1d6',
      brightBlack: '#565f89',
      brightRed: '#f7768e',
      brightGreen: '#9ece6a',
      brightYellow: '#e0af68',
      brightBlue: '#7aa2f7',
      brightMagenta: '#bb9af7',
      brightCyan: '#7dcfff',
      brightWhite: '#c0caf5'
    },
    fontFamily: '"JetBrains Mono", "SF Mono", Monaco, Menlo, Consolas, monospace',
    fontSize: 13,
    lineHeight: 1.5,
    cursorBlink: true,
    scrollback: 50000,
    allowTransparency: true,
    convertEol: true
  })

  fitAddon.value = new FitAddon()
  searchAddon.value = new SearchAddon()
  webLinksAddon.value = new WebLinksAddon()
  
  terminal.value.loadAddon(fitAddon.value)
  terminal.value.loadAddon(searchAddon.value)
  terminal.value.loadAddon(webLinksAddon.value)
  terminal.value.open(terminalRef.value)
  
  nextTick(() => { if (fitAddon.value) fitAddon.value.fit() })
  isTerminalReady.value = true

  terminal.value.writeln('\x1b[38;5;117m▶ 日志终端\x1b[0m')
  terminal.value.writeln(`  Pod: ${props.podName}`)
  terminal.value.writeln(`  NS:  ${props.namespace}`)
  if (props.container) terminal.value.writeln(`  C:   ${props.container}`)
  terminal.value.writeln('')
}

const connectWebSocket = () => {
  if (ws.value) ws.value.close()

  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.host || 'localhost:8080'
  let wsUrl = `${wsProtocol}//${wsHost}/api/k8s/pod/log/stream?cluster=${props.clusterAlias}&namespace=${props.namespace}&pod=${props.podName}&tail_lines=${props.tailLines}&follow=true`
  
  if (props.container) wsUrl += `&container=${props.container}`
  const token = localStorage.getItem('token')
  if (token) wsUrl += `&token=${token}`

  if (terminal.value) terminal.value.writeln('\x1b[38;5;214m⟳ 连接中...\x1b[0m')

  try {
    ws.value = new WebSocket(wsUrl)
    ws.value.onopen = () => {
      connected.value = true
      reconnectAttempts.value = 0
      connectionStartTime.value = Date.now()
      startDurationTimer()
      if (terminal.value) terminal.value.writeln('\x1b[38;5;46m✓ 已连接\x1b[0m\n')
    }

    ws.value.onmessage = (event) => {
      const data = event.data
      if (data && terminal.value) {
        logBuffer.value.push(data)
        terminal.value.write(data)
        const lines = data.split('\n')
        lines.forEach(line => { if (line.trim()) lineCount.value++ })
        if (autoScroll.value) scrollToBottom()
        if (searchKeyword.value) updateSearchMatches()
      }
    }

    ws.value.onerror = () => {
      if (terminal.value) terminal.value.writeln('\x1b[38;5;196m✗ 连接错误\x1b[0m')
      connected.value = false
    }

    ws.value.onclose = (event) => {
      connected.value = false
      stopDurationTimer()
      if (!terminal.value) return
      terminal.value.writeln('')
      if (event.code === 1000) {
        terminal.value.writeln('\x1b[38;5;46m✓ 流已结束\x1b[0m')
        reconnectAttempts.value = maxReconnectAttempts
      } else {
        terminal.value.writeln(`\x1b[38;5;214m■ 已断开 (${event.code})\x1b[0m`)
        if (reconnectAttempts.value < maxReconnectAttempts) {
          reconnectAttempts.value++
          terminal.value.writeln(`\x1b[38;5;214m⟳ 3秒后重连 (${reconnectAttempts.value}/${maxReconnectAttempts})\x1b[0m`)
          setTimeout(reconnect, 3000)
        } else {
          terminal.value.writeln('\x1b[38;5;196m✗ 请手动重连\x1b[0m')
        }
      }
    }
  } catch (error) {
    if (terminal.value) terminal.value.writeln(`\x1b[38;5;196m✗ ${error.message}\x1b[0m`)
    connected.value = false
  }
}

const scrollToBottom = () => { if (terminal.value) terminal.value.scrollToBottom() }
const disconnect = () => {
  if (ws.value) { ws.value.close(1000); ws.value = null }
  connected.value = false
  stopDurationTimer()
  if (terminal.value) terminal.value.writeln('\x1b[38;5;214m■ 已停止\x1b[0m')
}
const reconnect = () => {
  if (terminal.value) terminal.value.writeln('\n\x1b[38;5;214m⟳ 重连中...\x1b[0m')
  connectWebSocket()
}
const clearTerminal = () => {
  if (terminal.value) {
    terminal.value.clear()
    lineCount.value = 0
    logBuffer.value = []
    terminal.value.writeln('\x1b[38;5;46m✓ 已清屏\x1b[0m')
  }
}
const toggleAutoScroll = () => {
  autoScroll.value = !autoScroll.value
  if (autoScroll.value) scrollToBottom()
}
const toggleSearch = () => {
  showSearch.value = !showSearch.value
  if (!showSearch.value) { searchKeyword.value = ''; searchMatches.value = 0 }
}
const handleSearch = () => {
  if (searchAddon.value && searchKeyword.value) {
    searchAddon.value.findNext(searchKeyword.value, { caseSensitive: false })
    updateSearchMatches()
  }
}
const updateSearchMatches = () => {
  if (searchAddon.value && searchKeyword.value) searchAddon.value.findNext(searchKeyword.value, { caseSensitive: false })
}
const downloadLogs = () => {
  const logs = logBuffer.value.join('')
  const blob = new Blob([logs], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${props.podName}-${props.container || 'all'}-${new Date().toISOString().slice(0, 19)}.log`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已下载')
}
const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
  emit('fullscreen-change', isFullscreen.value)
  setTimeout(() => {
    requestAnimationFrame(() => {
      if (fitAddon.value && isTerminalReady.value) {
        try { fitAddon.value.fit(); if (terminal.value) terminal.value.scrollToBottom() } catch (e) {}
      }
    })
  }, 100)
}

const handleClose = () => {
  disconnect()
  emit('close')
}

const startDurationTimer = () => {
  stopDurationTimer()
  durationTimer = setInterval(() => {
    if (connectionStartTime.value) {
      const elapsed = Math.floor((Date.now() - connectionStartTime.value) / 1000)
      const m = Math.floor(elapsed / 60).toString().padStart(2, '0')
      const s = (elapsed % 60).toString().padStart(2, '0')
      connectionDuration.value = `${m}:${s}`
    }
  }, 1000)
}
const stopDurationTimer = () => { if (durationTimer) { clearInterval(durationTimer); durationTimer = null } }
const handleResize = () => {
  if (fitAddon.value && isTerminalReady.value) try { fitAddon.value.fit() } catch (e) {}
}

onMounted(() => { initTerminal(); connectWebSocket(); window.addEventListener('resize', handleResize) })
onBeforeUnmount(() => {
  disconnect()
  stopDurationTimer()
  if (terminal.value && isTerminalReady.value) try { terminal.value.dispose() } catch (e) {}
  terminal.value = null
  fitAddon.value = null
  searchAddon.value = null
  webLinksAddon.value = null
  isTerminalReady.value = false
  window.removeEventListener('resize', handleResize)
})
watch(() => props.container, () => {
  if (connected.value) disconnect()
  clearTerminal()
  nextTick(() => { initTerminal(); connectWebSocket() })
})
</script>

<style scoped>
.log-terminal-container {
  display: flex;
  flex-direction: column;
  height: 600px;
  background: #1a1b26;
}

.log-terminal-container.fullscreen {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
  border-radius: 0;
}

.terminal-header {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(180deg, #24283b 0%, #1a1b26 100%);
  border-bottom: 1px solid rgba(255,255,255,0.06);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.terminal-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(122, 162, 247, 0.15);
  border-radius: 8px;
  color: #7aa2f7;
  font-size: 13px;
  font-weight: 500;
}

.terminal-badge .el-icon { font-size: 14px; }

.terminal-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #565f89;
  font-size: 12px;
}

.meta-item.pod { color: #c0caf5; font-weight: 500; }
.meta-item.ns { color: #7dcfff; }
.meta-item.container { color: #bb9af7; }
.meta-sep { opacity: 0.5; }

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tool-group {
  display: flex;
  gap: 4px;
}

.tool-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255,255,255,0.04);
  border: none;
  border-radius: 8px;
  color: #565f89;
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover {
  background: rgba(255,255,255,0.08);
  color: #c0caf5;
}

.tool-btn.active {
  background: rgba(158, 206, 106, 0.2);
  color: #9ece6a;
}

.tool-btn.stop:hover:not(:disabled) {
  background: rgba(247, 118, 142, 0.2);
  color: #f7768e;
}

.tool-btn.close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.tool-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.tool-divider {
  width: 1px;
  height: 24px;
  background: rgba(255,255,255,0.08);
}

.search-bar {
  flex-shrink: 0;
  padding: 10px 16px;
  background: #24283b;
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-bar :deep(.el-input__wrapper) {
  background: #1a1b26;
  border: none;
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.08);
  border-radius: 8px;
}

.search-bar :deep(.el-input__wrapper:hover) {
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.12);
}

.search-bar :deep(.el-input__wrapper.is-focus) {
  box-shadow: inset 0 0 0 1px #7aa2f7;
}

.search-bar :deep(.el-input__inner) { color: #a9b1d6; }
.search-bar :deep(.el-input__prefix) { color: #565f89; }

.search-hint {
  color: #565f89;
  font-size: 12px;
}

.terminal-body {
  flex: 1;
  min-height: 0;
  background: #1a1b26;
  overflow: hidden;
}

.terminal-body :deep(.xterm) { padding: 12px 16px; height: 100%; }
.terminal-body :deep(.xterm-viewport) { height: 100% !important; }
.terminal-body :deep(.xterm-screen) { height: 100% !important; }

.terminal-footer {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: linear-gradient(180deg, #1a1b26 0%, #24283b 100%);
  border-top: 1px solid rgba(255,255,255,0.06);
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.status-dot.connected { background: #9ece6a; }
.status-dot.disconnected { background: #565f89; animation: none; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  color: #565f89;
  font-size: 12px;
}

.status-badge.scroll {
  padding: 2px 8px;
  background: rgba(234, 175, 104, 0.15);
  border-radius: 4px;
  color: #e0af68;
  font-size: 11px;
}

.footer-right {
  display: flex;
  gap: 16px;
}

.stat {
  color: #565f89;
  font-size: 12px;
}
</style>