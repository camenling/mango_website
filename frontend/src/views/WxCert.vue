<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Modal, message } from 'ant-design-vue'
import { wxCertCall, WX_CERT_LOGIN_URL } from '@/api/wxCert'

defineOptions({ name: 'WxCertView' })

const { t } = useI18n()
const route = useRoute()

const isWechat = /MicroMessenger/i.test(navigator.userAgent)
const LOGIN_TRY_KEY = 'wxCertLoginAt'

const loading = ref(true)
const needLogin = ref(false)
const errorKey = ref('')
const list = ref([])
const limit = ref(3)
const busy = ref(false)
const qrFailed = ref(false)
const QR_SRC = '/wx-mp-qrcode.png'
const channelInput = reactive({})
const ipInput = reactive({})

const createOpen = ref(false)
const createChannel = ref('')
const createIp = ref('')
const createIps = ref([])
const claimOpen = ref(false)
const claimKid = ref('')
const claimKey = ref('')

const full = computed(() => list.value.length >= limit.value)

function readTry() {
  try {
    return Number(sessionStorage.getItem(LOGIN_TRY_KEY) || 0)
  } catch {
    return 0
  }
}

function login() {
  try {
    sessionStorage.setItem(LOGIN_TRY_KEY, String(Date.now()))
  } catch {
    /* 隐私模式下忽略 */
  }
  window.location.href = WX_CERT_LOGIN_URL
}

// 微信里静默授权；30 秒内刚跳过一次就不再自动跳，避免 Cookie 写不进时来回循环
function onTicketExpired() {
  if (isWechat && Date.now() - readTry() > 30000) {
    login()
    return
  }
  needLogin.value = true
}

function apply(res, okText) {
  if (res.code === 0) {
    list.value = res.data?.list || []
    limit.value = res.data?.limit || 3
    needLogin.value = false
    if (okText) message.success(okText)
    return true
  }
  if (res.code === 401) {
    onTicketExpired()
    return false
  }
  message.error(res.message || t('wxCert.unavailable'))
  return false
}

async function run(action, body, okText) {
  busy.value = true
  try {
    return apply(await wxCertCall(action, body), okText)
  } finally {
    busy.value = false
  }
}

async function load() {
  loading.value = true
  try {
    apply(await wxCertCall('list'))
  } finally {
    loading.value = false
  }
}

// 只做格式粗查，最终以服务端校验为准
const IPV4 = /^(25[0-5]|2[0-4]\d|1?\d?\d)(\.(25[0-5]|2[0-4]\d|1?\d?\d)){3}$/
function validIp(v) {
  return IPV4.test(v) || (v.includes(':') && /^[0-9a-fA-F:.]+$/.test(v))
}

// 输入框里的 IP 变成一个标签；输入框为空时什么都不做
function addCreateIp() {
  const v = createIp.value.trim()
  if (!v) return true
  if (!validIp(v)) {
    message.warning(t('wxCert.badIp'))
    return false
  }
  if (!createIps.value.includes(v)) createIps.value.push(v)
  createIp.value = ''
  return true
}

function removeCreateIp(ip) {
  createIps.value = createIps.value.filter((v) => v !== ip)
}

async function submitCreate() {
  // 输入框里还没点「添加」的 IP 也算上
  if (!addCreateIp()) return
  const channel = createChannel.value.trim().toLowerCase()
  if (!channel || !createIps.value.length) {
    message.warning(t('wxCert.createRequired'))
    return
  }
  if (await run('create', { channel, ip: createIps.value.join(',') }, t('wxCert.createOk'))) {
    createOpen.value = false
    createChannel.value = ''
    createIp.value = ''
    createIps.value = []
  }
}

async function submitClaim() {
  const kid = claimKid.value.trim()
  const key = claimKey.value.trim()
  if (!kid || !key) {
    message.warning(t('wxCert.claimRequired'))
    return
  }
  if (await run('claim', { kid, public_key: key }, t('wxCert.claimOk'))) {
    claimOpen.value = false
    claimKid.value = ''
    claimKey.value = ''
  }
}

function bindChannel(row) {
  const channel = (channelInput[row.id] || '').trim().toLowerCase()
  if (!channel) return
  Modal.confirm({
    title: t('wxCert.bindChannelConfirm', { channel }),
    okText: t('wxCert.ok'),
    cancelText: t('wxCert.cancel'),
    onOk: async () => {
      if (await run('channel', { id: row.id, channel }, t('wxCert.channelOk'))) {
        channelInput[row.id] = ''
      }
    },
  })
}

