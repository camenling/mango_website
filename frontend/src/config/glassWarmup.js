import { GLASS_PRESETS } from '@/config/glassPresets'

/** 与 `.site-width-shell` 一致的 content 宽度估算 */
export function estimateShellContentWidth() {
  if (typeof window === 'undefined') return 792
  return Math.round(Math.min(Math.max(window.innerWidth - 48, 320), 840))
}

const VARIANT_HEIGHT_HINT = {
  dock: 52,
  card: 400,
  panel: 200,
  button: 44,
  badge: 32,
  pill: 32,
}

export function estimateGlassSize(variant = 'card') {
  return {
    width: estimateShellContentWidth(),
    height: VARIANT_HEIGHT_HINT[variant] ?? 120,
  }
}

/** F5 前预生成 shader，避免首帧占位图与真实 liquid 图切换 */
export function installRouteWarmupTargets() {
  const w = estimateShellContentWidth()
  const effect = GLASS_PRESETS.panel.effect
  const dockEffect = GLASS_PRESETS.dock.effect
  return [
    { width: w, height: 52, effect: dockEffect },
    { width: w, height: 134, effect },
    { width: w, height: 517, effect },
  ]
}

/** 挂载后按 DOM 实测尺寸补齐 shader 缓存，再展示页面 */
export async function primeGlassShadersFromDom() {
  const { prefetchShaderMap } = await import('@wxperia/liquid-glass-vue')
  const wraps = document.querySelectorAll('.app-glass-wrap:not(.app-glass-css-host)')
  const jobs = []

  wraps.forEach((wrap) => {
    const glass = wrap.querySelector('.glass')
    if (!glass) return
    const { width, height } = wrap.getBoundingClientRect()
    const w = Math.round(width)
    const h = Math.round(height)
    if (w < 2 || h < 2) return
    const mode = wrap.dataset.glassMode
    const effect = wrap.dataset.glassEffect
    if (mode !== 'shader' || !effect) return
    jobs.push(prefetchShaderMap(w, h, effect))
  })

  if (jobs.length) await Promise.all(jobs)
}
