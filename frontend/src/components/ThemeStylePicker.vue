<script setup>
import { useI18n } from 'vue-i18n'
import { THEMES } from '@/config/themes'

defineOptions({ name: 'ThemeStylePicker' })

const model = defineModel({ type: Number, default: 1 })
const { t } = useI18n()
</script>

<template>
  <div class="style-picker" role="radiogroup">
    <button
      v-for="theme in THEMES"
      :key="theme.id"
      type="button"
      role="radio"
      class="style-picker__item"
      :class="{ 'is-active': model === theme.id }"
      :aria-checked="model === theme.id"
      :style="{ '--swatch-accent': theme.accent }"
      @click="model = theme.id"
    >
      <img :src="`/themes/${theme.id}.png`" :alt="t(theme.nameKey)" width="48" height="48" />
      <span>{{ t(theme.nameKey) }}</span>
    </button>
  </div>
</template>

<style scoped>
.style-picker {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.style-picker__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 4px 8px;
  border: 1.5px solid transparent;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.22);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: border-color 0.2s ease, transform 0.18s var(--spring-snappy);
}

.style-picker__item img {
  width: 44px;
  height: 44px;
  border-radius: 11px;
  object-fit: cover;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.style-picker__item:active {
  transform: scale(0.96);
}

.style-picker__item.is-active {
  border-color: var(--swatch-accent);
  color: var(--text-primary);
  background: color-mix(in srgb, var(--swatch-accent) 12%, rgba(255, 255, 255, 0.4));
}

@media (max-width: 420px) {
  .style-picker__item img {
    width: 36px;
    height: 36px;
    border-radius: 9px;
  }
}
</style>