async function addIp(row) {
  const ip = (ipInput[row.id] || '').trim()
  if (!ip) return
  if (!validIp(ip)) {
    message.warning(t('wxCert.badIp'))
    return
  }
  if (await run('ip-add', { id: row.id, ip }, t('wxCert.ipAdded'))) {
    ipInput[row.id] = ''
  }
}

function removeIp(row, ip) {
  const last = row.ips.length === 1
  Modal.confirm({
    title: t(last ? 'wxCert.removeLastIpConfirm' : 'wxCert.removeIpConfirm', { ip }),
    okText: t('wxCert.ok'),
    okType: 'danger',
    cancelText: t('wxCert.cancel'),
    onOk: () => run('ip-remove', { id: row.id, ip }, t('wxCert.ipRemoved')),
  })
}

function resetKey(row) {
  Modal.confirm({
    title: t('wxCert.resetTitle'),
    content: t('wxCert.resetBody'),
    okText: t('wxCert.reset'),
    okType: 'danger',
    cancelText: t('wxCert.cancel'),
    onOk: () => run('reset', { id: row.id, confirm: true }, t('wxCert.resetOk')),
  })
}

// 渠道和 IP 都齐了就不提示，卡片更紧凑
function nextHint(row) {
  if (!row.channel) return t('wxCert.nextChannel')
  if (!row.ips.length) return t('wxCert.nextIp')
  return ''
}

async function copy(text) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(t('wxCert.copied'))
    return
  } catch {
    /* 微信内置浏览器常不支持，走下面的兜底 */
  }
  const el = document.createElement('textarea')
  el.value = text
  el.setAttribute('readonly', '')
  el.style.position = 'fixed'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.select()
  let ok = false
  try {
    ok = document.execCommand('copy')
  } catch {
    ok = false
  }
  document.body.removeChild(el)
  if (ok) message.success(t('wxCert.copied'))
  else message.warning(t('wxCert.copyFailed'))
}

onMounted(() => {
  // 不在微信里拿不到授权，直接展示公众号二维码，不必请求接口
  if (!isWechat) {
    needLogin.value = true
    loading.value = false
    return
  }
  const err = String(route.query.error || '')
  if (err) {
    errorKey.value = err === 'auth' ? 'wxCert.authFailed' : 'wxCert.unavailable'
    needLogin.value = true
    loading.value = false
    return
  }
  load()
})
</script>

