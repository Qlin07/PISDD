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

    <!-- 搜索添加 -->
    <div class="add-bar">
      <label for="contact-search" class="sr-only">搜索用户添加好友</label>
      <input id="contact-search" class="input" v-model="kw" placeholder="搜索用户添加好友" @input="doSearch" />
      <div v-if="kw" class="add-results">
        <div v-for="u in results" :key="u.user_id" class="add-item">
          <span class="avatar" aria-hidden="true">{{ (u.nickname||'?').slice(0,1) }}</span>
          <div class="a-info">
            <div class="a-name">{{ u.nickname }} <span class="a-acct">@{{ u.account }}</span></div>
          </div>
          <button class="mini ok" @click="apply(u.user_id)" type="button">添加</button>
        </div>
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

// friendship里是 id 列表(与 from_id/to_id) —— 由 Main 传入的是原始 friendship, 这里需转为用户
const friendUsers = computed(() => props.friends)

async function doSearch() {
  if (!kw.value) { results.value = []; return }
  try {
    const { data } = await api.get(`/contacts/search?keyword=${encodeURIComponent(kw.value)}`)
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
.pending { background:#fff8f0; border:1px solid #ffe0c0; border-radius:8px; padding:10px; margin-bottom:12px; }
.p-title,.f-title { font-size:13px; color:var(--text-2); margin:8px 4px; }
.p-item,.add-item { display:flex; align-items:center; gap:8px; padding:6px; }
.p-info { flex:1; }
.avatar { width:36px; height:36px; border-radius:50%; background:var(--primary); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:14px; flex-shrink:0; }
.mini { border:none; border-radius:4px; padding:4px 10px; font-size:12px; cursor:pointer; margin-left:6px; }
.mini.ok { background:var(--primary); color:#fff; }
.mini.no { background:#eee; color:#666; }
.mini:focus-visible { outline:2px solid var(--primary); outline-offset:1px; }
.add-bar { position:relative; margin-bottom:8px; }
.add-results { position:absolute; top:42px; left:0; right:0; background:#fff; border:1px solid var(--border);
  border-radius:8px; box-shadow:0 4px 12px rgba(0,0,0,.08); z-index:20; padding:6px; }
.add-item:hover { background:#f5f7fa; }
.a-info { flex:1; }
.a-name { font-size:14px; }
.a-acct { color:var(--text-2); font-size:12px; }
.f-item { display:flex; align-items:center; gap:10px; padding:10px; cursor:pointer; border-radius:8px;
  width:100%; text-align:left; border:none; background:transparent; font:inherit; color:inherit; }
.f-item:hover { background:#f5f7fa; }
.f-item:focus-visible { outline:2px solid var(--primary); outline-offset:-1px; background:#f5f7fa; }
.f-rem { font-size:15px; }
</style>