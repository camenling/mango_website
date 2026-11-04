<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  DownloadOutlined,
  CopyOutlined,
  BgColorsOutlined,
  InboxOutlined,
  FileZipOutlined,
  CloseOutlined,
  DownOutlined,
  RightOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import ThemeStylePicker from '@/components/ThemeStylePicker.vue'
import ThemeEditor from '@/components/ThemeEditor.vue'
import StepLocator from '@/components/StepLocator.vue'
import { useTheme } from '@/composables/useTheme'
import { HMTHEME_FILE_NAME, HMTHEME_MAX_BYTES, buildHmtheme, defaultThemeState, normalizeHex, styleAccent } from '@/config/hmTheme'
import { downloadChannelIpa, fetchChannelIpaTracks } from '@/api/sign'
import { encryptMerchantEnvelope } from '@/utils/merchantEnvelope'
import {
  apiPath,
  checkPath,
  checkReplyLatestSample,
  checkReplyUpdateSample,
  checkSample,
  checkPythonSample,
  downloadPythonSample,
  envelopeSample,
  errorSample,
  payloadSample,
  themePackageSample,
  themePresetSample,
} from '@/content/channelIpaSample'
import {
  checkGoSample,
  checkJavaSample,
  checkPhpSample,
  downloadGoSample,
  downloadJavaSample,
  downloadPhpSample,
} from '@/content/channelIpaLangSamples'

defineOptions({ name: 'ChannelIpaView' })

const { t } = useI18n()
const { currentThemeId } = useTheme()
const loading = ref(false)
const tracksLoading = ref(true)
const tracksError = ref(false)
const tracks = ref({ release: null, beta: null })
const track = ref('release')
const channel = ref('')
const kid = ref('')
const pubKey = ref('')
const checkLang = ref('php')
const downloadLang = ref('php')
const activeStep = ref('cipa-check')

const langOptions = [
  { label: 'PHP', value: 'php' },
  { label: 'Go', value: 'go' },
  { label: 'Java', value: 'java' },
  { label: 'Python', value: 'python' },
]

const trackOptions = computed(() => [
  { label: t('channelIpa.trackRelease'), value: 'release' },
  { label: t('channelIpa.trackBeta'), value: 'beta' },
])

const channelOk = computed(() => /^[a-z][a-z0-9_-]{0,31}$/i.test(channel.value.trim()))

// ---------- 内置主题 ----------

const themeEnabled = ref(false)
/** preset：在这里直接配置主题（和主题配置页同一套编辑器）；upload：上传现成的 .hmtheme */
const themeMode = ref('preset')
/** 展开背景、字体颜色、文件夹、底栏图标等详细配置 */
const themeDetail = ref(false)
const themeState = reactive(defaultThemeState(currentThemeId.value))
const themeFile = ref(null)
const themeExporting = ref(false)

/** 快捷行里的强调色：空值表示沿用该风格默认色 */
const themeAccent = computed({
  get: () => themeState.textColor.accentLight || styleAccent(themeState.style),
  set: (value) => {
    themeState.textColor.accentLight = normalizeHex(value) || value
  },
})

watch(() => themeState.style, () => {
  themeState.textColor.accentLight = null
})

const themeModeOptions = computed(() => [
  { label: t('channelIpa.themeModePreset'), value: 'preset' },
  { label: t('channelIpa.themeModeUpload'), value: 'upload' },
])

function formatBytes(n) {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

function pickThemeFile(file) {
  const name = String(file?.name || '').toLowerCase()
  if (!name.endsWith('.hmtheme') && !name.endsWith('.zip')) {
    message.error(t('channelIpa.errThemeFile'))
    return false
  }
  if (file.size > HMTHEME_MAX_BYTES) {
    message.error(t('channelIpa.errThemeTooLarge'))
    return false
  }
  themeFile.value = file
  return false
}

function clearThemeFile() {
  themeFile.value = null
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = String(reader.result || '')
      resolve(result.slice(result.indexOf(',') + 1))
    }
    reader.onerror = () => reject(reader.error || new Error('read_failed'))
    reader.readAsDataURL(file)
  })
}

