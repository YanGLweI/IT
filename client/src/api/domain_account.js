import axios from 'axios'

// 域账号自助平台 API（独立于主平台的 request 实例）

const domainAccountApi = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器：自动添加域账号自助平台的 token
domainAccountApi.interceptors.request.use(config => {
  const token = localStorage.getItem('domain_account_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：401 时清除 token
domainAccountApi.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('domain_account_token')
      localStorage.removeItem('domain_account_username')
    }
    return Promise.reject(error)
  }
)

/**
 * 域账号自助平台登录
 */
export function domainAccountLogin(username, password) {
  return domainAccountApi.post('/domain-account/login', { username, password })
}

/**
 * 获取域账号信息
 */
export function getDomainAccountInfo() {
  return domainAccountApi.get('/domain-account/info')
}

/**
 * 修改域账号密码
 */
export function changeDomainPassword(username, oldPassword, newPassword) {
  return domainAccountApi.post('/domain-account/change-password', {
    username,
    old_password: oldPassword,
    new_password: newPassword
  })
}
