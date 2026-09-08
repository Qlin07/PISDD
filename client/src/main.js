import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { loadTheme, applyTheme } from './store'
import './style.css'

// 启动即应用已保存的主题(默认夜航), 避免闪烁
applyTheme(loadTheme())

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')