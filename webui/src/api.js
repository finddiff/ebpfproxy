let _token = () => ''
let _onAuthFailed = null

export function initApi(getToken, onFailed) {
  _token = getToken
  _onAuthFailed = onFailed
}

export async function apiGet(url) {
  const headers = {}
  const t = _token()
  if (t) headers['X-Token'] = t
  const res = await fetch(url, { headers })
  if (res.status === 401) {
    if (_onAuthFailed) _onAuthFailed()
    throw new Error('auth_required')
  }
  return res.json()
}

export async function apiPut(url, body) {
  const headers = { 'Content-Type': 'application/json' }
  const t = _token()
  if (t) headers['X-Token'] = t
  const res = await fetch(url, { method: 'PUT', headers, body: JSON.stringify(body) })
  if (res.status === 401) {
    if (_onAuthFailed) _onAuthFailed()
    throw new Error('auth_required')
  }
  return res.json()
}

export async function apiPost(url, body) {
  const headers = { 'Content-Type': 'application/json' }
  const t = _token()
  if (t) headers['X-Token'] = t
  const res = await fetch(url, { method: 'POST', headers, body: JSON.stringify(body) })
  if (res.status === 401) {
    if (_onAuthFailed) _onAuthFailed()
    throw new Error('auth_required')
  }
  return res.json()
}
