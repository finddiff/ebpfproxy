<script setup>
import { ref, onMounted } from 'vue'
import { apiGet, apiPut } from '../api.js'

const config = ref('')
const loading = ref(true)
const saved = ref(false)

onMounted(async () => {
  try {
    const json = await apiGet('/api/config')
    config.value = json.config || ''
  } catch (e) { /* ignore */ }
  loading.value = false
})

async function saveConfig() {
  try {
    await apiPut('/api/config', { config: config.value })
    saved.value = true
    setTimeout(() => saved.value = false, 2000)
  } catch (e) {
    alert('Failed to save config')
  }
}
</script>

<template>
  <div class="config">
    <div class="config-header">
      <h2>Configuration</h2>
      <button class="save-btn" @click="saveConfig" :disabled="loading">
        {{ saved ? 'Saved!' : 'Save' }}
      </button>
    </div>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else>
      <textarea
        v-model="config"
        class="editor"
        spellcheck="false"
      ></textarea>
    </div>
  </div>
</template>

<style scoped>
.config h2 { margin: 0; color: #c9d1d9; }
.config-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.save-btn {
  padding: 8px 24px;
  background: #238636;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: background 0.15s;
}
.save-btn:hover { background: #2ea043; }
.save-btn:disabled { background: #30363d; color: #484f58; cursor: not-allowed; }
.loading { color: #8b949e; padding: 20px; text-align: center; }
.editor {
  width: 100%;
  height: 550px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 16px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
}
.editor:focus { border-color: #1f6feb; }
</style>
