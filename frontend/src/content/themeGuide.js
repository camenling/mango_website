/** 主题配置页「内置方法」：只讲怎么配、怎么打包、放哪里。 */

export const THEME_TREE = `hm-builtin-theme.hmtheme   (zip)
├── manifest.json
└── assets/
    ├── background/
    │   └── background.jpg
    └── tabIcon/
        └── pack/
            ├── info.json
            ├── app.png       app_selected.png
            ├── sign.png      sign_selected.png
            ├── cert.png      cert_selected.png
            ├── file.png      file_selected.png
            └── setting.png   setting_selected.png`

export const THEME_INSTALL_PATH = 'Payload/HillMountPlatform.app/website/hm-builtin-theme.hmtheme'

export const THEME_ZIP_COMMAND = `cd my-theme
zip -r -X ../hm-builtin-theme.hmtheme manifest.json assets -x "*.DS_Store"`

export const THEME_MANIFEST_SAMPLE = `{
  "schemaVersion": 1,
  "appVersion": "",
  "id": "8f2c1d7e-0b5a-4c1e-9d0f-2a6b3c4d5e6f",
  "name": "渠道主题",
  "createdAt": "2026-10-04T08:00:00Z",
  "updatedAt": "2026-10-04T08:00:00Z",
  "source": { "kind": "local" },
  "modules": {
    "appearance": { "v": 1, "mode": 0 },
    "style": { "v": 1, "style": 4, "glass": true },
    "background": {
      "v": 1,
      "mode": 1,
      "opacity": 0.55,
      "blur": 0.2,
      "base": 0,
      "stopsByStyle": {
        "4": {
          "hexes": ["#FFF0F5", "#FCE4EC", "#FFF5F7", "#F8E1FF"],
          "locations": [0, 0.35, 0.65, 1]
        }
      },
      "image": "assets/background/background.jpg"
    },
    "textColor": {
      "v": 1,
      "colors": { "accentLight": "#FF375F", "primaryDark": "#F5F5F7" }
    },
    "folderIcon": { "v": 1, "style": 2 },
    "tabIcon": {
      "v": 1,
      "id": "custom-3b1f0c2a-7d4e-4f5a-8b6c-9d0e1f2a3b4c",
      "revision": 1,
      "pack": "assets/tabIcon/pack"
    }
  }
}`

export const THEME_INFO_SAMPLE = `{
  "id": "custom-3b1f0c2a-7d4e-4f5a-8b6c-9d0e1f2a3b4c",
  "name": "渠道底栏",
  "summary": "",
  "revision": 1,
  "custom": true
}`

