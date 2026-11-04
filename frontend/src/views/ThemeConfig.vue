<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { DownloadOutlined, UndoOutlined } from '@ant-design/icons-vue'
import AppGlass from '@/components/AppGlass.vue'
import ThemeEditor from '@/components/ThemeEditor.vue'
import { useTheme } from '@/composables/useTheme'
import { HMTHEME_FILE_NAME, HMTHEME_MAX_BYTES, buildHmtheme, defaultThemeState } from '@/config/hmTheme'
import {
  THEME_INFO_SAMPLE,
  THEME_INSTALL_PATH,
  THEME_MANIFEST_SAMPLE,
  THEME_TREE,
  THEME_ZIP_COMMAND,
  themeGuide,
} from '@/content/themeGuide'

defineOptions({ name: 'ThemeConfigView' })

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentThemeId } = useTheme()

const tab = ref(route.query.tab === 'guide' ? 'guide' : 'config')
watch(tab, (value) => {
  router.replace({ query: value === 'guide' ? { tab: 'guide' } : {} })
})
watch(() => route.query.tab, (value) => {
  tab.value = value === 'guide' ? 'guide' : 'config'
})

const tabOptions = computed(() => [
  { label: t('themeCfg.tabConfig'), value: 'config' },
  { label: t('themeCfg.tabGuide'), value: 'guide' },
])

const state = reactive(defaultThemeState(currentThemeId.value))
const exporting = ref(false)
const editorRef = ref(null)

// ---------- 导出 ----------

function resetAll() {
  editorRef.value?.resetAll()
}

async function download() {
  exporting.value = true
  try {
    const bytes = await buildHmtheme(state, t('themeCfg.defaultName'))
    if (bytes.byteLength > HMTHEME_MAX_BYTES) {
      message.error(t('themeCfg.errTooLarge'))
      return
    }
    const url = URL.createObjectURL(new Blob([bytes], { type: 'application/zip' }))
    const a = document.createElement('a')
    a.href = url
    a.download = HMTHEME_FILE_NAME
    a.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    message.success(t('themeCfg.downloaded'))
  } catch (e) {
    message.error(e.message || t('themeCfg.errExport'))
  } finally {
    exporting.value = false
  }
}

const guide = computed(() => themeGuide(locale.value))
</script>

