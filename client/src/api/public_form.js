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

// ============ IT指南公开接口 ============

// 获取已发布IT指南列表
export function getPublicITGuides(params) {
  return publicRequest.get('/public/it-guides', { params })
}

// 获取单个已发布IT指南详情
export function getPublicITGuideDetail(id) {
  return publicRequest.get(`/public/it-guides/${id}`)
}

// 记录指南浏览量
export function recordITGuideView(id) {
  return publicRequest.post(`/public/it-guides/${id}/view`)
}

// 点赞/取消点赞
export function toggleITGuideLike(id, data) {
  return publicRequest.post(`/public/it-guides/${id}/like`, data)
}