const zh = {
  getTitle: '一、拿到主题文件',
  get: [
    '在「配置预览」里调好各项，点「下载主题文件」，得到 hm-builtin-theme.hmtheme。',
    '或者在 App 内：主题设置 → 主题方案 → 保存当前配置，在方案上左滑选择「导出」，得到 .hmtheme 文件（按方式一内置时需改名为 hm-builtin-theme.hmtheme）。',
    '也可以按下面第四步手动打包。',
  ],
  placeTitle: '二、导入方式',
  placeAuto: {
    title: '方式一：安装即导入',
    intro: '把 IPA 改成 .zip 解压，把文件放到下面的位置（文件名必须一致）：',
    lines: [
      '写代码内置到 Payload 到website目录下再签名。必须先放文件再签名，签名后改包会无法安装。',
      '在本站「证书安装」签名时，会按所选的风格、强调色、玻璃开关自动生成这个文件，并覆盖包内同名文件。',
      '用户装好后首次打开 App 自动套用，不需要任何操作。',
    ],
  },
  placeManual: {
    title: '方式二：下载后在 App 内手动导入',
    lines: [
      '把 .hmtheme 文件发到 iPhone 上，隔空投送、微信 / QQ 文件、存到「文件」App 都可以，文件名不限。',
      '直接打开：在文件上点「分享」或「用其他应用打开」，选择芒果签；App 提示「已导入」后点「立即套用」。',
      '或者在 App 内：主题设置 → 主题方案 → 导入，选择这个文件。',
      '导入的方案保存在「主题方案」列表里，可以随时切换、导出或删除。',
    ],
  },
  treeTitle: '三、文件结构',
  treeNote: '只有 manifest.json 是必需的。不用的模块可以整段不写，不写的项保持 App 默认；没有自选背景图、没有自定义底栏时，对应的 assets 目录也不用放。',
  fieldsTitle: '各配置项对应的字段',
  fieldsHead: ['配置项', 'manifest 字段', '取值', '文件'],
  fields: [
    ['外观', 'appearance.mode', '0 跟随系统，1 浅色，2 深色', '—'],
    ['风格', 'style.style', '1 沧澜，2 蜜晖，3 青岚，4 绛霞，5 流金（同时决定桌面图标和默认强调色）', '—'],
    ['玻璃渐变', 'style.glass', 'true / false，iOS 26 及以上生效', '—'],
    ['背景类型', 'background.mode', '0 渐变，1 图片（没有图片时按渐变）', '—'],
    ['渐变色标', 'background.stopsByStyle["风格编号"]', 'hexes 四个 #RRGGBB，locations 四个 0–1 的位置；不写就是该风格原配色', '—'],
    ['背景图', 'background.image', '固定写 assets/background/background.jpg', 'background.jpg（JPEG，竖图，建议不超过 1290×2796）'],
    ['图片不透明度', 'background.opacity', '0.1–1，默认 0.55', '—'],
    ['图片模糊', 'background.blur', '0–1，对应 0–40 像素', '—'],
    ['图片底色', 'background.base', '0 跟随系统，1 白底，2 黑底', '—'],
    ['字体颜色', 'textColor.colors', 'accentLight / accentDark / primaryLight / primaryDark / secondaryLight / secondaryDark，#RRGGBB；不写就是默认', '—'],
    ['文件夹样式', 'folderIcon.style', '0 经典，2 液态玻璃', '—'],
    ['底栏图标', 'tabIcon.id / revision / pack', 'id 以 custom- 开头，revision 写 1，pack 固定写 assets/tabIcon/pack；整段只写 {"v":1} 表示用默认底栏', 'pack 目录下 10 张 PNG + info.json'],
  ],
  manualTitle: '四、手动打包',
  manual: [
    '新建一个目录（例如 my-theme），按上面的结构放好 manifest.json 和 assets。',
    {
      text: '按示例填写 manifest.json。两处格式容易出错：',
      sub: [
        '① 每个模块里的 "v": 1 是这个模块配置的格式版本号，App 靠它识别字段，目前固定写 1；漏写的模块会被跳过，那一项保持 App 默认。',
        '② createdAt、updatedAt 是方案的创建和修改时间，App 内「主题方案」列表按修改时间排序；必须是 ISO8601 格式且不带毫秒，例如 2026-10-04T08:00:00Z，写成 2026-10-04T08:00:00.123Z 这种带毫秒的，整个文件都无法导入。',
      ],
    },
    '底栏 10 张图：正方形 PNG，建议 81×81、透明背景，文件名固定为 <位置>.png 和 <位置>_selected.png，位置依次是 app、sign、cert、file、setting；5 个位置都要放齐。',
    '在目录里把 manifest.json 和 assets 一起压缩，manifest.json 必须在压缩包第一层，然后把后缀改成 .hmtheme。macOS 可以用：',
    '整个文件不能超过 30MB。',
  ],
  manifestTitle: 'manifest.json 示例',
  infoTitle: 'assets/tabIcon/pack/info.json 示例',
}

