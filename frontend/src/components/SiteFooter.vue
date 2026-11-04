<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const year = new Date().getFullYear()
</script>

<template>
  <div class="site-width-shell site-footer">
    <footer class="footer-plain">
      <p>{{ $t('footer.copyright', { year }) }}</p>
      <p class="icp">
        <a
          class="icp-link icp-link--gongan"
          href="https://beian.mps.gov.cn/#/query/webSearch?code=45092102000216"
          target="_blank"
          rel="noreferrer"
        >
          <img class="icp-gongan" src="/gongan.png" width="16" height="16" alt="" />
          {{ $t('footer.gongan') }}
        </a>
        <span class="icp-sep" aria-hidden="true">|</span>
        <a
          class="icp-link"
          href="https://beian.miit.gov.cn/"
          target="_blank"
          rel="noreferrer"
        >{{ $t('footer.icp') }}</a>
      </p>
    </footer>
  </div>
</template>

<style scoped>
.site-footer {
  flex-shrink: 0;
  padding-top: clamp(20px, 4vw, 32px);
  padding-bottom: max(28px, env(safe-area-inset-bottom));
}

.footer-plain {
  width: 100%;
  margin: 0 auto;
  text-align: center;
  font-size: clamp(12px, 2.4vw, 13px);
  line-height: 1.85;
  color: var(--text-muted);
}

.footer-plain p {
  margin: 0 0 4px;
}

.footer-plain .icp {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 0;
  font-size: clamp(11px, 2.2vw, 12px);
}

.footer-plain .icp-sep {
  margin: 0 8px;
  color: var(--text-muted);
  opacity: 0.55;
}

.footer-plain .icp-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: inherit;
  text-decoration: none;
}

.footer-plain .icp-gongan {
  width: 16px;
  height: 16px;
  display: block;
  flex-shrink: 0;
}

.footer-plain .icp-link:hover {
  text-decoration: underline;
}
</style>
