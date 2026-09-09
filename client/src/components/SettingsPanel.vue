<template>
  <div class="settings">
    <div class="s-title">设置中心</div>

    <div class="s-section">
      <button class="btn btn-danger logout-btn" @click="$emit('logout')" type="button">退出登录</button>
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
import { ref } from 'vue'
import { api } from '../api/http'
import { THEMES } from '../store'

const props = defineProps({ store: Object })
const store = props.store

const oldPwd = ref('')
const newPwd = ref('')

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
.logout-btn { width:100%; padding:10px; border-radius:10px; font-size:14px; font-weight:600;
  background:var(--danger); color:#fff; transition: box-shadow .15s, filter .15s; }
.logout-btn:hover { box-shadow:0 0 0 3px var(--danger-glow); filter:brightness(1.05); }
.logout-btn:focus-visible { outline:2px solid var(--danger); outline-offset:2px; }
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