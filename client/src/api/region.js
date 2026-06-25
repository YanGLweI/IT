import request from './request'

export function getRegions() {
  return request.get('/regions')
}

export function createRegion(data) {
  return request.post('/regions', data)
}

export function updateRegion(id, data) {
  return request.put(`/regions/${id}`, data)
}

export function deleteRegion(id) {
  return request.delete(`/regions/${id}`)
}
