<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { SearchOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { lookupBlacklist } from '@/api/sign'

defineOptions({ name: 'LookupView' })

const { t } = useI18n()
const udid = ref('')
const loading = ref(false)
const result = ref(null)

function errorMessage(err) {
  const code = err?.response?.data?.error
  if (code === 'bad_request') return t('lookup.badUdid')
  return t('lookup.failed')
}

async function handleQuery() {
  const value = udid.value.trim()
  if (!value) {
    message.warning(t('lookup.warnRequired'))
    return
  }
  loading.value = true
  result.value = null
  try {
    const data = await lookupBlacklist(value)
    if (!data?.status) {
      message.error(t('lookup.failed'))
      return
    }
    result.value = data
  } catch (err) {
    message.error(errorMessage(err))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="lookup-page">
    <div class="site-width-shell lookup-shell">
      <section class="lookup-intro">
        <h1 class="lookup-title">{{ t('lookup.title') }}</h1>
        <p class="lookup-desc">{{ t('lookup.desc') }}</p>
      </section>

      <section class="lookup-disclaimer">
        <h2 class="lookup-disclaimer-title">{{ t('lookup.disclaimerTitle') }}</h2>
        <p class="lookup-disclaimer-text">{{ t('lookup.disclaimer') }}</p>
      </section>

      <section class="lookup-card">
        <label class="lookup-label" for="lookup-udid">{{ t('lookup.udidLabel') }}</label>
        <a-input
          id="lookup-udid"
          v-model:value="udid"
          size="large"
          :placeholder="t('lookup.udidPlaceholder')"
          :maxlength="64"
          allow-clear
          @pressEnter="handleQuery"
        />
        <a-button
          type="primary"
          size="large"
          block
          :loading="loading"
          class="lookup-btn"
          @click="handleQuery"
        >
          <SearchOutlined /> {{ t('lookup.submit') }}
        </a-button>

        <p v-if="result?.status === 'none'" class="lookup-result">
          {{ t('lookup.none') }}
        </p>
        <i18n-t
          v-else-if="result?.status === 'channel'"
          keypath="lookup.channel"
          tag="p"
          class="lookup-result"
        >
          <template #n>
            <strong class="lookup-emph">{{ result.channel_count }}</strong>
          </template>
        </i18n-t>
        <i18n-t
          v-else-if="result?.status === 'global'"
          keypath="lookup.global"
          tag="p"
          class="lookup-result"
        >
          <template #mark>
            <strong class="lookup-emph">{{ t('lookup.globalWord') }}</strong>
          </template>
        </i18n-t>
      </section>

      <p class="lookup-docs">
        <RouterLink class="lookup-docs-link" :to="{ path: '/docs', query: { tab: 'lookup' } }">
          {{ t('lookup.goDocs') }}
        </RouterLink>
      </p>
    </div>
  </div>
</template>

<style scoped>
.lookup-page {
  width: 100%;
  padding-bottom: 24px;
}

.lookup-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 40px;
  padding-bottom: 32px;
  box-sizing: border-box;
  max-width: 560px;
}

.lookup-intro {
  text-align: center;
}

.lookup-title {
  font-size: clamp(24px, 5vw, 30px);
  font-weight: 700;
  margin: 0 0 10px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.lookup-desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.lookup-disclaimer {
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: var(--radius-lg);
  background: rgba(255, 196, 90, 0.18);
  border: 0.5px solid rgba(180, 110, 0, 0.28);
}

.lookup-disclaimer-title {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 650;
  color: #8a5a00;
}

.lookup-disclaimer-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.75;
  color: var(--text-primary);
}

.lookup-card {
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

.lookup-label {
  display: block;
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 10px;
  color: var(--text-primary);
}

.lookup-btn {
  margin-top: 16px;
}

.lookup-result {
  margin: 20px 0 0;
  text-align: center;
  font-size: 16px;
  line-height: 1.7;
  color: var(--text-primary);
}

.lookup-emph {
  color: #e11d48;
  font-weight: 800;
  font-size: 1.45em;
  font-style: normal;
}

.lookup-docs {
  margin: 0;
  text-align: center;
}

.lookup-docs-link {
  color: var(--theme-accent, #0a84ff);
  text-decoration: underline;
  text-underline-offset: 3px;
  font-size: 14px;
  font-weight: 600;
}
</style>
