import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Main from '../views/Main.vue'
import { useStore } from '../store'

const routes = [
  { path: '/login', name: 'login', component: Login },
  { path: '/main', name: 'main', component: Main },
  { path: '/', redirect: '/main' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const store = useStore()
  if (to.name !== 'login' && !store.isLogin) {
    return { name: 'login' }
  }
  if (to.name === 'login' && store.isLogin) {
    return { name: 'main' }
  }
  return true
})

export default router