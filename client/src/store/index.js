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

export const useStore = defineStore('app', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null'),
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
    async login(account, password, remember) {
      const { data } = await api.post('/auth/login', { account, password, remember })
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      this.connectWs()
    },
    async register(account, nickname, password) {
      const { data } = await api.post('/auth/register', { account, nickname, password })
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      this.connectWs()
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
      this.conversations = data
    },
    async loadContacts() {
      const { data } = await api.get('/contacts')
      this.contacts = data
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