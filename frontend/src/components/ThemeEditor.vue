<script setup>
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { UndoOutlined, PictureOutlined, CloseOutlined, ScissorOutlined } from '@ant-design/icons-vue'
import AppGlass from '@/components/AppGlass.vue'
import ThemeStylePicker from '@/components/ThemeStylePicker.vue'
import ThemePhonePreview from '@/components/ThemePhonePreview.vue'
import BackgroundCropper from '@/components/BackgroundCropper.vue'
import {
  APPEARANCE_MODES,
  BACKGROUND_BASES,
  FOLDER_STYLES,
  TAB_SLOTS,
  builtinStops,
  defaultTabIconUrl,
  defaultThemeState,
  normalizeHex,
  styleAccent,
} from '@/config/hmTheme'

/**
 * 主题方案编辑器：手机预览 + 全部配置分区。
 * 主题配置页和下载页共用；state 由父组件持有（defaultThemeState() 的 reactive 对象），这里直接改它。
 * hide 可以隐藏某些分区：name、style、appearance、background、text、folder、tab。
 */
defineOptions({ name: 'ThemeEditor' })

const props = defineProps({
  state: { type: Object, required: true },
  hide: { type: Array, default: () => [] },
  showPreview: { type: Boolean, default: true },
  /** 嵌在别的玻璃卡片里时用纯色块，避免两层毛玻璃叠成灰色 */
  plain: { type: Boolean, default: false },
})

const { t } = useI18n()
const state = computed(() => props.state)
const show = (key) => !props.hide.includes(key)

const Block = computed(() => (props.plain ? 'section' : AppGlass))
const blockProps = computed(() => (props.plain ? {} : { variant: 'panel', padding: '20px 20px', trackMouse: false }))

// ---------- 预览 ----------

const previewMode = ref('light')
const previewDark = computed(() => previewMode.value === 'dark')

watch(() => state.value.appearance, (mode) => {
  if (mode === 1) previewMode.value = 'light'
  if (mode === 2) previewMode.value = 'dark'
})

watch(() => state.value.style, (style) => {
  state.value.background.stops = builtinStops(style)
})

const modeOptions = computed(() =>
  APPEARANCE_MODES.map((m) => ({ label: t(`themeCfg.appearance.${m.key}`), value: m.value })),
)
const baseOptions = computed(() =>
  BACKGROUND_BASES.map((m) => ({ label: t(`themeCfg.bg.base.${m.key}`), value: m.value })),
)
const bgModeOptions = computed(() => [
  { label: t('themeCfg.bg.gradient'), value: 0 },
  { label: t('themeCfg.bg.image'), value: 1 },
])
const previewOptions = computed(() => [
  { label: t('themeCfg.appearance.light'), value: 'light' },
  { label: t('themeCfg.appearance.dark'), value: 'dark' },
])

// ---------- 字体颜色 ----------

const colorRows = computed(() => [
  { role: 'accent', label: t('themeCfg.text.accent') },
  { role: 'primary', label: t('themeCfg.text.primary') },
  { role: 'secondary', label: t('themeCfg.text.secondary') },
])

function colorKey(role, dark) {
  return `${role}${dark ? 'Dark' : 'Light'}`
}

function colorDefault(role, dark) {
  const s = state.value
  if (role === 'accent') return dark ? s.textColor.accentLight || styleAccent(s.style) : styleAccent(s.style)
  if (role === 'primary') return dark ? '#FFFFFF' : '#000000'
  return dark ? '#EBEBF5' : '#3C3C43'
}

function colorValue(role, dark) {
  return state.value.textColor[colorKey(role, dark)] || colorDefault(role, dark)
}

function setColor(role, dark, value) {
  state.value.textColor[colorKey(role, dark)] = normalizeHex(value) || value
}

function resetColor(role, dark) {
  state.value.textColor[colorKey(role, dark)] = null
}

// ---------- 背景 ----------

const objectUrls = new Set()

function trackUrl(file) {
  const url = URL.createObjectURL(file)
  objectUrls.add(url)
  return url
}

function releaseUrl(url) {
  if (url && objectUrls.has(url)) {
    URL.revokeObjectURL(url)
    objectUrls.delete(url)
  }
}

