<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import StepLocator from '@/components/StepLocator.vue'
import { CopyOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import {
  apiPath,
  certFilePath,
  envelopeSample,
  payloadSample,
  replySample,
  pythonSample,
} from '@/content/encryptCertSample'
import {
  phpSample,
  goSample,
  javaSample,
} from '@/content/encryptCertLangSamples'
import {
  channelFilePath,
  channelPayloadSample,
  channelPythonSample,
  apiPath as blacklistApiPath,
  envelopeSample as blacklistEnvelopeSample,
  payloadSample as blacklistPayloadSample,
  replySample as blacklistReplySample,
  errorSample as blacklistErrorSample,
  pythonSample as blacklistPythonSample,
  unblacklistApiPath,
  unblacklistPayloadSample,
  unblacklistPythonSample,
} from '@/content/channelBlacklistSample'
import {
  channelPhpSample,
  channelGoSample,
  channelJavaSample,
  blacklistPhpSample,
  blacklistGoSample,
  blacklistJavaSample,
} from '@/content/channelBlacklistLangSamples'
import {
  unblacklistPhpSample,
  unblacklistGoSample,
  unblacklistJavaSample,
} from '@/content/channelUnblacklistLangSamples'
import {
  filePath as appSourceFilePath,
  singleSample,
  namedSample,
  multiSample,
} from '@/content/appSourceSample'
import {
  apiPath as lookupApiPath,
  payloadSample as lookupPayloadSample,
  noneSample as lookupNoneSample,
  channelSample as lookupChannelSample,
  globalSample as lookupGlobalSample,
  errorSample as lookupErrorSample,
  pythonSample as lookupPythonSample,
  phpSample as lookupPhpSample,
  goSample as lookupGoSample,
  javaSample as lookupJavaSample,
} from '@/content/lookupBlacklistSample'

defineOptions({ name: 'DocsView' })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const allowedTabs = ['cert', 'blacklist', 'lookup', 'source']

function tabFromRoute() {
  const raw = String(route.query.tab || '')
  return allowedTabs.includes(raw) ? raw : 'cert'
}

const tab = ref(tabFromRoute())
const certLang = ref('php')
const channelLang = ref('php')
const banLang = ref('php')
const unbanLang = ref('php')

const tabOptions = computed(() => [
  { label: t('docs.tabCert'), value: 'cert' },
  { label: t('docs.tabBlacklist'), value: 'blacklist' },
  { label: t('docs.tabLookup'), value: 'lookup' },
  { label: t('docs.tabSource'), value: 'source' },
])

const certLangOptions = [
  { label: 'PHP', value: 'php' },
  { label: 'Go', value: 'go' },
  { label: 'Java', value: 'java' },
  { label: 'Python', value: 'python' },
]

const certSample = computed(() => {
  if (certLang.value === 'php') return phpSample
  if (certLang.value === 'go') return goSample
  if (certLang.value === 'java') return javaSample
  return pythonSample
})

const certSampleIntro = computed(() => {
  if (certLang.value === 'php') return t('docs.certPhpIntro')
  if (certLang.value === 'go') return t('docs.certGoIntro')
  if (certLang.value === 'java') return t('docs.certJavaIntro')
  return t('docs.pythonIntro')
})

const certLocator = computed(() => [
  { id: 'cert-step-1', n: 1, title: t('docs.certFieldsTitle') },
  { id: 'cert-step-2', n: 2, title: t('docs.certEncryptTitle') },
  { id: 'cert-step-3', n: 3, title: t('docs.certWriteTitle') },
  { id: 'cert-step-sample', n: 4, title: t('docs.certSampleTitle') },
])

const blLocator = computed(() => [
  { id: 'bl-pre', n: t('docs.sectionOne'), title: t('docs.channelHowNav'), group: true },
  { id: 'bl-pre-1', n: 1, title: t('docs.channelPlainTitle') },
  { id: 'bl-pre-2', n: 2, title: t('docs.channelEncryptTitle') },
  { id: 'bl-pre-3', n: 3, title: t('docs.channelWriteTitle') },
  { id: 'bl-pre-4', n: 4, title: t('docs.channelSampleTitle') },
  { id: 'bl-ban', n: t('docs.sectionTwo'), title: t('docs.blacklistHowNav'), group: true },
  { id: 'bl-ban-1', n: 1, title: t('docs.blacklistPlainTitle') },
  { id: 'bl-ban-2', n: 2, title: t('docs.blacklistEncryptTitle') },
  { id: 'bl-ban-3', n: 3, title: t('docs.blacklistReplyTitle') },
  { id: 'bl-ban-4', n: 4, title: t('docs.blacklistSampleTitle') },
  { id: 'bl-unban', n: t('docs.sectionThree'), title: t('docs.unblacklistHowNav'), group: true },
  { id: 'bl-unban-1', n: 1, title: t('docs.unblacklistPlainTitle') },
  { id: 'bl-unban-2', n: 2, title: t('docs.unblacklistEncryptTitle') },
  { id: 'bl-unban-3', n: 3, title: t('docs.unblacklistReplyTitle') },
  { id: 'bl-unban-4', n: 4, title: t('docs.unblacklistSampleTitle') },
])

const lookupLang = ref('php')

const lookupLocator = computed(() => [
  { id: 'lookup-step-1', n: 1, title: t('docs.lookupStep1Title') },
  { id: 'lookup-step-2', n: 2, title: t('docs.lookupStep2Title') },
  { id: 'lookup-step-3', n: 3, title: t('docs.lookupStep3Title') },
  { id: 'lookup-step-sample', n: 4, title: t('docs.lookupSampleTitle') },
])

const pageLocator = computed(() => {
  if (tab.value === 'cert') return certLocator.value
  if (tab.value === 'blacklist') return blLocator.value
  if (tab.value === 'lookup') return lookupLocator.value
  return []
})

const channelSample = computed(() => {
  if (channelLang.value === 'php') return channelPhpSample
  if (channelLang.value === 'go') return channelGoSample
  if (channelLang.value === 'java') return channelJavaSample
  return channelPythonSample
})

const channelSampleIntro = computed(() => {
  if (channelLang.value === 'php') return t('docs.channelPhpIntro')
  if (channelLang.value === 'go') return t('docs.channelGoIntro')
  if (channelLang.value === 'java') return t('docs.channelJavaIntro')
  return t('docs.channelPythonIntro')
})

const banSample = computed(() => {
  if (banLang.value === 'php') return blacklistPhpSample
  if (banLang.value === 'go') return blacklistGoSample
  if (banLang.value === 'java') return blacklistJavaSample
  return blacklistPythonSample
})

const banSampleIntro = computed(() => {
  if (banLang.value === 'php') return t('docs.blacklistPhpIntro')
  if (banLang.value === 'go') return t('docs.blacklistGoIntro')
  if (banLang.value === 'java') return t('docs.blacklistJavaIntro')
  return t('docs.blacklistPythonIntro')
})

const unbanSample = computed(() => {
  if (unbanLang.value === 'php') return unblacklistPhpSample
  if (unbanLang.value === 'go') return unblacklistGoSample
  if (unbanLang.value === 'java') return unblacklistJavaSample
  return unblacklistPythonSample
})

const unbanSampleIntro = computed(() => {
  if (unbanLang.value === 'php') return t('docs.unblacklistPhpIntro')
  if (unbanLang.value === 'go') return t('docs.unblacklistGoIntro')
  if (unbanLang.value === 'java') return t('docs.unblacklistJavaIntro')
  return t('docs.unblacklistPythonIntro')
})

const lookupSample = computed(() => {
  if (lookupLang.value === 'php') return lookupPhpSample
  if (lookupLang.value === 'go') return lookupGoSample
  if (lookupLang.value === 'java') return lookupJavaSample
  return lookupPythonSample
})

const lookupSampleIntro = computed(() => {
  if (lookupLang.value === 'php') return t('docs.lookupPhpIntro')
  if (lookupLang.value === 'go') return t('docs.lookupGoIntro')
  if (lookupLang.value === 'java') return t('docs.lookupJavaIntro')
  return t('docs.lookupPythonIntro')
})

const activeCertStep = ref('cert-step-1')
let locatorLockUntil = 0
let locatorRaf = 0
let locatorClickGen = 0

/** 步骤标题要停在的视口位置：桌面端是顶栏下方；窄屏时定位栏贴在顶栏下，要再让出它的高度 */
function stepOffset() {
  if (!window.matchMedia('(max-width: 860px)').matches) return 104
  const bar = document.querySelector('.step-locator')
  const h = bar ? bar.getBoundingClientRect().height : 48
  return 78 + h + 12
}

function locatorLine() {
  return stepOffset() + 8
}

function syncActiveStep() {
  const items = pageLocator.value
  if (!items.length || performance.now() < locatorLockUntil) return
  const line = locatorLine()
  let current = items[0].id
  for (const item of items) {
    const el = document.getElementById(item.id)
    if (!el) continue
    if (el.getBoundingClientRect().top <= line) current = item.id
  }
  const root = document.documentElement
  if (window.scrollY + window.innerHeight >= root.scrollHeight - 32) {
    current = items[items.length - 1].id
  }
  activeCertStep.value = current
}

function onLocatorScroll() {
  if (locatorRaf) return
  locatorRaf = window.requestAnimationFrame(() => {
    locatorRaf = 0
    syncActiveStep()
  })
}

function unbindLocator() {
  window.removeEventListener('scroll', onLocatorScroll)
  window.removeEventListener('resize', onLocatorScroll)
  if (locatorRaf) {
    cancelAnimationFrame(locatorRaf)
    locatorRaf = 0
  }
}

function scrollToCertStep(id) {
  const gen = ++locatorClickGen
  activeCertStep.value = id
  locatorLockUntil = performance.now() + 1600
  const el = document.getElementById(id)
  if (!el) return
  window.scrollTo({ top: el.getBoundingClientRect().top + window.scrollY - stepOffset(), behavior: 'smooth' })
  const unlock = () => {
    if (gen !== locatorClickGen) return
    locatorLockUntil = 0
    window.removeEventListener('scrollend', unlock)
    syncActiveStep()
  }
  window.addEventListener('scrollend', unlock, { once: true })
  window.setTimeout(unlock, 1600)
}

function bindCertObserver() {
  unbindLocator()
  const items = pageLocator.value
  if (!items.length) return
  if (!items.some((item) => item.id === activeCertStep.value)) {
    activeCertStep.value = items[0].id
  }
  window.addEventListener('scroll', onLocatorScroll, { passive: true })
  window.addEventListener('resize', onLocatorScroll)
  nextTick(syncActiveStep)
}

watch(() => route.query.tab, () => {
  tab.value = tabFromRoute()
})

watch(tab, (value) => {
  const current = String(route.query.tab || '')
  if (current !== value) {
    router.replace({ query: { ...route.query, tab: value } })
  }
  nextTick(bindCertObserver)
})

onMounted(() => {
  nextTick(bindCertObserver)
})

onUnmounted(() => {
  unbindLocator()
})

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(t('docs.copied'))
  } catch {
    message.info(text)
  }
}
</script>

