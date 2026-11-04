/**
 * HillMount 整体主题方案（.hmtheme）的字段与默认值。
 * 与 HillMountPlatform Classes/Theme 下各模块的 Payload 一一对应，改字段前先对照 App 端。
 */
import { strToU8, zipSync } from 'fflate'
import { THEMES } from '@/config/themes'

export const HMTHEME_FILE_NAME = 'hm-builtin-theme.hmtheme'
export const HMTHEME_MAX_BYTES = 30 * 1024 * 1024

export const APPEARANCE_MODES = [
  { value: 0, key: 'system' },
  { value: 1, key: 'light' },
  { value: 2, key: 'dark' },
]

/** HMSandboxFileIconThemeStyle：0 经典，2 液态玻璃 */
export const FOLDER_STYLES = [
  { value: 2, key: 'liquid', image: '/theme-config/folder-liquid.png' },
  { value: 0, key: 'classic', image: '/theme-config/folder-classic.png' },
]

/** HMPageBackgroundBase：图片下面垫的底色 */
export const BACKGROUND_BASES = [
  { value: 0, key: 'system' },
  { value: 1, key: 'light' },
  { value: 2, key: 'dark' },
]

export const TAB_SLOTS = ['app', 'sign', 'cert', 'file', 'setting']
export const TAB_ICON_PIXELS = 81

export const DEFAULT_LOCATIONS = [0, 0.35, 0.65, 1]

/** 各风格主渐变四个色标（App 端 builtinGlassWallpaper 第一层） */
export const BUILTIN_STOPS = {
  1: ['#E8F4FF', '#F5E6FF', '#FFE8F0', '#D4F0FF'],
  2: ['#FAF6F2', '#F7EFE9', '#FAF4EF', '#F3EBE3'],
  3: ['#EAFFF5', '#E0F7FA', '#F0FFF8', '#D4F5EF'],
  4: ['#FFF0F5', '#FCE4EC', '#FFF5F7', '#F8E1FF'],
  5: ['#FAF8F3', '#F7F1E6', '#FAF5EB', '#F2EAD8'],
}

/** 主渐变上面叠的三层光斑：[角度, 颜色, 不透明度, 结束位置] */
const OVERLAYS = {
  1: [[45, '#78FFC8', 0.35, 0.4], [225, '#50A0FF', 0.5, 0.5], [135, '#FF78B4', 0.45, 0.45]],
  2: [[45, '#F5DCC8', 0.18, 0.4], [225, '#D97A55', 0.16, 0.5], [135, '#F0C8AE', 0.22, 0.45]],
  3: [[45, '#38F9D7', 0.35, 0.4], [225, '#38CCF2', 0.45, 0.5], [135, '#43E97B', 0.4, 0.45]],
  4: [[45, '#F5576C', 0.35, 0.4], [225, '#C471ED', 0.45, 0.5], [135, '#FF375F', 0.42, 0.45]],
  5: [[45, '#F0E0B0', 0.18, 0.4], [225, '#C9922E', 0.14, 0.5], [135, '#ECD898', 0.2, 0.45]],
}

const DARK_WALLPAPER_DIM = 0.45

export function styleAccent(style) {
  return (THEMES.find((t) => t.id === style) || THEMES[0]).accent.toUpperCase()
}

