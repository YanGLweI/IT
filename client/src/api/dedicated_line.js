import request from './request'

// 获取专线信息列表
export function getDedicatedLines(params) {
  return request.get('/dedicated-lines', { params })
}

// 创建专线信息
export function createDedicatedLine(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/dedicated-lines', data, config)
}

// 更新专线信息
export function updateDedicatedLine(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/dedicated-lines/${id}`, data, config)
}

// 删除专线信息
export function deleteDedicatedLine(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/dedicated-lines/${id}`, config)
}

// 上传专线图片
export function uploadDedicatedLineImage(formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/dedicated-lines/upload-image', formData, config)
}

// 删除专线图片
export function deleteDedicatedLineImage(imagePath, dualToken) {
  const config = { data: { image_path: imagePath } }
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete('/dedicated-lines/image', config)
}
