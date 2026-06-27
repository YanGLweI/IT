import request from './request'

export function getRegions() {
  return request.get('/regions')
}

export function createRegion(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/regions', data, config)
}

export function updateRegion(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/regions/${id}`, data, config)
}

export function deleteRegion(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/regions/${id}`, config)
}
