import request from './request'

export function getAssets(params) {
  return request.get('/assets', { params })
}

export function getAsset(id) {
  return request.get(`/assets/${id}`)
}

export function createAsset(data) {
  return request.post('/assets', data)
}

export function updateAsset(id, data) {
  return request.put(`/assets/${id}`, data)
}

export function deleteAsset(id) {
  return request.delete(`/assets/${id}`)
}
