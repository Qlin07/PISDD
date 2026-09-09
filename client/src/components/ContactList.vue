<template>
  <div class="contact-list">
    <!-- 待处理的好友申请 -->
    <div v-if="pending.length" class="pending">
      <div class="p-title">好友申请 <span aria-live="polite">{{ pending.length }}</span></div>
      <div v-for="f in pending" :key="f.id" class="p-item">
        <span class="avatar" aria-hidden="true">友</span>
        <div class="p-info">
          <div class="p-name">用户 #{{ f.from_id }}</div>
          <div class="p-ops">
            <button class="mini ok" @click="accept(f.from_id)" type="button">同意</button>
            <button class="mini no" @click="reject(f.from_id)" type="button">拒绝</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 搜索添加(UID 或 昵称/账号) -->
    <div class="search-bar">
      <label for="contact-search" class="sr-only">输入 UID 或昵称/账号搜索用户</label>
      <div class="search-row">
        <input id="contact-search" class="input" v-model="kw" placeholder="输入 UID 或昵称/账号搜索"
          aria-label="输入UID或昵称账号搜索" autocomplete="off" spellcheck="false"
          @input="onInput" @keyup.enter="doSearch" />
        <button class="mini ok" @click="doSearch" type="button">搜索</button>
      </div>

      <div v-if="kw && isUidKw" class="uid-result" v-show="uidResult || kw">
        <template v-if="uidResult">
          <span class="avatar" :style="uidAvatarStyle" aria-hidden="true">{{ (uidResult.nickname||'?').slice(0,1) }}</span>
          <div class="u-info">
            <div class="u-name">{{ uidResult.nickname }} <span class="u-id">UID {{ uidResult.user_id }}</span></div>
          </div>
          <button class="mini ok" @click="apply(uidResult.user_id)" type="button">添加</button>
        </template>
        <span v-else-if="uidMsg" class="uid-msg error">{{ uidMsg }}</span>
        <span v-else class="uid-msg">查找中…</span>
      </div>

      <div v-else-if="kw && !isUidKw" class="add-results">
        <div v-for="u in results" :key="u.user_id" class="add-item">
          <span class="avatar" aria-hidden="true">{{ (u.nickname||'?').slice(0,1) }}</span>
          <div class="a-info">
            <div class="a-name">{{ u.nickname }} <span class="a-acct">@{{ u.account }}</span></div>
          </div>
          <button class="mini ok" @click="apply(u.user_id)" type="button">添加</button>
        </div>
        <p v-if="!results.length" class="empty">无匹配用户</p>
      </div>
    </div>

    <!-- 好友列表 -->
    <div class="f-title">好友 ({{ friends.length }})</div>
    <button class="f-item" v-for="f in friendUsers" :key="f.user_id" @click="$emit('chat', f)" type="button">
      <span class="avatar" :style="avatarStyle(f.avatar_url)" aria-hidden="true">{{ (f.nickname||'?').slice(0,1) }}</span>
      <span class="f-rem">{{ f.remark || f.nickname }}</span>
    </button>
    <p v-if="!friends.length && !kw" class="empty">还没有好友，搜索添加</p>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/http'

const props = defineProps({ friends: Array, pending: Array })
const emit = defineEmits(['chat', 'refresh'])

const kw = ref('')
const results = ref([])
const uidResult = ref(null)
const uidMsg = ref('')
const uidError = ref(false)
let typingTimer = null

// friendship里是 id 列表(与 from_id/to_id) —— 由 Main 传入的是原始 friendship, 这里需转为用户
// 空值兜底: 后端无好友时 friends/pending 可能为 null, 归一到 [] 避免渲染崩溃
const friends = computed(() => props.friends || [])
const pending = computed(() => props.pending || [])
const friendUsers = computed(() => friends.value)

// 输入为纯数字 → 按 UID 精确查询; 否则 → 按昵称/账号模糊搜索
const isUidKw = computed(() => /^\d+$/.test(String(kw.value || '').trim()))

const uidAvatarStyle = computed(() => ({
  background: uidResult.value && uidResult.value.avatar_url ? `url(${uidResult.value.avatar_url}) center/cover` : 'var(--primary)'
}))

// 输入实时搜索: 非数字关键词做 400ms 防抖; 数字 UID 等用户点「搜索」/回车
function onInput() {
  uidResult.value = null
  uidMsg.value = ''
  uidError.value = false
  if (isUidKw.value) { results.value = [] }
  else {
    clearTimeout(typingTimer)
    typingTimer = setTimeout(() => doSearch(), 400)
  }
}

