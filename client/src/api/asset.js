import request from './request'

export function getAssets(params) {
  return request.get('/assets', { params })
}

export function getAsset(id) {
  return request.get(`/assets/${id}`)
}

export function createAsset(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/assets', data, config)
}

export function updateAsset(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/assets/${id}`, data, config)
}

export function deleteAsset(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/assets/${id}`, config)
}