<template>
  <div class="wxc-page">
    <div class="site-width-shell wxc-shell">
      <section class="wxc-intro">
        <h1 class="wxc-title">{{ t('wxCert.title') }}</h1>
        <p class="wxc-desc">{{ t('wxCert.desc') }}</p>
      </section>

      <section v-if="isWechat" class="wxc-warn">{{ t('wxCert.secret') }}</section>

      <section v-if="loading" class="wxc-card wxc-center">{{ t('wxCert.loading') }}</section>

      <section v-else-if="needLogin" class="wxc-card wxc-center">
        <p v-if="errorKey" class="wxc-muted">{{ t(errorKey) }}</p>
        <template v-if="isWechat">
          <a-button type="primary" size="large" @click="login">{{ t('wxCert.relogin') }}</a-button>
        </template>
        <template v-else>
          <p class="wxc-follow">{{ t('wxCert.notWechat') }}</p>
          <template v-if="!qrFailed">
            <!-- 用普通 img，手机上才能长按保存到相册 -->
            <img class="wxc-qr" :src="QR_SRC" :alt="t('wxCert.qrAlt')" @error="qrFailed = true" />
            <p class="wxc-muted">{{ t('wxCert.qrSave') }}</p>
            <a class="wxc-qr-dl" :href="QR_SRC" :download="t('wxCert.qrAlt') + '.png'">{{ t('wxCert.qrDownload') }}</a>
          </template>
        </template>
      </section>

      <template v-else>
        <div class="wxc-toolbar">
          <span class="wxc-muted">{{ t('wxCert.quota', { n: list.length, max: limit }) }}</span>
          <div class="wxc-toolbar-btns">
            <a-button :disabled="full || busy" @click="claimOpen = true">{{ t('wxCert.claim') }}</a-button>
            <a-button type="primary" :disabled="full || busy" @click="createOpen = true">
              {{ t('wxCert.create') }}
            </a-button>
          </div>
        </div>

        <section v-if="!list.length" class="wxc-card wxc-center wxc-muted">{{ t('wxCert.empty') }}</section>

        <section v-for="row in list" :key="row.id" class="wxc-card wxc-item">
          <div class="wxc-head">
            <strong class="wxc-channel">{{ row.channel || t('wxCert.noChannel') }}</strong>
            <a-tag v-if="!row.enabled" color="red" class="wxc-tag">{{ t('wxCert.disabled') }}</a-tag>
            <a class="wxc-reset" @click="!busy && resetKey(row)">{{ t('wxCert.reset') }}</a>
          </div>

          <div class="wxc-row">
            <span class="wxc-key">kid</span>
            <code class="wxc-kid">{{ row.kid }}</code>
          </div>

          <p v-if="nextHint(row)" class="wxc-next">{{ nextHint(row) }}</p>

          <div v-if="!row.channel" class="wxc-inline">
            <a-input
              v-model:value="channelInput[row.id]"
              :placeholder="t('wxCert.channelPlaceholder')"
              :maxlength="32"
              @pressEnter="bindChannel(row)"
            />
            <a-button :disabled="busy || !channelInput[row.id]" @click="bindChannel(row)">
              {{ t('wxCert.bindChannel') }}
            </a-button>
          </div>

          <div class="wxc-row wxc-row-top">
            <span class="wxc-key">IP</span>
            <div class="wxc-tags">
              <a-tag v-for="ip in row.ips" :key="ip" closable class="wxc-tag" @close.prevent="removeIp(row, ip)">
                {{ ip }}
              </a-tag>
              <span v-if="!row.ips.length" class="wxc-danger">{{ t('wxCert.noIp') }}</span>
            </div>
          </div>
          <div class="wxc-inline">
            <a-input
              v-model:value="ipInput[row.id]"
              :placeholder="t('wxCert.ipPlaceholder')"
              @pressEnter="addIp(row)"
            />
            <a-button :disabled="busy || !ipInput[row.id]" @click="addIp(row)">{{ t('wxCert.addIp') }}</a-button>
          </div>

          <div class="wxc-copy-btns">
            <a-button size="large" @click="copy(row.kid)">{{ t('wxCert.copyKid') }}</a-button>
            <a-button size="large" type="primary" @click="copy(row.public_key)">{{ t('wxCert.copyKey') }}</a-button>
          </div>
        </section>

        <p class="wxc-muted wxc-center">{{ t('wxCert.chatTip') }}</p>
      </template>
    </div>

    <a-modal
      v-model:open="createOpen"
      wrap-class-name="wxc-modal"
      :title="t('wxCert.createTitle')"
      :ok-text="t('wxCert.ok')"
      :cancel-text="t('wxCert.cancel')"
      :confirm-loading="busy"
      @ok="submitCreate"
    >
      <label class="wxc-label wxc-required">{{ t('wxCert.createChannelLabel') }}</label>
      <a-input v-model:value="createChannel" :maxlength="32" :placeholder="t('wxCert.channelPlaceholder')" />
      <p class="wxc-muted wxc-hint">{{ t('wxCert.createChannelHint') }}</p>
      <label class="wxc-label wxc-required wxc-gap">{{ t('wxCert.createIpLabel') }}</label>
      <div v-if="createIps.length" class="wxc-tags wxc-modal-tags">
        <a-tag v-for="ip in createIps" :key="ip" closable class="wxc-tag" @close.prevent="removeCreateIp(ip)">
          {{ ip }}
        </a-tag>
      </div>
      <div class="wxc-inline">
        <a-input v-model:value="createIp" :placeholder="t('wxCert.ipPlaceholder')" @pressEnter="addCreateIp" />
        <a-button :disabled="!createIp.trim()" @click="addCreateIp">{{ t('wxCert.addIp') }}</a-button>
      </div>
      <p class="wxc-muted wxc-hint">{{ t('wxCert.createIpHint') }}</p>
    </a-modal>

    <a-modal
      v-model:open="claimOpen"
      wrap-class-name="wxc-modal"
      :title="t('wxCert.claimTitle')"
      :ok-text="t('wxCert.ok')"
      :cancel-text="t('wxCert.cancel')"
      :confirm-loading="busy"
      @ok="submitClaim"
    >
      <p class="wxc-muted">{{ t('wxCert.claimHint') }}</p>
      <label class="wxc-label wxc-gap">{{ t('wxCert.claimKid') }}</label>
      <a-input v-model:value="claimKid" :maxlength="32" />
      <label class="wxc-label wxc-gap">{{ t('wxCert.claimKey') }}</label>
      <a-textarea v-model:value="claimKey" :rows="7" class="wxc-mono" />
    </a-modal>
  </div>
