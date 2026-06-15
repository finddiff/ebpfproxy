<script setup>
import { ref } from 'vue'

const props = defineProps({
  schema: { type: Object, required: true },
  modelValue: { type: Object, required: true }
})

const emit = defineEmits(['update:modelValue'])

const showAdvanced = ref(false)
const activeSub = ref('basic')

function updateField(name, value) {
  emit('update:modelValue', { ...props.modelValue, [name]: value })
}

function getVal(name) {
  return props.modelValue[name] ?? ''
}

function addMapItem(listName) {
  const list = [...(props.modelValue[listName] || [])]
  list.push({ key: '', value: '' })
  emit('update:modelValue', { ...props.modelValue, [listName]: list })
}

function updateMapItem(listName, index, field, value) {
  const list = [...(props.modelValue[listName] || [])]
  list[index] = { ...list[index], [field]: value }
  emit('update:modelValue', { ...props.modelValue, [listName]: list })
}

function removeMapItem(listName, index) {
  const list = props.modelValue[listName] ? [...props.modelValue[listName]] : []
  list.splice(index, 1)
  emit('update:modelValue', { ...props.modelValue, [listName]: list })
}

// DNS routing helpers
function addRequestRule() {
  const rules = [...props.modelValue.request_rules]
  rules.push({ rule: '', upstream: '' })
  emit('update:modelValue', { ...props.modelValue, request_rules: rules })
}

function updateRequestRule(index, field, value) {
  const rules = [...props.modelValue.request_rules]
  rules[index] = { ...rules[index], [field]: value }
  emit('update:modelValue', { ...props.modelValue, request_rules: rules })
}

function removeRequestRule(index) {
  const rules = [...props.modelValue.request_rules]
  rules.splice(index, 1)
  emit('update:modelValue', { ...props.modelValue, request_rules: rules })
}

function addResponseRule() {
  const rules = [...props.modelValue.response_rules]
  rules.push({ rule: '', action: '' })
  emit('update:modelValue', { ...props.modelValue, response_rules: rules })
}

function updateResponseRule(index, field, value) {
  const rules = [...props.modelValue.response_rules]
  rules[index] = { ...rules[index], [field]: value }
  emit('update:modelValue', { ...props.modelValue, response_rules: rules })
}

function removeResponseRule(index) {
  const rules = [...props.modelValue.response_rules]
  rules.splice(index, 1)
  emit('update:modelValue', { ...props.modelValue, response_rules: rules })
}
</script>

