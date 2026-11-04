<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterView } from 'vue-router'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import SiteHeader from '@/components/SiteHeader.vue'
import SiteFooter from '@/components/SiteFooter.vue'
import { isSafariEngine } from '@/utils/safari'

const { locale } = useI18n()
const ambientMotion = ref(false)

const antdLocale = computed(() => (locale.value === 'zh' ? zhCN : enUS))

onMounted(() => {
  document.getElementById('app-boot-wallpaper')?.remove()
  if (isSafariEngine()) return
  requestAnimationFrame(() => {
    ambientMotion.value = true
  })
})
</script>

<template>
  <a-config-provider :locale="antdLocale">
    <div class="app-shell" :class="{ 'app-shell--ambient': ambientMotion }">
      <div class="ios-wallpaper" aria-hidden="true">
        <div class="wallpaper-photo" />
        <div class="mesh mesh-1" />
        <div class="mesh mesh-2" />
        <div class="mesh mesh-3" />
        <div class="mesh mesh-4" />
      </div>
      <SiteHeader />
      <main class="app-main">
        <RouterView v-slot="{ Component }">
          <KeepAlive include="HomeView,InstallView">
            <component :is="Component" />
          </KeepAlive>
        </RouterView>
      </main>
      <SiteFooter />
    </div>
  </a-config-provider>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow-x: clip;
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  width: 100%;
  position: relative;
  z-index: 1;
}

.ios-wallpaper {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
}

.wallpaper-photo {
  position: absolute;
  inset: 0;
  background: var(--theme-wallpaper);
  transition: background 0.45s var(--spring-smooth);
}

.mesh {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: var(--theme-mesh-opacity, 0.58);
  transition: opacity 0.45s var(--spring-smooth);
}

.mesh-1 {
  width: 50vw;
  height: 50vw;
  max-width: 640px;
  background: radial-gradient(circle, var(--theme-mesh-1) 0%, transparent 70%);
  top: -10%;
  left: -8%;
}

.mesh-2 {
  width: 45vw;
  height: 45vw;
  max-width: 580px;
  background: radial-gradient(circle, var(--theme-mesh-2) 0%, transparent 70%);
  top: 20%;
  right: -10%;
  animation-delay: -8s;
}

.mesh-3 {
  width: 40vw;
  height: 40vw;
  background: radial-gradient(circle, var(--theme-mesh-3) 0%, transparent 70%);
  bottom: -5%;
  left: 25%;
  animation-delay: -16s;
}

.mesh-4 {
  width: 35vw;
  height: 35vw;
  background: radial-gradient(circle, var(--theme-mesh-4) 0%, transparent 70%);
  bottom: 20%;
  right: 15%;
  animation-delay: -12s;
}

@media (prefers-reduced-motion: no-preference) {
  .app-shell--ambient .mesh {
    animation: ios-drift 24s ease-in-out infinite;
  }
}

@keyframes ios-drift {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(3vw, -2vh) scale(1.05); }
  66% { transform: translate(-2vw, 2vh) scale(0.96); }
}
</style>
