<template>
  <div class="chat">
    <div class="chat-header">
      <div class="title">{{ store.currentConv.display_name }}</div>
      <div class="sub">{{ store.connected ? '在线' : '连接中...' }}</div>
    </div>

    <!-- 消息流 -->
    <div class="chat-body" ref="bodyRef" @scroll="onScroll">
      <div v-if="hasMore && !firstLoad" class="load-wrap">
        <button class="load-more" @click="loadMore" type="button">加载更早消息</button>
      </div>
      <div v-for="m in messages" :key="m.message_id" :id="'msg-'+m.message_id"
        class="msg-row" :class="{mine: m.sender_id===store.user.user_id}">
        <div class="msg-bubble">
          <div v-if="m.sender_id!==store.user.user_id" class="m-name">{{ m.sender_nickname || '用户' }}</div>
          <template v-if="m.type===0">{{ m.content }}</template>
          <template v-else-if="m.type===1">
            <img class="msg-img" v-if="m.media_url" :src="m.media_url" :alt="m.content || '图片消息'" @click="previewImg(m.media_url)" />
            <div v-else class="m-file">[图片] {{ m.content }}</div>
          </template>
          <template v-else-if="m.type===2">
            <div class="m-file">
              📎 <a :href="m.media_url" target="_blank" rel="noopener noreferrer" download>{{ m.content || '文件下载' }}</a>
            </div>
          </template>
          <template v-else-if="m.type===4"><em class="sys">{{ m.content }}</em></template>
          <template v-else>{{ m.content }}</template>
          <div class="m-status" v-if="m.sender_id===store.user.user_id">
            {{ m.status===3?'已读':m.status===2?'已送达':m.status===1?'已发送':'发送中' }}
          </div>
        </div>
        <div class="m-time">{{ fmtTime(m.sent_time) }}</div>
      </div>
      <div v-if="firstLoad" class="empty">加载中…</div>
    </div>

    <!-- 输入区 -->
    <div class="chat-input">
      <div class="toolbar">
        <label class="icon-btn" aria-label="发送图片" title="发送图片">
          🖼 <input type="file" accept="image/*" style="display:none" @change="uploadImage" />
        </label>
        <label class="icon-btn" aria-label="发送文件" title="发送文件">
          📎 <input type="file" style="display:none" @change="uploadFile" />
        </label>
        <div class="emoji-wrap">
        <button ref="emojiBtn" class="icon-btn" @click="toggleEmoji" aria-label="表情" aria-haspopup="true" :aria-expanded="emojiOpen ? 'true':'false'" title="表情" type="button">😊</button>
        <div ref="emojiPanel" class="emoji-panel" v-if="emojiOpen" role="menu" aria-label="选择表情">
          <div class="emoji-viewport">
            <div class="emoji-track" :style="{ transform: 'translateX(' + (0 - emojiPage * EMOJI_VIEWPORT_W) + 'px)' }">
              <div v-for="(page, pi) in emojiPages" :key="pi" class="emoji-page" role="presentation">
                <button v-for="e in page" :key="e" class="emoji-item" @click="insertEmoji(e)" type="button" role="menuitem">{{ e }}</button>
              </div>
            </div>
          </div>
          <div v-if="emojiPages.length > 1" class="emoji-dots" role="tablist" aria-label="表情翻页">
            <button v-for="(p, pi) in emojiPages" :key="pi" class="emoji-dot"
              :class="{active: pi === emojiPage}" @click="emojiPage = pi" type="button" role="tab"
              :aria-label="'第' + (pi + 1) + '页'" :aria-selected="pi === emojiPage ? 'true' : 'false'"></button>
          </div>
        </div>
      </div>
      </div>
      <textarea class="input-box" v-model="draft" placeholder="输入消息，Ctrl+Enter 发送" aria-label="消息内容"
        @keydown.ctrl.enter="sendText" @keydown.enter.exact.prevent="sendText" spellcheck="false"></textarea>
      <div class="send-row">
        <span class="tip">支持 Ctrl+Enter 发送</span>
        <button class="btn btn-primary" @click="sendText" :disabled="!draft.trim()">发送</button>
      </div>
    </div>

    <div v-if="previewUrl" class="img-preview" @click.self="previewUrl=''">
      <img :src="previewUrl" :alt="'图片预览'" />
      <button class="close-x" @click="previewUrl=''" aria-label="关闭预览" type="button">×</button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted, onBeforeUnmount, computed } from 'vue'
