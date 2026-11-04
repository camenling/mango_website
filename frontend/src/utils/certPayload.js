import { unzipSync } from 'fflate'

const MAX_ZIP_ENTRIES = 200

export function arrayBufferToBase64(buffer) {
  const bytes = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer)
  let binary = ''
  const chunkSize = 0x8000
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize))
  }
  return btoa(binary)
}

export function uint8ArrayToBase64(bytes) {
  return arrayBufferToBase64(bytes)
}

export function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(arrayBufferToBase64(reader.result))
    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsArrayBuffer(file)
  })
}

function normalizeZipEntryName(name) {
  const clean = name.replace(/\\/g, '/').replace(/^\/+/, '')
  if (clean.includes('..') || clean.startsWith('/')) {
    return null
  }
  return clean
}

function pickCertEntries(unzipped) {
  let p12Bytes = null
  let p12Name = ''
  let mpBytes = null
  let mpName = ''

  const paths = Object.keys(unzipped)
  if (paths.length > MAX_ZIP_ENTRIES) {
    throw new Error(`压缩包内文件过多（超过 ${MAX_ZIP_ENTRIES} 个）`)
  }

  for (const rawPath of paths) {
    const path = normalizeZipEntryName(rawPath)
    if (!path) continue
    const data = unzipped[rawPath]
    if (!data || data.length === 0) continue

    const base = path.split('/').pop()?.toLowerCase() ?? ''
    if (base.endsWith('.p12') || base.endsWith('.pfx')) {
      if (!p12Bytes || base.endsWith('.p12')) {
        p12Bytes = data
        p12Name = path.split('/').pop() ?? base
      }
    } else if (base.endsWith('.mobileprovision')) {
      if (!mpBytes) {
        mpBytes = data
        mpName = path.split('/').pop() ?? base
      }
    }
  }

  if (!p12Bytes) {
    throw new Error('压缩包中未找到 .p12 或 .pfx 证书文件')
  }
  if (!mpBytes) {
    throw new Error('压缩包中未找到 .mobileprovision 描述文件')
  }

  return {
    p12Base64: uint8ArrayToBase64(p12Bytes),
    mpBase64: uint8ArrayToBase64(mpBytes),
    p12Name,
    provisionName: mpName,
  }
}

export async function extractCertBase64FromZip(file) {
  const buffer = await file.arrayBuffer()
  let unzipped
  try {
    unzipped = unzipSync(new Uint8Array(buffer))
  } catch (e) {
    const detail = e instanceof Error ? e.message : String(e)
    throw new Error(`无法解压 ZIP 文件，请确认文件未损坏（${detail}）`)
  }
  return pickCertEntries(unzipped)
}

/** @returns {Promise<{ p12: string, mp: string, p12Name?: string, provisionName?: string }>} */
export async function buildCertUploadPayload({ mode, p12File, provisionFile, zipFile }) {
  if (mode === 'zip') {
    if (!zipFile) throw new Error('请上传 ZIP 压缩包')
    const extracted = await extractCertBase64FromZip(zipFile)
    return {
      p12: extracted.p12Base64,
      mp: extracted.mpBase64,
      p12Name: extracted.p12Name,
      provisionName: extracted.provisionName,
    }
  }

  if (!p12File) throw new Error('请上传 P12 证书')
  if (!provisionFile) throw new Error('请上传描述文件')

  const [p12, mp] = await Promise.all([
    fileToBase64(p12File),
    fileToBase64(provisionFile),
  ])

  return {
    p12,
    mp,
    p12Name: p12File.name,
    provisionName: provisionFile.name,
  }
}
