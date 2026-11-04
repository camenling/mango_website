<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  FOLDER_STYLES,
  TAB_SLOTS,
  defaultTabIconUrl,
  styleAccent,
  wallpaperCss,
} from '@/config/hmTheme'

defineOptions({ name: 'ThemePhonePreview' })

const props = defineProps({
  state: { type: Object, required: true },
  dark: { type: Boolean, default: false },
})

const { t } = useI18n()

/** 底栏可以点击切换，用来查看各图标选中和未选中的样子；默认停在「文件」。 */
const selectedTab = ref(3)

const accent = computed(() => {
  const light = props.state.textColor.accentLight || styleAccent(props.state.style)
  return props.dark ? props.state.textColor.accentDark || light : light
})

const primaryText = computed(() =>
  props.dark
    ? props.state.textColor.primaryDark || '#FFFFFF'
    : props.state.textColor.primaryLight || '#000000',
)

const secondaryText = computed(() =>
  props.dark
    ? props.state.textColor.secondaryDark || 'rgba(235, 235, 245, 0.6)'
    : props.state.textColor.secondaryLight || 'rgba(60, 60, 67, 0.6)',
)

const usesImage = computed(() => {
  const bg = props.state.background
  return props.state.glass && bg.mode === 1 && !!bg.image
})

const baseColor = computed(() => {
  const base = props.state.background.base
  if (base === 1) return '#FFFFFF'
  if (base === 2) return '#000000'
  return props.dark ? '#000000' : '#FFFFFF'
})

const screenStyle = computed(() => {
  if (!props.state.glass) {
    return { background: props.dark ? '#000000' : '#F2F2F7' }
  }
  if (usesImage.value) {
    return { background: baseColor.value }
  }
  return { background: wallpaperCss(props.state.style, props.state.background.stops, props.dark) }
})

const photoStyle = computed(() => ({
  backgroundImage: `url(${props.state.background.image})`,
  opacity: props.state.background.opacity,
  filter: `blur(${props.state.background.blur * 12}px)`,
}))

const cardClass = computed(() => [
  'pv-card',
  props.state.glass ? 'pv-card--glass' : 'pv-card--plain',
  props.dark ? 'is-dark' : '',
])

const folderImage = computed(
  () => (FOLDER_STYLES.find((f) => f.value === props.state.folderIcon) || FOLDER_STYLES[0]).image,
)

const tabNames = computed(() => TAB_SLOTS.map((slot) => t(`themeCfg.tabSlot.${slot}`)))

function tabIcon(slot, selected) {
  const tab = props.state.tabIcon
  const custom = tab.enabled ? tab.icons[slot][selected ? 'selected' : 'normal'] : null
  return custom || defaultTabIconUrl(slot, selected)
}
</script>

<template>
  <div class="pv-phone" :class="{ 'is-dark': dark }">
    <div class="pv-screen" :style="screenStyle">
      <div v-if="usesImage" class="pv-photo" :style="photoStyle" />

      <div class="pv-content">
        <div class="pv-status" :style="{ color: primaryText }">
          <span>9:41</span>
          <span class="pv-status__dots">●●● ▮</span>
        </div>

        <h3 class="pv-title" :style="{ color: primaryText }">{{ tabNames[selectedTab] }}</h3>

        <div :class="cardClass">
          <div class="pv-row">
            <img :src="folderImage" class="pv-folder" alt="" />
            <div class="pv-row__text">
              <strong :style="{ color: primaryText }">Documents</strong>
              <small :style="{ color: secondaryText }">{{ t('themeCfg.preview.items', { n: 12 }) }}</small>
            </div>
          </div>
          <div class="pv-divider" />
          <div class="pv-row">
            <img :src="folderImage" class="pv-folder" alt="" />
            <div class="pv-row__text">
              <strong :style="{ color: primaryText }">Signed</strong>
              <small :style="{ color: secondaryText }">{{ t('themeCfg.preview.items', { n: 3 }) }}</small>
            </div>
          </div>
        </div>

        <div :class="cardClass">
          <div class="pv-row pv-row--app">
            <img :src="`/themes/${state.style}.png`" class="pv-app-icon" alt="" />
            <div class="pv-row__text">
              <strong :style="{ color: primaryText }">{{ t('themeCfg.preview.appName') }}</strong>
              <small :style="{ color: secondaryText }">{{ t('themeCfg.preview.appMeta') }}</small>
            </div>
            <span class="pv-btn" :style="{ background: accent }">{{ t('themeCfg.preview.sign') }}</span>
          </div>
          <div class="pv-divider" />
          <div class="pv-link" :style="{ color: accent }">{{ t('themeCfg.preview.link') }}</div>
        </div>
      </div>

      <nav class="pv-tabbar" :class="[state.glass ? 'pv-tabbar--glass' : 'pv-tabbar--plain', dark ? 'is-dark' : '']">
        <button
          v-for="(slot, i) in TAB_SLOTS"
          :key="slot"
          type="button"
          class="pv-tab"
          :aria-pressed="i === selectedTab"
          @click="selectedTab = i"
        >
          <img :src="tabIcon(slot, i === selectedTab)" alt="" />
          <span :style="{ color: i === selectedTab ? accent : secondaryText }">{{ tabNames[i] }}</span>
        </button>
      </nav>
    </div>
  </div>
