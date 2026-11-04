import { createApp, nextTick } from 'vue'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'
import { i18n } from './i18n'
import { prefetchShaderMap } from '@wxperia/liquid-glass-vue'
import { installRouteWarmupTargets, primeGlassShadersFromDom } from '@/config/glassWarmup'
import { isSafariEngine, markSafariEngine } from '@/utils/safari'
import 'ant-design-vue/dist/reset.css'
import './styles/global.css'
import '@/composables/useTheme'

markSafariEngine()

const app = createApp(App)
app.use(router)
app.use(i18n)
app.use(Antd)

function revealApp() {
  document.documentElement.classList.add('app-revealed')
}

router.isReady().then(async () => {
  const safari = isSafariEngine()
  const path = router.currentRoute.value.path

  if (!safari && (path === '/install' || path.startsWith('/install/'))) {
    await Promise.all(
      installRouteWarmupTargets().map(({ width, height, effect }) =>
        prefetchShaderMap(width, height, effect),
      ),
    )
  }

  app.mount('#app')
  await nextTick()

  if (!safari) {
    try {
      await Promise.race([
        primeGlassShadersFromDom(),
        new Promise((resolve) => setTimeout(resolve, 1200)),
      ])
    } catch {
      /* 着色器预热失败仍展示页面 */
    }
    await nextTick()
  }

  revealApp()
}).catch((err) => {
  console.error(err)
  if (!document.getElementById('app')?.innerHTML) {
    app.mount('#app')
  }
  revealApp()
})
