import { ref, watch, onMounted, onUnmounted } from 'vue'
import { estimateGlassSize } from '@/config/glassWarmup'

export function useGlassLayout(wrapRef, variant = 'card') {
  const glassSize = ref(estimateGlassSize(variant))
  const layoutReady = ref(false)

  let frozen = false

  function readSize(el) {
    const width = el.offsetWidth
    const height = el.offsetHeight
    if (width >= 1 && height >= 1) {
      return { width: Math.round(width), height: Math.round(height) }
    }
    const rect = el.getBoundingClientRect()
    if (rect.width < 1 || rect.height < 1) return null
    return { width: Math.round(rect.width), height: Math.round(rect.height) }
  }

  function syncGlassSize() {
    if (frozen) return false
    const el = wrapRef.value
    if (!el) return false
    const next = readSize(el)
    if (!next) return false
    if (next.width === glassSize.value.width && next.height === glassSize.value.height) {
      layoutReady.value = true
      return true
    }
    glassSize.value = next
    layoutReady.value = true
    return true
  }

  function remeasureAfterResize() {
    frozen = false
    syncGlassSize()
    frozen = true
  }

  watch(
    wrapRef,
    (el) => {
      if (!el || frozen) return
      if (syncGlassSize()) frozen = true
    },
    { flush: 'post', immediate: true },
  )

  onMounted(() => {
    const trySync = () => {
      if (!frozen && syncGlassSize()) frozen = true
    }
    trySync()
    requestAnimationFrame(() => {
      trySync()
      requestAnimationFrame(trySync)
    })
    window.addEventListener('resize', remeasureAfterResize)
    window.visualViewport?.addEventListener('resize', remeasureAfterResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', remeasureAfterResize)
    window.visualViewport?.removeEventListener('resize', remeasureAfterResize)
  })

  return { glassSize, layoutReady, syncGlassSize }
}
