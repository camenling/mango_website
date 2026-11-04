import axios from 'axios'

const http = axios.create({
  baseURL: '/api',
  timeout: 120000,
})

http.interceptors.response.use(
  (res) => {
    const data = res.data
    if (data && typeof data.code !== 'undefined') {
      if (data.code !== 0) {
        return Promise.reject(new Error(data.message || '请求失败'))
      }
      return data.data
    }
    return data
  },
  (err) => Promise.reject(err),
)

/** 与主站 uploadCheck 请求体字段一致：p12、mp 为 Base64，password 为证书密码 */
export function signAppWithBase64({ p12, mp, password, theme = 1, accent, glass = true }) {
  return http.post('/sign/sign', {
    p12,
    mp,
    password,
    theme,
    accent,
    glass,
  })
}

export function healthCheck() {
  return http.get('/sign/health')
}

export function lookupBlacklist(udid) {
  return http.post('/sign/blacklist-lookup', { udid: String(udid || '').trim() })
}

export function fetchChannelIpaTracks() {
  return http.get('/sign/channel-ipa-tracks')
}

export async function checkChannelIpa(envelope) {
  const res = await fetch('/api/sign/channel-ipa-check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(envelope),
  })
  let data = {}
  try {
    data = await res.json()
  } catch {
    data = {}
  }
  if (!res.ok || data.error || data.ok === false) {
    const err = new Error(data.error || data.message || 'failed')
    err.code = data.error
    err.status = res.status
    throw err
  }
  return data
}

export async function downloadChannelIpa(envelope) {
  const res = await fetch('/api/sign/channel-ipa', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(envelope),
  })
  const ct = String(res.headers.get('content-type') || '').toLowerCase()
  if (ct.includes('json') || !res.ok) {
    let data = {}
    try {
      data = await res.json()
    } catch {
      data = {}
    }
    const err = new Error(data.error || data.message || 'failed')
    err.code = data.error
    err.status = res.status
    throw err
  }
  const blob = await res.blob()
  const disp = res.headers.get('content-disposition') || ''
  const m = /filename\*?=(?:UTF-8'')?["']?([^";]+)/i.exec(disp)
  return {
    blob,
    filename: m ? decodeURIComponent(m[1]) : 'mango.ipa',
  }
}
