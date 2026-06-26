import request from './request'

// === 服务器管理 ===
export function getSftpServers() {
  return request.get('/sftp-servers')
}

export function createSftpServer(data, dualToken) {
  return request.post('/sftp-servers', data, withDual(null, dualToken))
}

export function updateSftpServer(id, data, dualToken) {
  return request.put(`/sftp-servers/${id}`, data, withDual(null, dualToken))
}

export function deleteSftpServer(id, dualToken) {
  return request.delete(`/sftp-servers/${id}`, withDual(null, dualToken))
}

// === 账号管理 ===
export function getSftpAccounts(params) {
  return request.get('/sftp-accounts', { params })
}

export function createSftpAccount(data, dualToken) {
  return request.post('/sftp-accounts', data, withDual(null, dualToken))
}

export function updateSftpAccount(id, data, dualToken) {
  return request.put(`/sftp-accounts/${id}`, data, withDual(null, dualToken))
}

export function deleteSftpAccount(id, dualToken) {
  return request.delete(`/sftp-accounts/${id}`, withDual(null, dualToken))
}

export function exportSftpConfirmation(serverId) {
  return request.get('/sftp-accounts/export-confirmation', {
    params: { server_id: serverId },
    responseType: 'blob'
  })
}

function withDual(config, dualToken) {
  if (dualToken) {
    config = config || {}
    config.headers = { ...(config.headers || {}), 'X-Dual-Control-Token': dualToken }
  }
  return config
}