<template>
  <div class="docs-page">
    <div class="site-width-shell docs-shell">
      <section class="docs-card docs-disclaimer">
        <h2 class="docs-h2">{{ t('docs.disclaimerTitle') }}</h2>
        <p class="docs-disclaimer-text">{{ t('docs.disclaimer') }}</p>
      </section>

      <section class="docs-card docs-intro">
        <h1 class="docs-title">{{ t('docs.pageTitle') }}</h1>
        <div class="docs-tabs">
          <a-segmented v-model:value="tab" :options="tabOptions" block />
        </div>
      </section>

      <div v-if="tab === 'cert'" class="docs-with-nav">
        <div class="docs-main">
          <section id="cert-step-1" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.certFieldsTitle') }}</h2>
          </div>
          <ul class="docs-list docs-list--plain">
            <li>{{ t('docs.payloadPassword') }}</li>
            <li>{{ t('docs.payloadP12') }}</li>
            <li>{{ t('docs.payloadMp') }}</li>
          </ul>
          <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.certPayloadTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(payloadSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ payloadSample }}</pre>
          </section>

          <section id="cert-step-2" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.certEncryptTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.certEncryptHint') }}</p>
            <h3 class="docs-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="docs-list docs-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <h2 class="docs-h2 docs-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="docs-p">{{ t('docs.apiIntro') }}</p>
            <div class="docs-endpoint">
              <span class="docs-method">POST</span>
              <code>{{ apiPath }}</code>
              <button type="button" class="docs-copy" @click="copyText(apiPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
          <div class="docs-code-head docs-code-head--next">
            <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(envelopeSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ envelopeSample }}</pre>
          <div class="docs-code-head docs-code-head--next">
            <h2 class="docs-h2">{{ t('docs.replyTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(replySample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ replySample }}</pre>
            <p class="docs-note">{{ t('docs.certReplyHint') }}</p>
          </section>

          <section id="cert-step-3" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.certWriteTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.certWriteHint') }}</p>
            <div class="docs-endpoint">
              <code>{{ certFilePath }}</code>
              <button type="button" class="docs-copy" @click="copyText(certFilePath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
          </section>

          <section id="cert-step-sample" class="docs-card docs-step">
            <div class="docs-code-head">
              <div class="docs-step-head">
                <span class="docs-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="docs-h2">{{ t('docs.certSampleTitle') }}</h2>
              </div>
              <button type="button" class="docs-copy" @click="copyText(certSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="docs-tabs docs-tabs--lang">
              <a-segmented v-model:value="certLang" :options="certLangOptions" />
            </div>
            <p class="docs-p">{{ certSampleIntro }}</p>
            <pre class="docs-pre docs-pre--long">{{ certSample }}</pre>
          </section>
        </div>

        <StepLocator
          :items="certLocator"
          :active="activeCertStep"
          :title="t('docs.locatorTitle')"
          @select="scrollToCertStep"
        />
      </div>

      <div v-else-if="tab === 'blacklist'" class="docs-with-nav">
        <div class="docs-main">
          <section id="bl-pre" class="docs-card docs-step">
            <h2 class="docs-h2">{{ t('docs.channelHowTitle') }}</h2>
            <p class="docs-p">{{ t('docs.channelHowIntro') }}</p>
            <p class="docs-p">{{ t('docs.channelVersionHint') }}</p>
          </section>

          <section id="bl-pre-1" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.channelPlainTitle') }}</h2>
            </div>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.channelField') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.channelPayloadTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(channelPayloadSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ channelPayloadSample }}</pre>
          </section>

          <section id="bl-pre-2" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.channelEncryptTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.channelEncryptHint') }}</p>
            <h3 class="docs-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="docs-list docs-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistEnvelopeSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistEnvelopeSample }}</pre>
          </section>

          <section id="bl-pre-3" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.channelWriteTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.channelWriteHint') }}</p>
            <div class="docs-endpoint">
              <code>{{ channelFilePath }}</code>
              <button type="button" class="docs-copy" @click="copyText(channelFilePath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
          </section>

          <section id="bl-pre-4" class="docs-card docs-step">
            <div class="docs-code-head">
              <div class="docs-step-head">
                <span class="docs-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="docs-h2">{{ t('docs.channelSampleTitle') }}</h2>
              </div>
              <button type="button" class="docs-copy" @click="copyText(channelSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="docs-tabs docs-tabs--lang">
              <a-segmented v-model:value="channelLang" :options="certLangOptions" />
            </div>
            <p class="docs-p">{{ channelSampleIntro }}</p>
            <pre class="docs-pre docs-pre--long">{{ channelSample }}</pre>
          </section>

          <section id="bl-ban" class="docs-card docs-step">
            <h2 class="docs-h2">{{ t('docs.blacklistHowTitle') }}</h2>
            <p class="docs-p">{{ t('docs.blacklistApiIntro') }}</p>
            <h3 class="docs-h3">{{ t('docs.blacklistEffectTitle') }}</h3>
            <figure class="docs-figure">
              <img src="/docs-channel-blocked.png" :alt="t('docs.blacklistEffectTitle')" />
            </figure>
          </section>

          <section id="bl-ban-1" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.blacklistPlainTitle') }}</h2>
            </div>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.blacklistPayloadUdid') }}</li>
              <li>{{ t('docs.blacklistPayloadReason') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.payloadTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistPayloadSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistPayloadSample }}</pre>
          </section>

          <section id="bl-ban-2" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.blacklistEncryptTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.blacklistEncryptHint') }}</p>
            <h2 class="docs-h2 docs-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="docs-p">{{ t('docs.apiIntro') }}</p>
            <div class="docs-endpoint">
              <span class="docs-method">POST</span>
              <code>{{ blacklistApiPath }}</code>
              <button type="button" class="docs-copy" @click="copyText(blacklistApiPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <h3 class="docs-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="docs-list docs-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistEnvelopeSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistEnvelopeSample }}</pre>
          </section>

          <section id="bl-ban-3" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.blacklistReplyTitle') }}</h2>
            </div>
            <div class="docs-code-head">
              <h2 class="docs-h2">{{ t('docs.replyTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistReplySample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistReplySample }}</pre>
            <p class="docs-note">{{ t('docs.blacklistOk') }}</p>
            <h3 class="docs-h3">{{ t('docs.blacklistFailTitle') }}</h3>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.blacklistErrBadRequest') }}</li>
              <li>{{ t('docs.blacklistErrDecrypt') }}</li>
              <li>{{ t('docs.blacklistErrPayload') }}</li>
              <li>{{ t('docs.blacklistErrDenied') }}</li>
              <li>{{ t('docs.blacklistErrUnavailable') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistErrorSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistErrorSample }}</pre>
          </section>

          <section id="bl-ban-4" class="docs-card docs-step">
            <div class="docs-code-head">
              <div class="docs-step-head">
                <span class="docs-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="docs-h2">{{ t('docs.blacklistSampleTitle') }}</h2>
              </div>
              <button type="button" class="docs-copy" @click="copyText(banSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="docs-tabs docs-tabs--lang">
              <a-segmented v-model:value="banLang" :options="certLangOptions" />
            </div>
            <p class="docs-p">{{ banSampleIntro }}</p>
            <pre class="docs-pre docs-pre--long">{{ banSample }}</pre>
          </section>

          <section id="bl-unban" class="docs-card docs-step">
            <h2 class="docs-h2">{{ t('docs.unblacklistHowTitle') }}</h2>
            <p class="docs-p">{{ t('docs.unblacklistApiIntro') }}</p>
          </section>

          <section id="bl-unban-1" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.unblacklistPlainTitle') }}</h2>
            </div>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.unblacklistPayloadUdid') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.payloadTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(unblacklistPayloadSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ unblacklistPayloadSample }}</pre>
          </section>

          <section id="bl-unban-2" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.unblacklistEncryptTitle') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.unblacklistEncryptHint') }}</p>
            <h2 class="docs-h2 docs-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="docs-p">{{ t('docs.apiIntro') }}</p>
            <div class="docs-endpoint">
              <span class="docs-method">POST</span>
              <code>{{ unblacklistApiPath }}</code>
              <button type="button" class="docs-copy" @click="copyText(unblacklistApiPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <h3 class="docs-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="docs-list docs-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistEnvelopeSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistEnvelopeSample }}</pre>
          </section>

          <section id="bl-unban-3" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.unblacklistReplyTitle') }}</h2>
            </div>
            <div class="docs-code-head">
              <h2 class="docs-h2">{{ t('docs.replyTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistReplySample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistReplySample }}</pre>
            <p class="docs-note">{{ t('docs.unblacklistOk') }}</p>
            <h3 class="docs-h3">{{ t('docs.blacklistFailTitle') }}</h3>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.blacklistErrBadRequest') }}</li>
              <li>{{ t('docs.blacklistErrDecrypt') }}</li>
              <li>{{ t('docs.blacklistErrPayload') }}</li>
              <li>{{ t('docs.blacklistErrDenied') }}</li>
              <li>{{ t('docs.blacklistErrUnavailable') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(blacklistErrorSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ blacklistErrorSample }}</pre>
          </section>

          <section id="bl-unban-4" class="docs-card docs-step">
            <div class="docs-code-head">
              <div class="docs-step-head">
                <span class="docs-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="docs-h2">{{ t('docs.unblacklistSampleTitle') }}</h2>
              </div>
              <button type="button" class="docs-copy" @click="copyText(unbanSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="docs-tabs docs-tabs--lang">
              <a-segmented v-model:value="unbanLang" :options="certLangOptions" />
            </div>
            <p class="docs-p">{{ unbanSampleIntro }}</p>
            <pre class="docs-pre docs-pre--long">{{ unbanSample }}</pre>
          </section>
        </div>

        <StepLocator
          :items="blLocator"
          :active="activeCertStep"
          :title="t('docs.locatorTitle')"
          @select="scrollToCertStep"
        />
      </div>

      <div v-else-if="tab === 'lookup'" class="docs-with-nav">
        <div class="docs-main">
          <section class="docs-card docs-disclaimer">
            <h2 class="docs-h2">{{ t('lookup.disclaimerTitle') }}</h2>
            <p class="docs-disclaimer-text">{{ t('lookup.disclaimer') }}</p>
          </section>

          <section id="lookup-step-1" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.lookupStep1Title') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.lookupIntro') }}</p>
            <p class="docs-p">{{ t('docs.lookupStep1Hint') }}</p>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.lookupPayloadUdid') }}</li>
            </ul>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.payloadTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(lookupPayloadSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ lookupPayloadSample }}</pre>
          </section>

          <section id="lookup-step-2" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.lookupStep2Title') }}</h2>
            </div>
            <p class="docs-p">{{ t('docs.lookupStep2Hint') }}</p>
            <h2 class="docs-h2 docs-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="docs-p">{{ t('docs.apiIntro') }}</p>
            <div class="docs-endpoint">
              <span class="docs-method">POST</span>
              <code>{{ lookupApiPath }}</code>
              <button type="button" class="docs-copy" @click="copyText(lookupApiPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
          </section>

          <section id="lookup-step-3" class="docs-card docs-step">
            <div class="docs-step-head">
              <span class="docs-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="docs-h2">{{ t('docs.lookupStep3Title') }}</h2>
            </div>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.lookupReplyNone') }}</li>
              <li>{{ t('docs.lookupReplyChannel') }}</li>
              <li>{{ t('docs.lookupReplyGlobal') }}</li>
            </ul>
            <div class="docs-code-head">
              <h2 class="docs-h2">{{ t('docs.lookupReplyNoneTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(lookupNoneSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ lookupNoneSample }}</pre>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.lookupReplyChannelTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(lookupChannelSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ lookupChannelSample }}</pre>
            <div class="docs-code-head docs-code-head--next">
              <h2 class="docs-h2">{{ t('docs.lookupReplyGlobalTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(lookupGlobalSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ lookupGlobalSample }}</pre>
            <h3 class="docs-h3">{{ t('docs.lookupFailTitle') }}</h3>
            <ul class="docs-list docs-list--plain">
              <li>{{ t('docs.lookupErrBadRequest') }}</li>
              <li>{{ t('docs.lookupErrUnavailable') }}</li>
            </ul>
            <div class="docs-code-head">
              <h2 class="docs-h2">{{ t('docs.replyTitle') }}</h2>
              <button type="button" class="docs-copy" @click="copyText(lookupErrorSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="docs-pre">{{ lookupErrorSample }}</pre>
        </section>

          <section id="lookup-step-sample" class="docs-card docs-step">
          <div class="docs-code-head">
              <div class="docs-step-head">
                <span class="docs-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="docs-h2">{{ t('docs.lookupSampleTitle') }}</h2>
              </div>
              <button type="button" class="docs-copy" @click="copyText(lookupSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
            <div class="docs-tabs docs-tabs--lang">
              <a-segmented v-model:value="lookupLang" :options="certLangOptions" />
            </div>
            <p class="docs-p">{{ lookupSampleIntro }}</p>
            <pre class="docs-pre docs-pre--long">{{ lookupSample }}</pre>
        </section>
        </div>

        <StepLocator
          :items="lookupLocator"
          :active="activeCertStep"
          :title="t('docs.locatorTitle')"
          @select="scrollToCertStep"
        />
      </div>

      <template v-else>
        <section class="docs-card">
          <h2 class="docs-h2">{{ t('docs.sourceHowTitle') }}</h2>
          <p class="docs-p">{{ t('docs.sourceHowIntro') }}</p>
          <ol class="docs-list docs-list--steps">
            <li>{{ t('docs.sourceStep1') }}</li>
            <li>{{ t('docs.sourceStep2') }}</li>
            <li>{{ t('docs.sourceStep3') }}</li>
          </ol>
          <div class="docs-endpoint">
            <code>{{ appSourceFilePath }}</code>
            <button type="button" class="docs-copy" @click="copyText(appSourceFilePath)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <p class="docs-note">{{ t('docs.sourcePlainHint') }}</p>
        </section>

        <section class="docs-card">
          <h2 class="docs-h2">{{ t('docs.sourceFieldsTitle') }}</h2>
          <ul class="docs-list docs-list--plain">
            <li>{{ t('docs.sourceFieldUrl') }}</li>
            <li>{{ t('docs.sourceFieldSources') }}</li>
            <li>{{ t('docs.sourceFieldName') }}</li>
          </ul>
          <div class="docs-code-head docs-code-head--next">
            <h2 class="docs-h2">{{ t('docs.sourceSingleTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(singleSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ singleSample }}</pre>
          <div class="docs-code-head docs-code-head--next">
            <h2 class="docs-h2">{{ t('docs.sourceNamedTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(namedSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ namedSample }}</pre>
          <div class="docs-code-head docs-code-head--next">
            <h2 class="docs-h2">{{ t('docs.sourceMultiTitle') }}</h2>
            <button type="button" class="docs-copy" @click="copyText(multiSample)">
              <CopyOutlined /> {{ t('docs.copy') }}
            </button>
          </div>
          <pre class="docs-pre">{{ multiSample }}</pre>
          <p class="docs-note">{{ t('docs.sourceNote') }}</p>
        </section>
      </template>
    </div>
  </div>
</template>

<style scoped>
.docs-page {
  width: 100%;
  padding-bottom: 24px;
}

.docs-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 40px;
  padding-bottom: 32px;
  box-sizing: border-box;
}

.docs-with-nav {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 176px;
  gap: 16px;
  align-items: start;
}

.docs-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.docs-step {
  scroll-margin-top: 104px;
}

.docs-step-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.docs-step-head .docs-h2 {
  margin: 0;
}

.docs-step-no {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #fff;
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  border-radius: 999px;
  padding: 3px 9px;
}

@media (max-width: 860px) {
  .docs-with-nav {
    grid-template-columns: 1fr;
  }

  .docs-with-nav > .step-locator {
    order: -1;
  }
}

.docs-intro {
  text-align: center;
}

.docs-title {
  font-size: clamp(24px, 5vw, 30px);
  font-weight: 700;
  margin: 0 0 18px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.docs-tabs {
  max-width: 720px;
  margin: 0 auto;
}

.docs-tabs :deep(.ant-segmented) {
  width: 100%;
}

/* 标签完整显示，不截成省略号；窄屏缩小字号保持一行 */
.docs-tabs:not(.docs-tabs--lang) :deep(.ant-segmented-item-label) {
  padding: 0 6px;
  overflow: visible;
  text-overflow: clip;
}

@media (max-width: 480px) {
  .docs-tabs:not(.docs-tabs--lang) :deep(.ant-segmented-item-label) {
    padding: 0 2px;
    font-size: 13px;
    letter-spacing: -0.02em;
  }
}

.docs-card {
  width: 100%;
  box-sizing: border-box;
  padding: 24px 22px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
}

.docs-card.docs-disclaimer {
  background: rgba(255, 196, 90, 0.18);
  border-color: rgba(180, 110, 0, 0.28);
}

.docs-disclaimer .docs-h2 {
  color: #8a5a00;
}

.docs-disclaimer-text {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.75;
  color: var(--text-primary);
}

.docs-h2 {
  margin: 0 0 12px;
  font-size: 16px;
  font-weight: 650;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

.docs-h2--after {
  margin-top: 22px;
}

.docs-h3 {
  margin: 8px 0 10px;
  font-size: 14px;
  font-weight: 650;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

.docs-tabs--lang {
  width: fit-content;
  max-width: 100%;
  margin: 0 0 12px;
}

.docs-tabs--lang :deep(.ant-segmented) {
  width: auto;
}

.docs-tabs--lang :deep(.ant-segmented-group) {
  display: flex;
}

.docs-tabs--lang :deep(.ant-segmented-item) {
  flex: 1 0 0;
  min-width: 76px;
  justify-content: center;
}

.docs-tabs--lang :deep(.ant-segmented-item-label) {
  width: 100%;
  padding-inline: 10px;
  text-align: center;
}

.docs-p,
.docs-note {
  margin: 0 0 12px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.docs-note {
  margin: 12px 0 0;
}

.docs-hint {
  margin: 12px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-muted);
}

.docs-list {
  margin: 0 0 14px;
  padding-left: 1.2em;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
}

.docs-list--plain {
  list-style: disc;
}

.docs-list--steps {
  list-style: decimal;
}

.docs-list li + li {
  margin-top: 4px;
}

.docs-endpoint {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 10px 12px;
  border-radius: 14px;
  background: rgba(0, 0, 0, 0.05);
  margin: 0;
}

.docs-method {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #fff;
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  border-radius: 999px;
  padding: 3px 8px;
}

.docs-endpoint code {
  flex: 1 1 180px;
  min-width: 0;
  font-size: 12.5px;
  word-break: break-all;
  color: var(--text-primary);
}

.docs-copy {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.35);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
  transition: transform 0.18s var(--spring-snappy);
}

.docs-copy:active {
  transform: scale(0.96);
}

.docs-code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.docs-code-head .docs-h2 {
  margin: 0;
}

.docs-code-head .docs-step-head {
  margin-bottom: 0;
}

.docs-code-head--next {
  margin-top: 22px;
}

.docs-pre {
  margin: 0;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(0, 0, 0, 0.06);
  font-size: 12.5px;
  line-height: 1.55;
  overflow-x: auto;
  white-space: pre;
  color: var(--text-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.docs-pre--long {
  max-height: 520px;
  overflow: auto;
}

.docs-figure {
  margin: 12px 0 0;
}

.docs-figure img {
  display: block;
  width: min(100%, 360px);
  border-radius: 16px;
  background: #f7f7f8;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.06);
}
</style>
