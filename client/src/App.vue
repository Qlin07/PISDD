<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from './store'

const store = useStore()
const router = useRouter()

onMounted(() => {
  // 已有登录态则恢复WebSocket
  if (store.isLogin) {
    store.connectWs()
    if (router.currentRoute.value.name === 'main' || router.currentRoute.value.path === '/') {
      store.refreshConversations().catch(() => {})
      store.loadContacts().catch(() => {})
      store.loadGroups().catch(() => {})
      store.loadPending().catch(() => {})
    }
  }
})
</script>