import request from './request'

// 获取防火墙检查记录列表
export function getFirewallChecks(params) {
  return request.get('/firewall-checks', { params })
}

// 创建防火墙检查记录
export function createFirewallCheck(data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.post('/firewall-checks', data, config)
}

// 更新防火墙检查记录
export function updateFirewallCheck(id, data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.put(`/firewall-checks/${id}`, data, config)
}

// 删除防火墙检查记录
export function deleteFirewallCheck(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/firewall-checks/${id}`, config)
}

// 上传整改报告
export function uploadFirewallRectReport(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/firewall-checks/${id}/rect`, formData, config)
}

// 删除整改报告
export function deleteFirewallRectReport(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/firewall-checks/${id}/rect`, config)
}

// 获取整改报告预览URL
export function getFirewallRectPreviewUrl(id) {
  return `/api/firewall-checks/${id}/rect-preview`
}

// 获取整改报告下载URL
export function getFirewallRectDownloadUrl(id) {
  return `/api/firewall-checks/${id}/rect-download`
}
