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
export function createITGuide(data, dualToken) {
  const formData = new FormData()
  formData.append('title', data.title)
  formData.append('description', data.description || '')
  formData.append('guide_type', data.guide_type)
  formData.append('category', data.category || '')
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/it-guides', formData, config)
}

// 更新指南基本信息
export function updateITGuide(id, data, dualToken) {
  const formData = new FormData()
  if (data.title) formData.append('title', data.title)
  formData.append('description', data.description || '')
  formData.append('category', data.category || '')
  if (data.sort_order !== undefined) formData.append('sort_order', data.sort_order)
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/it-guides/${id}`, formData, config)
}

// 删除指南
export function deleteITGuide(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/it-guides/${id}`, config)
}

// 发布指南
export function publishITGuide(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/it-guides/${id}/publish`, {}, config)
}

// 取消发布
export function unpublishITGuide(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/it-guides/${id}/unpublish`, {}, config)
}

// 获取指南步骤列表
export function getITGuideSteps(guideId) {
  return request.get(`/it-guides/${guideId}/steps`)
}

// 创建步骤
export function createITGuideStep(guideId, data, dualToken) {
  const formData = new FormData()
  formData.append('title', data.title || '')
  formData.append('description', data.description || '')
  formData.append('sort_order', data.sort_order || 0)
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post(`/it-guides/${guideId}/steps`, formData, config)
}

// 更新步骤
export function updateITGuideStep(guideId, stepId, data, dualToken) {
  const formData = new FormData()
  if (data.title) formData.append('title', data.title)
  formData.append('description', data.description || '')
  if (data.sort_order !== undefined) formData.append('sort_order', data.sort_order)
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/it-guides/${guideId}/steps/${stepId}`, formData, config)
}

// 删除步骤
export function deleteITGuideStep(guideId, stepId, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/it-guides/${guideId}/steps/${stepId}`, config)
}

// 批量更新步骤排序
export function reorderITGuideSteps(guideId, stepIds) {
  return request.post(`/it-guides/${guideId}/steps/reorder`, { step_ids: stepIds })
}

// 上传媒体文件
export function uploadITGuideMedia(guideId, formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 120000 }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post(`/it-guides/${guideId}/media`, formData, config)
}

// 删除媒体文件
export function deleteITGuideMedia(guideId, mediaId, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/it-guides/${guideId}/media/${mediaId}`, config)
}

// 获取指南附件列表
export function getITGuideAttachments(guideId) {
  return request.get(`/it-guides/${guideId}/attachments`)
}

// 上传附件（文件或链接）
export function uploadITGuideAttachment(guideId, formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 120000 }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post(`/it-guides/${guideId}/attachments`, formData, config)
}

// 删除附件
export function deleteITGuideAttachment(guideId, attachId, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/it-guides/${guideId}/attachments/${attachId}`, config)
}
