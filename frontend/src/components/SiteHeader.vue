<script setup>
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  MenuOutlined,
  DownOutlined,
  ApiOutlined,
  CloudDownloadOutlined,
  WechatOutlined,
} from '@ant-design/icons-vue'
import AppGlass from '@/components/AppGlass.vue'
import LanguageSwitch from '@/components/LanguageSwitch.vue'
import ThemeSwitch from '@/components/ThemeSwitch.vue'
import { useTheme } from '@/composables/useTheme'

const { currentTheme } = useTheme()

const route = useRoute()
const { t } = useI18n()
const mobileOpen = ref(false)

/** 与地址栏同步，避免首帧 route 未对齐时误高亮「首页」 */
const navPath = computed(() => {
  const fromRouter = route.path
  if (typeof window !== 'undefined') {
    const fromLocation = window.location.pathname
    if (fromRouter === '/' && fromLocation !== '/' && fromLocation.startsWith('/install')) {
      return fromLocation
    }
  }
  return fromRouter
})

/** 顶层标签 */
const navItems = computed(() => [
  { label: t('nav.home'), to: '/' },
  { label: t('nav.install'), to: '/install' },
  { label: t('nav.theme'), to: '/theme' },
  { label: t('nav.lookup'), to: '/lookup' },
])

/** 「下载 & 对接」分组 */
const integrateItems = computed(() => [
  { label: t('nav.channelIpa'), desc: t('nav.channelIpaDesc'), to: '/channel-ipa', icon: CloudDownloadOutlined },
  { label: t('nav.docs'), desc: t('nav.docsDesc'), to: '/docs', icon: ApiOutlined },
  { label: t('nav.wxCert'), desc: t('nav.wxCertDesc'), to: '/wx-cert', icon: WechatOutlined },
])

function isNavActive(to) {
  const path = navPath.value
  if (to === '/') return path === '/'
  return path === to || path.startsWith(`${to}/`)
}

const integrateActive = computed(() => integrateItems.value.some((item) => isNavActive(item.to)))

/* ---------- 桌面下拉菜单 ---------- */
const menuOpen = ref(false)
const triggerRef = ref(null)
const menuStyle = ref({ top: '0px', left: '0px' })
const MENU_WIDTH = 272
let closeTimer = null

function updateMenuPosition() {
  const el = triggerRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const vw = window.innerWidth
  const left = Math.min(Math.max(12, rect.left + rect.width / 2 - MENU_WIDTH / 2), vw - MENU_WIDTH - 12)
  menuStyle.value = { top: `${rect.bottom + 10}px`, left: `${left}px`, width: `${MENU_WIDTH}px` }
}

function cancelClose() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

async function openMenu() {
  cancelClose()
  menuOpen.value = true
  await nextTick()
  updateMenuPosition()
}

function scheduleClose() {
  cancelClose()
  closeTimer = setTimeout(() => {
    menuOpen.value = false
  }, 140)
}

/** 鼠标悬停已先打开菜单，点击只负责保持打开（触屏设备上点击即打开） */
function toggleMenu(e) {
  e.stopPropagation()
  openMenu()
}

function closeMenu() {
  cancelClose()
  menuOpen.value = false
}

function onDocPointerDown(e) {
  if (!menuOpen.value) return
  const menu = document.getElementById('nav-integrate-menu')
  if (triggerRef.value?.contains(e.target) || menu?.contains(e.target)) return
  closeMenu()
}

function onKeydown(e) {
  if (e.key === 'Escape') closeMenu()
}

onMounted(() => {
  document.addEventListener('mousedown', onDocPointerDown)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', updateMenuPosition)
  window.addEventListener('scroll', updateMenuPosition, true)
})

onUnmounted(() => {
  cancelClose()
  document.removeEventListener('mousedown', onDocPointerDown)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
})

watch(
  () => route.path,
  () => {
    closeMenu()
    mobileOpen.value = false
  },
)
</script>