onBeforeUnmount(() => {
  objectUrls.forEach((url) => URL.revokeObjectURL(url))
  objectUrls.clear()
})

/** 原图留着，可以反复调整裁切；state.background.image 是裁切后的图，预览和打包都用它。 */
const bgSource = ref(null)
const bgCrop = ref(null)
const cropOpen = ref(false)

function pickBackground(file) {
  if (!file.type.startsWith('image/')) {
    message.error(t('themeCfg.errImage'))
    return false
  }
  releaseUrl(bgSource.value)
  bgSource.value = trackUrl(file)
  bgCrop.value = null
  cropOpen.value = true
  return false
}

function onCropConfirm({ blob, crop }) {
  releaseUrl(state.value.background.image)
  state.value.background.image = trackUrl(blob)
  state.value.background.mode = 1
  bgCrop.value = crop
}

function clearBackground() {
  releaseUrl(state.value.background.image)
  releaseUrl(bgSource.value)
  state.value.background.image = null
  state.value.background.mode = 0
  bgSource.value = null
  bgCrop.value = null
}

function resetStops() {
  state.value.background.stops = builtinStops(state.value.style)
}

const percent = (value) => `${Math.round(value * 100)}%`

// ---------- 底栏 ----------

function pickTabIcon(slot, selected, file) {
  if (!file.type.startsWith('image/')) {
    message.error(t('themeCfg.errImage'))
    return false
  }
  const key = selected ? 'selected' : 'normal'
  releaseUrl(state.value.tabIcon.icons[slot][key])
  state.value.tabIcon.icons[slot][key] = trackUrl(file)
  return false
}

function resetTabIcon(slot, selected) {
  const key = selected ? 'selected' : 'normal'
  releaseUrl(state.value.tabIcon.icons[slot][key])
  state.value.tabIcon.icons[slot][key] = null
}

function tabIconSrc(slot, selected) {
  return state.value.tabIcon.icons[slot][selected ? 'selected' : 'normal'] || defaultTabIconUrl(slot, selected)
}

// ---------- 重置 ----------

function resetAll() {
  objectUrls.forEach((url) => URL.revokeObjectURL(url))
  objectUrls.clear()
  bgSource.value = null
  bgCrop.value = null
  Object.assign(state.value, defaultThemeState(state.value.style))
}

defineExpose({ resetAll })
</script>

