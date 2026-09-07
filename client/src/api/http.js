import axios from 'axios'

export const api = axios.create({
  baseURL: '/api',
  timeout: 15000
})

// 请求拦截: 附加 JWT
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截: 统一错误提示, 401跳登录
api.interceptors.response.use(
  (resp) => {
    const d = resp.data
    if (d && typeof d.code !== 'undefined' && d.code !== 0) {
      return Promise.reject(new Error(d.msg || '请求失败'))
    }
    return d
  },
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/login'
      }
    }
    const msg = err.response?.data?.msg || err.message || '网络错误'
    return Promise.reject(new Error(msg))
  }
)