<template>
  <header class="site-header">
    <div class="site-width-shell site-header__column">
      <AppGlass variant="dock" class="nav-bar">
      <nav class="nav-inner">
        <RouterLink to="/" class="brand" @click="mobileOpen = false">
          <span class="brand-icon">
            <img
              :src="`/themes/${currentTheme.id}.png`"
              :alt="$t('brand.name')"
              class="brand-logo"
              width="40"
              height="40"
            />
          </span>
          <span class="brand-text">{{ $t('brand.name') }}</span>
        </RouterLink>

        <div class="nav-desktop">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="nav-link"
            active-class=""
            exact-active-class=""
            :class="{ active: isNavActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>

          <button
            ref="triggerRef"
            type="button"
            class="nav-link nav-link--dropdown"
            :class="{ active: integrateActive, open: menuOpen }"
            :aria-expanded="menuOpen"
            aria-haspopup="menu"
            @click="toggleMenu"
            @mouseenter="openMenu"
            @mouseleave="scheduleClose"
          >
            <span>{{ $t('nav.integrate') }}</span>
            <DownOutlined class="nav-link__caret" />
          </button>
        </div>

        <div class="nav-actions">
          <ThemeSwitch />
          <LanguageSwitch />
          <RouterLink to="/install" class="cta-btn">{{ $t('nav.installNow') }}</RouterLink>
        </div>

        <div class="nav-actions-mobile">
          <ThemeSwitch />
          <LanguageSwitch />
          <button class="menu-toggle" :aria-expanded="mobileOpen" @click="mobileOpen = !mobileOpen">
            <MenuOutlined />
          </button>
        </div>
      </nav>
    </AppGlass>

    <AppGlass v-if="mobileOpen" variant="panel" class="nav-mobile" padding="10px">
      <div class="nav-mobile-inner">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-mobile-link"
          :class="{ active: isNavActive(item.to) }"
          @click="mobileOpen = false"
        >
          {{ item.label }}
        </RouterLink>

        <div class="nav-mobile-group">
          <div class="nav-mobile-group__title">{{ $t('nav.integrate') }}</div>
          <RouterLink
            v-for="item in integrateItems"
            :key="item.to"
            :to="item.to"
            class="nav-mobile-link nav-mobile-link--sub"
            :class="{ active: isNavActive(item.to) }"
            @click="mobileOpen = false"
          >
            <component :is="item.icon" class="nav-mobile-link__icon" />
            <span class="nav-mobile-link__text">
              <span>{{ item.label }}</span>
              <small>{{ item.desc }}</small>
            </span>
          </RouterLink>
        </div>

        <RouterLink to="/install" class="cta-btn mobile-cta" @click="mobileOpen = false">
          {{ $t('nav.installNow') }}
        </RouterLink>
      </div>
    </AppGlass>
    </div>

    <Teleport to="body">
      <Transition name="nav-menu">
        <div
          v-if="menuOpen"
          id="nav-integrate-menu"
          class="nav-menu"
          role="menu"
          :style="menuStyle"
          @mousedown.stop
          @mouseenter="cancelClose"
          @mouseleave="scheduleClose"
        >
          <RouterLink
            v-for="item in integrateItems"
            :key="item.to"
            :to="item.to"
            class="nav-menu__item"
            :class="{ active: isNavActive(item.to) }"
            role="menuitem"
            @click="closeMenu"
          >
            <span class="nav-menu__icon"><component :is="item.icon" /></span>
            <span class="nav-menu__text">
              <span class="nav-menu__label">{{ item.label }}</span>
              <span class="nav-menu__desc">{{ item.desc }}</span>
            </span>
          </RouterLink>
        </div>
      </Transition>
    </Teleport>
  </header>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 14px;
  z-index: 100;
  padding: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.site-header__column {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-bar {
  width: 100%;
}

.nav-bar :deep(.glass) {
  width: 100% !important;
}

.nav-inner {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px 4px 10px;
  width: 100%;
  min-width: 0;
}

.nav-bar :deep(.app-glass--dock .glass) {
  overflow: visible !important;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 650;
  font-size: 15px;
  letter-spacing: -0.025em;
  flex-shrink: 0;
  transition: transform 0.28s var(--spring);
}

.brand:active {
  transform: scale(0.96);
}

.brand-icon {
  width: 44px;
  height: 44px;
  border-radius: 11px;
  padding: 0;
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.brand-logo {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  display: block;
}

.nav-desktop {
  display: flex;
  align-items: center;
  gap: 2px;
  flex: 1 1 auto;
  min-width: 0;
  justify-content: center;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.nav-actions-mobile {
  display: none;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  flex-shrink: 0;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
  padding: 7px 15px;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--text-secondary);
  text-decoration: none;
  font: inherit;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: -0.015em;
  line-height: 1.4;
  cursor: pointer;
  transition:
    color 0.28s var(--spring-smooth),
    background 0.28s var(--spring-smooth),
    transform 0.18s var(--spring-snappy);
}

.nav-link:hover,
.nav-link.open {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.15);
}

.nav-link:active {
  transform: scale(0.96);
  transition: transform 0.18s var(--spring-snappy);
}

.nav-link.active {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.28);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.75),
    0 1px 4px rgba(0, 0, 0, 0.04);
}

.nav-link--dropdown {
  padding-right: 12px;
}

.nav-link__caret {
  font-size: 9px;
  opacity: 0.7;
  transition: transform 0.28s var(--spring-smooth);
}

