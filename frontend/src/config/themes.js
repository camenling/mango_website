/** 与 HillMountPlatform HMAppIconOption 五套图标一致 */
export const THEMES = [
  {
    id: 1,
    nameKey: 'theme.canglan',
    accent: '#0a84ff',
    accentSoft: 'rgba(10, 132, 255, 0.88)',
    bgBase: '#e8f4ff',
    mesh1: '#ff6b9d',
    mesh2: '#4facfe',
    mesh3: '#a18cd1',
    mesh4: '#43e97b',
    wallpaper: `
      linear-gradient(135deg, rgba(255, 120, 180, 0.45) 0%, transparent 45%),
      linear-gradient(225deg, rgba(80, 160, 255, 0.5) 0%, transparent 50%),
      linear-gradient(45deg, rgba(120, 255, 200, 0.35) 0%, transparent 40%),
      linear-gradient(160deg, #e8f4ff 0%, #f5e6ff 35%, #ffe8f0 65%, #d4f0ff 100%)`,
  },
  {
    id: 2,
    nameKey: 'theme.mihui',
    accent: '#d97a55',
    accentSoft: 'rgba(217, 122, 85, 0.82)',
    bgBase: '#faf6f2',
    meshOpacity: 0.48,
    mesh1: '#f0c8ae',
    mesh2: '#e8b098',
    mesh3: '#f5dcc8',
    mesh4: '#ddb896',
    wallpaper: `
      linear-gradient(135deg, rgba(240, 200, 174, 0.22) 0%, transparent 45%),
      linear-gradient(225deg, rgba(217, 122, 85, 0.16) 0%, transparent 50%),
      linear-gradient(45deg, rgba(245, 220, 200, 0.18) 0%, transparent 40%),
      linear-gradient(160deg, #faf6f2 0%, #f7efe9 35%, #faf4ef 65%, #f3ebe3 100%)`,
  },
  {
    id: 3,
    nameKey: 'theme.qinglan',
    accent: '#30b0c7',
    accentSoft: 'rgba(48, 176, 199, 0.9)',
    bgBase: '#e8faf5',
    mesh1: '#43e97b',
    mesh2: '#38f9d7',
    mesh3: '#56ccf2',
    mesh4: '#7bed9f',
    wallpaper: `
      linear-gradient(135deg, rgba(67, 233, 123, 0.4) 0%, transparent 45%),
      linear-gradient(225deg, rgba(56, 204, 242, 0.45) 0%, transparent 50%),
      linear-gradient(45deg, rgba(56, 249, 215, 0.35) 0%, transparent 40%),
      linear-gradient(160deg, #eafff5 0%, #e0f7fa 35%, #f0fff8 65%, #d4f5ef 100%)`,
  },
  {
    id: 4,
    nameKey: 'theme.jiangxia',
    accent: '#ff375f',
    accentSoft: 'rgba(255, 55, 95, 0.9)',
    bgBase: '#fff0f3',
    mesh1: '#ff6b9d',
    mesh2: '#c471ed',
    mesh3: '#ff375f',
    mesh4: '#f5576c',
    wallpaper: `
      linear-gradient(135deg, rgba(255, 55, 95, 0.42) 0%, transparent 45%),
      linear-gradient(225deg, rgba(196, 113, 237, 0.45) 0%, transparent 50%),
      linear-gradient(45deg, rgba(245, 87, 108, 0.35) 0%, transparent 40%),
      linear-gradient(160deg, #fff0f5 0%, #fce4ec 35%, #fff5f7 65%, #f8e1ff 100%)`,
  },
  {
    id: 5,
    nameKey: 'theme.liujin',
    accent: '#c9922e',
    accentSoft: 'rgba(201, 146, 46, 0.82)',
    bgBase: '#faf8f3',
    meshOpacity: 0.48,
    mesh1: '#ecd898',
    mesh2: '#d4b068',
    mesh3: '#f0e0b0',
    mesh4: '#c8a858',
    wallpaper: `
      linear-gradient(135deg, rgba(236, 216, 152, 0.2) 0%, transparent 45%),
      linear-gradient(225deg, rgba(201, 146, 46, 0.14) 0%, transparent 50%),
      linear-gradient(45deg, rgba(240, 224, 176, 0.18) 0%, transparent 40%),
      linear-gradient(160deg, #faf8f3 0%, #f7f1e6 35%, #faf5eb 65%, #f2ead8 100%)`,
  },
]

export const DEFAULT_THEME_ID = 1
export const THEME_STORAGE_KEY = 'mangosign-theme'

export function getThemeById(id) {
  return THEMES.find((t) => t.id === id) || THEMES[0]
}
