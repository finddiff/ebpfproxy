<script setup>
import { ref, onMounted, computed } from 'vue'
import { apiGet, apiPut, apiPost } from '../api.js'
import { configSchema } from '../configSchema.js'
import { parseConfig } from '../configParser.js'
import { generateConfig } from '../configGenerator.js'
import SectionGlobal from './SectionGlobal.vue'
import SectionDns from './SectionDns.vue'
import SectionNodeList from './SectionNodeList.vue'
import SectionGroup from './SectionGroup.vue'
import SectionRouting from './SectionRouting.vue'

const tabs = ['global', 'dns', 'subscription', 'node', 'group', 'routing']
const activeTab = ref('global')
const state = ref(null)
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const configRaw = ref('')
const showPreview = ref(true)
const validateError = ref('')

onMounted(async () => {
  try {
    const json = await apiGet('/api/config')
    configRaw.value = json.config || ''
    if (configRaw.value) {
      state.value = parseConfig(configRaw.value)
    } else {
      state.value = getDefaultState()
    }
  } catch {
    state.value = getDefaultState()
  }
  loading.value = false
})

function getDefaultState() {
  return {
    global: {},
    dns: {
      upstream: [],
      hosts: [],
      fixed_domain_ttl: [],
      request_rules: [],
      response_rules: [],
      request_fallback: 'asis',
      response_fallback: 'accept',
    },
    subscription: [],
    node: [],
    group: [],
    routing: {
      rules: [],
      fallback: 'direct'
    }
  }
}

const previewText = computed(() => {
  if (!state.value) return ''
  try {
    return generateConfig(state.value)
  } catch {
    return '/* Error generating config preview */'
  }
})

const schema = computed(() => configSchema)

const outboundNames = computed(() => {
  const names = ['direct', 'block', 'must_direct']
  if (state.value?.group) {
    for (const g of state.value.group) {
      if (g.name) names.push(g.name)
    }
  }
  return names
})

function updateSection(name, value) {
  if (state.value) {
    state.value[name] = value
  }
}

async function doValidate() {
  validateError.value = ''
  const text = previewText.value
  if (!text.trim()) {
    validateError.value = 'Config is empty'
    return false
  }
  try {
    const json = await apiPost('/api/config/validate', { config: text })
    if (json.valid) {
      validateError.value = ''
      return true
    } else {
      validateError.value = json.error || 'Validation failed'
      return false
    }
  } catch (e) {
    validateError.value = 'Validation request failed: ' + e.message
    return false
  }
}

async function saveConfig() {
  saving.value = true
  saved.value = false
  validateError.value = ''

  try {
    // First validate
    const valid = await doValidate()
    if (!valid) {
      saving.value = false
      return
    }

    const text = previewText.value
    const res = await apiPut('/api/config', { config: text })
    configRaw.value = text
    saved.value = true
    setTimeout(() => saved.value = false, 3000)
    validateError.value = ''
    if (!res.reloaded) {
      validateError.value = 'Saved but reload may have failed. Check logs.'
    }
  } catch (e) {
    validateError.value = 'Save failed: ' + (e.message || 'Network error')
  }
  saving.value = false
}
</script>

<template>
  <div class="config-editor">
    <div class="header">
      <h2>Configuration</h2>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="doValidate" :disabled="saving || !state">
          Validate
        </button>
        <button class="btn btn-primary" @click="saveConfig" :disabled="saving || !state">
          {{ saving ? 'Saving...' : saved ? 'Saved!' : 'Save & Reload' }}
        </button>
      </div>
    </div>

    <div v-if="validateError" class="error-box">{{ validateError }}</div>

    <div v-if="loading" class="loading">Loading config...</div>

    <template v-else>
      <div class="section-tabs">
        <button v-for="tab in tabs" :key="tab"
          :class="['tab', { active: activeTab === tab }]"
          @click="activeTab = tab">
          {{ schema[tab]?.label || tab }}
        </button>
        <button class="tab preview-tab" :class="{ active: showPreview }" @click="showPreview = !showPreview">
          Preview
        </button>
      </div>

      <div class="section-content">
        <SectionGlobal
          v-if="activeTab === 'global' && schema.global"
          :schema="schema.global"
          :modelValue="state.global"
          @update:modelValue="v => updateSection('global', v)"
        />
        <SectionDns
          v-if="activeTab === 'dns'"
          :schema="schema.dns"
          :modelValue="state.dns"
          @update:modelValue="v => updateSection('dns', v)"
        />
        <SectionNodeList
          v-if="activeTab === 'subscription'"
          :schema="schema.subscription"
          :modelValue="state.subscription"
          @update:modelValue="v => updateSection('subscription', v)"
        />
        <SectionNodeList
          v-if="activeTab === 'node'"
          :schema="schema.node"
          :modelValue="state.node"
          @update:modelValue="v => updateSection('node', v)"
        />
        <SectionGroup
          v-if="activeTab === 'group'"
          :schema="schema.group"
          :modelValue="state.group"
          @update:modelValue="v => updateSection('group', v)"
          :outboundNames="outboundNames"
        />
        <SectionRouting
          v-if="activeTab === 'routing'"
          :schema="schema.routing"
          :modelValue="state.routing"
          @update:modelValue="v => updateSection('routing', v)"
          :outboundNames="outboundNames"
        />
      </div>

      <div v-if="showPreview" class="preview-section">
        <div class="preview-header">
          <h3>Generated Config Preview</h3>
        </div>
        <pre class="preview-text">{{ previewText }}</pre>
      </div>
    </template>
  </div>
</template>

<style scoped>
.config-editor { display: flex; flex-direction: column; gap: 12px; }
.header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.header h2 { margin: 0; color: #c9d1d9; }
.header-actions { display: flex; gap: 8px; }

.btn {
  padding: 8px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: background 0.15s;
}
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: #238636; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #2ea043; }
.btn-secondary { background: #21262d; color: #c9d1d9; border: 1px solid #30363d; }
.btn-secondary:hover:not(:disabled) { background: #30363d; }

.error-box {
  background: #f8514920;
  border: 1px solid #f85149;
  color: #f85149;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  white-space: pre-wrap;
}

.loading { color: #8b949e; padding: 40px; text-align: center; }

.section-tabs {
  display: flex;
  gap: 2px;
  flex-wrap: wrap;
  border-bottom: 1px solid #30363d;
  padding-bottom: 0;
}
.tab {
  padding: 8px 16px;
  background: transparent;
  color: #8b949e;
  border: 1px solid transparent;
  border-bottom: none;
  border-radius: 6px 6px 0 0;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}
.tab:hover { color: #c9d1d9; }
.tab.active {
  background: #1f6feb;
  color: #fff;
  border-color: #1f6feb;
}
.preview-tab { margin-left: auto; }

.section-content {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 6px;
  padding: 16px;
}

.preview-section {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 6px;
  overflow: hidden;
}
.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #161b22;
  border-bottom: 1px solid #30363d;
}
.preview-header h3 { margin: 0; font-size: 13px; color: #8b949e; }
.preview-text {
  margin: 0;
  padding: 12px 16px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #c9d1d9;
  max-height: 350px;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
