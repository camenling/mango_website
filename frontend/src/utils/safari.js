/** iOS WebKit（含 iOS Chrome）以及桌面 Safari：液态玻璃 / 100vh / overflow 表现与 Chromium 不同 */

export function detectSafariEngine() {
  if (typeof navigator === 'undefined') return { safari: false, ios: false }
  const ua = navigator.userAgent || ''
  const ios =
    /iP(hone|ad|od)/.test(ua) ||
    (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
  const desktopSafari = /Safari/i.test(ua) && !/Chrome|Chromium|Android/i.test(ua)
  return { safari: ios || desktopSafari, ios }
}

export function markSafariEngine() {
  if (typeof document === 'undefined') return { safari: false, ios: false }
  const { safari, ios } = detectSafariEngine()
  if (safari) document.documentElement.classList.add('is-safari')
  if (ios) document.documentElement.classList.add('is-ios')
  return { safari, ios }
}

export function isSafariEngine() {
  if (typeof document === 'undefined') return false
  if (document.documentElement.classList.contains('is-safari')) return true
  return detectSafariEngine().safari
}
