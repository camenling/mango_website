import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'
import Install from '@/views/Install.vue'
import Docs from '@/views/Docs.vue'
import Lookup from '@/views/Lookup.vue'
import ChannelIpa from '@/views/ChannelIpa.vue'
import ThemeConfig from '@/views/ThemeConfig.vue'
import WxCert from '@/views/WxCert.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: Home },
    { path: '/install', name: 'install', component: Install },
    { path: '/theme', name: 'theme', component: ThemeConfig },
    { path: '/docs', name: 'docs', component: Docs },
    { path: '/lookup', name: 'lookup', component: Lookup },
    { path: '/channel-ipa', name: 'channel-ipa', component: ChannelIpa },
    { path: '/wx-cert', name: 'wx-cert', component: WxCert },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

export default router
