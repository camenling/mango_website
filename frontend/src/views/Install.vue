<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  InboxOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
  CheckCircleOutlined,
  CopyOutlined,
  MobileOutlined,
  FileZipOutlined,
  BgColorsOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { RouterLink } from 'vue-router'
import { signAppWithBase64 } from '@/api/sign'
import { buildCertUploadPayload } from '@/utils/certPayload'
import AppGlass from '@/components/AppGlass.vue'
import ThemeStylePicker from '@/components/ThemeStylePicker.vue'
import { useTheme } from '@/composables/useTheme'
import { normalizeHex, styleAccent } from '@/config/hmTheme'

const { currentThemeId } = useTheme()

defineOptions({ name: 'InstallView' })

const { t } = useI18n()

const loading = ref(false)
const resultVisible = ref(false)

const uploadMode = ref('separate')

const form = reactive({
  password: '',
})

/** 签名时内置到 App 的主题：风格、强调色、玻璃渐变开关 */
const appTheme = reactive({
  style: currentThemeId.value,
  accent: styleAccent(currentThemeId.value),
  glass: true,
})

watch(() => appTheme.style, (style) => {
  appTheme.accent = styleAccent(style)
})

const p12File = ref(null)
const provisionFile = ref(null)
const zipFile = ref(null)

const signResult = ref(null)

const uploadModeOptions = computed(() => [
  { label: t('install.modeCert'), value: 'separate' },
  { label: t('install.modeZip'), value: 'zip' },
])

const noticeItems = computed(() => [
  t('install.notice1'),
  t('install.notice2'),
  t('install.notice3'),
])

function clearSeparateFiles() {
  p12File.value = null
  provisionFile.value = null
}

function clearZipFile() {
  zipFile.value = null
}

watch(uploadMode, (mode) => {
  if (mode === 'separate') {
    clearZipFile()
  } else {
    clearSeparateFiles()
  }
})

function beforeP12Upload(file) {
  const ext = file.name.toLowerCase()
  if (!ext.endsWith('.p12') && !ext.endsWith('.pfx')) {
    message.error(t('install.errP12Format'))
    return false
  }
  p12File.value = file
  return false
}

function beforeProvisionUpload(file) {
  const ext = file.name.toLowerCase()
  if (!ext.endsWith('.mobileprovision')) {
    message.error(t('install.errProvisionFormat'))
    return false
  }
  provisionFile.value = file
  return false
}

function beforeZipUpload(file) {
  const ext = file.name.toLowerCase()
  if (!ext.endsWith('.zip')) {
    message.error(t('install.errZipFormat'))
    return false
  }
  zipFile.value = file
  return false
}

async function resolveUploadPayload() {
  return buildCertUploadPayload({
    mode: uploadMode.value,
    p12File: p12File.value,
    provisionFile: provisionFile.value,
    zipFile: zipFile.value,
  })
}

function validateForm() {
  if (uploadMode.value === 'zip') {
    if (!zipFile.value) {
      message.warning(t('install.warnZipRequired'))
      return false
    }
  } else {
    if (!p12File.value) {
      message.warning(t('install.warnP12Required'))
      return false
    }
    if (!provisionFile.value) {
      message.warning(t('install.warnProvisionRequired'))
      return false
    }
  }
  if (!form.password) {
    message.warning(t('install.warnPasswordRequired'))
    return false
  }
  if (!/^[a-zA-Z0-9]+$/.test(form.password)) {
    message.warning(t('install.warnPasswordFormat'))
    return false
  }
  return true
}

async function handleSign() {
  if (!validateForm()) return
  loading.value = true
  try {
    const payload = await resolveUploadPayload()
    const password = form.password.trim()

    signResult.value = await signAppWithBase64({
      p12: payload.p12,
      mp: payload.mp,
      password,
      theme: appTheme.style,
      accent: normalizeHex(appTheme.accent) || styleAccent(appTheme.style),
      glass: appTheme.glass,
    })
    if (payload.p12Name) {
      signResult.value = {
        ...signResult.value,
        p12Name: signResult.value?.p12Name || payload.p12Name,
        provisionName: signResult.value?.provisionName || payload.provisionName,
      }
    }
    resultVisible.value = true
    message.success(t('install.signSuccess'))
  } catch (e) {
    message.error(e.message || t('install.signFailed'))
  } finally {
    loading.value = false
  }
}

function handleInstall() {
  if (signResult.value?.installUrl) {
    window.location.href = signResult.value.installUrl
  }
}

async function copyLink() {
  const url = signResult.value?.installUrl
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    message.success(t('install.linkCopied'))
  } catch {
    message.info(url)
  }
}
</script>

