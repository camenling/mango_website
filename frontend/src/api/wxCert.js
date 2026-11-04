// 公众号凭证页的接口。返回 { code, message, data }；code 401 表示需要重新微信授权。
export async function wxCertCall(action, body = {}) {
  let res
  try {
    res = await fetch(`/api/wx-cert/${action}`, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    return { code: -1, message: '' }
  }
  try {
    return await res.json()
  } catch {
    return { code: -1, message: '' }
  }
}

export const WX_CERT_LOGIN_URL = '/api/wx-cert/login'