</template>

<style>
/* iOS（含微信）里输入框字号小于 16px 时，聚焦会自动放大页面。
   弹窗渲染在 body 下，scoped 样式覆盖不到，所以用全局样式并限定在本页和本页弹窗内。 */
.wxc-page input,
.wxc-page textarea,
.wxc-modal input,
.wxc-modal textarea {
  font-size: 16px !important;
}

/* 弹窗沿用全站玻璃样式；只有不支持 backdrop-filter 的 WebView 才加实底，避免表单看不清 */
@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
  .wxc-modal .ant-modal-content {
    background: rgba(255, 255, 255, 0.92) !important;
  }
}

.wxc-modal .wxc-label {
  display: block;
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 6px;
}

.wxc-modal .wxc-modal-tags {
  margin-bottom: 8px;
}

.wxc-modal .wxc-gap {
  margin-top: 14px;
}

.wxc-modal .wxc-hint {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.6;
  opacity: 0.7;
}

.wxc-modal .wxc-mono textarea {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.wxc-required::before {
  content: '*';
  margin-right: 4px;
  color: #e11d48;
}
</style>

<style scoped>
.wxc-page {
  width: 100%;
  padding-bottom: 24px;
}

.wxc-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 32px;
  padding-bottom: 32px;
  box-sizing: border-box;
  max-width: 640px;
}

.wxc-intro {
  text-align: center;
}

.wxc-title {
  font-size: clamp(22px, 5vw, 28px);
  font-weight: 700;
  margin: 0 0 8px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.wxc-desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.wxc-warn {
  box-sizing: border-box;
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  background: rgba(255, 196, 90, 0.18);
  border: 0.5px solid rgba(180, 110, 0, 0.28);
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
}

.wxc-card {
  box-sizing: border-box;
  padding: 18px 16px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  transform: translateZ(0);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.wxc-center {
  text-align: center;
  align-items: center;
}

.wxc-muted {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.wxc-follow {
  margin: 0;
  font-size: 15px;
  line-height: 1.75;
  color: var(--text-primary);
}

/* 横版「扫码 + 搜一搜」物料，自带圆角绿底，不再加白框 */
.wxc-qr {
  display: block;
  width: 100%;
  max-width: 420px;
  height: auto;
  -webkit-touch-callout: default;
  -webkit-user-select: auto;
  user-select: auto;
}

.wxc-qr-dl {
  font-size: 14px;
  font-weight: 600;
  color: var(--theme-accent, #0a84ff);
}

.wxc-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.wxc-toolbar-btns {
  display: flex;
  gap: 8px;
}

.wxc-item {
  gap: 10px;
  padding: 14px;
}

.wxc-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wxc-channel {
  font-size: 17px;
  color: var(--text-primary);
  word-break: break-all;
}

.wxc-tag {
  margin: 0;
}

.wxc-reset {
  margin-left: auto;
  flex-shrink: 0;
  font-size: 13px;
  color: #e11d48;
  cursor: pointer;
}

.wxc-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.wxc-row-top {
  align-items: flex-start;
}

.wxc-key {
  flex-shrink: 0;
  width: 28px;
  font-size: 12px;
  font-weight: 600;
  line-height: 22px;
  color: var(--text-secondary);
}

.wxc-kid {
  font: 13px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace;
  color: var(--text-primary);
  word-break: break-all;
  user-select: all;
  -webkit-user-select: all;
}

.wxc-next {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--theme-accent, #0a84ff);
}

.wxc-label {
  display: block;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.wxc-inline {
  display: flex;
  gap: 8px;
}

.wxc-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 0;
}

.wxc-danger {
  font-size: 12px;
  line-height: 22px;
  color: #e11d48;
}

.wxc-copy-btns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 2px;
}

.wxc-copy-btns :deep(.ant-btn) {
  height: 46px;
  font-size: 16px;
  font-weight: 600;
}
</style>