<template>
  <div class="config-layout" :class="{ 'config-layout--no-preview': !showPreview }">
    <aside v-if="showPreview" class="config-preview">
      <ThemePhonePreview :state="state" :dark="previewDark" />
      <a-segmented
        v-if="state.appearance === 0"
        v-model:value="previewMode"
        :options="previewOptions"
        size="small"
        class="preview-switch"
      />
      <p class="preview-note">{{ t('themeCfg.previewNote') }}</p>
    </aside>

    <div class="config-form">
      <component :is="Block" v-if="show('name')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.nameLabel') }}</h2>
        <a-input v-model:value="state.name" :placeholder="t('themeCfg.defaultName')" :maxlength="40" />
      </component>

      <component :is="Block" v-if="show('style')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.styleLabel') }}</h2>
        <p class="section-hint">{{ t('themeCfg.styleHint') }}</p>
        <ThemeStylePicker v-model="state.style" />
        <div class="row">
          <span class="row__label">
            {{ t('themeCfg.glassLabel') }}
            <small>{{ t('themeCfg.glassHint') }}</small>
          </span>
          <a-switch v-model:checked="state.glass" />
        </div>
      </component>

      <component :is="Block" v-if="show('appearance')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.appearanceLabel') }}</h2>
        <a-segmented v-model:value="state.appearance" :options="modeOptions" block />
      </component>

      <component :is="Block" v-if="show('background')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.bg.title') }}</h2>
        <p class="section-hint">{{ t('themeCfg.bg.hint') }}</p>
        <a-segmented v-model:value="state.background.mode" :options="bgModeOptions" block />

        <template v-if="state.background.mode === 0">
          <div v-for="(hex, i) in state.background.stops.hexes" :key="i" class="row">
            <span class="row__label">{{ t('themeCfg.bg.stop', { n: i + 1 }) }}</span>
            <div class="row__control">
              <a-input-number
                v-model:value="state.background.stops.locations[i]"
                :min="0"
                :max="1"
                :step="0.05"
                size="small"
                class="loc-input"
              />
              <input v-model="state.background.stops.hexes[i]" type="color" class="color-swatch" />
              <a-input v-model:value="state.background.stops.hexes[i]" class="color-text" :maxlength="7" />
            </div>
          </div>
          <a-button size="small" class="inline-btn" @click="resetStops">
            <UndoOutlined /> {{ t('themeCfg.bg.resetStops') }}
          </a-button>
        </template>

        <template v-else>
          <div class="bg-picker">
            <a-upload :show-upload-list="false" :before-upload="pickBackground" accept="image/*">
              <button type="button" class="bg-thumb">
                <img v-if="state.background.image" :src="state.background.image" alt="" />
                <span v-else><PictureOutlined /> {{ t('themeCfg.bg.pick') }}</span>
              </button>
            </a-upload>
            <div v-if="state.background.image" class="bg-actions">
              <a-button size="small" @click="cropOpen = true">
                <ScissorOutlined /> {{ t('themeCfg.bg.recrop') }}
              </a-button>
              <a-button size="small" @click="clearBackground">
                <CloseOutlined /> {{ t('themeCfg.bg.remove') }}
              </a-button>
            </div>
          </div>
          <BackgroundCropper
            v-model:open="cropOpen"
            :src="bgSource || ''"
            :initial="bgCrop"
            @confirm="onCropConfirm"
          />
          <div class="slider-row">
            <span>{{ t('themeCfg.bg.opacity') }} · {{ percent(state.background.opacity) }}</span>
            <a-slider v-model:value="state.background.opacity" :min="0.1" :max="1" :step="0.01" :tooltip-open="false" />
          </div>
          <div class="slider-row">
            <span>{{ t('themeCfg.bg.blur') }} · {{ percent(state.background.blur) }}</span>
            <a-slider v-model:value="state.background.blur" :min="0" :max="1" :step="0.01" :tooltip-open="false" />
          </div>
          <div class="slider-row">
            <span>{{ t('themeCfg.bg.baseLabel') }}</span>
            <a-segmented v-model:value="state.background.base" :options="baseOptions" block />
          </div>
        </template>
      </component>

      <component :is="Block" v-if="show('text')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.text.title') }}</h2>
        <p class="section-hint">{{ t('themeCfg.text.hint') }}</p>
        <div class="color-grid">
          <span />
          <span class="color-grid__head">{{ t('themeCfg.appearance.light') }}</span>
          <span class="color-grid__head">{{ t('themeCfg.appearance.dark') }}</span>
          <template v-for="row in colorRows" :key="row.role">
            <span class="color-grid__label">{{ row.label }}</span>
            <div v-for="dark in [false, true]" :key="String(dark)" class="color-cell">
              <input
                type="color"
                class="color-swatch"
                :value="colorValue(row.role, dark)"
                @input="setColor(row.role, dark, $event.target.value)"
              />
              <button
                v-if="state.textColor[colorKey(row.role, dark)]"
                type="button"
                class="link-btn"
                @click="resetColor(row.role, dark)"
              >
                {{ t('themeCfg.text.reset') }}
              </button>
              <span v-else class="muted">{{ t('themeCfg.text.default') }}</span>
            </div>
          </template>
        </div>
      </component>

      <component :is="Block" v-if="show('folder')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <h2 class="section-title">{{ t('themeCfg.folder.title') }}</h2>
        <div class="folder-options">
          <button
            v-for="f in FOLDER_STYLES"
            :key="f.value"
            type="button"
            class="folder-option"
            :class="{ 'is-active': state.folderIcon === f.value }"
            @click="state.folderIcon = f.value"
          >
            <img :src="f.image" alt="" />
            <span>{{ t(`themeCfg.folder.${f.key}`) }}</span>
          </button>
        </div>
      </component>

      <component :is="Block" v-if="show('tab')" v-bind="blockProps" class="theme-block" :class="{ 'theme-block--plain': plain }">
        <div class="row row--flat">
          <span class="row__label">
            {{ t('themeCfg.tab.title') }}
            <small>{{ t('themeCfg.tab.hint') }}</small>
          </span>
          <a-switch v-model:checked="state.tabIcon.enabled" />
        </div>
        <template v-if="state.tabIcon.enabled">
          <a-input
            v-model:value="state.tabIcon.name"
            :placeholder="t('themeCfg.tab.namePlaceholder')"
            :maxlength="40"
            class="tab-name"
          />
          <div class="tab-grid">
            <span />
            <span class="color-grid__head">{{ t('themeCfg.tab.normal') }}</span>
            <span class="color-grid__head">{{ t('themeCfg.tab.selected') }}</span>
            <template v-for="slot in TAB_SLOTS" :key="slot">
              <span class="color-grid__label">{{ t(`themeCfg.tabSlot.${slot}`) }}</span>
              <div v-for="selected in [false, true]" :key="String(selected)" class="tab-cell">
                <a-upload
                  :show-upload-list="false"
                  :before-upload="(file) => pickTabIcon(slot, selected, file)"
                  accept="image/png,image/*"
                >
                  <button type="button" class="tab-thumb">
                    <img :src="tabIconSrc(slot, selected)" alt="" />
                  </button>
                </a-upload>
                <button
                  v-if="state.tabIcon.icons[slot][selected ? 'selected' : 'normal']"
                  type="button"
                  class="link-btn"
                  @click="resetTabIcon(slot, selected)"
                >
                  {{ t('themeCfg.text.reset') }}
                </button>
              </div>
            </template>
          </div>
          <p class="section-hint">{{ t('themeCfg.tab.sizeHint') }}</p>
        </template>
      </component>

      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.config-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
}

