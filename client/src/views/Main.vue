<template>
  <div class="main">
    <!-- 左侧导航(上三下一: 头像置顶, 聊天/联系人紧随, 设置置底) -->
    <div class="sidebar">
      <button class="me" @click="showProfile=true" aria-label="我的资料" :style="meStyle" type="button">{{ store.user?.avatar_url ? '' : (store.user?.nickname || '我').slice(0,1) }}</button>
      <div class="nav-items">
        <button class="nav-item" :class="{active: tab==='chat'}" @click="tab='chat'" aria-label="聊天" type="button">💬</button>
        <button class="nav-item" :class="{active: tab==='contacts'}" @click="tab='contacts'" aria-label="联系人" type="button">👥</button>
      </div>
      <button class="nav-item" :class="{active: tab==='settings'}" @click="tab='settings'" aria-label="设置" type="button">⚙️</button>
    </div>

    <!-- 中间列 -->
    <div class="mid-panel">
      <!-- 顶部: 搜索 + tab切换 -->
      <div class="mid-header">
        <input class="input search-input" v-model="searchKw" placeholder="搜索" aria-label="搜索" @focus="searchFocus=true" @blur="onSearchBlur" @input="onSearch" />
        <div class="search-result" v-if="searchKw">
          <div class="sr-title">联系人</div>
          <button v-for="u in searchUsers" :key="u.user_id" class="sr-item" @click="startChat(u)" type="button">
            <span class="sr-name">{{ u.nickname }}</span><span class="sr-acct">@{{ u.account }}</span>
          </button>
          <div class="sr-title">消息</div>
          <button v-for="m in searchMsgs" :key="m.message_id" class="sr-item" @click="jumpMessage(m)" type="button">
            <span class="sr-name">{{ m.sender_nickname }}</span>
            <span class="sr-text">{{ m.content }}</span>
            <span class="sr-conv">{{ m.conversation_name }}</span>
          </button>
          <div v-if="!searchUsers.length && !searchMsgs.length" class="empty">无结果</div>
          <button class="sr-close" @click="clearSearch" type="button">× 收起</button>
        </div>
        <div class="mid-tabs">
          <button class="tab" :class="{active: convTab==='msg'}" @click="convTab='msg'" type="button">消息</button>
          <button class="tab" :class="{active: convTab==='group'}" @click="convTab='group';loadGroups()" type="button">群聊</button>
        </div>
      </div>

      <!-- 会话/群列表 -->
      <div class="mid-list" @click="searchFocus=false">
        <template v-if="convTab==='msg'">
          <ConvList :conversations="store.conversations"
            :current="store.currentConv"
            @select="openConv" @create-group="showCreateGroup=true" />
        </template>
        <template v-else>
          <GroupList :groups="store.groups" @select="openGroup" @create-group="showCreateGroup=true" />
        </template>
      </div>
      <!-- 联系人tab -->
      <div class="mid-list" v-if="tab==='contacts'">
        <ContactList :friends="store.contacts" :pending="store.pendingApplies"
          @chat="startChat" @refresh="refreshAll" />
      </div>
    </div>

    <!-- 右侧内容 -->
    <div class="right-panel">
      <!-- 设置页: 独立于会话状态, 始终可访问 -->
      <SettingsPanel v-if="tab==='settings'" :store="store" @logout="logout" />
      <!-- 聊天页 -->
      <template v-else>
        <ChatWindow v-if="store.currentConv" :store="store" :wsConnected="store.connected" />
        <div v-else class="right-empty">
          <div class="logo-big">简聊</div>
          <p>选择一个会话开始聊天</p>
        </div>
      </template>
    </div>

    <!-- 建群弹窗 -->
    <CreateGroupModal v-if="showCreateGroup" :friends="store.contacts" @close="showCreateGroup=false" @created="onGroupCreated" />
    <!-- 资料编辑弹窗 -->
    <ProfileModal v-if="showProfile" :store="store" @close="showProfile=false" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from '../store'