<template>
  <div class="install-page">
    <div class="site-width-shell install-shell">
      <AppGlass variant="panel" class="install-glass-block install-intro" padding="24px 22px" :track-mouse="false">
        <h1 class="install-title">{{ $t('install.title') }}</h1>
        <p class="install-desc">{{ $t('install.desc') }}</p>
      </AppGlass>

      <AppGlass variant="panel" class="install-glass-block install-form-card" padding="24px 22px" :track-mouse="false">
        <div class="mode-switch">
          <a-segmented
            v-model:value="uploadMode"
            :options="uploadModeOptions"
            block
          />
        </div>

        <div v-if="uploadMode === 'zip'" class="field-block field-block--upload">
          <label class="field-label">
            <FileZipOutlined /> {{ $t('install.zipLabel') }}
          </label>
          <p class="field-hint">{{ $t('install.zipHint') }}</p>
          <a-upload-dragger
            :show-upload-list="false"
            :before-upload="beforeZipUpload"
            accept=".zip"
          >
            <p class="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p class="ant-upload-text">{{ $t('install.zipDrag') }}</p>
            <p v-if="zipFile" class="file-name">{{ zipFile.name }}</p>
          </a-upload-dragger>
        </div>

        <div v-else class="upload-split">
          <div class="field-block field-block--upload">
            <label class="field-label">
              <SafetyCertificateOutlined /> {{ $t('install.p12Label') }}
            </label>
            <p class="field-hint">{{ $t('install.p12Hint') }}</p>
            <a-upload-dragger
              :show-upload-list="false"
              :before-upload="beforeP12Upload"
              accept=".p12,.pfx"
            >
              <p class="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p class="ant-upload-text">{{ $t('install.dragFile') }}</p>
              <p v-if="p12File" class="file-name">{{ p12File.name }}</p>
            </a-upload-dragger>
          </div>

          <div class="field-block field-block--upload">
            <label class="field-label">
              <SafetyCertificateOutlined /> {{ $t('install.provisionLabel') }}
            </label>
            <p class="field-hint">{{ $t('install.provisionHint') }}</p>
            <a-upload-dragger
              :show-upload-list="false"
              :before-upload="beforeProvisionUpload"
              accept=".mobileprovision"
            >
              <p class="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p class="ant-upload-text">{{ $t('install.dragFile') }}</p>
              <p v-if="provisionFile" class="file-name">{{ provisionFile.name }}</p>
            </a-upload-dragger>
          </div>
        </div>

        <div class="field-block field-block--password">
          <label class="field-label"><LockOutlined /> {{ $t('install.passwordLabel') }}</label>
          <p class="field-hint">
            {{ uploadMode === 'zip' ? $t('install.passwordHintZip') : $t('install.passwordHintCert') }}
          </p>
          <a-input-password
            v-model:value="form.password"
            :placeholder="$t('install.passwordPlaceholder')"
            size="large"
            :maxlength="128"
          />
        </div>

        <div class="field-block field-block--theme">
          <label class="field-label"><BgColorsOutlined /> {{ $t('install.themeLabel') }}</label>
          <p class="field-hint">{{ $t('install.themeHint') }}</p>
          <ThemeStylePicker v-model="appTheme.style" />
          <div class="theme-row">
            <span class="theme-row__label">{{ $t('install.accentLabel') }}</span>
            <div class="theme-row__control">
              <input v-model="appTheme.accent" type="color" class="color-swatch" />
              <a-input v-model:value="appTheme.accent" class="color-text" :maxlength="7" />
            </div>
          </div>
          <div class="theme-row">
            <span class="theme-row__label">
              {{ $t('install.glassLabel') }}
              <small>{{ $t('install.glassHint') }}</small>
            </span>
            <a-switch v-model:checked="appTheme.glass" />
          </div>
          <p class="field-hint field-hint--more">
            {{ $t('install.themeMore') }}
            <RouterLink to="/theme" class="theme-more-link">{{ $t('install.themeMoreLink') }}</RouterLink>
          </p>
        </div>

        <p class="security-note">{{ $t('install.securityNote') }}</p>

        <div class="form-actions">
          <a-button
            type="primary"
            size="large"
            block
            :loading="loading"
            @click="handleSign"
            >
              {{ $t('install.signBtn') }}
            </a-button>
          </div>
        </AppGlass>

      <section class="install-notice">
        <h4 class="install-notice-title">{{ $t('install.noticeTitle') }}</h4>
        <ul class="install-notice-list">
          <li v-for="(item, i) in noticeItems" :key="i">{{ item }}</li>
        </ul>
      </section>
    </div>

    <a-modal
      v-model:open="resultVisible"
      :title="$t('install.resultTitle')"
      :footer="null"
      centered
      width="420"
    >
      <div class="result-modal">
        <CheckCircleOutlined class="result-icon success" />
        <h3>{{ $t('install.resultHeading') }}</h3>
        <p>{{ signResult?.message }}</p>
        <div v-if="signResult?.p12Name || signResult?.provisionName" class="sign-cert-files">
          <div v-if="signResult.p12Name" class="zip-result-item">
            <span class="zip-label">P12</span>
            <span class="zip-value">{{ signResult.p12Name }}</span>
          </div>
          <div v-if="signResult.provisionName" class="zip-result-item">
            <span class="zip-label">{{ $t('install.provisionFile') }}</span>
            <span class="zip-value">{{ signResult.provisionName }}</span>
          </div>
        </div>
        <div class="result-actions">
          <a-button type="primary" size="large" block @click="handleInstall">
            <MobileOutlined /> {{ $t('install.installDirect') }}
          </a-button>
          <a-button size="large" block @click="copyLink">
            <CopyOutlined /> {{ $t('install.copyLink') }}
          </a-button>
        </div>
        <p class="result-tip">{{ $t('install.resultTip') }}</p>
      </div>
    </a-modal>
  </div>
