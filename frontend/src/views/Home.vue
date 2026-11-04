<script setup>
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import {
  ThunderboltFilled,
  SafetyCertificateOutlined,
  StopOutlined,
  SkinOutlined,
} from '@ant-design/icons-vue'
import GlassButton from '@/components/GlassButton.vue'
import { useTheme } from '@/composables/useTheme'

defineOptions({ name: 'HomeView' })

const { t } = useI18n()
const { currentTheme } = useTheme()
</script>

<template>
  <div class="home-landing">
    <div class="site-width-shell home-landing__shell">
      <div class="home-landing__main">
        <div class="home-brand-mark">
          <span class="home-brand-icon">
            <img
              :src="`/themes/${currentTheme.id}.png`"
              :alt="t('home.title')"
              class="home-brand-logo"
              width="120"
              height="120"
            />
          </span>
        </div>

        <h1 class="home-title">{{ t('home.title') }}</h1>

        <div class="home-copy">
          <p class="home-desc">{{ t('home.desc') }}</p>
          <p class="home-desc">{{ t('home.descFree') }}</p>
        </div>

        <div class="home-actions">
          <GlassButton to="/install" primary block>
            {{ t('home.uploadInstall') }}
          </GlassButton>
        </div>
      </div>

      <section class="home-features" aria-labelledby="home-features-title">
        <h2 id="home-features-title" class="home-features-title">{{ t('home.featuresTitle') }}</h2>
        <div class="home-feature-list-wrap">
          <article class="home-feature-card">
            <div class="home-feature-icon home-feature-icon--flash" aria-hidden="true">
              <ThunderboltFilled />
            </div>
            <div class="home-feature-body">
              <h3>{{ t('home.speedTitle') }}</h3>
              <p>{{ t('home.speedDesc') }}</p>
            </div>
          </article>

          <article class="home-feature-card">
            <div class="home-feature-icon" aria-hidden="true">
              <SafetyCertificateOutlined />
            </div>
            <div class="home-feature-body">
              <h3>{{ t('home.certTitle') }}</h3>
              <p>{{ t('home.certDesc') }}</p>
            </div>
          </article>

          <article class="home-feature-card">
            <div class="home-feature-icon" aria-hidden="true">
              <SkinOutlined />
            </div>
            <div class="home-feature-body">
              <h3>{{ t('home.themeTitle') }}</h3>
              <ol class="home-feature-list">
                <i18n-t keypath="home.themeLine1" tag="li">
                  <template #config>
                    <RouterLink class="home-text-link home-text-link--plain" to="/theme">
                      {{ t('home.goThemeConfig') }}
                    </RouterLink>
                  </template>
                </i18n-t>
                <i18n-t keypath="home.themeLine2" tag="li">
                  <template #guide>
                    <RouterLink
                      class="home-text-link home-text-link--plain"
                      :to="{ path: '/theme', query: { tab: 'guide' } }"
                    >
                      {{ t('home.goThemeGuide') }}
                    </RouterLink>
                  </template>
                </i18n-t>
              </ol>
            </div>
          </article>

          <article class="home-feature-card">
            <div class="home-feature-icon" aria-hidden="true">
              <StopOutlined />
            </div>
            <div class="home-feature-body">
              <h3>{{ t('home.blTitle') }}</h3>
              <ol class="home-feature-list">
                <i18n-t keypath="home.blLine1" tag="li">
                  <template #channel>
                    <RouterLink
                      class="home-text-link"
                      :to="{ path: '/docs', query: { tab: 'blacklist' } }"
                      :data-tip="t('home.goIntegrate')"
                    >
                      {{ t('home.channelBan') }}
                    </RouterLink>
                  </template>
                  <template #global>{{ t('home.globalBan') }}</template>
                  <template #pending>
                    <span class="home-pending">{{ t('home.globalPending') }}</span>
                  </template>
                </i18n-t>
                <i18n-t keypath="home.blLine2" tag="li">
                  <template #lookup>
                    <RouterLink class="home-text-link home-text-link--plain" to="/lookup">
                      {{ t('home.goQuery') }}
                    </RouterLink>
                  </template>
                </i18n-t>
                <i18n-t keypath="home.blLine3" tag="li">
                  <template #api>
                    <RouterLink
                      class="home-text-link home-text-link--plain"
                      :to="{ path: '/docs', query: { tab: 'lookup' } }"
                    >
                      {{ t('home.goLookupApi') }}
                    </RouterLink>
                  </template>
                </i18n-t>
              </ol>
            </div>
          </article>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.home-landing {
  flex: 1;
  width: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.home-landing__shell {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding-top: clamp(16px, 4vw, 32px);
  padding-bottom: clamp(24px, 5vw, 48px);
  gap: clamp(28px, 5vh, 48px);
}

.home-landing__main {
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: clamp(18px, 3vh, 28px);
  box-sizing: border-box;
  min-height: min(52vh, 560px);
}

.home-brand-mark {
  margin-bottom: clamp(4px, 1vh, 12px);
}

.home-brand-icon {
  width: clamp(112px, 28vw, 144px);
  height: clamp(112px, 28vw, 144px);
  border-radius: clamp(26px, 6.5vw, 32px);
  padding: 0;
  background: rgba(255, 255, 255, 0.22);
  border: 0.5px solid rgba(255, 255, 255, 0.45);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.home-brand-logo {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  display: block;
}

.home-title {
  margin: 0;
  font-size: clamp(32px, 7vw, 44px);
  font-weight: 700;
  letter-spacing: -0.04em;
  line-height: 1.1;
  color: var(--text-primary);
}

.home-copy {
  width: 100%;
  max-width: 520px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.home-desc {
  width: 100%;
  margin: 0;
  font-size: clamp(14px, 2.9vw, 16px);
  line-height: 1.75;
  color: var(--text-secondary);
}

.home-actions {
  width: 100%;
  max-width: 320px;
  margin-top: clamp(8px, 2vh, 16px);
  display: flex;
  justify-content: center;
}

.home-features {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.home-features-title {
  margin: 0 0 18px;
  text-align: center;
  font-size: clamp(20px, 4vw, 26px);
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.home-feature-list-wrap {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.home-feature-card {
  box-sizing: border-box;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  width: 100%;
  padding: 20px 22px;
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.28);
  border: 0.5px solid rgba(255, 255, 255, 0.5);
  box-shadow: var(--lg-shadow-env);
  backdrop-filter: blur(28px) saturate(160%);
  -webkit-backdrop-filter: blur(28px) saturate(160%);
  transform: translateZ(0);
  text-align: left;
}

.home-feature-body {
  flex: 1;
  min-width: 0;
}

.home-feature-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--theme-accent, #0a84ff);
  background: color-mix(in srgb, var(--theme-accent, #0a84ff) 14%, rgba(255, 255, 255, 0.55));
}

.home-feature-icon--flash {
  color: #f5a524;
  background: rgba(245, 165, 36, 0.18);
}

.home-feature-card h3 {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

.home-feature-card p {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.home-feature-list {
  margin: 0;
  padding-left: 18px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
}

.home-feature-list li + li {
  margin-top: 8px;
}

.home-pending {
  color: var(--text-muted);
}

.home-text-link {
  position: relative;
  color: var(--theme-accent, #0a84ff);
  text-decoration: underline;
  text-underline-offset: 3px;
  font-weight: 600;
}

.home-text-link--plain {
  margin-left: 6px;
  white-space: nowrap;
}

.home-text-link[data-tip]::after {
  content: attr(data-tip);
  position: absolute;
  left: 50%;
  bottom: calc(100% + 8px);
  transform: translateX(-50%) translateY(4px);
  padding: 5px 8px;
  border-radius: 8px;
  background: rgba(22, 22, 24, 0.88);
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.2;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.16s ease, transform 0.16s ease;
  z-index: 2;
}

.home-text-link[data-tip]:hover::after,
.home-text-link[data-tip]:focus-visible::after {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

@media (max-width: 480px) {
  .home-landing__main {
    min-height: 0;
    padding-top: 12px;
  }

  .home-copy {
    max-width: 100%;
    padding: 0 4px;
  }
}

@media (min-width: 900px) {
  .home-landing__main {
    max-width: 720px;
  }
}
</style>
