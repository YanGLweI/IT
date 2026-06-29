import forge from 'node-forge'

let publicKeyCache = null

/**
 * 从后端获取RSA公钥
 */
async function fetchPublicKey() {
  const response = await fetch('/api/public-key')
  const result = await response.json()
  if (result.code !== 200) {
    throw new Error('获取公钥失败')
  }
  return result.data.public_key
}

/**
 * 使用RSA公钥加密明文（OAEP + SHA-256）
 * @param {string} plainText 明文
 * @returns {Promise<string>} base64编码的密文
 */
export async function encryptPassword(plainText) {
  if (!publicKeyCache) {
    publicKeyCache = await fetchPublicKey()
  }

  const publicKey = forge.pki.publicKeyFromPem(publicKeyCache)
  const encrypted = publicKey.encrypt(plainText, 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: { md: forge.md.sha256.create() }
  })
  return forge.util.encode64(encrypted)
}

/**
 * 清除公钥缓存（可选，用于密钥轮换场景）
 */
export function clearPublicKeyCache() {
  publicKeyCache = null
}