<template>
  <div class="theme-page">
    <div class="site-width-shell theme-shell">
      <AppGlass variant="panel" class="theme-block" padding="24px 22px" :track-mouse="false">
        <h1 class="theme-title">{{ t('themeCfg.title') }}</h1>
        <p class="theme-desc">{{ t('themeCfg.desc') }}</p>
        <a-segmented v-model:value="tab" :options="tabOptions" block class="theme-tabs" />
      </AppGlass>

      <!-- 配置预览 -->
      <ThemeEditor v-if="tab === 'config'" ref="editorRef" :state="state">
        <template #actions>
          <AppGlass variant="panel" class="theme-block theme-actions" padding="20px 20px" :track-mouse="false">
            <a-button type="primary" size="large" block :loading="exporting" @click="download">
              <DownloadOutlined /> {{ t('themeCfg.download') }}
            </a-button>
            <a-button size="large" block @click="resetAll">
              <UndoOutlined /> {{ t('themeCfg.resetAll') }}
            </a-button>
            <p class="section-hint">
              {{ t('themeCfg.downloadHint', { file: HMTHEME_FILE_NAME }) }}
              <a class="text-link" @click="tab = 'guide'">{{ t('themeCfg.goGuide') }}</a>
            </p>
          </AppGlass>
        </template>
      </ThemeEditor>

      <!-- 内置方法 -->
      <div v-else class="guide">
        <AppGlass variant="panel" class="theme-block" padding="22px 22px" :track-mouse="false">
          <h2 class="section-title">{{ guide.getTitle }}</h2>
          <ul class="guide-list">
            <li v-for="(line, i) in guide.get" :key="i">{{ line }}</li>
          </ul>
        </AppGlass>

        <AppGlass variant="panel" class="theme-block" padding="22px 22px" :track-mouse="false">
          <h2 class="section-title">{{ guide.placeTitle }}</h2>
          <h3 class="sub-title sub-title--first">{{ guide.placeAuto.title }}</h3>
          <p class="guide-p">{{ guide.placeAuto.intro }}</p>
          <pre class="code">{{ THEME_INSTALL_PATH }}</pre>
          <ul class="guide-list">
            <li v-for="(line, i) in guide.placeAuto.lines" :key="i">{{ line }}</li>
          </ul>
          <h3 class="sub-title">{{ guide.placeManual.title }}</h3>
          <ul class="guide-list">
            <li v-for="(line, i) in guide.placeManual.lines" :key="i">{{ line }}</li>
          </ul>
        </AppGlass>

        <AppGlass variant="panel" class="theme-block" padding="22px 22px" :track-mouse="false">
          <h2 class="section-title">{{ guide.treeTitle }}</h2>
          <pre class="code">{{ THEME_TREE }}</pre>
          <p class="guide-p">{{ guide.treeNote }}</p>
          <h3 class="sub-title">{{ guide.fieldsTitle }}</h3>
          <div class="table-wrap">
            <table class="field-table">
              <thead>
                <tr>
                  <th v-for="h in guide.fieldsHead" :key="h">{{ h }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in guide.fields" :key="row[1]">
                  <td>{{ row[0] }}</td>
                  <td><code>{{ row[1] }}</code></td>
                  <td>{{ row[2] }}</td>
                  <td>{{ row[3] }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </AppGlass>

        <AppGlass variant="panel" class="theme-block" padding="22px 22px" :track-mouse="false">
          <h2 class="section-title">{{ guide.manualTitle }}</h2>
          <ol class="guide-list">
            <li v-for="(line, i) in guide.manual" :key="i">
              {{ typeof line === 'string' ? line : line.text }}
              <ul v-if="line.sub" class="guide-sublist">
                <li v-for="(sub, j) in line.sub" :key="j">{{ sub }}</li>
              </ul>
              <pre v-if="i === 3" class="code">{{ THEME_ZIP_COMMAND }}</pre>
            </li>
          </ol>
          <h3 class="sub-title">{{ guide.manifestTitle }}</h3>
          <pre class="code">{{ THEME_MANIFEST_SAMPLE }}</pre>
          <h3 class="sub-title">{{ guide.infoTitle }}</h3>
          <pre class="code">{{ THEME_INFO_SAMPLE }}</pre>
        </AppGlass>
      </div>
    </div>
  </div>
</template>

<style scoped>
.theme-page {
  width: 100%;
  padding-bottom: 24px;
}

.theme-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 40px;
  padding-bottom: 32px;
}

@media (max-width: 720px) {
  .theme-shell {
    padding-top: 28px;
  }
}

.theme-title {
  margin: 0 0 8px;
  font-size: clamp(24px, 5vw, 30px);
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.theme-desc {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.guide {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.section-title {
  margin: 0 0 6px;
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
}

.sub-title {
  margin: 18px 0 8px;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.sub-title--first {
  margin-top: 10px;
}

.section-hint {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.theme-actions :deep(.ant-btn + .ant-btn) {
  margin-top: 10px;
}

.theme-actions .section-hint {
  margin: 12px 0 0;
}

.text-link {
  margin-left: 4px;
  color: var(--theme-accent, #0a84ff);
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
}

.guide-p {
  margin: 0 0 10px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
}

.guide-list {
  margin: 0;
  padding-left: 20px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
}

.guide-list li + li {
  margin-top: 6px;
}

.guide-sublist {
  margin: 4px 0 0;
  padding-left: 1.2em;
  list-style: none;
}

.guide-sublist li {
  padding-left: 1.2em;
  text-indent: -1.2em;
}

.code {
  margin: 8px 0 10px;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(22, 22, 24, 0.86);
  color: #f5f5f7;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
}

.table-wrap {
  overflow-x: auto;
}

.field-table {
  width: 100%;
  min-width: 720px;
  border-collapse: collapse;
  font-size: 13px;
  color: var(--text-secondary);
}

.field-table th:first-child,
.field-table td:first-child {
  min-width: 112px;
  white-space: nowrap;
  color: var(--text-primary);
  font-weight: 600;
}

.field-table th:nth-child(2),
.field-table td:nth-child(2) {
  min-width: 180px;
}

.field-table th,
.field-table td {
  padding: 8px 10px;
  text-align: left;
  vertical-align: top;
  border-bottom: 0.5px solid rgba(0, 0, 0, 0.1);
}

.field-table th {
  color: var(--text-primary);
  font-weight: 700;
}

.field-table code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: var(--text-primary);
}
</style>