import { api } from '../api/http'
import { sendWsMessage } from '../api/ws'
import ConvList from '../components/ConvList.vue'
import GroupList from '../components/GroupList.vue'
import ContactList from '../components/ContactList.vue'
import ChatWindow from '../components/ChatWindow.vue'
import SettingsPanel from '../components/SettingsPanel.vue'
import CreateGroupModal from '../components/CreateGroupModal.vue'
import ProfileModal from '../components/ProfileModal.vue'

const store = useStore()
const router = useRouter()

// 侧边栏头像: 有头像图则显示图片(保留底色兜底), 否则显示渐变+首字母
const meStyle = computed(() => ({
  background: store.user?.avatar_url
    ? `var(--primary) url(${store.user.avatar_url}) center/cover no-repeat`
    : 'var(--primary)'
}))

const tab = ref('chat')
const convTab = ref('msg')
const showCreateGroup = ref(false)
const showProfile = ref(false)
const searchKw = ref('')
const searchFocus = ref(false)
const searchUsers = ref([])
const searchMsgs = ref([])

let searchTimer = null

onMounted(() => { ensureSession(); refreshAll() })
watch(() => store.token, (v) => { if (v) refreshAll() })

// 有 token 但本地 user 缺失/脏数据时, 从服务端恢复, 避免渲染崩溃
async function ensureSession() {
  if (store.token && !store.user) {
    try {
      await store.loadProfile()
    } catch (e) {
      store.logout()
      router.push('/login')
    }
  }
}

function refreshAll() {
  store.refreshConversations().catch(() => {})
  store.loadContacts().catch(() => {})
  store.loadGroups().catch(() => {})
  store.loadPending().catch(() => {})
}

async function loadGroups() {
  await store.loadGroups().catch(() => {})
}

function openConv(conv) {
  store.currentConv = conv
  markRead(conv)
}

async function openGroup(group) {
  const convId = await api.get(`/groups/${group.group_id}`).then(r => r.data).catch(() => null)
  // 群会话ID由群消息返回; 用例走消息: 直接请求会话接口定位
  findGroupConv(group)
}

async function findGroupConv(group) {
  // 通过历史消息兜底: 群不提供直接会话ID, 使用约定: 群会话与群ID不同.
  // 简化: 向 /conversations 查找 type=group 且 group 匹配
  const list = store.conversations
  const found = list.find(c => c.group_id === group.group_id)
  if (found) { openConv(found) }
  else { store.ensureSingle(0) } // no-op
}

async function startChat(user) {
  const convId = await store.ensureSingle(user.user_id).catch(() => null)
  if (!convId) return
  searchFocus.value = false
  searchKw.value = ''
  // 构造会话对象
  const conv = {
    conversation_id: convId,
    type: 0,
    display_name: user.nickname,
    avatar: user.avatar_url,
    peer_user_id: user.user_id
  }
  store.currentConv = conv
}

function markRead(conv) {
  api.post(`/conversations/${conv.conversation_id}/read`).catch(() => {})
}

async function onSearch() {
  searchFocus.value = true
  clearTimeout(searchTimer)
  searchTimer = setTimeout(async () => {
    if (!searchKw.value) { searchUsers.value = []; searchMsgs.value = []; return }
    const kw = searchKw.value
    try {
      let { data: users } = await api.get(`/contacts/search?keyword=${encodeURIComponent(kw)}`)
      searchUsers.value = users || []
      let { data: msgs } = await api.get(`/search/messages?keyword=${encodeURIComponent(kw)}`)
      searchMsgs.value = msgs || []
    } catch (e) {}
  }, 400)
}

// 失焦后延迟收起, 给点击搜索结果留出时间(mousedown 优先已阻止默认转跳)
function onSearchBlur() {
  setTimeout(() => { searchFocus.value = false }, 120)
}

function clearSearch() {
  searchFocus.value = false
  searchKw.value = ''
  searchUsers.value = []
  searchMsgs.value = []
}

