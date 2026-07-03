import axios from 'axios'
import { Message } from 'element-ui'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
  withCredentials: true // 允许携带 Cookie（用于 refreshToken）
})

// ========== 刷新状态管理 ==========
let isRefreshing = false        // 是否正在刷新中
let refreshSubscribers = []     // 等待刷新完成的请求队列

// 将等待刷新的请求加入队列
function subscribeTokenRefresh(cb) {
  refreshSubscribers.push(cb)
}

// 刷新完成后，用新 Token 重放所有排队请求
function onTokenRefreshed(newToken) {
  refreshSubscribers.forEach(cb => cb(newToken))
  refreshSubscribers = []
}

// 刷新失败，拒绝所有排队请求
function onTokenRefreshFailed() {
  refreshSubscribers.forEach(cb => cb(null))
  refreshSubscribers = []
}

// 执行刷新 Token 的请求（使用独立 axios 实例，避免被拦截器循环）
function doRefreshToken() {
  return axios.post('/api/refresh-token', null, { withCredentials: true })
}

// 清除登录状态并跳转登录页
function handleLogout(msg) {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('display_name')
  const currentPath = router.currentRoute.path
  if (currentPath !== '/login') {
    if (msg) {
      Message.error(msg)
    }
    router.push('/login').catch(() => {})
  }
}

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
    const originalRequest = error.config

    // 401 且不是刷新请求本身 且没有重试过
    if (
      error.response &&
      error.response.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url.includes('/refresh-token')
    ) {
      // logout 请求不走刷新逻辑，直接清除状态
      if (originalRequest.url.includes('/logout')) {
        handleLogout()
        return Promise.reject(error)
      }

      // 如果已经在刷新中，将请求加入等待队列
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh(newToken => {
            if (newToken) {
              originalRequest.headers['Authorization'] = `Bearer ${newToken}`
              originalRequest._retry = true
              resolve(request(originalRequest))
            } else {
              reject(new Error('Token刷新失败'))
            }
          })
        })
      }

      // 标记正在刷新，防止重复触发
      isRefreshing = true
      originalRequest._retry = true

      return new Promise((resolve, reject) => {
        doRefreshToken()
          .then(res => {
            const newToken = res.data && res.data.data && res.data.data.token
            if (newToken) {
              // 更新 localStorage 中的 accessToken
              localStorage.setItem('token', newToken)
              // 用新 Token 重放当前失败的请求
              originalRequest.headers['Authorization'] = `Bearer ${newToken}`
              // 通知所有排队中的请求
              onTokenRefreshed(newToken)
              resolve(request(originalRequest))
            } else {
              // 刷新接口返回了非预期格式，视为失败
              onTokenRefreshFailed()
              handleLogout('登录已过期，请重新登录')
              reject(new Error('Token刷新失败'))
            }
          })
          .catch(() => {
            // refreshToken 也过期或网络错误
            onTokenRefreshFailed()
            handleLogout('登录已过期，请重新登录')
            reject(error)
          })
          .finally(() => {
            isRefreshing = false
          })
      })
    }

    // 其他错误显示提示
    Message.error(error.response?.data?.message || error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
