import request from './request'

// 核准软件目录
export function getApprovedSoftware() {
  return request.get('/approved-software')
}

export function createApprovedSoftware(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/approved-software', data, config)
}

export function updateApprovedSoftware(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/approved-software/${id}`, data, config)
}

export function deleteApprovedSoftware(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/approved-software/${id}`, config)
}

// 资产对应表
export function getAssetSoftwareList(params) {
  return request.get('/asset-software', { params })
}

export function getAssetSoftwareLinks(assetId) {
  return request.get(`/asset-software/${assetId}/links`)
}

export function updateAssetSoftwareLinks(assetId, softwareIds, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/asset-software/${assetId}/links`, { software_ids: softwareIds }, config)
}

export function exportPatchUpdateRecord() {
  return request.get('/asset-software/export-patch-update', {
    responseType: 'blob'
  })
}
