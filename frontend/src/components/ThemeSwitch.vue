<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { BgColorsOutlined } from '@ant-design/icons-vue'
import { useTheme } from '@/composables/useTheme'

const { t } = useI18n()
const { themes, currentThemeId, currentTheme, setTheme } = useTheme()
const open = ref(false)
const rootRef = ref(null)
const menuStyle = ref({ top: '0px', left: '0px' })

function updateMenuPosition() {
  const el = rootRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  menuStyle.value = {
    top: `${rect.bottom + 8}px`,
    left: `${Math.max(8, rect.right - 180)}px`,
  }
}

function pick(id) {
  setTheme(id)
  open.value = false
}

async function toggle(e) {
  e.stopPropagation()
  open.value = !open.value
  if (open.value) {
    await nextTick()
    updateMenuPosition()
  }
}

function onDocPointerDown(e) {
  if (!open.value) return
  const el = rootRef.value
  const menu = document.getElementById('theme-switch-menu')
  if (el?.contains(e.target) || menu?.contains(e.target)) return
  open.value = false
}

onMounted(() => {
  document.addEventListener('mousedown', onDocPointerDown)
  window.addEventListener('resize', updateMenuPosition)
  window.addEventListener('scroll', updateMenuPosition, true)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocPointerDown)
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
})

watch(open, (v) => {
  if (v) nextTick(updateMenuPosition)
})
</script>

<template>
  <div ref="rootRef" class="theme-switch">
    <button
      type="button"
      class="theme-trigger"
      :title="t('theme.switch')"
      :aria-label="t('theme.switch')"
      :aria-expanded="open"
      @click="toggle"
    >
      <img
        :src="`/themes/${currentTheme.id}.png`"
        :alt="t(currentTheme.nameKey)"
        class="theme-trigger-icon"
        width="28"
        height="28"
      />
      <BgColorsOutlined class="theme-trigger-badge" />
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        id="theme-switch-menu"
        class="theme-menu"
        :style="menuStyle"
        @mousedown.stop
      >
        <button
          v-for="item in themes"
          :key="item.id"
          type="button"
          class="theme-option"
          :class="{ active: item.id === currentThemeId }"
          @click="pick(item.id)"
        >
          <img
            :src="`/themes/${item.id}.png`"
            :alt="t(item.nameKey)"
            class="theme-option-icon"
            width="40"
            height="40"
          />
          <span class="theme-option-label">{{ t(item.nameKey) }}</span>
          <span class="theme-option-dot" :style="{ background: item.accent }" />
        </button>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.theme-switch {
  position: relative;
  flex-shrink: 0;
  z-index: 2;
}

.theme-trigger {
  position: relative;
  border: none;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(16px);
  border-radius: var(--radius-pill);
  width: 44px;
  height: 40px;
  padding: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.28s var(--spring);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45);
}

.theme-trigger:hover {
  background: rgba(255, 255, 255, 0.28);
  transform: scale(1.04);
}

.theme-trigger:active {
  transform: scale(0.96);
}

.theme-trigger-icon {
  border-radius: 9px;
  object-fit: cover;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}

.theme-trigger-badge {
  position: absolute;
  right: 2px;
  bottom: 2px;
  font-size: 10px;
  color: var(--theme-accent, var(--ios-blue));
  background: rgba(255, 255, 255, 0.85);
  border-radius: 50%;
  padding: 1px;
  pointer-events: none;
}
</style>

<style>
/* Teleport 到 body，需全局样式 */
.theme-menu {
  position: fixed;
  min-width: 180px;
  padding: 6px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(24px) saturate(180%);
  box-shadow:
    0 0 0 0.5px rgba(255, 255, 255, 0.6) inset,
    0 8px 28px rgba(0, 0, 0, 0.14);
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-option {
  border: none;
  background: transparent;
  border-radius: 12px;
  padding: 8px 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  text-align: left;
  transition: background 0.2s ease;
  width: 100%;
}

.theme-option:hover {
  background: rgba(255, 255, 255, 0.55);
}

.theme-option.active {
  background: rgba(255, 255, 255, 0.95);
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.06);
}

.theme-option-icon {
  border-radius: 10px;
  flex-shrink: 0;
  object-fit: cover;
}

.theme-option-label {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.86);
}

.theme-option-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08);
}
</style>
