<script setup>
import { ref, computed } from 'vue'
import { GlassContainer } from '@wxperia/liquid-glass-vue'
import { GLASS_PRESETS } from '@/config/glassPresets'
import { useGlassLayout } from '@/composables/useGlassLayout'
import { isSafariEngine } from '@/utils/safari'

const props = defineProps({
  variant: {
    type: String,
    default: 'card',
  },
  className: {
    type: String,
    default: '',
  },
  padding: {
    type: String,
    default: '',
  },
  cornerRadius: {
    type: Number,
    default: undefined,
  },
  onClick: {
    type: Function,
    default: undefined,
  },
  trackMouse: {
    type: Boolean,
    default: true,
  },
  pressable: {
    type: Boolean,
    default: false,
  },
})

const cssGlass = isSafariEngine()
const preset = computed(() => GLASS_PRESETS[props.variant] || GLASS_PRESETS.card)

const wrapRef = ref(null)
const { glassSize } = useGlassLayout(wrapRef, props.variant)
const mouseOffset = ref({ x: 0, y: 0 })
const active = ref(false)

function handleMouseMove(e) {
  if (cssGlass || !props.trackMouse) return
  const el = wrapRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  mouseOffset.value = {
    x: ((e.clientX - rect.left) / rect.width - 0.5) * 100,
    y: ((e.clientY - rect.top) / rect.height - 0.5) * 100,
  }
}

function resetMouse() {
  mouseOffset.value = { x: 0, y: 0 }
  active.value = false
}

const bindProps = computed(() => ({
  mode: preset.value.mode,
  effect: preset.value.effect,
  displacementScale: preset.value.displacementScale,
  blurAmount: preset.value.blurAmount,
  saturation: preset.value.saturation,
  aberrationIntensity: preset.value.aberrationIntensity,
  cornerRadius: props.cornerRadius ?? preset.value.cornerRadius,
  overLight: preset.value.overLight,
  padding: props.padding || preset.value.padding,
  mouseOffset: mouseOffset.value,
  active: active.value,
  glassSize: glassSize.value,
  onClick: props.onClick,
  onMouseLeave: resetMouse,
  onMouseDown: () => {
    if (props.pressable || props.onClick) active.value = true
  },
  onMouseUp: () => {
    active.value = false
  },
}))

const isDock = computed(() => props.variant === 'dock')
const isInline = computed(() => ['badge', 'button'].includes(props.variant))

const glassLayoutClass = computed(() => {
  if (isDock.value) return 'app-glass--dock'
  if (isInline.value) return 'app-glass--inline'
  return 'app-glass--block'
})

const cssRadius = computed(() => `${props.cornerRadius ?? preset.value.cornerRadius}px`)
const cssPadding = computed(() => props.padding || preset.value.padding)
</script>

<template>
  <div
    ref="wrapRef"
    class="app-glass-wrap"
    :data-glass-mode="cssGlass ? 'css' : preset.mode"
    :data-glass-effect="preset.effect"
    :class="[
      className,
      {
        'app-glass-inline': isInline,
        'app-glass-dock-wrap': isDock,
        'app-glass-css-host': cssGlass,
      },
    ]"
    @mousemove="trackMouse ? handleMouseMove : undefined"
    @mouseleave="resetMouse"
  >
    <div
      v-if="cssGlass"
      class="app-glass"
      :class="glassLayoutClass"
      :style="{ width: isInline ? 'auto' : '100%', display: isInline ? 'inline-block' : 'block' }"
    >
      <div
        class="glass app-glass-css"
        :class="{ active }"
        :style="{
          borderRadius: cssRadius,
          padding: cssPadding,
        }"
        @click="onClick"
      >
        <div>
          <slot />
        </div>
      </div>
    </div>
    <GlassContainer
      v-else
      v-bind="bindProps"
      :class-name="`app-glass ${glassLayoutClass}`"
      :style="{ width: isInline ? 'auto' : '100%', display: isInline ? 'inline-block' : 'block' }"
    >
      <slot />
    </GlassContainer>
  </div>
</template>

<style scoped>
.app-glass-wrap {
  width: 100%;
  max-width: 100%;
  margin-left: auto;
  margin-right: auto;
}

.app-glass-inline {
  width: auto;
  display: inline-block;
}

.app-glass-dock-wrap {
  width: 100%;
}

.app-glass--dock :deep(.glass),
.app-glass--dock .glass {
  display: flex !important;
  flex-direction: row !important;
  align-items: center !important;
  width: 100% !important;
  box-sizing: border-box;
}

.app-glass--dock :deep(.glass > div),
.app-glass--dock .glass > div {
  width: 100% !important;
  display: flex !important;
  align-items: center !important;
  box-sizing: border-box;
}

.app-glass--block :deep(.relative),
.app-glass--block {
  width: 100% !important;
  max-width: 100%;
  display: block !important;
  box-sizing: border-box;
}

.app-glass--block :deep(.glass),
.app-glass--block .glass {
  display: flex !important;
  flex-direction: column;
  align-items: stretch;
  width: 100% !important;
  max-width: 100%;
  box-sizing: border-box;
  margin-left: auto;
  margin-right: auto;
}

.app-glass--block :deep(.glass > div),
.app-glass--block .glass > div {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.app-glass--inline :deep(.glass),
.app-glass--inline .glass {
  display: inline-flex !important;
}

.app-glass :deep(.glass > div),
.app-glass .glass > div {
  width: 100%;
  color: var(--text-primary) !important;
  text-shadow: none !important;
}

.app-glass--inline :deep(.glass > div),
.app-glass--inline .glass > div {
  width: auto;
}

.app-glass-css {
  position: relative;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
}

.app-glass--dock .app-glass-css {
  overflow: visible;
}
</style>