const en = {
  getTitle: '1. Get the theme file',
  get: [
    'Configure everything under "Configure & Preview" and click "Download theme file" to get hm-builtin-theme.hmtheme.',
    'Or in the app: Theme Settings → Theme Profiles → Save current, swipe left on the profile and tap Export to get a .hmtheme file (rename it to hm-builtin-theme.hmtheme when bundling with option 1).',
    'You can also pack it by hand (step 4).',
  ],
  placeTitle: '2. How to import',
  placeAuto: {
    title: 'Option 1: Applied on install',
    intro: 'Rename the IPA to .zip, unzip it, and put the file here (the file name must match):',
    lines: [
      'Zip the Payload folder back, rename it to .ipa, then sign. Add the file before signing — changing the app after signing breaks installation.',
      'When you sign on this site ("Install"), this file is generated from the selected style, accent color and glass switch, replacing any file with the same name.',
      'The theme is applied automatically the first time the user opens the app.',
    ],
  },
  placeManual: {
    title: 'Option 2: Download and import in the app',
    lines: [
      'Send the .hmtheme file to the iPhone — AirDrop, a chat app, or save it to the Files app. Any file name works.',
      'Open it directly: tap Share or "Open in" on the file and choose MangoSign; when the app shows "Imported", tap Apply now.',
      'Or in the app: Theme Settings → Theme Profiles → Import, then pick the file.',
      'Imported profiles stay in the Theme Profiles list, where you can switch, export or delete them anytime.',
    ],
  },
  treeTitle: '3. File layout',
  treeNote: 'Only manifest.json is required. Leave out any module you do not need — omitted settings keep the app default. Skip the assets folders when there is no background image or custom tab bar.',
  fieldsTitle: 'Settings and fields',
  fieldsHead: ['Setting', 'manifest field', 'Values', 'File'],
  fields: [
    ['Appearance', 'appearance.mode', '0 system, 1 light, 2 dark', '—'],
    ['Style', 'style.style', '1–5 (also sets the home screen icon and default accent)', '—'],
    ['Glass', 'style.glass', 'true / false, iOS 26 and later', '—'],
    ['Background type', 'background.mode', '0 gradient, 1 image (falls back to gradient without an image)', '—'],
    ['Gradient stops', 'background.stopsByStyle["style"]', 'hexes: four #RRGGBB, locations: four values 0–1; omit for the original colors', '—'],
    ['Background image', 'background.image', 'Always assets/background/background.jpg', 'background.jpg (JPEG, portrait, up to 1290×2796)'],
    ['Image opacity', 'background.opacity', '0.1–1, default 0.55', '—'],
    ['Image blur', 'background.blur', '0–1, maps to 0–40 px', '—'],
    ['Image base', 'background.base', '0 system, 1 white, 2 black', '—'],
    ['Text colors', 'textColor.colors', 'accentLight / accentDark / primaryLight / primaryDark / secondaryLight / secondaryDark as #RRGGBB; omit for defaults', '—'],
    ['Folder style', 'folderIcon.style', '0 classic, 2 liquid glass', '—'],
    ['Tab bar icons', 'tabIcon.id / revision / pack', 'id starts with custom-, revision 1, pack is assets/tabIcon/pack; {"v":1} alone means the default tab bar', '10 PNGs + info.json in pack'],
  ],
  manualTitle: '4. Pack by hand',
  manual: [
    'Create a folder (e.g. my-theme) and lay out manifest.json and assets as above.',
    {
      text: 'Fill in manifest.json following the example. Two formats are easy to get wrong:',
      sub: [
        '(1) "v": 1 inside each module is that module\'s format version, which the app uses to read its fields — always 1 for now; a module without it is skipped and keeps the app default.',
        '(2) createdAt / updatedAt are when the profile was created and last changed; the app\'s Theme Profiles list is sorted by updatedAt. They must be ISO8601 without milliseconds, e.g. 2026-10-04T08:00:00Z — a value like 2026-10-04T08:00:00.123Z makes the whole file fail to import.',
      ],
    },
    'Tab bar: 10 square PNGs, 81×81 with transparency recommended, named <slot>.png and <slot>_selected.png for app, sign, cert, file, setting. All five slots are required.',
    'Zip manifest.json and assets together with manifest.json at the top level, then change the extension to .hmtheme. On macOS:',
    'The file must stay under 30MB.',
  ],
  manifestTitle: 'manifest.json example',
  infoTitle: 'assets/tabIcon/pack/info.json example',
}

export function themeGuide(locale) {
  return String(locale).startsWith('zh') ? zh : en
}
