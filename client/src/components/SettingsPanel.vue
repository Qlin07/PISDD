<template>
  <div class="settings">
    <div class="s-title">设置中心</div>

    <div class="s-section">
      <div class="s-item">
        <label for="nickname">昵称</label>
        <input id="nickname" class="input" v-model="form.nickname" aria-label="昵称" />
      </div>
      <div class="s-item">
        <label for="signature">个性签名</label>
        <input id="signature" class="input" v-model="form.signature" placeholder="最多50字符" aria-label="个性签名" spellcheck="false" />
      </div>
      <div class="s-item">
        <label>头像</label>
        <label class="avatar-upload" aria-label="上传头像">
          <span class="avatar" :style="avatarStyle">{{ (form.nickname||'我').slice(0,1) }}</span>
          <input type="file" accept="image/*" style="display:none" @change="uploadAvatar" aria-label="选择头像文件" />
          <span class="up-tip">点击上传</span>
        </label>
      </div>
      <button class="btn btn-primary" @click="saveProfile">保存资料</button>
    </div>

    <div class="s-section">
      <div class="s-item">
        <label for="oldPwd">修改密码</label>
        <input id="oldPwd" class="input" type="password" v-model="oldPwd" placeholder="原密码" aria-label="原密码" autocomplete="current-password" />
        <input id="newPwd" class="input" type="password" v-model="newPwd" placeholder="新密码" aria-label="新密码" autocomplete="new-password" style="margin-top:6px" />
        <button class="btn" @click="changePwd" style="margin-top:8px">修改密码</button>
      </div>
    </div>

    <div class="s-section">
      <div class="s-item gap" style="flex-direction:column; align-items:stretch;">
        <label>主题外观</label>
        <div class="theme-picker" role="radiogroup" aria-label="选择主题">
          <button v-for="t in THEMES" :key="t.key" type="button" class="theme-opt"
            :class="{active: store.theme === t.key}" @click="store.setTheme(t.key)"
            :aria-checked="store.theme === t.key" role="radio">
            <span class="theme-swatch" :data-theme="t.key" aria-hidden="true"></span>
            <span class="theme-label">{{ t.label }}</span>
          </button>
        </div>
      </div>
    </div>

    <div class="s-section about">
      <div class="s-item"><label>账号</label><span>{{ store.user?.account || '' }}</span></div>
      <div class="s-item"><label>UID</label><span>{{ store.user?.user_id ?? '' }}</span></div>
      <div class="s-item"><label>连接状态</label>
        <span :class="{ok: store.connected}">{{ store.connected ? '● 已连接' : '○ 断开' }}</span>
      </div>
      <div class="s-item"><label>版本</label><span>SimpleChat v0.1.0</span></div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { api } from '../api/http'
import { THEMES } from '../store'

const props = defineProps({ store: Object })
const store = props.store

const form = reactive({
  nickname: store.user?.nickname || '',
  signature: store.user?.signature || '',
  avatar_url: store.user?.avatar_url || ''
})
const oldPwd = ref('')
const newPwd = ref('')

const avatarStyle = computed(() => ({
  background: form.avatar_url ? `url(${form.avatar_url}) center/cover` : 'var(--primary)'
}))

async function uploadAvatar(e) {
  const file = e.target.files[0]
  if (!file) return
  const fd = new FormData()
  fd.append('image', file)
  try {
    const { data } = await api.post('/files/upload?type=image', fd)
    form.avatar_url = data.file_url
  } catch (err) { alert(err.message) }
  e.target.value = ''
}
async function saveProfile() {
  try {
    await api.put('/user/profile', {
      nickname: form.nickname,
      signature: form.signature,
      avatar_url: form.avatar_url
    })
    store.user.nickname = form.nickname
    store.user.signature = form.signature
    store.user.avatar_url = form.avatar_url
    localStorage.setItem('user', JSON.stringify(store.user))
    alert('保存成功')
  } catch (err) { alert(err.message) }
}
async function changePwd() {
  if (!oldPwd.value || !newPwd.value) return alert('请填写密码')
  try {
    await api.put('/user/password', { old_password: oldPwd.value, new_password: newPwd.value })
    oldPwd.value = ''; newPwd.value = ''
    alert('密码修改成功')
  } catch (err) { alert(err.message) }
}
</script>

<style scoped>
.settings { padding:20px; max-width:440px; }
.s-title { font-size:18px; font-weight:600; margin-bottom:16px; color:var(--text); }
.s-section { background:var(--surface); border:1px solid var(--border-soft); border-radius:14px; padding:16px; margin-bottom:12px;
  display:flex; flex-direction:column; gap:10px; }
.s-item { display:flex; align-items:center; justify-content:space-between; gap:10px; }
.s-item label { font-size:14px; color:var(--text-2); min-width:70px; }
.s-item .input { flex:1; }
.s-item > span:last-child { color:var(--text); }
.avatar { width:48px; height:48px; border-radius:50%; background:linear-gradient(135deg,var(--primary),var(--primary-dim)); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:18px; font-weight:600; box-shadow:0 2px 10px var(--primary-glow); }
.avatar-upload { display:flex; align-items:center; gap:10px; cursor:pointer; }
.up-tip { color:var(--primary); font-size:13px; }
.theme-picker { display:flex; gap:10px; flex-wrap:wrap; }
.theme-opt { display:flex; align-items:center; gap:8px; padding:8px 14px; border:1px solid var(--border);
  border-radius:10px; background:var(--surface-2); cursor:pointer; color:var(--text-2); transition: border-color .18s, color .18s, box-shadow .18s; }
.theme-opt:hover { border-color:var(--primary); color:var(--text); }
.theme-opt:focus-visible { outline:2px solid var(--primary); outline-offset:2px; }
.theme-opt.active { border-color:var(--primary); color:var(--primary); box-shadow:0 0 0 3px var(--primary-glow); font-weight:600; }
.theme-swatch { width:18px; height:18px; border-radius:50%; border:1px solid var(--border-soft); }
/* 各主题的迷你色板(用于选择器预览) */
.theme-swatch[data-theme="night"] { background:radial-gradient(circle at 35% 35%, #4cc9f0, #0b0f16 70%); }
.theme-swatch[data-theme="win11"] { background:radial-gradient(circle at 35% 35%, #f0c0f4, #202020 70%); }
.about .s-item span.ok { color:var(--success); font-weight:600; }
</style>