.nav-link.open .nav-link__caret {
  transform: rotate(180deg);
}

.cta-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 7px 18px;
  border-radius: 9999px !important;
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  color: #fff !important;
  text-decoration: none;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: -0.015em;
  line-height: 1.2;
  box-shadow:
    0 2px 12px color-mix(in srgb, var(--theme-accent, #007aff) 38%, transparent),
    inset 0 1px 0 rgba(255, 255, 255, 0.35);
  transition: all 0.28s var(--spring);
  flex-shrink: 0;
  white-space: nowrap;
  isolation: isolate;
  overflow: hidden;
  border: none;
}

.cta-btn:hover {
  transform: scale(1.04);
  box-shadow:
    0 4px 18px color-mix(in srgb, var(--theme-accent, #007aff) 45%, transparent),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.cta-btn:active {
  transform: scale(0.96);
  transition: transform 0.18s var(--spring-snappy);
}

.menu-toggle {
  border: none;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-radius: 50%;
  width: 36px;
  height: 36px;
  font-size: 16px;
  cursor: pointer;
  color: var(--text-primary);
  flex-shrink: 0;
  transition: transform 0.28s var(--spring);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.menu-toggle:active {
  transform: scale(0.92);
}

.nav-mobile {
  width: 100%;
}

.nav-mobile-inner {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 4px;
}

.nav-mobile-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  color: var(--text-primary);
  text-decoration: none;
  border-radius: var(--radius-md);
  font-size: 15px;
  font-weight: 500;
  transition: background 0.25s var(--spring-smooth), transform 0.25s var(--spring);
}

.nav-mobile-link.active {
  background: rgba(255, 255, 255, 0.22);
}

.nav-mobile-link:active {
  transform: scale(0.98);
  background: rgba(255, 255, 255, 0.15);
}

.nav-mobile-group {
  margin-top: 6px;
  padding-top: 8px;
  border-top: 0.5px solid rgba(255, 255, 255, 0.45);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-mobile-group__title {
  padding: 4px 16px 6px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--text-secondary);
}

.nav-mobile-link--sub {
  padding: 10px 16px;
}

.nav-mobile-link__icon {
  font-size: 18px;
  color: var(--theme-accent, var(--ios-blue));
  flex-shrink: 0;
}

.nav-mobile-link__text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.nav-mobile-link__text small {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.mobile-cta {
  margin-top: 6px;
  text-align: center;
  display: block;
}

/* 窄桌面：收紧标签间距并隐藏品牌文字，保证右侧按钮不出框 */
@media (min-width: 769px) and (max-width: 1040px) {
  .nav-link {
    padding: 7px 10px;
  }
  .nav-link--dropdown {
    padding-right: 8px;
  }
}

@media (min-width: 769px) and (max-width: 900px) {
  .brand-text {
    display: none;
  }
  .nav-link {
    padding: 7px 8px;
    font-size: 12.5px;
  }
  .cta-btn {
    padding: 7px 14px;
  }
}

@media (max-width: 768px) {
  .site-header {
    top: 8px;
  }
  .nav-desktop,
  .nav-actions {
    display: none;
  }
  .nav-actions-mobile {
    display: flex;
  }
}
</style>

<style>
/* Teleport 到 body，需全局样式 */
.nav-menu {
  position: fixed;
  padding: 6px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow:
    0 0 0 0.5px rgba(255, 255, 255, 0.6) inset,
    0 8px 28px rgba(0, 0, 0, 0.14);
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 2px;
  transform-origin: top center;
}

.nav-menu-enter-active,
.nav-menu-leave-active {
  transition: opacity 0.18s ease, transform 0.22s var(--spring-smooth);
}

.nav-menu-enter-from,
.nav-menu-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
}

.nav-menu__item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 14px;
  text-decoration: none;
  color: var(--text-primary, #1d1d1f);
  transition: background 0.2s ease, transform 0.18s var(--spring-snappy);
}

.nav-menu__item:hover {
  background: rgba(0, 0, 0, 0.05);
}

.nav-menu__item:active {
  transform: scale(0.98);
}

.nav-menu__item.active {
  background: color-mix(in srgb, var(--theme-accent, #007aff) 10%, transparent);
}

.nav-menu__icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
  color: var(--theme-accent, #007aff);
  background: color-mix(in srgb, var(--theme-accent, #007aff) 12%, #fff);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
  flex-shrink: 0;
}

.nav-menu__text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.nav-menu__label {
  font-size: 13.5px;
  font-weight: 600;
  letter-spacing: -0.015em;
  line-height: 1.3;
}

.nav-menu__desc {
  font-size: 12px;
  line-height: 1.35;
  color: var(--text-secondary, #6e6e73);
}
</style>