.config-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.preview-note {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

@media (min-width: 900px) {
  .config-layout:not(.config-layout--no-preview) {
    grid-template-columns: 320px minmax(0, 1fr);
    align-items: start;
  }

  .config-preview {
    position: sticky;
    top: 96px;
  }
}

.theme-block--plain {
  box-sizing: border-box;
  padding: 16px 16px 18px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.3);
  border: 0.5px solid rgba(255, 255, 255, 0.55);
}

.theme-block--plain .row,
.theme-block--plain .folder-option {
  background: rgba(255, 255, 255, 0.34);
}

.section-title {
  margin: 0 0 6px;
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
}

.section-hint {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 10px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.22);
}

.row--flat {
  margin-top: 0;
  padding: 0;
  background: none;
}

.row__label {
  display: flex;
  flex-direction: column;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.row__label small {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.row__control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.loc-input {
  width: 76px;
}

.color-swatch {
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 9px;
  background: none;
  cursor: pointer;
}

.color-text {
  width: 96px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.inline-btn {
  margin-top: 10px;
}

.bg-picker {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin: 12px 0 4px;
}

.bg-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bg-thumb {
  width: 96px;
  aspect-ratio: 9 / 19.5;
  border: 1px dashed rgba(0, 0, 0, 0.2);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.3);
  overflow: hidden;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 12px;
}

.bg-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.slider-row {
  margin-top: 10px;
  font-size: 13px;
  color: var(--text-secondary);
}

.color-grid,
.tab-grid {
  display: grid;
  grid-template-columns: auto 1fr 1fr;
  gap: 10px 14px;
  align-items: center;
}

.color-grid__head {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.color-grid__label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.color-cell,
.tab-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.link-btn {
  border: none;
  background: none;
  padding: 0;
  font-size: 12px;
  color: var(--theme-accent, #0a84ff);
  cursor: pointer;
}

.muted {
  font-size: 12px;
  color: var(--text-muted);
}

.folder-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.folder-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px;
  border: 1.5px solid transparent;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.22);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.folder-option img {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.folder-option.is-active {
  border-color: var(--theme-accent, #0a84ff);
  color: var(--text-primary);
}

.tab-name {
  margin: 12px 0;
}

.tab-thumb {
  width: 44px;
  height: 44px;
  padding: 6px;
  border: 1px dashed rgba(0, 0, 0, 0.2);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.4);
  cursor: pointer;
}

.tab-thumb img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.tab-grid + .section-hint {
  margin-top: 12px;
}
</style>
