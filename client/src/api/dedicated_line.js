import request from './request'

// 获取专线信息列表
export function getDedicatedLines(params) {
  return request.get('/dedicated-lines', { params })
}

// 创建专线信息（FormData）
export function createDedicatedLine(formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/dedicated-lines', formData, config)
}

// 更新专线信息（FormData）
export function updateDedicatedLine(id, formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/dedicated-lines/${id}`, formData, config)
}

// 删除专线信息
export function deleteDedicatedLine(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/dedicated-lines/${id}`, config)
}

// 删除专线图片
export function deleteDedicatedLineImage(imagePath, lineId, dualToken) {
  const config = { data: { image_path: imagePath, line_id: lineId } }
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete('/dedicated-lines/image', config)
}