import { api } from '../api/http'
import { sendWsMessage, sendRead } from '../api/ws'

const props = defineProps({ store: Object, wsConnected: Boolean })
const store = props.store

// 消息渲染统一以 store.messages[conversation_id] 为唯一数据源,
// 历史分页与 WS 实时消息都由 store 维护, 此处 watch 驱动 UI 刷新
const bodyRef = ref(null)
const draft = ref('')
const emojiOpen = ref(false)
const previewUrl = ref('')
const firstLoad = ref(true)
const hasMore = ref(true)
const beforeId = ref(0)

const messages = ref([])
const emojis = ['😀','😃','😄','😁','😆','🤣','😊','😇','🙂','😉','😍','🤩','😘','😜','🤪','😎','🥳','😭','😢','😡','🥰','😌','🤔','🤫','😴','👍','👎','👏','🙌','🙏','💪','🎉','🔥','❤️','💙','💚','💛','⭐','🌈','🍀','🎁','🍎','☕','⚡','✨','🎈','🕶','🐱','🐶']

// 表情面板: 每页最多 4 行, 每行 8 个 → 每页 32 个; 视口宽度(与 .emoji-viewport 一致)
const EMOJI_COLS = 8
const EMOJI_ROWS = 4
const PAGESIZE = EMOJI_COLS * EMOJI_ROWS
const EMOJI_VIEWPORT_W = 276
const emojiPage = ref(0)
// 将 emojis 切分为若干页
const emojiPages = computed(() => {
  const pages = []
  for (let i = 0; i < emojis.length; i += PAGESIZE) {
    pages.push(emojis.slice(i, i + PAGESIZE))
  }
  return pages
})

// 表情面板点击外部关闭所需 ref
const emojiBtn = ref(null)
const emojiPanel = ref(null)

// 当前会话消息
const cid = () => store.currentConv?.conversation_id
const convMessages = () => (cid() ? store.messages[cid()] || [] : [])

onMounted(() => loadMessages())
watch(() => store.currentConv?.conversation_id, () => loadMessages())
// 实时消息进来后同步本地视图(按 message_id 去重, 保持顺序)
watch(convMessages, (list) => {
  syncList(messages.value, list)
}, { deep: true })

// 将 store 消息合并进本地渲染列表
function syncList(local, remote) {
  if (remote.length === 0) {
    messages.value = []
    return
  }
  const seen = new Set()
  const merged = [...local]
  for (let i = 0; i < merged.length; i++) seen.add(merged[i].message_id)
  for (const m of remote) {
    // 已存在的消息更新其发送状态(已送达/已读), 避免重复行
    const idx = merged.findIndex(x => x.message_id === m.message_id)
    if (idx >= 0) {
      merged[idx] = m
    } else {
      seen.add(m.message_id)
      merged.push(m)
    }
  }
  const needScroll = messages.value !== merged
  messages.value = merged
  if (needScroll && firstLoad.value === false && autoScroll.value) {
    nextTick(() => scrollBottom())
  }
}

const autoScroll = ref(true)

async function loadMessages() {
  const convId = cid()
  if (!convId) { messages.value = []; return }
  firstLoad.value = true
  autoScroll.value = true
  try {
    const data = await store.openConversation(convId, 0)
    beforeId.value = (data[0] && data[0].message_id) || 0
    hasMore.value = data.length >= 50
    messages.value = store.messages[convId] || []
    await nextTick()
    scrollBottom()
    // 发送已读
    const list = messages.value
    sendRead(convId, list.length ? list[list.length - 1].message_id : 0)
  } catch (e) {
    console.warn(e)
  } finally {
    firstLoad.value = false
  }
}

