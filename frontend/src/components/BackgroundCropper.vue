<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

defineOptions({ name: 'BackgroundCropper' })

/** 裁切框固定为手机屏幕比例（1290×2796），和 App 内选背景图的裁切一致：拖动、缩放选取要铺到页面上的那一块。 */
const OUTPUT_W = 1290
const OUTPUT_H = 2796
const MAX_ZOOM = 4

const props = defineProps({
  open: { type: Boolean, default: false },
  src: { type: String, default: '' },
  /** 上次的裁切位置：{ zoom, cx, cy }，cx/cy 是裁切框中心在原图上的比例位置 */
  initial: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'confirm'])

const { t } = useI18n()

const frame = ref(null)
const natural = ref({ w: 0, h: 0 })
const box = ref({ w: 0, h: 0 })
const zoom = ref(1)
const offset = ref({ x: 0, y: 0 })
const saving = ref(false)

/** 缩放为 1 时刚好铺满裁切框 */
const baseScale = computed(() => {
  const { w, h } = natural.value
  if (!w || !h || !box.value.w) return 1
  return Math.max(box.value.w / w, box.value.h / h)
})
const scale = computed(() => baseScale.value * zoom.value)

const imageStyle = computed(() => ({
  width: `${natural.value.w * scale.value}px`,
  height: `${natural.value.h * scale.value}px`,
  transform: `translate(${offset.value.x}px, ${offset.value.y}px)`,
}))

function clampOffset(x, y) {
  const w = natural.value.w * scale.value
  const h = natural.value.h * scale.value
  return {
    x: Math.min(0, Math.max(box.value.w - w, x)),
    y: Math.min(0, Math.max(box.value.h - h, y)),
  }
}

/** 让原图上 (cx, cy) 这一点落在裁切框中心 */
function centerOn(cx, cy) {
  offset.value = clampOffset(
    box.value.w / 2 - cx * natural.value.w * scale.value,
    box.value.h / 2 - cy * natural.value.h * scale.value,
  )
}

function currentCenter() {
  return {
    cx: (box.value.w / 2 - offset.value.x) / (natural.value.w * scale.value),
    cy: (box.value.h / 2 - offset.value.y) / (natural.value.h * scale.value),
  }
}

async function setup() {
  if (!props.src) return
  const img = new Image()
  img.src = props.src
  await img.decode()
  natural.value = { w: img.naturalWidth, h: img.naturalHeight }
  await nextTick()
  const rect = frame.value.getBoundingClientRect()
  box.value = { w: rect.width, h: rect.height }
  settingUp = true
  zoom.value = props.initial?.zoom ?? 1
  centerOn(props.initial?.cx ?? 0.5, props.initial?.cy ?? 0.5)
  await nextTick()
  settingUp = false
}

watch(() => [props.open, props.src], ([open]) => {
  if (open) setup()
})

let settingUp = false
let drag = null

/** 缩放时保持裁切框中心对着原图同一点 */
watch(zoom, (next, prev) => {
  if (settingUp || !natural.value.w) return
  const prevScale = baseScale.value * prev
  centerOn(
    (box.value.w / 2 - offset.value.x) / (natural.value.w * prevScale),
    (box.value.h / 2 - offset.value.y) / (natural.value.h * prevScale),
  )
})

function onPointerDown(e) {
  drag = { x: e.clientX, y: e.clientY, ox: offset.value.x, oy: offset.value.y }
  e.currentTarget.setPointerCapture(e.pointerId)
}

function onPointerMove(e) {
  if (!drag) return
  offset.value = clampOffset(drag.ox + e.clientX - drag.x, drag.oy + e.clientY - drag.y)
}

function onPointerUp() {
  drag = null
}

function onWheel(e) {
  zoom.value = Math.min(MAX_ZOOM, Math.max(1, zoom.value * (e.deltaY < 0 ? 1.08 : 1 / 1.08)))
}

function close() {
  emit('update:open', false)
}

async function confirm() {
  saving.value = true
  try {
    const img = new Image()
    img.src = props.src
    await img.decode()
    const sx = -offset.value.x / scale.value
    const sy = -offset.value.y / scale.value
    const sw = box.value.w / scale.value
    const sh = box.value.h / scale.value
    // 原图选区比输出尺寸小时不放大，保持原分辨率
    const outW = Math.min(OUTPUT_W, Math.round(sw))
    const outH = Math.round((outW * OUTPUT_H) / OUTPUT_W)
    const canvas = document.createElement('canvas')
    canvas.width = outW
    canvas.height = outH
    canvas.getContext('2d').drawImage(img, sx, sy, sw, sh, 0, 0, outW, outH)
    const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.92))
    emit('confirm', { blob, crop: { zoom: zoom.value, ...currentCenter() } })
    close()
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    :title="t('themeCfg.crop.title')"
    centered
    :width="380"
    :mask-closable="false"
    destroy-on-close
    @cancel="close"
  >
    <p class="crop-hint">{{ t('themeCfg.crop.hint') }}</p>
    <div
      ref="frame"
      class="crop-frame"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
      @wheel.prevent="onWheel"
    >
      <img v-if="src" :src="src" :style="imageStyle" class="crop-image" alt="" draggable="false" />
    </div>
    <div class="crop-zoom">
      <span>{{ t('themeCfg.crop.zoom') }}</span>
      <a-slider
        v-model:value="zoom"
        :min="1"
        :max="MAX_ZOOM"
        :step="0.01"
        :tooltip-open="false"
      />
    </div>
    <template #footer>
      <a-button @click="close">{{ t('themeCfg.crop.cancel') }}</a-button>
      <a-button type="primary" :loading="saving" @click="confirm">{{ t('themeCfg.crop.done') }}</a-button>
    </template>
  </a-modal>
</template>

<style scoped>
.crop-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.crop-frame {
  position: relative;
  width: 240px;
  aspect-ratio: 1290 / 2796;
  margin: 0 auto;
  overflow: hidden;
  border-radius: 20px;
  background: #111;
  cursor: grab;
  touch-action: none;
  user-select: none;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.15);
}

.crop-frame:active {
  cursor: grabbing;
}

.crop-image {
  position: absolute;
  top: 0;
  left: 0;
  max-width: none;
  pointer-events: none;
}

.crop-zoom {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
  font-size: 13px;
  color: var(--text-secondary);
}

.crop-zoom :deep(.ant-slider) {
  flex: 1;
}
</style>
