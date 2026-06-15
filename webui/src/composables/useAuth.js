import { ref, onMounted } from 'vue'

const token = ref(localStorage.getItem('dae_token') || '')
const tokenRequired = ref(false)
const tokenValid = ref(true)
const tokenChecked = ref(false)
let onAuthFailed = null

export function useAuth() {
  function registerAuthFailedCallback(fn) {
    onAuthFailed = fn
  }

  async function checkToken() {
    try {
      const res = await fetch('/api/token', {
        headers: token.value ? { 'X-Token': token.value } : {},
      })
      const data = await res.json()
      tokenRequired.value = data.configured
      tokenValid.value = data.valid
      tokenChecked.value = true
      if (!data.configured) {
        tokenValid.value = true
      }
    } catch (e) {
      tokenChecked.value = true
      tokenValid.value = true
    }
  }

  function setToken(t) {
    token.value = t
    localStorage.setItem('dae_token', t)
    checkToken()
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('dae_token')
    tokenChecked.value = false
    tokenValid.value = false
    checkToken()
  }

  function getToken() {
    return token.value
  }

  async function apiFetch(url, options = {}) {
    const headers = { ...(options.headers || {}) }
    if (token.value) {
      headers['X-Token'] = token.value
    }
    const res = await fetch(url, { ...options, headers })
    if (res.status === 401 && onAuthFailed) {
      onAuthFailed()
      throw new Error('auth_required')
    }
    return res
  }

  function wsUrl(path) {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const base = `${protocol}//${location.host}${path}`
    if (token.value) {
      return base + '?token=' + encodeURIComponent(token.value)
    }
    return base
  }

  onMounted(() => {
    checkToken()
  })

  return { token, tokenRequired, tokenValid, tokenChecked, setToken, clearToken, getToken, apiFetch, wsUrl, checkToken, registerAuthFailedCallback }
}