</template>

<style scoped>
.field-block--theme {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.theme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.22);
}

.theme-row__label {
  display: flex;
  flex-direction: column;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.theme-row__label small {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.theme-row__control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-swatch {
  width: 34px;
  height: 34px;
  padding: 0;
  border: none;
  border-radius: 10px;
  background: none;
  cursor: pointer;
}

.color-text {
  width: 104px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.field-hint--more {
  margin: 0;
}

.theme-more-link {
  margin-left: 4px;
  color: var(--theme-accent, #0a84ff);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 3px;
}

.install-page {
  width: 100%;
  padding-bottom: 24px;
}

.install-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 40px;
  padding-bottom: 32px;
  box-sizing: border-box;
}

.install-intro {
  text-align: center;
}

.install-title {
  font-size: clamp(24px, 5vw, 30px);
  font-weight: 700;
  margin: 0 0 10px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.install-desc {
  font-size: 14px;
  line-height: 1.65;
  color: var(--text-secondary);
  margin: 0 auto;
  max-width: 36em;
}

.install-glass-block {
  width: 100%;
  box-sizing: border-box;
}

.install-glass-block :deep(.glass.active) {
  transform: none !important;
}

.mode-switch {
  margin-bottom: 20px;
}

.mode-switch :deep(.ant-segmented) {
  width: 100%;
}

.mode-switch :deep(.ant-segmented-group) {
  display: flex;
  width: 100%;
}

.mode-switch :deep(.ant-segmented-item) {
  flex: 1 1 50%;
  justify-content: center;
}

.mode-switch :deep(.ant-segmented-item-label) {
  width: 100%;
  text-align: center;
  font-size: 13px;
}

.upload-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
  align-items: start;
}

.upload-split .field-block--upload {
  min-width: 0;
}

.field-block {
  margin-bottom: 0;
  position: relative;
  z-index: 1;
}

.field-block--upload {
  margin-bottom: 4px;
}

.field-block--password {
  margin-top: 0;
  margin-bottom: 8px;
  padding-top: 20px;
  border-top: 0.5px solid rgba(255, 255, 255, 0.18);
  position: relative;
  z-index: 10;
  isolation: isolate;
}

.field-block--password :deep(.ant-input-affix-wrapper) {
  position: relative;
  z-index: 1;
}

.install-form-card :deep(.ant-upload-wrapper) {
  display: block;
}

.install-form-card :deep(.ant-upload-drag) {
  padding: 20px 16px !important;
  margin: 0 !important;
}

.install-form-card :deep(.ant-upload-drag .ant-upload-drag-icon) {
  margin-bottom: 6px !important;
}

.install-form-card :deep(.ant-upload-drag .anticon) {
  font-size: 36px !important;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
  color: var(--text-primary);
}

.field-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0 0 10px;
  line-height: 1.5;
}

.file-name {
  color: var(--ios-blue);
  font-size: 13px;
  margin-top: 8px;
}

.security-note {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0 0 16px;
  line-height: 1.5;
}

.form-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 温馨提示不用 liquid glass：Safari 对折页下 SVG filter + backdrop-filter 会延迟合成，先空白再突然出现 */
.install-notice {
  width: 100%;
  box-sizing: border-box;
  padding: 24px 22px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  transform: translateZ(0);
}

.install-notice-title {
  margin: 0 0 10px;
  color: #c93400;
  font-weight: 650;
  font-size: 14px;
}

.install-notice-list {
  margin: 0;
  padding-left: 18px;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.7;
}

.result-modal {
  text-align: center;
  padding: 8px 0;
}

.result-icon {
  font-size: 56px;
  margin-bottom: 16px;
}

.result-icon.success {
  color: #34c759;
}

.result-modal h3 {
  margin: 0 0 8px;
  color: var(--text-primary);
}

.result-modal p {
  color: var(--text-secondary);
  margin: 0 0 16px;
}

.sign-cert-files {
  margin: 0 0 20px;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  background: rgba(52, 199, 89, 0.08);
  border: 0.5px solid rgba(52, 199, 89, 0.2);
  text-align: left;
}

.sign-cert-files .zip-result-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  padding: 3px 0;
}

.sign-cert-files .zip-label {
  color: var(--text-muted);
}

.sign-cert-files .zip-value {
  color: var(--text-primary);
  word-break: break-all;
  text-align: right;
}

.result-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.result-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 16px !important;
}

@media (max-width: 720px) {
  .install-shell {
    gap: 16px;
    padding-top: 28px;
  }

  .install-intro {
    text-align: left;
  }

  .install-desc {
    max-width: none;
  }

  .upload-split {
    grid-template-columns: 1fr;
    gap: 20px;
  }
}
</style>
