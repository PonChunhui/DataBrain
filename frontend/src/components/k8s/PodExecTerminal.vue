<template>
  <div class="exec-terminal-container" :class="{ fullscreen: isFullscreen }">
    <div class="terminal-header">
      <div class="header-left">
        <div class="terminal-badge exec">
          <el-icon><Monitor /></el-icon>
          <span>终端</span>
        </div>
        <div class="terminal-meta">
          <span class="meta-item pod">{{ podName }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item ns">{{ namespace }}</span>
          <span v-if="selectedContainer" class="meta-sep">·</span>
          <span v-if="selectedContainer" class="meta-item container">{{ selectedContainer }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item shell">{{ selectedShell }}</span>
        </div>
      </div>
      
      <div class="header-right">
        <div class="tool-group" v-if="containers.length > 1">
          <el-select 
            v-model="selectedContainer" 
            size="small"
            style="width: 120px"
            @change="handleContainerChange"
          >
            <el-option 
              v-for="c in containers" 
              :key="c.name" 
              :label="c.name" 
              :value="c.name"
            />
          </el-select>
        </div>
        
        <div class="tool-group">
          <el-select 
            v-model="selectedShell" 
            size="small"
            style="width: 100px"
            @change="handleShellChange"
          >
            <el-option label="sh" value="/bin/sh" />
            <el-option label="bash" value="/bin/bash" />
          </el-select>
        </div>
        
        <div class="tool-divider"></div>
        
        <div class="tool-group">
          <el-tooltip :content="connected ? '已连接' : '连接'" placement="bottom">
            <button class="tool-btn connect" :class="{ active: connected }" @click="connected ? disconnect() : connectWebSocket()">
              <el-icon><Connection /></el-icon>
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
          <el-tooltip content="断开" placement="bottom">
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
    
    <div ref="terminalRef" class="terminal-body" :style="terminalBodyStyle"></div>
    
    <div class="terminal-footer">
      <div class="footer-left">
        <span class="status-dot" :class="connected ? 'connected' : 'disconnected'"></span>
        <span class="status-text">{{ connected ? '已连接' : '未连接' }}</span>
      </div>
      
      <div class="footer-right">
        <span class="stat">{{ shell }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Monitor, Connection, FullScreen, Close, VideoPause } from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({
  clusterAlias: { type: String, required: true },
  namespace: { type: String, required: true },
  podName: { type: String, required: true },
  container: { type: String, default: '' },
  shell: { type: String, default: '/bin/sh' },
  containers: { type: Array, default: () => [] }
})

const emit = defineEmits(['close', 'fullscreen-change', 'container-change', 'shell-change'])

const selectedContainer = ref(props.container)
const selectedShell = ref(props.shell)

const terminalRef = ref(null)
const connected = ref(false)
const terminal = ref(null)
const fitAddon = ref(null)
const webLinksAddon = ref(null)
const ws = ref(null)
const isTerminalReady = ref(false)
const isFullscreen = ref(false)

const terminalBodyStyle = computed(() => {
  if (isFullscreen.value) {
    const headerHeight = 48
    const footerHeight = 36
    return { height: `calc(100vh - ${headerHeight + footerHeight}px)` }
  }
  return {}
})

const initTerminal = () => {
  if (terminal.value) {
    try { terminal.value.dispose() } catch (e) { console.warn(e) }
    terminal.value = null
    fitAddon.value = null
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
    fontSize: 14,
    lineHeight: 1.4,
    cursorBlink: true,
    cursorStyle: 'bar',
    cursorWidth: 1,
    scrollback: 10000,
    allowTransparency: true
  })

  fitAddon.value = new FitAddon()
  webLinksAddon.value = new WebLinksAddon()
  
  terminal.value.loadAddon(fitAddon.value)
  terminal.value.loadAddon(webLinksAddon.value)
  terminal.value.open(terminalRef.value)
  
  nextTick(() => { if (fitAddon.value) { fitAddon.value.fit(); sendResize() } })
  isTerminalReady.value = true

  if (terminal.value) {
    terminal.value.writeln('\x1b[38;5;117m▶ 交互终端\x1b[0m')
    terminal.value.writeln(`  Pod: ${props.podName}`)
    terminal.value.writeln(`  NS:  ${props.namespace}`)
    if (props.container) terminal.value.writeln(`  C:   ${props.container}`)
    terminal.value.writeln(`  Sh:  ${props.shell}`)
    terminal.value.writeln('')
    terminal.value.writeln('\x1b[38;5;214m点击连接按钮开始会话...\x1b[0m\n')

    terminal.value.onData((data) => { if (ws.value && connected.value) ws.value.send(data) })
    terminal.value.onResize(({ cols, rows }) => { sendResize(cols, rows) })
  }
}

const sendResize = (cols, rows) => {
  if (ws.value && connected.value && terminal.value) {
    const c = cols || terminal.value.cols
    const r = rows || terminal.value.rows
    ws.value.send(`\x01${c},${r}`)
  }
}

const connectWebSocket = () => {
  if (ws.value) ws.value.close()

  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.host || 'localhost:8080'
  let wsUrl = `${wsProtocol}//${wsHost}/api/k8s/pod/exec/stream?cluster=${props.clusterAlias}&namespace=${props.namespace}&pod=${props.podName}&shell=${selectedShell.value}`
  
  if (selectedContainer.value) wsUrl += `&container=${selectedContainer.value}`
  const token = localStorage.getItem('token')
  if (token) wsUrl += `&token=${token}`

  if (terminal.value) terminal.value.writeln('\x1b[38;5;214m⟳ 连接中...\x1b[0m')

  try {
    ws.value = new WebSocket(wsUrl)
    ws.value.binaryType = 'arraybuffer'

    ws.value.onopen = () => {
      connected.value = true
      if (terminal.value) {
        terminal.value.write('\x1b[2J\x1b[H')
        terminal.value.writeln('\x1b[38;5;46m✓ 终端已连接\x1b[0m')
        // terminal.value.writeln(`            Pod: ${props.podName} | Container: ${selectedContainer.value || props.container} | Shell: ${selectedShell.value}`)
      }
      nextTick(() => { if (fitAddon.value) { fitAddon.value.fit(); sendResize() } })
    }

    ws.value.onmessage = (event) => {
      if (terminal.value && isTerminalReady.value) {
        const data = event.data
        if (typeof data === 'string') terminal.value.write(data)
        else if (data instanceof ArrayBuffer) terminal.value.write(new TextDecoder().decode(data))
      }
    }

    ws.value.onerror = () => {
      if (terminal.value) terminal.value.writeln('\x1b[38;5;196m✗ 连接错误\x1b[0m')
      connected.value = false
    }

    ws.value.onclose = (event) => {
      connected.value = false
      if (terminal.value) {
        terminal.value.writeln('')
        if (event.code === 1000) terminal.value.writeln('\x1b[38;5;46m✓ 已关闭\x1b[0m')
        else terminal.value.writeln(`\x1b[38;5;196m✗ 断开 (${event.code})\x1b[0m`)
        terminal.value.writeln('\x1b[38;5;214m点击连接按钮重连\x1b[0m')
      }
    }
  } catch (error) {
    if (terminal.value) terminal.value.writeln(`\x1b[38;5;196m✗ ${error.message}\x1b[0m`)
    connected.value = false
  }
}

const disconnect = () => {
  if (ws.value) { ws.value.close(1000); ws.value = null }
  connected.value = false
  if (terminal.value) terminal.value.writeln('\x1b[38;5;214m■ 已断开\x1b[0m')
}

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
  emit('fullscreen-change', isFullscreen.value)
  nextTick(() => {
    setTimeout(() => {
      requestAnimationFrame(() => {
        if (fitAddon.value && isTerminalReady.value) {
          try { fitAddon.value.fit(); sendResize() } catch (e) {}
        }
      })
    }, 100)
  })
}

