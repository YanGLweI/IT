import axios from 'axios'
import { Message } from 'element-ui'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器：添加 JWT Token
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code && res.code !== 200) {
      Message.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  error => {
    if (error.response && error.response.status === 401) {
      // Token 无效或过期，清除并跳转登录
      const currentPath = router.currentRoute.path
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('display_name')
      
      // 如果是 logout 请求，静默处理（不显示错误提示）
      const isLogoutRequest = error.config.url === '/logout'
      
      // 如果当前不在登录页才跳转，避免重复导航
      if (currentPath !== '/login' && !isLogoutRequest) {
        router.push('/login').catch(() => {})
      }
      
      // 非 logout 请求才显示错误提示
      if (!isLogoutRequest) {
        Message.error('登录已过期，请重新登录')
      }
    } else {
      Message.error(error.response?.data?.message || error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
