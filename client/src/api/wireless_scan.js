import request from './request'

// 获取无线扫描报告列表
export function getWirelessScans(params) {
  return request.get('/wireless-scans', { params })
}

// 上传无线扫描报告
export function createWirelessScan(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/wireless-scans', formData, config)
}

// 更新无线扫描报告
export function updateWirelessScan(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/wireless-scans/${id}`, formData, config)
}

// 删除无线扫描报告
export function deleteWirelessScan(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/wireless-scans/${id}`, config)
}

// 获取预览URL
export function getWirelessScanPreviewUrl(id) {
  return `/api/wireless-scans/${id}/preview`
}

// 获取下载URL
export function getWirelessScanDownloadUrl(id) {
  return `/api/wireless-scans/${id}/download`
}
