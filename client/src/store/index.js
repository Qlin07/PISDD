import { defineStore } from 'pinia'
import { api } from '../api/http'
import { setupWebSocket, closeWs } from '../api/ws'

// 可用的主题定义(易于扩展: 在此追加新主题即可)
export const THEMES = [
  { key: 'night', label: '夜航' },
  { key: 'win11', label: 'Win11 暗色' }
]

const THEME_KEY = 'simplechat_theme'
export { THEME_KEY }
const DEFAULT_THEME = 'night'

// loadProfile 的 in-flight 标记, 并发调用去重(仅保留一个在途请求)
let loadProfilePromise = null

// 将主题应用到根元素 data-theme, 驱动 style.css 的变量切换
export function applyTheme(theme) {
  const root = document.documentElement
  if (theme) {
    root.setAttribute('data-theme', theme)
  } else {
    root.removeAttribute('data-theme')
  }
}

export function loadTheme() {
  const saved = localStorage.getItem(THEME_KEY)
  const valid = THEMES.some(t => t.key === saved)
  return valid ? saved : DEFAULT_THEME
}

// 安全地从 localStorage 解析用户对象; 异常/非对象/缺 user_id 一律返回 null,
// 避免页面上 store.user 为空对象或脏数据导致渲染崩溃。
function parseStoredUser() {
  const raw = localStorage.getItem('user')
  if (!raw) return null
  try {
    const u = JSON.parse(raw)
    return (u && typeof u === 'object' && typeof u.user_id !== 'undefined') ? u : null
  } catch (e) {
    return null
  }
}

export const useStore = defineStore('app', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: parseStoredUser(),
    conversations: [],
    contacts: [],
    groups: [],
    currentConv: null,
    messages: {},
    pendingApplies: [],
    connected: false,
    theme: loadTheme()
  }),
  getters: {
    isLogin: (s) => !!s.token
  },
  actions: {
    // 校验登录/注册响应的用户结构, 缺失则抛错避免半登录态
    applyAuth(_token, _user) {
      if (!_user || typeof _user !== 'object' || typeof _user.user_id === 'undefined') {
        throw new Error('登录/注册失败: 响应缺少有效的用户信息')
      }
      this.token = _token
      this.user = _user
      localStorage.setItem('token', _token)
      localStorage.setItem('user', JSON.stringify(_user))
    },
    async login(account, password, remember) {
      const { data } = await api.post('/auth/login', { account, password, remember })
      this.applyAuth(data.token, data.user)
      this.connectWs()
    },
    async register(account, nickname, password) {
      const { data } = await api.post('/auth/register', { account, nickname, password })
      this.applyAuth(data.token, data.user)
      this.connectWs()
    },
    // 会话恢复: 仅在已有 token 但本地 user 缺失/脏数据时, 从服务端拉取并补全
    // 带 in-flight 去重, 并发调用只发一次请求
    loadProfile() {
      if (loadProfilePromise) return loadProfilePromise
      loadProfilePromise = api.get('/user/profile')
        .then(({ data }) => {
          if (data && data.user_id) {
            this.user = data
            localStorage.setItem('user', JSON.stringify(data))
          }
          return data
        })
        .finally(() => { loadProfilePromise = null })
      return loadProfilePromise
    },
    logout() {
      api.post('/auth/logout').catch(() => {})
      closeWs()
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },
    connectWs() {
      if (!this.token) return
      setupWebSocket(
        (message) => this.onMessage(message),
        () => { this.connected = true },
        () => { this.connected = false }
      )
    },
    setTheme(key) {
      if (!THEMES.some(t => t.key === key)) return
      this.theme = key
      localStorage.setItem(THEME_KEY, key)
      applyTheme(key)
    },
    onMessage(payload) {
      // payload: {action, data}
      if (payload.action === 'message') {
        const msg = payload.data
        this.appendMessage(msg)
      } else if (payload.action === 'conv_update') {
        this.refreshConversations()
      }
    },
    appendMessage(msg) {
      const cid = msg.conversation_id
      if (!this.messages[cid]) this.messages[cid] = []
      // 去重: 避免乐观插入与 WS 回显重复
      if (!this.messages[cid].some(m => m.message_id === msg.message_id)) {
        this.messages[cid].push(msg)
      }
      this.refreshConversations()
    },
    async refreshConversations() {
      const { data } = await api.get('/conversations')
      this.conversations = data || []
    },
    async loadContacts() {
      const { data } = await api.get('/contacts')
      // 无好友时后端可能返回 null, 归一到 [] 避免列表渲染崩溃
      this.contacts = data || []
    },
    async loadGroups() {
      const { data } = await api.get('/groups/mine')
      this.groups = data || []
    },
    async loadPending() {
      const { data } = await api.get('/contacts/pending')
      this.pendingApplies = data || []
    },
    async openConversation(convId, beforeId = 0) {
      const { data } = await api.get(`/conversations/${convId}/messages?before_id=${beforeId}&limit=50`)
      // 统一消息源: 首次加载覆盖, 分页加载前置插入; 实时消息经 appendMessage 合并
      if (beforeId === 0) {
        this.messages[convId] = data
      } else {
        this.messages[convId] = [...data, ...(this.messages[convId] || [])]
      }
      return data
    },
    async ensureSingle(peerId) {
      const { data } = await api.post('/conversations/single', { peer_id: peerId })
      return data.conversation_id
    }
  }
})