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

    <div class="s-section about">
      <div class="s-item"><label>账号</label><span>{{ store.user.account }}</span></div>
      <div class="s-item"><label>UID</label><span>{{ store.user.user_id }}</span></div>
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
.settings { padding:20px; max-width:420px; }
.s-title { font-size:18px; font-weight:600; margin-bottom:16px; }
.s-section { background:#fff; border-radius:8px; padding:16px; margin-bottom:12px;
  display:flex; flex-direction:column; gap:10px; }
.s-item { display:flex; align-items:center; justify-content:space-between; gap:10px; }
.s-item label { font-size:14px; color:#555; min-width:70px; }
.s-item .input { flex:1; }
.avatar { width:48px; height:48px; border-radius:50%; background:var(--primary); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:18px; }
.avatar-upload { display:flex; align-items:center; gap:10px; cursor:pointer; }
.up-tip { color:var(--primary); font-size:13px; }
.about .s-item span.ok { color:#2aa94f; font-weight:600; }
</style>