</template>

<style scoped>
.pv-phone {
  width: 100%;
  max-width: 300px;
  margin: 0 auto;
  padding: 10px;
  border-radius: 46px;
  background: #1c1c1e;
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.22), inset 0 0 0 2px #3a3a3c;
}

.pv-screen {
  position: relative;
  aspect-ratio: 9 / 19.5;
  border-radius: 36px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.pv-photo {
  position: absolute;
  inset: -24px;
  background-size: cover;
  background-position: center;
}

.pv-content {
  position: relative;
  flex: 1;
  padding: 12px 14px 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 0;
}

.pv-status {
  display: flex;
  justify-content: space-between;
  padding: 2px 12px 0;
  font-size: 12px;
  font-weight: 600;
}

.pv-status__dots {
  font-size: 9px;
  letter-spacing: 1px;
}

.pv-title {
  margin: 8px 2px 2px;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.pv-card {
  border-radius: 16px;
  padding: 4px 12px;
}

.pv-card--glass {
  background: rgba(255, 255, 255, 0.42);
  border: 0.5px solid rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
}

.pv-card--glass.is-dark {
  background: rgba(40, 40, 44, 0.45);
  border-color: rgba(255, 255, 255, 0.12);
}

.pv-card--plain {
  background: #ffffff;
}

.pv-card--plain.is-dark {
  background: #1c1c1e;
}

.pv-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
}

.pv-row__text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.pv-row__text strong {
  font-size: 13px;
  font-weight: 600;
}

.pv-row__text small {
  font-size: 11px;
}

.pv-folder {
  width: 30px;
  height: 30px;
  object-fit: contain;
}

.pv-app-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
}

.pv-btn {
  padding: 4px 12px;
  border-radius: 999px;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
}

.pv-divider {
  height: 0.5px;
  margin-left: 40px;
  background: rgba(128, 128, 128, 0.3);
}

.pv-link {
  padding: 9px 0;
  font-size: 12px;
  font-weight: 600;
}

.pv-tabbar {
  position: relative;
  display: flex;
  justify-content: space-around;
  margin: 0 10px 14px;
  padding: 6px 4px;
  border-radius: 26px;
}

.pv-tabbar--glass {
  background: rgba(255, 255, 255, 0.5);
  border: 0.5px solid rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
}

.pv-tabbar--glass.is-dark {
  background: rgba(40, 40, 44, 0.5);
  border-color: rgba(255, 255, 255, 0.12);
}

.pv-tabbar--plain {
  margin: 0;
  border-radius: 0;
  padding-bottom: 14px;
  background: rgba(249, 249, 249, 0.96);
  border-top: 0.5px solid rgba(0, 0, 0, 0.12);
}

.pv-tabbar--plain.is-dark {
  background: rgba(22, 22, 24, 0.96);
  border-top-color: rgba(255, 255, 255, 0.12);
}

.pv-tab {
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-size: 9px;
  font-weight: 600;
}

.pv-tab img {
  width: 22px;
  height: 22px;
  object-fit: contain;
}
</style>
