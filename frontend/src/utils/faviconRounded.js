/**
 * iOS 桌面图标 squircle 近似：超椭圆 |x/a|^n + |y/b|^n = 1
 * n≈4 接近系统观感（非 Apple 精确贝塞尔蒙版）
 */
export const SQUIRCLE_EXPONENT = 4

const dataUrlCache = new Map()

export function squirclePath(ctx, x, y, size, n = SQUIRCLE_EXPONENT) {
  const cx = x + size / 2
  const cy = y + size / 2
  const a = size / 2
  const b = size / 2
  const steps = 128
  const power = 2 / n

  ctx.beginPath()
  for (let i = 0; i <= steps; i++) {
    const t = (i / steps) * Math.PI * 2
    const cos = Math.cos(t)
    const sin = Math.sin(t)
    const px = cx + a * Math.sign(cos) * Math.pow(Math.abs(cos), power)
    const py = cy + b * Math.sign(sin) * Math.pow(Math.abs(sin), power)
    if (i === 0) {
      ctx.moveTo(px, py)
    } else {
      ctx.lineTo(px, py)
    }
  }
  ctx.closePath()
}

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.decoding = 'async'
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error(`favicon load failed: ${src}`))
    img.src = src
  })
}

export function themeIconSrc(themeId) {
  return `/themes/${themeId}.png?v=${themeId}`
}

export async function roundedFaviconDataUrl(src, size = 64, cacheKey = src) {
  const key = `${cacheKey}@${size}`
  if (dataUrlCache.has(key)) {
    return dataUrlCache.get(key)
  }

  const task = (async () => {
    const img = await loadImage(src)
    const canvas = document.createElement('canvas')
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      throw new Error('canvas unsupported')
    }
    squirclePath(ctx, 0, 0, size)
    ctx.clip()
    ctx.drawImage(img, 0, 0, size, size)
    return canvas.toDataURL('image/png')
  })()

  dataUrlCache.set(key, task)
  return task
}

export function applyFaviconToDocument({ iconHref, appleHref }) {
  document
    .querySelectorAll('link[rel="icon"][type="image/svg+xml"], link[href*="favicon.svg"]')
    .forEach((el) => el.remove())

  const specs = [
    { rel: 'icon', href: iconHref },
    { rel: 'shortcut icon', href: iconHref },
    { rel: 'apple-touch-icon', href: appleHref || iconHref },
  ]

  for (const { rel, href } of specs) {
    let link = document.querySelector(`link[rel="${rel}"]`)
    if (!link) {
      link = document.createElement('link')
      link.rel = rel
      document.head.appendChild(link)
    }
    link.type = 'image/png'
    link.href = href
  }
}
