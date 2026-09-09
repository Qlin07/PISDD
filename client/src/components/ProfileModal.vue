<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="profile-modal-title">
      <div class="m-title" id="profile-modal-title">我的资料</div>

      <div class="field">
        <label for="pf-avatar" class="avatar-upload" aria-label="上传头像">
          <span class="avatar" :style="avatarStyle">{{ form.avatar_url ? '' : (form.nickname||'我').slice(0,1) }}</span>
          <input id="pf-avatar" type="file" accept="image/*" style="display:none" @change="uploadAvatar" aria-label="选择头像文件" />
          <span class="up-tip">点击上传头像</span>
        </label>
      </div>

      <label for="pf-nickname" class="sr-only">昵称</label>
      <input id="pf-nickname" class="input" v-model="form.nickname" placeholder="昵称(必填,≤20字符)" aria-label="昵称" />

      <label for="pf-signature" class="sr-only">个性签名</label>
      <input id="pf-signature" class="input" v-model="form.signature" placeholder="个性签名(≤50字符)" aria-label="个性签名" spellcheck="false" />

      <div class="m-ops">
        <button class="btn" @click="$emit('close')" type="button">取消</button>
        <button class="btn btn-primary" @click="save" type="button">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { api } from '../api/http'

const props = defineProps({ store: Object })
const store = props.store
const emit = defineEmits(['close', 'saved'])

const form = reactive({
  nickname: store.user?.nickname || '',
  signature: store.user?.signature || '',
  avatar_url: store.user?.avatar_url || ''
})

const avatarStyle = computed(() => ({
  background: form.avatar_url
    ? `var(--primary) url(${form.avatar_url}) center/cover no-repeat`
    : 'var(--primary)'
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

async function save() {
  if (!form.nickname) return alert('昵称不能为空')
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
    emit('saved')
    emit('close')
  } catch (err) { alert(err.message) }
}
</script>

<style scoped>
.sr-only { position:absolute; width:1px; height:1px; margin:-1px; padding:0; clip:rect(0,0,0,0); border:0; overflow:hidden; white-space:nowrap; }
.modal-mask { position:fixed; inset:0; background:var(--overlay-soft); backdrop-filter:blur(3px); display:flex;
  align-items:center; justify-content:center; z-index:200; }
.modal { width:360px; background:var(--surface-2); border-radius:16px; padding:22px; border:1px solid var(--border); box-shadow:0 20px 60px var(--shadow-deep); }
.m-title { font-size:17px; font-weight:600; margin-bottom:16px; color:var(--text); }
.field { display:flex; justify-content:center; margin-bottom:14px; }
.avatar-upload { display:flex; align-items:center; gap:10px; cursor:pointer; }
.avatar { width:60px; height:60px; border-radius:50%; background:linear-gradient(135deg,var(--primary),var(--primary-dim)); color:#fff;
  display:flex; align-items:center; justify-content:center; font-size:20px; font-weight:600; box-shadow:0 2px 10px var(--primary-glow); }
.up-tip { color:var(--primary); font-size:13px; }
.input { width:100%; height:38px; border:1px solid var(--border); border-radius:10px; padding:0 12px; margin-bottom:10px;
  font-size:14px; font-family:inherit; outline:none; color:var(--text); background:var(--surface); transition: border-color .18s, box-shadow .18s; box-sizing:border-box; }
.input:focus-visible { border-color:var(--primary); box-shadow:0 0 0 3px var(--primary-glow); }
.m-ops { display:flex; justify-content:flex-end; gap:10px; margin-top:6px; }
.btn { border:none; border-radius:8px; padding:8px 16px; font-size:14px; cursor:pointer; color:var(--text-2); background:var(--surface-3); transition: background .15s, box-shadow .15s; }
.btn:hover { background:var(--hover-bg); }
.btn.btn-primary { background:var(--primary); color:var(--primary-ink); font-weight:600; }
.btn.btn-primary:hover { box-shadow:0 0 0 3px var(--primary-glow); }
</style>