const handleClose = () => {
  disconnect()
  emit('close')
}

const handleContainerChange = () => {
  disconnect()
  emit('container-change', selectedContainer.value)
  nextTick(() => {
    setTimeout(() => connectWebSocket(), 500)
  })
}

const handleShellChange = () => {
  disconnect()
  emit('shell-change', selectedShell.value)
  nextTick(() => {
    setTimeout(() => connectWebSocket(), 500)
  })
}

const handleResize = () => {
  if (fitAddon.value && isTerminalReady.value) try { fitAddon.value.fit(); sendResize() } catch (e) {}
}

onMounted(() => { initTerminal(); window.addEventListener('resize', handleResize) })
onBeforeUnmount(() => {
  disconnect()
  if (terminal.value && isTerminalReady.value) try { terminal.value.dispose() } catch (e) {}
  terminal.value = null
  fitAddon.value = null
  webLinksAddon.value = null
  isTerminalReady.value = false
  window.removeEventListener('resize', handleResize)
})

watch(() => props.container, (val) => { if (val) selectedContainer.value = val })
watch(() => props.shell, (val) => { if (val) selectedShell.value = val })
</script>

<style scoped>
.exec-terminal-container {
  display: flex;
  flex-direction: column;
  height: 500px;
  background: #1a1b26;
}

.exec-terminal-container.fullscreen {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
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

.terminal-badge.exec {
  background: rgba(187, 154, 247, 0.15);
  color: #bb9af7;
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
.meta-item.shell { color: #e0af68; }
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

.tool-btn.connect.active {
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

.footer-right {
  display: flex;
  gap: 16px;
}

.stat {
  color: #565f89;
  font-size: 12px;
}
</style>