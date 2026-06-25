import request from './request'

export function verifyDualControl(data) {
  return request.post('/dual-control/verify', data)
}
