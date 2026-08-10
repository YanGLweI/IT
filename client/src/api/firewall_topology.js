import request from './request'

export function getFirewallTopologyTree() {
  return request.get('/firewall-topology/tree')
}

export function createFirewallNode(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/firewall-topology/nodes', data, config)
}

export function updateFirewallNode(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/firewall-topology/nodes/${id}`, data, config)
}

export function deleteFirewallNode(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/firewall-topology/nodes/${id}`, config)
}

export function saveRegionFirewall(items, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put('/firewall-topology/regions', { items }, config)
}
