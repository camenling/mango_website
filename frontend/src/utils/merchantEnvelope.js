function bytesToB64(u8) {
  let s = ''
  for (let i = 0; i < u8.length; i += 1) s += String.fromCharCode(u8[i])
  return btoa(s)
}

function pemToBytes(pem) {
  const cleaned = String(pem || '')
    .replace(/-----BEGIN [A-Z ]+-----/g, '')
    .replace(/-----END [A-Z ]+-----/g, '')
    .replace(/\s+/g, '')
  const bin = atob(cleaned)
  const buf = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i += 1) buf[i] = bin.charCodeAt(i)
  return buf.buffer
}

function importPublicKey(pem) {
  return crypto.subtle.importKey(
    'spki',
    pemToBytes(pem),
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
}

export async function encryptMerchantEnvelope(kid, pubPem, plainObj) {
  const trimmedKid = String(kid || '').trim()
  if (!trimmedKid) throw new Error('bad_kid')
  const key = await importPublicKey(pubPem)
  const plain = new TextEncoder().encode(JSON.stringify(plainObj))
  const aesKeyBytes = crypto.getRandomValues(new Uint8Array(32))
  const nonce = crypto.getRandomValues(new Uint8Array(12))
  const aesKey = await crypto.subtle.importKey('raw', aesKeyBytes, 'AES-GCM', false, ['encrypt'])
  const aad = new TextEncoder().encode(`1|${trimmedKid}`)
  const sealed = new Uint8Array(
    await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv: nonce, additionalData: aad, tagLength: 128 },
      aesKey,
      plain,
    ),
  )
  const ciphertext = sealed.subarray(0, sealed.length - 16)
  const tag = sealed.subarray(sealed.length - 16)
  const wrapped = new Uint8Array(aesKeyBytes.length + nonce.length)
  wrapped.set(aesKeyBytes, 0)
  wrapped.set(nonce, aesKeyBytes.length)
  const encKey = new Uint8Array(await crypto.subtle.encrypt({ name: 'RSA-OAEP' }, key, wrapped))
  return {
    v: 1,
    kid: trimmedKid,
    enc_key: bytesToB64(encKey),
    ciphertext: bytesToB64(ciphertext),
    tag: bytesToB64(tag),
  }
}