<template>
  <div class="dns-section">
    <div class="sub-tabs">
      <button :class="['sub-tab', { active: activeSub === 'basic' }]" @click="activeSub = 'basic'">Basic</button>
      <button :class="['sub-tab', { active: activeSub === 'upstream' }]" @click="activeSub = 'upstream'">Upstream ({{ modelValue.upstream.length }})</button>
      <button :class="['sub-tab', { active: activeSub === 'hosts' }]" @click="activeSub = 'hosts'">Hosts ({{ modelValue.hosts.length }})</button>
      <button :class="['sub-tab', { active: activeSub === 'routing' }]" @click="activeSub = 'routing'">Routing</button>
    </div>

    <!-- Basic -->
    <div v-if="activeSub === 'basic'" class="sub-content">
      <div class="field-row" v-for="field in schema.fields" :key="field.name">
        <label class="field-label">{{ field.label }}</label>
        <input v-if="field.type === 'boolean'" type="checkbox" class="toggle"
          :checked="getVal(field.name) === true"
          @change="updateField(field.name, $event.target.checked)" />
        <select v-else-if="field.type === 'enum'" class="input" :value="getVal(field.name)"
          @change="updateField(field.name, $event.target.value)">
          <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
        </select>
        <input v-else class="input" type="text"
          :value="getVal(field.name)"
          :placeholder="String(field.default ?? '')"
          @input="updateField(field.name, $event.target.value === '' ? undefined : $event.target.value)" />
        <span class="field-desc">{{ field.desc }}</span>
      </div>
    </div>

    <!-- Upstream -->
    <div v-if="activeSub === 'upstream'" class="sub-content">
      <div class="list-header">
        <span>DNS Upstream Servers</span>
        <button class="btn-sm" @click="addMapItem('upstream')">+ Add</button>
      </div>
      <div v-if="modelValue.upstream.length === 0" class="empty">No upstream servers defined</div>
      <div v-for="(item, i) in modelValue.upstream" :key="i" class="map-row">
        <input class="input input-small" placeholder="Name" :value="item.key"
          @input="updateMapItem('upstream', i, 'key', $event.target.value)" />
        <input class="input input-medium" placeholder="tcp+udp://dns.google:53" :value="item.value"
          @input="updateMapItem('upstream', i, 'value', $event.target.value)" />
        <button class="btn-remove" @click="removeMapItem('upstream', i)">✕</button>
      </div>
    </div>

    <!-- Hosts -->
    <div v-if="activeSub === 'hosts'" class="sub-content">
      <div class="list-header">
        <span>DNS Hosts (domain → IP)</span>
        <button class="btn-sm" @click="addMapItem('hosts')">+ Add</button>
      </div>
      <div v-if="modelValue.hosts.length === 0" class="empty">No hosts entries</div>
      <div v-for="(item, i) in modelValue.hosts" :key="i" class="map-row">
        <input class="input input-small" placeholder="domain.com" :value="item.key"
          @input="updateMapItem('hosts', i, 'key', $event.target.value)" />
        <span class="arrow">→</span>
        <input class="input input-small" placeholder="192.168.1.1" :value="item.value"
          @input="updateMapItem('hosts', i, 'value', $event.target.value)" />
        <button class="btn-remove" @click="removeMapItem('hosts', i)">✕</button>
      </div>
    </div>

    <!-- DNS Routing -->
    <div v-if="activeSub === 'routing'" class="sub-content">
      <div class="dns-routing-section">
        <h4>Request Routing</h4>
        <div class="list-header">
          <span>Rules (qname(...) → upstream)</span>
          <button class="btn-sm" @click="addRequestRule()">+ Add</button>
        </div>
        <div v-if="modelValue.request_rules.length === 0" class="empty">No request rules</div>
        <div v-for="(rule, i) in modelValue.request_rules" :key="'req' + i" class="map-row">
          <input class="input input-medium" placeholder="qname(geosite:cn)" :value="rule.rule"
            @input="updateRequestRule(i, 'rule', $event.target.value)" />
          <span class="arrow">→</span>
          <input class="input input-small" placeholder="alinodns" :value="rule.upstream"
            @input="updateRequestRule(i, 'upstream', $event.target.value)" />
          <button class="btn-remove" @click="removeRequestRule(i)">✕</button>
        </div>
        <div class="field-row" style="margin-top: 8px">
          <label class="field-label">Fallback</label>
          <input class="input input-small" :value="modelValue.request_fallback || 'asis'"
            @input="updateField('request_fallback', $event.target.value)" />
        </div>
      </div>

      <div class="dns-routing-section" style="margin-top: 16px">
        <h4>Response Routing</h4>
        <div class="list-header">
          <span>Rules (ip(...) / upstream(...) → accept/reject)</span>
          <button class="btn-sm" @click="addResponseRule()">+ Add</button>
        </div>
        <div v-if="modelValue.response_rules.length === 0" class="empty">No response rules</div>
        <div v-for="(rule, i) in modelValue.response_rules" :key="'resp' + i" class="map-row">
          <input class="input input-medium" placeholder="upstream(google)" :value="rule.rule"
            @input="updateResponseRule(i, 'rule', $event.target.value)" />
          <span class="arrow">→</span>
          <input class="input input-small" placeholder="accept" :value="rule.action"
            @input="updateResponseRule(i, 'action', $event.target.value)" />
          <button class="btn-remove" @click="removeResponseRule(i)">✕</button>
        </div>
        <div class="field-row" style="margin-top: 8px">
          <label class="field-label">Fallback</label>
          <input class="input input-small" :value="modelValue.response_fallback || 'accept'"
            @input="updateField('response_fallback', $event.target.value)" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dns-section { display: flex; flex-direction: column; gap: 12px; }

.sub-tabs { display: flex; gap: 2px; border-bottom: 1px solid #30363d; }
.sub-tab {
  padding: 6px 14px;
  background: transparent;
  color: #8b949e;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}
.sub-tab:hover { color: #c9d1d9; }
.sub-tab.active { color: #58a6ff; border-bottom-color: #58a6ff; }

.sub-content { padding-top: 8px; }

.field-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 0;
  border-bottom: 1px solid #21262d;
  flex-wrap: wrap;
}
.field-label { font-size: 12px; font-weight: 600; color: #8b949e; min-width: 140px; }
.field-desc { font-size: 11px; color: #484f58; width: 100%; margin-left: 152px; }

.input {
  padding: 5px 10px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  flex: 1;
  max-width: 400px;
}
.input:focus { border-color: #1f6feb; }
.input-small { max-width: 160px; }
.input-medium { max-width: 280px; }
select.input { cursor: pointer; }

.toggle { accent-color: #1f6feb; }

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  color: #8b949e;
  font-size: 12px;
}

.btn-sm {
  padding: 3px 12px;
  background: #21262d;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}
.btn-sm:hover { background: #30363d; }

.btn-remove {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 6px;
}
.btn-remove:hover { color: #ff7b72; }

.map-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}
.arrow { color: #484f58; font-size: 13px; }

.empty { color: #484f58; font-size: 13px; padding: 12px 0; }
.dns-routing-section h4 { color: #c9d1d9; font-size: 14px; margin: 0 0 8px 0; }
</style>