function jumpMessage(m) {
  // 定位到该消息所在会话
  const conv = store.conversations.find(c => c.conversation_id === m.conversation_id)
  if (conv) {
    store.currentConv = conv
    setTimeout(() => {
      const el = document.getElementById('msg-' + m.message_id)
      el && el.scrollIntoView({ block: 'center' })
    }, 300)
  }
  searchFocus.value = false
  clearSearch()
}

function onGroupCreated(g) {
  showCreateGroup.value = false
  store.loadGroups()
  store.refreshConversations()
}

function logout() {
  store.logout()
  router.push('/login')
}
</script>

<style scoped>
.main { display:flex; height:100vh; }
.sidebar { width:64px; background:var(--surface); border-right:1px solid var(--border-soft);
  display:flex; flex-direction:column; align-items:center; padding:12px 0; }
.nav-items { flex:1; display:flex; flex-direction:column; gap:10px; width:100%; align-items:center; justify-content:flex-start; padding-top:10px; }
.nav-item { width:44px; height:44px; display:flex; align-items:center; justify-content:center;
  font-size:21px; border-radius:12px; cursor:pointer; color:var(--text-2); border:none; background:transparent; transition: color .18s, background .18s; }
.nav-item:hover { background:var(--surface-3); color:var(--text); }
.nav-item:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; }
.nav-item.active { background:var(--primary); color:var(--primary-ink); box-shadow: 0 0 14px var(--primary-glow); }
.me { width:40px; height:40px; border-radius:50%; background:linear-gradient(135deg,var(--primary),var(--primary-dim)); color:#fff;
  display:flex; align-items:center; justify-content:center; cursor:pointer; font-size:15px; border:none; font-weight:600; }
.me:hover { opacity:.92; }
.me:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.mid-panel { width:280px; background:var(--surface); border-right:1px solid var(--border-soft); display:flex; flex-direction:column; }
.mid-header { padding:12px; border-bottom:1px solid var(--border-soft); position:relative; }
.search-input { height:34px; }
.search-result { position:absolute; top:52px; left:12px; right:12px; background:var(--surface-2);
  border:1px solid var(--border); border-radius:12px; box-shadow:0 8px 24px var(--shadow), 0 0 0 1px var(--border-soft);
  z-index:20; max-height:360px; overflow:auto; padding:8px; }
.sr-title { font-size:12px; color:var(--text-2); margin:8px 4px 2px; }
.sr-item { display:flex; gap:8px; padding:9px 8px; border-radius:8px; cursor:pointer; font-size:13px;
  width:100%; text-align:left; border:none; background:transparent; font:inherit; color:inherit; transition: background .15s; }
.sr-item:hover { background:var(--surface-3); }
.sr-item:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; }
.sr-name { color:var(--primary); font-weight:600; }
.sr-acct { color:var(--text-3); }
.sr-text { color:var(--text); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width:120px; }
.sr-conv { margin-left:auto; color:var(--text-3); font-size:12px; }
.sr-close { text-align:center; color:var(--primary); cursor:pointer; padding:7px; font-size:13px;
  border:none; background:transparent; font:inherit; width:100%; border-radius:8px; }
.sr-close:hover { background:var(--surface-3); }
.sr-close:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; }
.mid-tabs { display:flex; gap:18px; margin-top:12px; }
.tab { font-size:14px; color:var(--text-2); cursor:pointer; padding-bottom:5px;
  border:none; background:transparent; font:inherit; transition: color .18s; }
.tab:hover { color:var(--text); }
.tab:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.tab.active { color:var(--primary); border-bottom:2px solid var(--primary); font-weight:600; }
.mid-list { flex:1; overflow:auto; }
.right-panel { flex:1; background:var(--bg); position:relative; overflow:hidden; }
.right-empty { height:100%; display:flex; flex-direction:column; align-items:center; justify-content:center; color:var(--text-3); }
.logo-big { font-size:52px; color:var(--primary); margin-bottom:12px;
  text-shadow: 0 0 24px var(--primary-glow); }
</style>