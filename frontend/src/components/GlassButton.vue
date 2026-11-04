<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { GlassContainer } from '@wxperia/liquid-glass-vue'
import { GLASS_PRESETS } from '@/config/glassPresets'
import { useGlassLayout } from '@/composables/useGlassLayout'
import { isSafariEngine } from '@/utils/safari'

const props = defineProps({
  to: { type: [String, Object], default: null },
  href: { type: String, default: null },
  primary: { type: Boolean, default: false },
  block: { type: Boolean, default: false },
})

const router = useRouter()
const cssGlass = isSafariEngine()
const preset = computed(() => ({
  ...GLASS_PRESETS.button,
  ...(props.primary ? { displacementScale: 72, effect: 'liquidGlass' } : {}),
}))

const wrapRef = ref(null)
const { glassSize } = useGlassLayout(wrapRef)
const mouseOffset = ref({ x: 0, y: 0 })
const active = ref(false)

function handleMouseMove(e) {
  if (cssGlass) return
  const el = wrapRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  mouseOffset.value = {
    x: ((e.clientX - rect.left) / rect.width - 0.5) * 100,
    y: ((e.clientY - rect.top) / rect.height - 0.5) * 100,
  }
}

function handleClick() {
  if (props.to) router.push(props.to)
  else if (props.href) {
    if (props.href.startsWith('#')) {
      document.querySelector(props.href)?.scrollIntoView({ behavior: 'smooth' })
    } else {
      window.location.href = props.href
    }
  }
}

const hasAction = computed(() => Boolean(props.to || props.href))
</script>

<template>
  <div
    ref="wrapRef"
    class="glass-btn-wrap"
    :class="{ block }"
    @mousemove="handleMouseMove"
    @mouseleave="mouseOffset = { x: 0, y: 0 }; active = false"
  >
    <button
      v-if="cssGlass"
      type="button"
      class="glass-btn-css"
      :class="{ primary, active, block }"
      @click="hasAction ? handleClick() : undefined"
    >
      <span class="glass-btn-label" :class="{ primary }">
        <slot />
      </span>
    </button>
    <GlassContainer
      v-else
      :mode="preset.mode"
      :effect="preset.effect"
      :displacement-scale="preset.displacementScale"
      :blur-amount="preset.blurAmount"
      :saturation="preset.saturation"
      :aberration-intensity="preset.aberrationIntensity"
      :corner-radius="preset.cornerRadius"
      :over-light="preset.overLight"
      :padding="preset.padding"
      :mouse-offset="mouseOffset"
      :glass-size="glassSize"
      :active="active"
      :on-click="hasAction ? handleClick : undefined"
      class-name="glass-btn"
      :style="{ width: block ? '100%' : 'auto', display: block ? 'block' : 'inline-block' }"
      @mousedown="active = true"
      @mouseup="active = false"
    >
      <span class="glass-btn-label" :class="{ primary }">
        <slot />
      </span>
    </GlassContainer>
  </div>
</template>

<style scoped>
.glass-btn-wrap {
  display: inline-block;
}

.glass-btn-wrap.block {
  display: block;
  width: 100%;
}

.glass-btn-label {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  white-space: nowrap;
}

.glass-btn-label.primary {
  color: #fff;
}

.glass-btn :deep(.glass) {
  cursor: pointer;
  transition: transform 0.2s ease-out !important;
  border-radius: 9999px !important;
  overflow: hidden;
}

.glass-btn :deep(.glass.active) {
  transform: scale(0.96);
}

.glass-btn :deep(.glass > div) {
  display: flex !important;
  align-items: center;
  justify-content: center;
  text-shadow: none !important;
  font: inherit !important;
  color: var(--text-primary) !important;
  transition: none !important;
}

.glass-btn:has(.glass-btn-label.primary) :deep(.glass > div) {
  color: #fff !important;
}

.glass-btn :deep(.glass__warp) {
  transition: none !important;
}

.glass-btn:has(.primary) :deep(.glass) {
  background: linear-gradient(135deg, rgba(0, 122, 255, 0.78), rgba(88, 86, 214, 0.68));
}

.glass-btn-css {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: auto;
  margin: 0;
  padding: 12px 28px;
  border: 0.5px solid rgba(255, 255, 255, 0.45);
  border-radius: 9999px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.28);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(24px) saturate(160%);
  -webkit-backdrop-filter: blur(24px) saturate(160%);
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
  appearance: none;
  -webkit-appearance: none;
}

.glass-btn-css.block {
  display: flex;
  width: 100%;
}

.glass-btn-css.primary {
  background: linear-gradient(135deg, rgba(0, 122, 255, 0.78), rgba(88, 86, 214, 0.68));
  border-color: rgba(255, 255, 255, 0.35);
  color: #fff;
}

.glass-btn-css.active,
.glass-btn-css:active {
  transform: scale(0.96);
}
</style>