async function loadMore() {
  const convId = cid()
  if (!convId || !hasMore.value || firstLoad.value) return
  const prevHeight = bodyRef.value ? bodyRef.value.scrollHeight : 0
  const before = beforeId.value
  const data = await store.openConversation(convId, before)
  if (data.length) {
    beforeId.value = data[0].message_id
  }
  hasMore.value = data.length >= 50
  await nextTick()
  // 保持滚动位置: 新内容插入顶部后, 回到原相对位置
  if (bodyRef.value) bodyRef.value.scrollTop = bodyRef.value.scrollHeight - prevHeight
}

function sendText() {
  const content = draft.value.trim()
  const convId = cid()
  if (!content || !convId) return
  sendWsMessage({ conversation_id: convId, type: 0, content })
  autoScroll.value = true
  draft.value = ''
  emojiOpen.value = false
  nextTick(() => scrollBottom())
}

async function uploadImage(e) {
  const file = e.target.files[0]
  if (!file) return
  const resp = await upload('/files/upload?type=image', file, 'image')
  if (resp) {
    sendWsMessage({ conversation_id: store.currentConv.conversation_id, type: 1, content: file.name, media_url: resp.file_url })
  }
  e.target.value = ''
}
async function uploadFile(e) {
  const file = e.target.files[0]
  if (!file) return
  const resp = await upload('/files/upload?type=file', file, 'file')
  if (resp) {
    sendWsMessage({ conversation_id: store.currentConv.conversation_id, type: 2, content: file.name, media_url: resp.file_url })
  }
  e.target.value = ''
}
async function upload(url, file, field) {
  const fd = new FormData()
  fd.append(field, file)
  try {
    const { data } = await api.post(url, fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    return data
  } catch (err) {
    alert(err.message)
    return null
  }
}

function scrollBottom() {
  bodyRef.value && (bodyRef.value.scrollTop = bodyRef.value.scrollHeight)
}
function onScroll() {
  if (bodyRef.value && bodyRef.value.scrollTop < 10 && hasMore.value) loadMore()
}
function toggleEmoji() { 
  emojiOpen.value = !emojiOpen.value
  emojiPage.value = 0
}
function insertEmoji(e) {
  draft.value += e
  // 选择表情后保留面板(便于连续插入), 通过点击外部关闭
}
function previewImg(url) { previewUrl.value = url }

// 点击面板外区域时关闭表情面板; 点击面板本身(选择/翻页)不关闭
function onDocPointerDown(ev) {
  if (!emojiOpen.value) return
  const panel = emojiPanel.value
  const btn = emojiBtn.value
  const inPanel = panel && panel.contains(ev.target)
  const inBtn = btn && btn.contains(ev.target)
  if (!inPanel && !inBtn) emojiOpen.value = false
}

onMounted(() => document.addEventListener('pointerdown', onDocPointerDown, true))
onBeforeUnmount(() => document.removeEventListener('pointerdown', onDocPointerDown, true))
function fmtTime(t) {
  if (!t) return ''
  const d = new Date(t)
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}
</script>

<style scoped>
.chat { height:100%; display:flex; flex-direction:column; }
.chat-header { padding:14px 18px; background:var(--surface); border-bottom:1px solid var(--border-soft); }
.title { font-size:16px; font-weight:600; color:var(--text); }
.sub { font-size:12px; color:var(--text-2); }
.chat-body { flex:1; overflow:auto; padding:18px 20px; background:var(--bg); }
.msg-row { display:flex; flex-direction:column; margin-bottom:14px; }
.msg-row.mine { align-items:flex-end; }
.msg-row.mine .msg-bubble { background:var(--bubble-self-bg); color:var(--bubble-self-color); align-self:flex-end; }
.msg-bubble { max-width:70%; align-self:flex-start; background:var(--surface-2); padding:10px 14px;
  border-radius:14px; font-size:14px; line-height:1.6; word-break:break-word; color:var(--text);
  border:1px solid var(--border-soft); }
.m-name { font-size:12px; color:var(--primary); margin-bottom:4px; font-weight:600; }
.m-time { font-size:11px; color:var(--text-3); margin-top:4px; align-self:flex-start; }
.msg-row.mine .m-time { align-self:flex-end; }
.m-status { font-size:11px; opacity:.85; color:var(--text-3); text-align:right; margin-top:2px; }
.msg-row.mine .m-status { color:var(--bubble-self-color); opacity:.75; }
.msg-img { max-width:300px; border-radius:10px; display:block; cursor:pointer; border:1px solid var(--border-soft); }
.m-file { font-size:13px; }
.m-file a { color:var(--primary); }
.sys { color:var(--text-2); font-size:12px; }
.load-wrap { text-align:center; padding:4px; }
.load-more { border:none; background:transparent; color:var(--primary); cursor:pointer;
  font-size:13px; padding:8px 16px; border-radius:8px; }
.load-more:hover { background:var(--surface-3); }
.load-more:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.chat-input { background:var(--surface); border-top:1px solid var(--border-soft); padding:12px 16px; position:relative; }
.toolbar { position:relative; display:flex; gap:6px; margin-bottom:8px; align-items:center; }
.emoji-wrap { position:relative; display:inline-flex; }
.icon-btn { font-size:20px; cursor:pointer; position:relative; border:none; background:transparent;
  padding:4px; border-radius:8px; color:var(--text-2); transition: background .15s, color .15s; }
.icon-btn:hover { background:var(--surface-3); color:var(--text); }
.icon-btn:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.emoji-panel { position:absolute; left:100%; bottom:calc(100% + 6px); margin-left:2px; transform-origin: bottom left; background:var(--surface-2);
  border:1px solid var(--border); border-radius:14px; box-shadow:0 12px 40px var(--shadow-deep); padding:10px 12px 6px;
  width:300px; z-index:40; overscroll-behavior: contain; animation: emojiPop .16s ease; }
@keyframes emojiPop { from { opacity:0; transform: translateY(6px) scale(.97); } to { opacity:1; transform: translateY(0) scale(1); } }
@media (prefers-reduced-motion: reduce) { .emoji-panel { animation: none; } .emoji-track { transition: none; } }
.emoji-viewport { width: 276px; overflow: hidden; }
.emoji-track { display:flex; transition: transform .18s ease; will-change: transform; }
.emoji-page { display:grid; grid-template-columns: repeat(8, 1fr); gap:2px; flex:0 0 276px; width:276px; }
.emoji-item { cursor:pointer; font-size:22px; border:none; background:transparent; padding:4px 0;
  border-radius:8px; transition: background .15s; display:flex; align-items:center; justify-content:center; line-height:1; }
.emoji-item:hover { background:var(--surface-3); }
.emoji-item:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.emoji-dots { display:flex; justify-content:center; gap:6px; margin-top:8px; padding-bottom:2px; }
.emoji-dot { width:6px; height:6px; border-radius:50%; border:none; cursor:pointer; padding:0;
  background:var(--text-3); opacity:.5; transition: background .15s, opacity .15s; }
.emoji-dot:hover { opacity:1; }
.emoji-dot.active { background:var(--primary); opacity:1; }
.emoji-dot:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.input-box { width:100%; min-height:60px; max-height:150px; resize:none; border:1px solid var(--border);
  border-radius:10px; padding:10px 12px; font-size:14px; font-family:inherit; outline:none; color:var(--text);
  background:var(--surface-2); transition: border-color .18s, box-shadow .18s; }
.input-box::placeholder { color:var(--text-3); }
.input-box:focus-visible { border-color:var(--primary); box-shadow: 0 0 0 3px var(--primary-glow); }
.send-row { display:flex; justify-content:space-between; align-items:center; margin-top:8px; }
.tip { font-size:12px; color:var(--text-2); }
.img-preview { position:fixed; inset:0; background:var(--overlay); backdrop-filter:blur(4px); display:flex; align-items:center;
  justify-content:center; z-index:100; }
.img-preview img { max-width:92vw; max-height:92vh; border-radius:10px; box-shadow:0 12px 50px var(--shadow-deep); }
.close-x { position:absolute; top:16px; right:24px; color:#fff; font-size:38px; cursor:pointer;
  border:none; background:transparent; line-height:1; opacity:.9; }
.close-x:hover { opacity:1; }
.close-x:focus-visible { outline:2px solid #fff; outline-offset:2px; }
</style>