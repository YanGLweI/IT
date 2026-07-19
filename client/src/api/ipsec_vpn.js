import request from './request'

// 获取IPsec VPN列表
export function getIPsecVpns(params) {
  return request.get('/ipsec-vpns', { params })
}

// 创建IPsec VPN（FormData）
export function createIPsecVpn(formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/ipsec-vpns', formData, config)
}

// 更新IPsec VPN（FormData）
export function updateIPsecVpn(id, formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/ipsec-vpns/${id}`, formData, config)
}

// 删除IPsec VPN
export function deleteIPsecVpn(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/ipsec-vpns/${id}`, config)
}
