import request from './request'

export function getTopologies() {
  return request.get('/topologies')
}

export function createTopology(formData) {
  return request.post('/topologies', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function updateTopology(id, data) {
  return request.put(`/topologies/${id}`, data)
}

export function replaceTopologyFile(id, formData) {
  return request.put(`/topologies/${id}/file`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function deleteTopology(id) {
  return request.delete(`/topologies/${id}`)
}

export function getTopologyPreviewUrl(id) {
  return `/api/topologies/${id}/preview`
}

export function getTopologyDownloadUrl(id) {
  return `/api/topologies/${id}/download`
}
