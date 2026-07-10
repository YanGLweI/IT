import axios from 'axios'

// 独立的 axios 实例，不附加 JWT token（免登录访问）
const publicRequest = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 获取已发布表单列表
export function getPublicForms(params) {
  return publicRequest.get('/public/forms', { params })
}

// 公开预览 URL
export function getPublicPreviewUrl(id) {
  return `/api/public/forms/${id}/preview`
}

// 公开预览 — 获取文件 blob（无认证）
export function getPublicPreviewBlob(id) {
  return publicRequest.get(`/public/forms/${id}/preview`, {
    responseType: 'blob'
  })
}
