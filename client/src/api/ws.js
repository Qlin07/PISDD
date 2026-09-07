// 原生 WebSocket 客户端(对接服务端 /ws)
let ws = null
let connected = false
let reconnectTimer = null
let onMessageCb = null
let onOpenCb = null
let onCloseCb = null

export function setupWebSocket(onMessage, onOpen, onClose) {
  onMessageCb = onMessage
  onOpenCb = onOpen
  onCloseCb = onClose
  if (ws) return
  connect()
}

function connect() {
  closeWsInternal()
  const token = encodeURIComponent(localStorage.getItem('token') || '')
  // 走vite代理, 通过query鉴权
  const url = `/ws?token=${token}`
  ws = new WebSocket(url)
  ws.onopen = () => {
    connected = true
    onOpenCb && onOpenCb()
  }
  ws.onmessage = (e) => {
    try {
      const payload = JSON.parse(e.data)
      onMessageCb && onMessageCb(payload)
    } catch (err) { console.warn('ws parse error', err) }
  }
  ws.onclose = () => {
    connected = false
    onCloseCb && onCloseCb()
    scheduleReconnect()
  }
  ws.onerror = () => { ws && ws.close() }
}

function scheduleReconnect() {
  if (reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    if (localStorage.getItem('token')) connect()
  }, 3000)
}

export function send(obj) {
  if (ws && connected) {
    ws.send(JSON.stringify(obj))
    return true
  }
  return false
}

// 发送消息
export function sendWsMessage(payload) {
  return send({ action: 'send', ...payload })
}

// 已读回执
export function sendRead(conversationId, upToMsgId) {
  send({ action: 'read', conversation_id: conversationId, message_id: upToMsgId || 0 })
}

export function getWsConnected() { return connected }

function closeWsInternal() {
  if (ws) {
    try { ws.onclose = null; ws.close() } catch (e) {}
    ws = null
  }
}

export function closeWs() {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  closeWsInternal()
}