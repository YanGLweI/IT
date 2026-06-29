import request from './request'
import { encryptPassword } from '@/utils/rsa'

export async function verifyDualControl(data) {
  const encryptedPassword = await encryptPassword(data.password)
  return request.post('/dual-control/verify', {
    username: data.username,
    password: encryptedPassword
  })
}
