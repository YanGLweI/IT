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

// ========== Retry Configuration ==========
const MAX_REFRESH_RETRIES = 3          // Maximum retry attempts
const REFRESH_RETRY_DELAY = 1000       // Initial delay in milliseconds
let refreshRetries = 0                 // Retry counter (session-scoped)

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

// 创建独立的 axios 实例用于刷新 token（避免被主实例的拦截器循环）
const refreshAxios = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: true // 必须携带 HttpOnly Cookie
})

// 执行刷新 Token 的请求（使用独立 axios 实例，避免被拦截器循环）
function doRefreshToken() {
  return refreshAxios.post('/refresh-token')
}

// 清除登录状态并跳转登录页
function handleLogout(msg) {
  // 先调用登出接口清除 HttpOnly Cookie（access_token + refresh_token）
  // 即使请求失败（如 token 已过期）也不影响本地清除流程
  try {
    axios.post('/api/logout', null, { withCredentials: true })
  } catch (_) {}
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

// Decode JWT token to extract claims (no signature verification, payload only)
function decodeToken(token) {
  try {
    // Validate input type
    if (!token || typeof token !== 'string') {
      console.warn('Invalid token type:', typeof token)
      return null
    }
    
    const parts = token.split('.')
    if (parts.length !== 3) {
      console.warn('Invalid token format: expected 3 parts, got', parts.length)
      return null
    }
    
    const base64Url = parts[1]
    if (!base64Url || base64Url.length === 0) {
      console.warn('Invalid token format: payload part is empty')
      return null
    }
    
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)).join(''))
    return JSON.parse(jsonPayload)
  } catch (e) {
    console.error('Failed to decode token:', e)
    return null
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
              // Decode token to get username and display_name
              const claims = decodeToken(newToken)
                    
              // Safe extraction with defaults
              const username = claims?.username || 'unknown'
              const displayName = claims?.display_name || claims?.username || 'Unknown User'
                  
              // Update localStorage with all necessary information for session recovery
              localStorage.setItem('token', newToken)
              localStorage.setItem('username', username)
              localStorage.setItem('display_name', displayName)
                    
              // Set authorization header for current request
              originalRequest.headers['Authorization'] = `Bearer ${newToken}`
                    
              // Notify waiting subscribers with new token
              onTokenRefreshed(newToken)
              resolve(request(originalRequest))
            } else {
              // Invalid response format
              onTokenRefreshFailed()
              handleLogout('登录已过期，请重新登录')
              reject(new Error('Token 刷新失败'))
            }
          })
          .catch(() => {
            // refreshToken 也过期或网络错误
            refreshRetries++
            
            if (refreshRetries >= MAX_REFRESH_RETRIES) {
              // Max retries exceeded, force logout
              console.error('Token 刷新失败次数已达上限，强制登出')
              onTokenRefreshFailed()
              handleLogout('登录已过期，请重新登录')
              reject(error)
            } else {
              // Retry with exponential backoff
              const delay = REFRESH_RETRY_DELAY * Math.pow(2, refreshRetries - 1)
              console.warn(`Token 刷新失败 (${refreshRetries}/${MAX_REFRESH_RETRIES}), ${delay}ms 后重试...`)
              
              setTimeout(() => {
                // Reset retry counter and retry the original request
                refreshRetries = 0
                originalRequest._retry = false
                request(originalRequest).then(resolve).catch(reject)
              }, delay)
            }
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
