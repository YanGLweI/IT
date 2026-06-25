import request from './request'

export function getTopologies() {
  return request.get('/topologies')
}

export function createTopology(formData) {
  return request.post('/topologies', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function updateTopology(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/topologies/${id}`, data, config)
}

export function replaceTopologyFile(id, formData, dualToken) {
  const headers = { 'Content-Type': 'multipart/form-data' }
  if (dualToken) headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/topologies/${id}/file`, formData, { headers })
}

export function deleteTopology(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/topologies/${id}`, config)
}

export function getTopologyPreviewUrl(id) {
  return `/api/topologies/${id}/preview`
}

export function getTopologyDownloadUrl(id) {
  return `/api/topologies/${id}/download`
}
