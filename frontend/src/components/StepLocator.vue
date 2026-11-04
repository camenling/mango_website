<script setup>
import { computed, nextTick, watch } from 'vue'

/**
 * 文档页右侧的步骤定位栏。
 * items 为扁平列表，带 group: true 的条目是分组标题，其后的条目归到该组；没有分组时所有条目归一组。
 * 桌面端纵向列出；窄屏时第一行是分组标签（只有一组则不显示），第二行只放当前分组的步骤，横向可滚。
 */
defineOptions({ name: 'StepLocator' })

const props = defineProps({
  items: { type: Array, default: () => [] },
  active: { type: String, default: '' },
  title: { type: String, default: '' },
})

const emit = defineEmits(['select'])

const groups = computed(() => {
  const out = []
  for (const item of props.items) {
    if (item.group) {
      out.push({ ...item, items: [] })
    } else {
      if (!out.length) out.push({ id: '__default', n: '', title: '', items: [], virtual: true })
      out[out.length - 1].items.push(item)
    }
  }
  return out
})

const hasGroups = computed(() => groups.value.some((g) => !g.virtual))

const currentGroupId = computed(() => {
  const hit = groups.value.find(
    (g) => g.id === props.active || g.items.some((item) => item.id === props.active),
  )
  return hit?.id || groups.value[0]?.id
})

/** 窄屏下步骤行横向可滚，当前步骤变化时把它滚到中间 */
watch(
  () => props.active,
  (id) => {
    nextTick(() => {
      const center = (el, strip) => {
        if (!el || !strip || strip.scrollWidth <= strip.clientWidth) return
        const target = el.offsetLeft - (strip.clientWidth - el.offsetWidth) / 2
        strip.scrollTo({ left: Math.max(0, target), behavior: 'smooth' })
      }
      const step = document.querySelector(`.step-locator__item[data-step="${id}"]`)
      center(step, step?.closest('.step-locator__steps'))
      const tab = document.querySelector(`.step-locator__tab[data-tab="${currentGroupId.value}"]`)
      center(tab, tab?.closest('.step-locator__tabs'))
    })
  },
)
</script>

<template>
  <nav class="step-locator" :class="{ 'step-locator--grouped': hasGroups }" :aria-label="title">
    <p v-if="title" class="step-locator__title">{{ title }}</p>

    <!-- 窄屏：分组标签独占一行，可横向滚动 -->
    <div v-if="hasGroups" class="step-locator__tabs">
      <button
        v-for="group in groups"
        :key="`tab-${group.id}`"
        type="button"
        class="step-locator__tab"
        :class="{ current: currentGroupId === group.id }"
        :data-tab="group.id"
        @click="emit('select', group.id)"
      >
        {{ group.title }}
      </button>
    </div>

    <div
      v-for="group in groups"
      :key="group.id"
      class="step-locator__group"
      :class="{ 'is-current': currentGroupId === group.id }"
    >
      <button
        v-if="!group.virtual"
        type="button"
        class="step-locator__item step-locator__item--group"
        :class="{ active: active === group.id, current: currentGroupId === group.id }"
        :data-step="group.id"
        @click="emit('select', group.id)"
      >
        <span class="step-locator__dot">{{ group.n }}</span>
        <span class="step-locator__text">{{ group.title }}</span>
      </button>
      <div class="step-locator__steps">
        <button
          v-for="item in group.items"
          :key="item.id"
          type="button"
          class="step-locator__item"
          :class="{ active: active === item.id }"
          :data-step="item.id"
          @click="emit('select', item.id)"
        >
          <span class="step-locator__dot">{{ item.n }}</span>
          <span class="step-locator__text">{{ item.title }}</span>
        </button>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.step-locator {
  position: sticky;
  top: 96px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 12px 10px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
}

.step-locator__tabs {
  display: none;
}

.step-locator__title {
  margin: 0 6px 8px;
  font-size: 12px;
  font-weight: 650;
  color: var(--text-muted);
}

.step-locator__group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.step-locator__group + .step-locator__group {
  margin-top: 8px;
}

.step-locator__steps {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.step-locator__item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  padding: 7px 6px;
  border-radius: 10px;
  color: var(--text-secondary);
  font: inherit;
  transition: background 0.16s ease, color 0.16s ease;
}

.step-locator__item:hover,
.step-locator__item.active {
  background: rgba(255, 255, 255, 0.4);
  color: var(--text-primary);
}

.step-locator__dot {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  background: rgba(0, 0, 0, 0.08);
  color: var(--text-secondary);
}

.step-locator__item.active .step-locator__dot {
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  color: #fff;
}

.step-locator__item--group .step-locator__dot {
  width: auto;
  min-width: 20px;
  padding: 0 5px;
}

.step-locator__item--group .step-locator__text {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
}

.step-locator__item--group.current .step-locator__text {
  color: var(--text-primary);
}

.step-locator__text {
  font-size: 12.5px;
  line-height: 1.4;
  font-weight: 600;
}

/* 窄屏：第一行分组标签（可横向滚动），第二行只放当前分组的步骤 */
@media (max-width: 860px) {
  .step-locator {
    top: 78px;
    gap: 6px;
    padding: 8px;
    z-index: 8;
    /* 两行都是横向滚动条，不能让它们的内容宽度把页面撑开 */
    min-width: 0;
    max-width: 100%;
    contain: inline-size;
  }

  .step-locator__tabs,
  .step-locator__steps,
  .step-locator__group.is-current {
    min-width: 0;
    max-width: 100%;
  }

  .step-locator__title,
  .step-locator__item--group {
    display: none;
  }

  .step-locator__tabs,
  .step-locator__steps {
    display: flex;
    flex-direction: row;
    gap: 4px;
    overflow-x: auto;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    scroll-snap-type: x proximity;
  }

  .step-locator__tabs::-webkit-scrollbar,
  .step-locator__steps::-webkit-scrollbar {
    display: none;
  }

  .step-locator__tabs {
    gap: 6px;
  }

  .step-locator__tab {
    flex: 1 0 auto;
    border: none;
    padding: 8px 12px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.18);
    color: var(--text-secondary);
    font: inherit;
    font-size: 13px;
    font-weight: 650;
    white-space: nowrap;
    cursor: pointer;
    scroll-snap-align: center;
    transition: background 0.16s ease, color 0.16s ease;
  }

  .step-locator__tab.current {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.5);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.8),
      0 1px 4px rgba(0, 0, 0, 0.05);
  }

  .step-locator__group {
    display: none;
  }

  .step-locator__group.is-current {
    display: flex;
  }

  .step-locator__group + .step-locator__group {
    margin-top: 0;
  }

  .step-locator__steps .step-locator__item {
    flex: 0 0 auto;
    width: auto;
    align-items: center;
    padding: 7px 10px;
    white-space: nowrap;
    scroll-snap-align: center;
  }
}
</style>
