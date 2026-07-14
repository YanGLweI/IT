import request from './request'

// 获取指南列表
export function getITGuides(params) {
  return request.get('/it-guides', { params })
}

// 获取单个指南详情
export function getITGuide(id) {
  return request.get(`/it-guides/${id}`)
}

// 创建指南
export function createITGuide(data) {
  const formData = new FormData()
  formData.append('title', data.title)
  formData.append('description', data.description || '')
  formData.append('guide_type', data.guide_type)
  formData.append('category', data.category || '')
  return request.post('/it-guides', formData)
}

// 更新指南基本信息
export function updateITGuide(id, data) {
  const formData = new FormData()
  if (data.title) formData.append('title', data.title)
  formData.append('description', data.description || '')
  formData.append('category', data.category || '')
  if (data.sort_order !== undefined) formData.append('sort_order', data.sort_order)
  return request.put(`/it-guides/${id}`, formData)
}

// 删除指南
export function deleteITGuide(id) {
  return request.delete(`/it-guides/${id}`)
}

// 发布指南
export function publishITGuide(id) {
  return request.put(`/it-guides/${id}/publish`)
}

// 取消发布
export function unpublishITGuide(id) {
  return request.put(`/it-guides/${id}/unpublish`)
}

// 获取指南步骤列表
export function getITGuideSteps(guideId) {
  return request.get(`/it-guides/${guideId}/steps`)
}

// 创建步骤
export function createITGuideStep(guideId, data) {
  const formData = new FormData()
  formData.append('title', data.title || '')
  formData.append('description', data.description || '')
  formData.append('sort_order', data.sort_order || 0)
  return request.post(`/it-guides/${guideId}/steps`, formData)
}

// 更新步骤
export function updateITGuideStep(guideId, stepId, data) {
  const formData = new FormData()
  if (data.title) formData.append('title', data.title)
  formData.append('description', data.description || '')
  if (data.sort_order !== undefined) formData.append('sort_order', data.sort_order)
  return request.put(`/it-guides/${guideId}/steps/${stepId}`, formData)
}

// 删除步骤
export function deleteITGuideStep(guideId, stepId) {
  return request.delete(`/it-guides/${guideId}/steps/${stepId}`)
}

// 批量更新步骤排序
export function reorderITGuideSteps(guideId, stepIds) {
  return request.post(`/it-guides/${guideId}/steps/reorder`, { step_ids: stepIds })
}

// 上传媒体文件
export function uploadITGuideMedia(guideId, formData) {
  return request.post(`/it-guides/${guideId}/media`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}

// 删除媒体文件
export function deleteITGuideMedia(guideId, mediaId) {
  return request.delete(`/it-guides/${guideId}/media/${mediaId}`)
}