// 统一搜索入口: 纯数字走 UID, 否则昵称/账号
function doSearch() {
  const v = String(kw.value || '').trim()
  if (!v) { results.value = []; uidResult.value = null; uidMsg.value = ''; return }
  if (isUidKw.value) { queryByUid(v) }
  else { queryByName(v) }
}

async function queryByUid(uidVal) {
  uidResult.value = null
  uidMsg.value = ''
  uidError.value = false
  try {
    const { data } = await api.get(`/contacts/uid?uid=${uidVal}`)
    uidResult.value = data
  } catch (e) {
    uidResult.value = null
    uidMsg.value = e.message || '未找到该用户'
    uidError.value = true
  }
}

async function queryByName(name) {
  if (!name) { results.value = []; return }
  try {
    const { data } = await api.get(`/contacts/search?keyword=${encodeURIComponent(name)}`)
    results.value = data || []
  } catch (e) { results.value = [] }
}
async function apply(toId) {
  try {
    await api.post('/contacts/apply', { to_id: toId })
    alert('已发送好友申请')
  } catch (e) { alert(e.message) }
}
async function accept(fromId) {
  try { await api.post('/contacts/accept', { from_id: fromId }); emit('refresh') }
  catch (e) { alert(e.message) }
}
async function reject(fromId) {
  try { await api.post('/contacts/reject', { from_id: fromId }); emit('refresh') }
  catch (e) { alert(e.message) }
}
function avatarStyle(url) { return { background: url ? `url(${url}) center/cover` : 'var(--primary)' } }
</script>

<style scoped>
.contact-list { padding:12px; }
.sr-only { position:absolute; width:1px; height:1px; margin:-1px; padding:0; clip:rect(0,0,0,0); border:0; overflow:hidden; white-space:nowrap; }
.pending { background:var(--warning-soft); border:1px solid var(--warning-border); border-radius:10px; padding:10px; margin-bottom:12px; }
.p-title,.f-title { font-size:13px; color:var(--text-2); margin:8px 4px; }
.p-item,.add-item { display:flex; align-items:center; gap:8px; padding:6px; }
.p-info { flex:1; }
.p-name { color:var(--text); }
.avatar { width:36px; height:36px; border-radius:50%; background:linear-gradient(135deg,var(--primary),var(--primary-dim)); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:14px; flex-shrink:0; font-weight:600; }
.mini { border:none; border-radius:6px; padding:4px 12px; font-size:12px; cursor:pointer; margin-left:6px; transition: background .15s, box-shadow .15s; }
.mini.ok { background:var(--primary); color:var(--primary-ink); font-weight:600; }
.mini.ok:hover { box-shadow:0 0 0 3px var(--primary-glow); }
.mini.no { background:var(--surface-3); color:var(--text-2); }
.mini.no:hover { background:var(--hover-bg); }
.mini:focus-visible { outline:2px solid var(--primary); outline-offset:1px; }
.search-bar { position:relative; margin-bottom:10px; }
.search-row { display:flex; gap:8px; }
.search-row .input { flex:1; min-width:0; }
.search-row .mini { margin-left:0; white-space:nowrap; }
.uid-msg { font-size:12px; margin-top:6px; color:var(--success); }
.uid-msg.error { color:var(--danger); }
.uid-result { display:flex; align-items:center; gap:10px; padding:8px; margin-top:8px;
  background:var(--surface-2); border:1px solid var(--border); border-radius:10px; }
.u-info { flex:1; min-width:0; }
.u-name { font-size:14px; color:var(--text); }
.u-id { color:var(--text-3); font-size:12px; margin-left:4px; }
.add-results { display:flex; flex-direction:column; margin-top:8px; background:var(--surface-2); border:1px solid var(--border);
  border-radius:10px; padding:6px; }
.add-item:hover { background:var(--surface-3); }
.a-info { flex:1; }
.a-name { font-size:14px; color:var(--text); }
.a-acct { color:var(--text-3); font-size:12px; }
.f-item { display:flex; align-items:center; gap:10px; padding:10px; cursor:pointer; border-radius:8px;
  width:100%; text-align:left; border:none; background:transparent; font:inherit; color:inherit; transition: background .15s; }
.f-item:hover { background:var(--surface-3); }
.f-item:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; background:var(--surface-3); }
.f-rem { font-size:15px; color:var(--text); }
</style>