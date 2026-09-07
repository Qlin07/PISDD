<template>
  <div class="login-wrap">
    <div class="login-card card">
      <h1 class="logo">简聊</h1>
      <p class="sub">SimpleChat · 简洁纯净的实时沟通</p>

      <template v-if="mode === 'login'">
        <input class="input" v-model="account" placeholder="账号" @keyup.enter="doLogin" />
        <input class="input" type="password" v-model="password" placeholder="密码" @keyup.enter="doLogin" />
        <label class="remember">
          <input type="checkbox" v-model="remember" /> 记住我(30天)
        </label>
        <button class="btn btn-primary btn-block" @click="doLogin" :disabled="loading">
          {{ loading ? '登录中...' : '登 录' }}
        </button>
        <div class="switch-links">
          <a @click="mode='register'">注册账号</a>
          <a>忘记密码(联系管理员重置)</a>
        </div>
      </template>

      <template v-else>
        <input class="input" v-model="regAccount" placeholder="账号(字母+数字,3-32位)" />
        <input class="input" v-model="regNickname" placeholder="昵称(默认同账号,2-20字符)" />
        <input class="input" type="password" v-model="regPassword" placeholder="密码" />
        <button class="btn btn-primary btn-block" @click="doRegister" :disabled="loading">
          {{ loading ? '注册中...' : '注 册' }}
        </button>
        <div class="switch-links">
          <a @click="mode='login'">返回登录</a>
        </div>
      </template>

      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from '../store'

const store = useStore()
const router = useRouter()

const mode = ref('login')
const loading = ref(false)
const error = ref('')
const account = ref('')
const password = ref('')
const remember = ref(false)
const regAccount = ref('')
const regNickname = ref('')
const regPassword = ref('')

async function doLogin() {
  error.value = ''
  loading.value = true
  try {
    await store.login(account.value, password.value, remember.value)
    router.push('/main')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function doRegister() {
  error.value = ''
  loading.value = true
  try {
    await store.register(regAccount.value, regNickname.value, regPassword.value)
    router.push('/main')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap { height:100%; display:flex; align-items:center; justify-content:center;
  background:linear-gradient(135deg,#e8f1ff,#f2f2f2); }
.login-card { width:400px; padding:40px 36px; display:flex; flex-direction:column; gap:14px; }
.logo { font-size:34px; color:var(--primary); text-align:center; }
.sub { text-align:center; color:var(--text-2); margin-bottom:8px; }
.remember { font-size:13px; color:var(--text-2); display:flex; align-items:center; gap:6px; }
.switch-links { display:flex; justify-content:space-between; font-size:13px; }
.switch-links a { color:var(--primary); cursor:pointer; }
.error { color:var(--danger); font-size:13px; text-align:center; }
</style>