export function normalizeHex(hex) {
  const raw = String(hex || '').trim().replace(/^#/, '')
  return /^[0-9a-fA-F]{6}$/.test(raw) ? `#${raw.toUpperCase()}` : null
}

function rgba(hex, alpha) {
  const n = Number.parseInt(hex.slice(1), 16)
  return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, ${alpha})`
}

/** 和 App 端绘制顺序一致：主渐变在最下，三层光斑依次叠上；深色模式再压一层黑。 */
export function wallpaperCss(style, stops, dark) {
  const layers = [...OVERLAYS[style]].reverse().map(
    ([angle, hex, alpha, end]) => `linear-gradient(${angle}deg, ${rgba(hex, alpha)} 0%, ${rgba(hex, 0)} ${end * 100}%)`,
  )
  const primary = stops.hexes
    .map((hex, i) => `${hex} ${Math.round(stops.locations[i] * 100)}%`)
    .join(', ')
  const dim = dark ? [`linear-gradient(rgba(0, 0, 0, ${DARK_WALLPAPER_DIM}), rgba(0, 0, 0, ${DARK_WALLPAPER_DIM}))`] : []
  return [...dim, ...layers, `linear-gradient(160deg, ${primary})`].join(', ')
}

export function builtinStops(style) {
  return { hexes: [...BUILTIN_STOPS[style]], locations: [...DEFAULT_LOCATIONS] }
}

export function isBuiltinStops(style, stops) {
  return stops.hexes.every((hex, i) => hex.toUpperCase() === BUILTIN_STOPS[style][i])
    && stops.locations.every((loc, i) => Math.abs(loc - DEFAULT_LOCATIONS[i]) < 0.001)
}

/** 编辑器的初始状态，等同 App 装好后什么都没改。 */
export function defaultThemeState(style = 1) {
  return {
    name: '',
    appearance: 0,
    style,
    glass: true,
    background: {
      mode: 0,
      stops: builtinStops(style),
      image: null,
      opacity: 0.55,
      blur: 0,
      base: 0,
    },
    textColor: {
      accentLight: null,
      accentDark: null,
      primaryLight: null,
      primaryDark: null,
      secondaryLight: null,
      secondaryDark: null,
    },
    folderIcon: 2,
    tabIcon: {
      enabled: false,
      name: '',
      icons: Object.fromEntries(TAB_SLOTS.map((slot) => [slot, { normal: null, selected: null }])),
    },
  }
}

export function defaultTabIconUrl(slot, selected) {
  return `/theme-config/tab/${slot}${selected ? '_selected' : ''}.png`
}

// ---------- 图片处理 ----------

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = reject
    img.src = src
  })
}

function canvasBytes(canvas, type, quality) {
  return new Promise((resolve, reject) => {
    canvas.toBlob(async (blob) => {
      if (!blob) return reject(new Error('encode failed'))
      resolve(new Uint8Array(await blob.arrayBuffer()))
    }, type, quality)
  })
}

/** 底栏图标：居中裁成正方形，缩到 81×81 PNG，保留透明。和 App 内自定义图标的处理一致。 */
export async function normalizeTabIcon(src) {
  const img = await loadImage(src)
  const side = Math.min(img.naturalWidth, img.naturalHeight)
  const canvas = document.createElement('canvas')
  canvas.width = TAB_ICON_PIXELS
  canvas.height = TAB_ICON_PIXELS
  canvas.getContext('2d').drawImage(
    img,
    (img.naturalWidth - side) / 2,
    (img.naturalHeight - side) / 2,
    side,
    side,
    0,
    0,
    TAB_ICON_PIXELS,
    TAB_ICON_PIXELS,
  )
  return canvasBytes(canvas, 'image/png')
}

/** 背景图：长边不超过 2796 像素，转成 JPEG。 */
export async function encodeBackground(src) {
  const img = await loadImage(src)
  const ratio = Math.min(1, 2796 / Math.max(img.naturalWidth, img.naturalHeight))
  const canvas = document.createElement('canvas')
  canvas.width = Math.round(img.naturalWidth * ratio)
  canvas.height = Math.round(img.naturalHeight * ratio)
  canvas.getContext('2d').drawImage(img, 0, 0, canvas.width, canvas.height)
  return canvasBytes(canvas, 'image/jpeg', 0.82)
}

// ---------- 打包 ----------

function isoNow() {
  // App 端按 ISO8601 解析，不带毫秒
  return new Date().toISOString().replace(/\.\d{3}Z$/, 'Z')
}

function uuid() {
  if (crypto.randomUUID) return crypto.randomUUID()
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

/** 生成 .hmtheme（zip）：根目录 manifest.json，文件放在 assets/<模块>/ 下。 */
export async function buildHmtheme(state, fallbackName) {
  const files = {}
  const now = isoNow()
  const bg = state.background

  let image
  if (bg.mode === 1 && bg.image) {
    files['assets/background/background.jpg'] = await encodeBackground(bg.image)
    image = 'assets/background/background.jpg'
  }
  const background = {
    v: 1,
    mode: image ? 1 : 0,
    opacity: bg.opacity,
    blur: bg.blur,
    base: bg.base,
  }
  const stops = {
    hexes: bg.stops.hexes.map((h, i) => normalizeHex(h) || BUILTIN_STOPS[state.style][i]),
    locations: bg.stops.locations.map((l, i) => (Number.isFinite(l) ? Math.min(Math.max(l, 0), 1) : DEFAULT_LOCATIONS[i])),
  }
  if (!isBuiltinStops(state.style, stops)) {
    background.stopsByStyle = { [String(state.style)]: stops }
  }
  if (image) background.image = image

  const colors = {}
  for (const [key, value] of Object.entries(state.textColor)) {
    const hex = normalizeHex(value)
    if (hex) colors[key] = hex
  }

  const tabIcon = { v: 1 }
  if (state.tabIcon.enabled) {
    const id = `custom-${uuid()}`
    const dir = 'assets/tabIcon/pack'
    for (const slot of TAB_SLOTS) {
      for (const selected of [false, true]) {
        const custom = state.tabIcon.icons[slot][selected ? 'selected' : 'normal']
        files[`${dir}/${slot}${selected ? '_selected' : ''}.png`] = await normalizeTabIcon(
          custom || defaultTabIconUrl(slot, selected),
        )
      }
    }
    const info = { id, name: state.tabIcon.name || fallbackName, summary: '', revision: 1, custom: true }
    files[`${dir}/info.json`] = strToU8(JSON.stringify(info))
    Object.assign(tabIcon, { id, revision: 1, pack: dir })
  }

  const manifest = {
    schemaVersion: 1,
    appVersion: '',
    id: uuid(),
    name: state.name.trim() || fallbackName,
    createdAt: now,
    updatedAt: now,
    source: { kind: 'local' },
    modules: {
      appearance: { v: 1, mode: state.appearance },
      style: { v: 1, style: state.style, glass: state.glass },
      background,
      textColor: { v: 1, colors },
      folderIcon: { v: 1, style: state.folderIcon },
      tabIcon,
    },
  }
  files['manifest.json'] = strToU8(JSON.stringify(manifest, null, 2))
  return zipSync(files, { level: 6 })
}