/** 把当前配置打包成 .hmtheme 下载备份，下次直接走「上传主题包」 */
async function exportTheme() {
  themeExporting.value = true
  try {
    const bytes = await buildHmtheme(themeState, t('themeCfg.defaultName'))
    if (bytes.byteLength > HMTHEME_MAX_BYTES) {
      message.error(t('channelIpa.errThemeTooLarge'))
      return
    }
    saveBlob(new Blob([bytes], { type: 'application/zip' }), HMTHEME_FILE_NAME)
    message.success(t('themeCfg.downloaded'))
  } catch (e) {
    message.error(e?.message || t('themeCfg.errExport'))
  } finally {
    themeExporting.value = false
  }
}

function bytesToBase64(bytes) {
  let bin = ''
  for (let i = 0; i < bytes.length; i += 0x8000) {
    bin += String.fromCharCode.apply(null, bytes.subarray(i, i + 0x8000))
  }
  return btoa(bin)
}

/** 组装请求体里的 theme 字段；关闭开关时返回 null，校验不过返回 undefined */
async function buildThemeField() {
  if (!themeEnabled.value) return null
  if (themeMode.value === 'upload') {
    if (!themeFile.value) {
      message.warning(t('channelIpa.warnThemeFile'))
      return undefined
    }
    return { package: await fileToBase64(themeFile.value) }
  }
  const bytes = await buildHmtheme(themeState, t('themeCfg.defaultName'))
  if (bytes.byteLength > HMTHEME_MAX_BYTES) {
    message.error(t('channelIpa.errThemeTooLarge'))
    return undefined
  }
  return { package: bytesToBase64(bytes) }
}

const checkSampleCode = computed(() => {
  if (checkLang.value === 'php') return checkPhpSample
  if (checkLang.value === 'go') return checkGoSample
  if (checkLang.value === 'java') return checkJavaSample
  return checkPythonSample
})

const checkSampleIntro = computed(() => {
  if (checkLang.value === 'php') return t('channelIpa.checkPhpIntro')
  if (checkLang.value === 'go') return t('channelIpa.checkGoIntro')
  if (checkLang.value === 'java') return t('channelIpa.checkJavaIntro')
  return t('channelIpa.checkPythonIntro')
})

const downloadSampleCode = computed(() => {
  if (downloadLang.value === 'php') return downloadPhpSample
  if (downloadLang.value === 'go') return downloadGoSample
  if (downloadLang.value === 'java') return downloadJavaSample
  return downloadPythonSample
})

const downloadSampleIntro = computed(() => {
  if (downloadLang.value === 'php') return t('channelIpa.downloadPhpIntro')
  if (downloadLang.value === 'go') return t('channelIpa.downloadGoIntro')
  if (downloadLang.value === 'java') return t('channelIpa.downloadJavaIntro')
  return t('channelIpa.downloadPythonIntro')
})

const pageLocator = computed(() => [
  { id: 'cipa-check', n: t('docs.sectionOne'), title: t('channelIpa.checkHowNav'), group: true },
  { id: 'cipa-check-1', n: 1, title: t('channelIpa.checkPlainTitle') },
  { id: 'cipa-check-2', n: 2, title: t('channelIpa.checkEncryptTitle') },
  { id: 'cipa-check-3', n: 3, title: t('channelIpa.checkReplyTitle') },
  { id: 'cipa-check-sample', n: 4, title: t('channelIpa.checkSampleTitle') },
  { id: 'cipa-dl', n: t('docs.sectionTwo'), title: t('channelIpa.downloadHowNav'), group: true },
  { id: 'cipa-dl-1', n: 1, title: t('channelIpa.downloadPlainTitle') },
  { id: 'cipa-dl-2', n: 2, title: t('channelIpa.downloadEncryptTitle') },
  { id: 'cipa-dl-3', n: 3, title: t('channelIpa.downloadSaveTitle') },
  { id: 'cipa-dl-sample', n: 4, title: t('channelIpa.downloadSampleTitle') },
])

function formatTrack(item) {
  const version = String(item?.version || '').trim()
  const date = String(item?.hm_build_date || '').trim()
  if (version && date) return `${version}（${date}）`
  return version || date
}

function latestBuildFor(tr) {
  return String(tracks.value[tr]?.hm_build_date || '').trim()
}

const selectedLatestText = computed(() => formatTrack(tracks.value[track.value]))

