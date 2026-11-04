<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale } from '@/i18n'

const { locale } = useI18n()

const nextLocale = computed(() => (locale.value === 'zh' ? 'en' : 'zh'))
const label = computed(() => (locale.value === 'zh' ? 'EN' : '中文'))

function toggle() {
  setLocale(nextLocale.value)
}
</script>

<template>
  <button
    type="button"
    class="lang-switch"
    :title="$t('lang.switch')"
    :aria-label="$t('lang.switch')"
    @click="toggle"
  >
    {{ label }}
  </button>
</template>

<style scoped>
.lang-switch {
  border: none;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(16px);
  border-radius: var(--radius-pill);
  padding: 7px 12px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--text-primary);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.28s var(--spring);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45);
  min-width: 44px;
}

.lang-switch:hover {
  background: rgba(255, 255, 255, 0.28);
  transform: scale(1.04);
}

.lang-switch:active {
  transform: scale(0.96);
  transition: transform 0.18s var(--spring-snappy);
}
</style>
