<script setup>
import { ref, onMounted, provide } from 'vue'
import { useRoute } from 'vue-router'
import { useWebSocket } from './composables/useWebSocket'
import { useAuth } from './composables/useAuth'
import { initApi } from './api.js'

const tabs = [
  { id: 'overview',    label: 'Overview',    to: '/overview' },
  { id: 'dhcp',        label: 'DHCP',        to: '/dhcp' },
  { id: 'sensors',     label: 'Sensors',     to: '/sensors' },
  { id: 'proxy',       label: 'Proxy',       to: '/proxy' },
  { id: 'rules',       label: 'Rules',       to: '/rules' },
  { id: 'dns',         label: 'DNS',         to: '/dns' },
  { id: 'connections', label: 'Connections', to: '/connections' },
  { id: 'config',      label: 'Config',      to: '/config' },
  { id: 'logs',        label: 'Logs',        to: '/logs' },
]

const route = useRoute()
const auth = useAuth()
const { connected, overviewData, logEntry } = useWebSocket(auth.wsUrl)

const tokenInput = ref('')
const tokenError = ref(false)

initApi(() => auth.getToken(), () => auth.clearToken())

onMounted(() => {
  auth.registerAuthFailedCallback(() => { auth.clearToken() })
})

provide('overviewData', overviewData)
provide('logEntry', logEntry)

function doLogin() {
  const t = tokenInput.value.trim()
  if (!t) return
  fetch('/api/token', { headers: { 'X-Token': t } })
    .then(r => r.json())
    .then(data => {
      if (data.valid) { auth.setToken(t); tokenError.value = false }
      else { tokenError.value = true }
    })
    .catch(() => { tokenError.value = true })
}

function doLogout() { auth.clearToken() }
</script>

<template>
  <div v-if="auth.tokenChecked.value && auth.tokenRequired.value && !auth.tokenValid.value" class="login-overlay">
    <div class="login-box">
      <h2>Dae WebUI</h2>
      <p class="login-desc">Authentication required</p>
      <input v-model="tokenInput" class="login-input" type="password" placeholder="Enter access token" @keyup.enter="doLogin" autofocus />
      <button class="login-btn" @click="doLogin">Login</button>
      <p v-if="tokenError" class="login-error">Invalid token</p>
    </div>
  </div>

  <div v-else class="app">
    <header class="header">
      <h1 class="title">Dae WebUI</h1>
      <div class="header-right">
        <span v-if="auth.token.value" class="token-badge">🔒<button class="logout-btn" @click="doLogout">Clear</button></span>
        <span class="status" :class="{ online: connected, offline: !connected }">
          <span class="dot"></span>{{ connected ? 'Connected' : 'Disconnected' }}
        </span>
      </div>
    </header>
    <nav class="tabs">
      <router-link v-for="tab in tabs" :key="tab.id" :to="tab.to" class="tab" active-class="active">
        {{ tab.label }}
      </router-link>
    </nav>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.login-overlay { position: fixed; inset: 0; background: #0d1117; display: flex; align-items: center; justify-content: center; z-index: 1000; }
.login-box { background: #161b22; border: 1px solid #30363d; border-radius: 12px; padding: 40px; text-align: center; min-width: 340px; }
.login-box h2 { font-size: 24px; color: #58a6ff; margin: 0 0 8px 0; }
.login-desc { color: #8b949e; margin-bottom: 20px; font-size: 14px; }
.login-input { width: 100%; padding: 10px 14px; background: #0d1117; color: #c9d1d9; border: 1px solid #30363d; border-radius: 8px; font-size: 15px; outline: none; box-sizing: border-box; }
.login-input:focus { border-color: #1f6feb; }
.login-btn { margin-top: 12px; width: 100%; padding: 10px; background: #238636; color: #fff; border: none; border-radius: 8px; font-size: 15px; font-weight: 600; cursor: pointer; }
.login-btn:hover { background: #2ea043; }
.login-error { color: #f85149; margin-top: 10px; font-size: 13px; }
.app { max-width: 1400px; margin: 0 auto; padding: 16px; }
.header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid #30363d; }
.header-right { display: flex; align-items: center; gap: 12px; }
.title { font-size: 24px; font-weight: 600; color: #58a6ff; margin: 0; }
.token-badge { display: flex; align-items: center; gap: 6px; font-size: 12px; color: #3fb950; }
.logout-btn { padding: 2px 8px; background: #21262d; color: #8b949e; border: 1px solid #30363d; border-radius: 4px; cursor: pointer; font-size: 11px; }
.logout-btn:hover { background: #f8514920; color: #f85149; border-color: #f8514940; }
.status { display: flex; align-items: center; gap: 6px; font-size: 13px; padding: 4px 12px; border-radius: 20px; }
.status.online { background: rgba(63,185,80,0.15); color: #3fb950; }
.status.offline { background: rgba(248,81,73,0.15); color: #f85149; }
.dot { width: 8px; height: 8px; border-radius: 50%; }
.online .dot { background: #3fb950; }
.offline .dot { background: #f85149; }
.tabs { display: flex; gap: 2px; margin-bottom: 16px; flex-wrap: wrap; }
.tab { padding: 8px 18px; background: #21262d; border: 1px solid #30363d; border-radius: 6px 6px 0 0; color: #8b949e; cursor: pointer; font-size: 13px; transition: all 0.15s; border-bottom: none; text-decoration: none; }
.tab:hover { background: #30363d; color: #c9d1d9; }
.tab.active { background: #1f6feb; border-color: #1f6feb; color: #fff; }
.content { background: #161b22; border: 1px solid #30363d; border-radius: 0 8px 8px 8px; padding: 20px; min-height: 500px; }
</style>