function errorMessage(err) {
  const code = err?.code || err?.message
  if (code === 'denied') return t('channelIpa.denied')
  if (code === 'decrypt_failed') return t('channelIpa.decryptFailed')
  if (code === 'not_found') return t('channelIpa.notFound')
  if (code === 'bad_request' || code === 'bad_payload') return t('channelIpa.badRequest')
  if (code === 'bad_theme') return t('channelIpa.badTheme')
  return t('channelIpa.failed')
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(t('docs.copied'))
  } catch {
    message.info(text)
  }
}

function saveBlob(blob, filename) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function formPlain() {
  const ch = channel.value.trim().toLowerCase()
  const k = kid.value.trim()
  const pem = pubKey.value.trim()
  const bd = latestBuildFor(track.value)
  if (tracksError.value || tracksLoading.value) {
    message.warning(t('channelIpa.latestFailed'))
    return null
  }
  if (!bd) {
    message.warning(t('channelIpa.warnNoLatest'))
    return null
  }
  if (!ch) {
    message.warning(t('channelIpa.warnChannel'))
    return null
  }
  if (!channelOk.value) {
    message.warning(t('channelIpa.badChannel'))
    return null
  }
  if (!k) {
    message.warning(t('channelIpa.warnKid'))
    return null
  }
  if (!pem) {
    message.warning(t('channelIpa.warnKey'))
    return null
  }
  return { track: track.value, channel: ch, build: bd, kid: k, pem }
}

async function handleDownload() {
  const form = formPlain()
  if (!form) return
  const theme = await buildThemeField()
  if (theme === undefined) return
  loading.value = true
  try {
    const envelope = await encryptMerchantEnvelope(form.kid, form.pem, {
      track: form.track,
      channel: form.channel,
      build: form.build,
    })
    if (theme) envelope.theme = theme
    const { blob, filename } = await downloadChannelIpa(envelope)
    saveBlob(blob, filename)
  } catch (err) {
    if (err?.message === 'bad_kid' || /public key|spki|import/i.test(String(err?.message || ''))) {
      message.error(t('channelIpa.badKey'))
      return
    }
    message.error(errorMessage(err))
  } finally {
    loading.value = false
  }
}

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
  activeStep.value = current
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

function scrollToStep(id) {
  const gen = ++locatorClickGen
  activeStep.value = id
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

function bindLocator() {
  unbindLocator()
  window.addEventListener('scroll', onLocatorScroll, { passive: true })
  window.addEventListener('resize', onLocatorScroll)
  nextTick(syncActiveStep)
}

onMounted(async () => {
  bindLocator()
  tracksLoading.value = true
  tracksError.value = false
  try {
    const data = await fetchChannelIpaTracks()
    tracks.value = {
      release: data?.release || null,
      beta: data?.beta || null,
    }
  } catch {
    tracksError.value = true
  } finally {
    tracksLoading.value = false
  }
})

onUnmounted(() => {
  unbindLocator()
})
</script>

<template>
  <div class="cipa-page">
    <div class="site-width-shell cipa-shell">
      <section class="cipa-intro">
        <h1 class="cipa-title">{{ t('channelIpa.title') }}</h1>
        <p class="cipa-desc">{{ t('channelIpa.desc') }}</p>
      </section>

      <section class="cipa-notice">
        <h2 class="cipa-h2">{{ t('channelIpa.noticeTitle') }}</h2>
        <i18n-t keypath="channelIpa.notice" tag="p" class="cipa-notice-text">
          <template #mark>
            <strong class="cipa-emph">{{ t('channelIpa.noticeMark') }}</strong>
          </template>
        </i18n-t>
      </section>

      <section class="cipa-card">
        <h2 class="cipa-h2">{{ t('channelIpa.latestTitle') }}</h2>
        <p v-if="tracksLoading" class="cipa-p">{{ t('channelIpa.latestLoading') }}</p>
        <p v-else-if="tracksError" class="cipa-p">{{ t('channelIpa.latestFailed') }}</p>
        <ul v-else class="cipa-latest">
          <li>
            <span class="cipa-latest-track">{{ t('channelIpa.trackRelease') }}</span>
            <span class="cipa-latest-ver">{{ formatTrack(tracks.release) || t('channelIpa.latestNone') }}</span>
          </li>
          <li>
            <span class="cipa-latest-track">{{ t('channelIpa.trackBeta') }}</span>
            <span class="cipa-latest-ver">{{ formatTrack(tracks.beta) || t('channelIpa.latestNone') }}</span>
          </li>
        </ul>
      </section>

      <section class="cipa-card">
        <label class="cipa-label">{{ t('channelIpa.trackLabel') }}</label>
        <a-radio-group v-model:value="track" :options="trackOptions" option-type="button" button-style="solid" />
        <p v-if="selectedLatestText" class="cipa-will">{{ t('channelIpa.willDownload', { version: selectedLatestText }) }}</p>
        <p v-else-if="!tracksLoading" class="cipa-will">{{ t('channelIpa.warnNoLatest') }}</p>

        <div class="cipa-label-row cipa-label--after">
          <label class="cipa-label" for="cipa-channel">{{ t('channelIpa.channelLabel') }}</label>
          <RouterLink to="/wx-cert" class="cipa-get-channel">
            {{ t('channelIpa.getChannel') }} <RightOutlined />
          </RouterLink>
        </div>
        <a-input
          id="cipa-channel"
          v-model:value="channel"
          size="large"
          :placeholder="t('channelIpa.channelPlaceholder')"
          :maxlength="32"
          allow-clear
        />

        <label class="cipa-label cipa-label--after" for="cipa-kid">{{ t('channelIpa.kidLabel') }}</label>
        <a-input
          id="cipa-kid"
          v-model:value="kid"
          size="large"
          :placeholder="t('channelIpa.kidPlaceholder')"
          :maxlength="64"
          allow-clear
        />

        <label class="cipa-label cipa-label--after" for="cipa-key">{{ t('channelIpa.keyLabel') }}</label>
        <a-textarea
          id="cipa-key"
          v-model:value="pubKey"
          :placeholder="t('channelIpa.keyPlaceholder')"
          :rows="8"
          class="cipa-key"
        />

        <div class="cipa-theme">
          <div class="cipa-theme-head">
            <span class="cipa-theme-head__label">
              <span class="cipa-theme-head__title"><BgColorsOutlined /> {{ t('channelIpa.themeLabel') }}</span>
              <small>{{ t('channelIpa.themeHint') }}</small>
            </span>
            <a-switch v-model:checked="themeEnabled" />
          </div>

          <div v-show="themeEnabled" class="cipa-theme-body">
            <a-segmented v-model:value="themeMode" :options="themeModeOptions" block />

            <div v-show="themeMode === 'preset'" class="cipa-theme-preset">
              <ThemeStylePicker v-model="themeState.style" />
              <div class="cipa-theme-row">
                <span class="cipa-theme-row__label">{{ t('channelIpa.accentLabel') }}</span>
                <div class="cipa-theme-row__control">
                  <input v-model="themeAccent" type="color" class="cipa-color-swatch" />
                  <a-input v-model:value="themeAccent" class="cipa-color-text" :maxlength="7" />
                </div>
              </div>
              <div class="cipa-theme-row">
                <span class="cipa-theme-row__label">
                  {{ t('channelIpa.glassLabel') }}
                  <small>{{ t('channelIpa.glassHint') }}</small>
                </span>
                <a-switch v-model:checked="themeState.glass" />
              </div>

              <button
                type="button"
                class="cipa-theme-expand"
                :class="{ open: themeDetail }"
                :aria-expanded="themeDetail"
                @click="themeDetail = !themeDetail"
              >
                <span>{{ themeDetail ? t('channelIpa.themeCollapse') : t('channelIpa.themeExpand') }}</span>
                <DownOutlined class="cipa-theme-expand__caret" />
              </button>
              <div v-show="themeDetail" class="cipa-theme-detail">
                <ThemeEditor :state="themeState" :hide="['name', 'style']" plain />
                <div class="cipa-theme-export">
                  <a-button size="large" block :loading="themeExporting" @click="exportTheme">
                    <DownloadOutlined /> {{ t('channelIpa.themeExport') }}
                  </a-button>
                  <p class="cipa-theme-export__hint">{{ t('channelIpa.themeExportHint') }}</p>
                </div>
              </div>
            </div>

            <div v-show="themeMode === 'upload'" class="cipa-theme-upload">
              <a-upload-dragger
                v-if="!themeFile"
                accept=".hmtheme,.zip"
                :max-count="1"
                :show-upload-list="false"
                :before-upload="pickThemeFile"
                class="cipa-upload"
              >
                <p class="cipa-upload__icon"><InboxOutlined /></p>
                <p class="cipa-upload__title">{{ t('channelIpa.uploadTitle') }}</p>
                <p class="cipa-upload__hint">{{ t('channelIpa.uploadHint') }}</p>
              </a-upload-dragger>
              <div v-else class="cipa-theme-file">
                <span class="cipa-theme-file__icon"><FileZipOutlined /></span>
                <span class="cipa-theme-file__text">
                  <span class="cipa-theme-file__name">{{ themeFile.name }}</span>
                  <small>{{ formatBytes(themeFile.size) }}</small>
                </span>
                <button type="button" class="cipa-theme-file__remove" :aria-label="t('channelIpa.uploadRemove')" @click="clearThemeFile">
                  <CloseOutlined />
                </button>
              </div>
            </div>
          </div>
        </div>

        <a-button
          type="primary"
          size="large"
          block
          :loading="loading"
          class="cipa-btn"
          @click="handleDownload"
        >
          <DownloadOutlined /> {{ loading ? t('channelIpa.downloading') : t('channelIpa.submit') }}
        </a-button>
      </section>

      <section class="cipa-card">
        <h2 class="cipa-h2">{{ t('channelIpa.howTitle') }}</h2>
        <p class="cipa-p">{{ t('channelIpa.howIntro') }}</p>
      </section>

      <div class="cipa-with-nav">
        <div class="cipa-docs">
          <section id="cipa-check" class="cipa-card cipa-step">
            <h2 class="cipa-h2">{{ t('channelIpa.checkHowTitle') }}</h2>
            <p class="cipa-p">{{ t('channelIpa.checkHowIntro') }}</p>
          </section>

          <section id="cipa-check-1" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.checkPlainTitle') }}</h2>
            </div>
            <ul class="cipa-list">
              <li>{{ t('channelIpa.fieldTrack') }}</li>
              <li>{{ t('channelIpa.fieldChannel') }}</li>
              <li>{{ t('channelIpa.fieldBuildCheck') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.checkPayloadTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(checkSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ checkSample }}</pre>
          </section>

          <section id="cipa-check-2" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.checkEncryptTitle') }}</h2>
            </div>
            <p class="cipa-p">{{ t('channelIpa.checkEncryptHint') }}</p>
            <h2 class="cipa-h2 cipa-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="cipa-p">{{ t('docs.apiIntro') }}</p>
            <div class="cipa-endpoint">
              <span class="cipa-method">POST</span>
              <code>{{ checkPath }}</code>
              <button type="button" class="cipa-copy-btn" @click="copyText(checkPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <h3 class="cipa-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="cipa-list cipa-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(envelopeSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ envelopeSample }}</pre>
          </section>

          <section id="cipa-check-3" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.checkReplyTitle') }}</h2>
            </div>
            <p class="cipa-note">{{ t('channelIpa.checkOk') }}</p>
            <ul class="cipa-list">
              <li>{{ t('channelIpa.checkOkLatest') }}</li>
              <li>{{ t('channelIpa.checkOkUpdate') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.checkReplyUpdateTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(checkReplyUpdateSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ checkReplyUpdateSample }}</pre>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.checkReplyLatestTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(checkReplyLatestSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ checkReplyLatestSample }}</pre>
            <h3 class="cipa-h3">{{ t('docs.blacklistFailTitle') }}</h3>
            <ul class="cipa-list">
              <li>{{ t('docs.blacklistErrBadRequest') }}</li>
              <li>{{ t('docs.blacklistErrDecrypt') }}</li>
              <li>{{ t('docs.blacklistErrPayload') }}</li>
              <li>{{ t('docs.blacklistErrDenied') }}</li>
              <li>{{ t('docs.blacklistErrUnavailable') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(errorSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ errorSample }}</pre>
          </section>

          <section id="cipa-check-sample" class="cipa-card cipa-step">
            <div class="cipa-code-head">
              <div class="cipa-step-head">
                <span class="cipa-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="cipa-h2">{{ t('channelIpa.checkSampleTitle') }}</h2>
              </div>
              <button type="button" class="cipa-copy-btn" @click="copyText(checkSampleCode)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="cipa-lang">
              <a-segmented v-model:value="checkLang" :options="langOptions" />
            </div>
            <p class="cipa-p">{{ checkSampleIntro }}</p>
            <pre class="cipa-pre cipa-pre--code">{{ checkSampleCode }}</pre>
          </section>

          <section id="cipa-dl" class="cipa-card cipa-step">
            <h2 class="cipa-h2">{{ t('channelIpa.downloadHowTitle') }}</h2>
            <p class="cipa-p">{{ t('channelIpa.downloadHowIntro') }}</p>
          </section>

          <section id="cipa-dl-1" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 1 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.downloadPlainTitle') }}</h2>
            </div>
            <ul class="cipa-list">
              <li>{{ t('channelIpa.fieldTrack') }}</li>
              <li>{{ t('channelIpa.fieldChannel') }}</li>
              <li>{{ t('channelIpa.fieldBuildDownload') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.downloadPayloadTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(payloadSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ payloadSample }}</pre>
          </section>

          <section id="cipa-dl-2" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 2 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.downloadEncryptTitle') }}</h2>
            </div>
            <p class="cipa-p">{{ t('channelIpa.downloadEncryptHint') }}</p>
            <h2 class="cipa-h2 cipa-h2--after">{{ t('docs.apiTitle') }}</h2>
            <p class="cipa-p">{{ t('docs.apiIntro') }}</p>
            <div class="cipa-endpoint">
              <span class="cipa-method">POST</span>
              <code>{{ apiPath }}</code>
              <button type="button" class="cipa-copy-btn" @click="copyText(apiPath)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <h3 class="cipa-h3">{{ t('docs.certLogicTitle') }}</h3>
            <ol class="cipa-list cipa-list--steps">
              <li>{{ t('docs.certLogic1') }}</li>
              <li>{{ t('docs.certLogic2') }}</li>
              <li>{{ t('docs.certLogic3') }}</li>
              <li>{{ t('docs.certLogic4') }}</li>
              <li>{{ t('docs.certLogic5') }}</li>
            </ol>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(envelopeSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ envelopeSample }}</pre>
            <h2 class="cipa-h2 cipa-h2--after">{{ t('channelIpa.themeDocTitle') }}</h2>
            <p class="cipa-p">{{ t('channelIpa.themeDocIntro') }}</p>
            <ul class="cipa-list">
              <li>{{ t('channelIpa.themeDocStyle') }}</li>
              <li>{{ t('channelIpa.themeDocAccent') }}</li>
              <li>{{ t('channelIpa.themeDocGlass') }}</li>
              <li>{{ t('channelIpa.themeDocPackage') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.themeDocPresetTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(themePresetSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ themePresetSample }}</pre>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('channelIpa.themeDocPackageTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(themePackageSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ themePackageSample }}</pre>
            <p class="cipa-note">{{ t('channelIpa.themeDocErr') }}</p>
          </section>

          <section id="cipa-dl-3" class="cipa-card cipa-step">
            <div class="cipa-step-head">
              <span class="cipa-step-no">{{ t('docs.stepNo', { n: 3 }) }}</span>
              <h2 class="cipa-h2">{{ t('channelIpa.downloadSaveTitle') }}</h2>
            </div>
            <p class="cipa-p">{{ t('channelIpa.downloadSaveHint') }}</p>
            <p class="cipa-note">{{ t('channelIpa.downloadOk') }}</p>
            <h3 class="cipa-h3">{{ t('docs.blacklistFailTitle') }}</h3>
            <ul class="cipa-list">
              <li>{{ t('docs.blacklistErrBadRequest') }}</li>
              <li>{{ t('docs.blacklistErrDecrypt') }}</li>
              <li>{{ t('docs.blacklistErrPayload') }}</li>
              <li>{{ t('docs.blacklistErrDenied') }}</li>
              <li>{{ t('channelIpa.downloadErrNotFound') }}</li>
              <li>{{ t('docs.blacklistErrUnavailable') }}</li>
            </ul>
            <div class="cipa-code-head">
              <h2 class="cipa-h2">{{ t('docs.envelopeTitle') }}</h2>
              <button type="button" class="cipa-copy-btn" @click="copyText(errorSample)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <pre class="cipa-pre">{{ errorSample }}</pre>
          </section>

          <section id="cipa-dl-sample" class="cipa-card cipa-step">
            <div class="cipa-code-head">
              <div class="cipa-step-head">
                <span class="cipa-step-no">{{ t('docs.stepNo', { n: 4 }) }}</span>
                <h2 class="cipa-h2">{{ t('channelIpa.downloadSampleTitle') }}</h2>
              </div>
              <button type="button" class="cipa-copy-btn" @click="copyText(downloadSampleCode)">
                <CopyOutlined /> {{ t('docs.copy') }}
              </button>
            </div>
            <div class="cipa-lang">
              <a-segmented v-model:value="downloadLang" :options="langOptions" />
            </div>
            <p class="cipa-p">{{ downloadSampleIntro }}</p>
            <pre class="cipa-pre cipa-pre--code">{{ downloadSampleCode }}</pre>
          </section>
        </div>

        <StepLocator
          :items="pageLocator"
          :active="activeStep"
          :title="t('channelIpa.locatorTitle')"
          @select="scrollToStep"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.cipa-page {
  width: 100%;
  padding-bottom: 24px;
}

.cipa-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 40px;
  padding-bottom: 32px;
  box-sizing: border-box;
}

.cipa-intro {
  text-align: center;
}

.cipa-title {
  font-size: clamp(24px, 5vw, 30px);
  font-weight: 700;
  margin: 0 0 10px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.cipa-desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.cipa-notice {
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: var(--radius-lg);
  background: rgba(255, 196, 90, 0.18);
  border: 0.5px solid rgba(180, 110, 0, 0.28);
}

.cipa-h2 {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 650;
  color: var(--text-primary);
}

.cipa-h2--after {
  margin-top: 22px;
}

.cipa-notice .cipa-h2 {
  color: #8a5a00;
}

.cipa-notice-text,
.cipa-p,
.cipa-note {
  margin: 0 0 12px;
  font-size: 13.5px;
  line-height: 1.75;
  color: var(--text-primary);
}

.cipa-emph {
  color: #e11d48;
  font-weight: 800;
  font-size: 1.45em;
  font-style: normal;
}

.cipa-p,
.cipa-note {
  color: var(--text-secondary);
}

.cipa-card {
  box-sizing: border-box;
  padding: 22px 20px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  transform: translateZ(0);
}

.cipa-latest {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cipa-latest li {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.cipa-latest-track {
  font-size: 14px;
  color: var(--text-secondary);
}

.cipa-latest-ver {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.cipa-label {
  display: block;
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 10px;
  color: var(--text-primary);
}

.cipa-label--after {
  margin-top: 16px;
}

.cipa-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.cipa-label-row .cipa-label {
  margin-bottom: 0;
}

.cipa-get-channel {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: var(--radius-pill);
  background: color-mix(in srgb, var(--theme-accent, #0a84ff) 12%, rgba(255, 255, 255, 0.5));
  color: var(--theme-accent, #0a84ff);
  font-size: 12.5px;
  font-weight: 600;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.2s ease, transform 0.18s var(--spring-snappy);
}

.cipa-get-channel :deep(svg) {
  font-size: 10px;
}

.cipa-get-channel:hover {
  background: color-mix(in srgb, var(--theme-accent, #0a84ff) 20%, rgba(255, 255, 255, 0.5));
}

.cipa-get-channel:active {
  transform: scale(0.96);
}

.cipa-will {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.cipa-key :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.cipa-theme {
  margin-top: 18px;
  padding: 14px 14px 12px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.22);
  border: 0.5px solid rgba(255, 255, 255, 0.45);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cipa-theme-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.cipa-theme-head__label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.cipa-theme-head__title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.cipa-theme-head__label small,
.cipa-theme-row__label small {
  font-size: 12px;
  font-weight: 400;
  line-height: 1.5;
  color: var(--text-secondary);
}

.cipa-theme-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 2px;
}

.cipa-theme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.22);
}

.cipa-theme-row__label {
  display: flex;
  flex-direction: column;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.cipa-theme-row__control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cipa-color-swatch {
  width: 34px;
  height: 34px;
  padding: 0;
  border: none;
  border-radius: 10px;
  background: none;
  cursor: pointer;
}

.cipa-color-text {
  width: 104px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.cipa-theme-preset,
.cipa-theme-upload {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cipa-theme-expand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 10px 14px;
  border: 1px dashed rgba(0, 0, 0, 0.16);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.16);
  color: var(--theme-accent, #0a84ff);
  font: inherit;
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.cipa-theme-expand:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: var(--theme-accent, #0a84ff);
}

.cipa-theme-expand__caret {
  font-size: 11px;
  transition: transform 0.28s var(--spring-smooth);
}

.cipa-theme-expand.open .cipa-theme-expand__caret {
  transform: rotate(180deg);
}

.cipa-theme-detail {
  padding-top: 4px;
}

.cipa-theme-export {
  margin-top: 14px;
}

.cipa-theme-export__hint {
  margin: 8px 0 0;
  font-size: 12.5px;
  line-height: 1.6;
  text-align: center;
  color: var(--text-secondary);
}

.cipa-upload :deep(.ant-upload-drag) {
  border-radius: var(--radius-sm);
  border-color: rgba(0, 0, 0, 0.14);
  background: rgba(255, 255, 255, 0.2);
}

.cipa-upload :deep(.ant-upload-drag:hover) {
  border-color: var(--theme-accent, #0a84ff);
}

.cipa-upload :deep(.ant-upload-btn) {
  padding: 18px 12px;
}

.cipa-upload__icon {
  margin: 0 0 6px;
  font-size: 32px;
  color: var(--theme-accent, #0a84ff);
}

.cipa-upload__title {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.cipa-upload__hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.cipa-theme-file {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.26);
}

.cipa-theme-file__icon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: var(--theme-accent, #0a84ff);
  background: color-mix(in srgb, var(--theme-accent, #0a84ff) 12%, #fff);
  flex-shrink: 0;
}

.cipa-theme-file__text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  flex: 1;
}

.cipa-theme-file__name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cipa-theme-file__text small {
  font-size: 12px;
  color: var(--text-secondary);
}

.cipa-theme-file__remove {
  border: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.cipa-theme-file__remove:hover {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-primary);
}

.cipa-btn {
  margin-top: 16px;
}

.cipa-with-nav {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 176px;
  gap: 16px;
  align-items: start;
}

.cipa-docs {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.cipa-step {
  scroll-margin-top: 104px;
}

.cipa-step-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.cipa-step-head .cipa-h2 {
  margin: 0;
}

.cipa-step-no {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #fff;
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  border-radius: 999px;
  padding: 3px 9px;
}

.cipa-list {
  margin: 0 0 14px;
  padding-left: 1.2em;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
}

.cipa-list--steps {
  list-style: decimal;
}

.cipa-list li + li {
  margin-top: 4px;
}

.cipa-endpoint {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 10px 12px;
  border-radius: 14px;
  background: rgba(0, 0, 0, 0.05);
  margin: 0 0 16px;
}

.cipa-method {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #fff;
  background: var(--theme-accent-soft, rgba(0, 122, 255, 0.88));
  border-radius: 999px;
  padding: 3px 8px;
}

.cipa-endpoint code {
  flex: 1 1 180px;
  min-width: 0;
  font-size: 12.5px;
  word-break: break-all;
  color: var(--text-primary);
}

.cipa-h3 {
  margin: 16px 0 8px;
  font-size: 14px;
  font-weight: 650;
  color: var(--text-primary);
}

.cipa-code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 22px 0 10px;
}

.cipa-code-head .cipa-h2,
.cipa-code-head .cipa-step-head {
  margin: 0;
}

.cipa-copy-btn {
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
}

.cipa-pre {
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

.cipa-pre--code {
  max-height: 520px;
  overflow: auto;
}

.cipa-lang {
  width: fit-content;
  max-width: 100%;
  margin: 0 0 12px;
}

.cipa-lang :deep(.ant-segmented) {
  width: auto;
}

.cipa-lang :deep(.ant-segmented-group) {
  display: flex;
}

.cipa-lang :deep(.ant-segmented-item) {
  flex: 1 0 0;
  min-width: 76px;
  justify-content: center;
}

.cipa-lang :deep(.ant-segmented-item-label) {
  width: 100%;
  padding-inline: 10px;
  text-align: center;
}

@media (max-width: 860px) {
  .cipa-with-nav {
    grid-template-columns: 1fr;
  }

  .cipa-with-nav > .step-locator {
    order: -1;
  }
}
</style>
