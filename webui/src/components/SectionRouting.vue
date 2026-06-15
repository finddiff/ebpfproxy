<script setup>
import { computed, reactive } from 'vue'

const props = defineProps({
  schema: { type: Object, required: true },
  modelValue: { type: Object, required: true },
  outboundNames: { type: Array, required: false, default: () => [] }
})

const emit = defineEmits(['update:modelValue'])

const MAX_VISIBLE_FUNCS = 4
const MAX_VISIBLE_PARAMS = 3

const allOutbounds = computed(() => {
  return [...(props.schema.builtinOutbounds || []).map(o => o.value), ...(props.outboundNames || [])]
})

function findFuncDef(name) {
  return (props.schema.functions || []).find(f => f.name === name) || null
}

function insertRule(atIndex) {
  const rules = [...props.modelValue.rules]
  rules.splice(atIndex, 0, {
    functions: [{ name: 'domain', params: [{ key: '', value: '' }], not: false }],
    outbound: 'direct'
  })
  emit('update:modelValue', { ...props.modelValue, rules })
}

function removeRule(index) {
  const rules = [...props.modelValue.rules]
  rules.splice(index, 1)
  emit('update:modelValue', { ...props.modelValue, rules })
}

function moveRule(index, dir) {
  const rules = [...props.modelValue.rules]
  const newIdx = index + dir
  if (newIdx < 0 || newIdx >= rules.length) return
  const temp = rules[index]
  rules[index] = rules[newIdx]
  rules[newIdx] = temp
  emit('update:modelValue', { ...props.modelValue, rules })
}

function addFunction(ruleIndex) {
  const rules = [...props.modelValue.rules]
  rules[ruleIndex] = {
    ...rules[ruleIndex],
    functions: [...rules[ruleIndex].functions, { name: 'domain', params: [{ key: '', value: '' }], not: false }]
  }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function removeFunction(ruleIndex, funcIndex) {
  const rules = [...props.modelValue.rules]
  rules[ruleIndex] = {
    ...rules[ruleIndex],
    functions: rules[ruleIndex].functions.filter((_, i) => i !== funcIndex)
  }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function updateFunction(ruleIndex, funcIndex, field, value) {
  const rules = [...props.modelValue.rules]
  const funcs = [...rules[ruleIndex].functions]
  if (field === 'name') {
    const def = findFuncDef(value)
    funcs[funcIndex] = { ...funcs[funcIndex], name: value, params: [{ key: '', value: '' }] }
    if (def && def.params) {
      funcs[funcIndex].params = def.params.map(p => ({ key: p.key || '', value: '' }))
    }
  } else {
    funcs[funcIndex] = { ...funcs[funcIndex], [field]: value }
  }
  rules[ruleIndex] = { ...rules[ruleIndex], functions: funcs }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function updateFuncParamKey(ruleIndex, funcIndex, paramIndex, newKey) {
  const rules = [...props.modelValue.rules]
  const funcs = [...rules[ruleIndex].functions]
  const params = [...funcs[funcIndex].params]
  params[paramIndex] = { ...params[paramIndex], key: newKey }
  funcs[funcIndex] = { ...funcs[funcIndex], params }
  rules[ruleIndex] = { ...rules[ruleIndex], functions: funcs }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function updateFuncParamValue(ruleIndex, funcIndex, paramIndex, value) {
  const rules = [...props.modelValue.rules]
  const funcs = [...rules[ruleIndex].functions]
  const params = [...funcs[funcIndex].params]
  params[paramIndex] = { ...params[paramIndex], value }
  funcs[funcIndex] = { ...funcs[funcIndex], params }
  rules[ruleIndex] = { ...rules[ruleIndex], functions: funcs }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function addFuncParam(ruleIndex, funcIndex) {
  const rules = [...props.modelValue.rules]
  const funcs = [...rules[ruleIndex].functions]
  const params = [...funcs[funcIndex].params]
  params.push({ key: '', value: '' })
  funcs[funcIndex] = { ...funcs[funcIndex], params }
  rules[ruleIndex] = { ...rules[ruleIndex], functions: funcs }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function removeFuncParam(ruleIndex, funcIndex, paramIndex) {
  const rules = [...props.modelValue.rules]
  const funcs = [...rules[ruleIndex].functions]
  const params = [...funcs[funcIndex].params]
  params.splice(paramIndex, 1)
  funcs[funcIndex] = { ...funcs[funcIndex], params }
  rules[ruleIndex] = { ...rules[ruleIndex], functions: funcs }
  emit('update:modelValue', { ...props.modelValue, rules })
}

const expandedParams = reactive({})
function isParamsExpanded(ruleIndex, funcIndex) {
  return !!expandedParams[`${ruleIndex}-${funcIndex}`]
}
function toggleParams(ruleIndex, funcIndex) {
  const key = `${ruleIndex}-${funcIndex}`
  expandedParams[key] = !expandedParams[key]
}

function updateOutbound(ruleIndex, value) {
  const rules = [...props.modelValue.rules]
  rules[ruleIndex] = { ...rules[ruleIndex], outbound: value }
  emit('update:modelValue', { ...props.modelValue, rules })
}

function updateFallback(value) {
  emit('update:modelValue', { ...props.modelValue, fallback: value })
}

function getParamPlaceholder(funcDef, param) {
  if (!funcDef) return 'value'
  const defParam = (funcDef.params || [])[0]
  return (defParam && defParam.placeholder) || 'value'
}

function funcSummary(func) {
  const def = findFuncDef(func.name)
  const label = def ? def.label : func.name
  const params = func.params.filter(p => p.value).map(p => p.key ? `${p.key}:${p.value}` : p.value).join(', ')
  return (func.not ? '!' : '') + label + (params ? `(${params})` : '()')
}

function ruleSummary(rule) {
  return rule.functions.map(f => funcSummary(f)).join(' && ') + ' → ' + rule.outbound
}
</script>

<template>
  <div class="routing-section">
    <div class="list-header">
      <span>Routing Rules ({{ modelValue.rules.length }})</span>
      <div class="header-btns">
        <span class="hint">rules are evaluated top → bottom, first match wins</span>
        <button class="btn-add" @click="insertRule(modelValue.rules.length)">+ Add Rule</button>
      </div>
    </div>

    <div v-if="modelValue.rules.length === 0" class="empty">No routing rules. Add a rule to match traffic.</div>

    <div v-for="(rule, ri) in modelValue.rules" :key="ri" class="rule-card">
      <div class="rule-header">
        <div class="rule-header-left">
          <span class="rule-num">#{{ ri + 1 }}</span>
          <button class="btn-order" :disabled="ri === 0" @click="moveRule(ri, -1)" title="Move up">▲</button>
          <button class="btn-order" :disabled="ri === modelValue.rules.length - 1" @click="moveRule(ri, 1)" title="Move down">▼</button>
          <button class="btn-insert" @click="insertRule(ri + 1)" title="Insert rule after">＋</button>
        </div>
        <div class="rule-header-right">
          <span class="rule-summary">{{ ruleSummary(rule) }}</span>
          <button class="btn-remove" @click="removeRule(ri)">✕</button>
        </div>
      </div>

      <div class="rule-body">
        <!-- Functions chain -->
        <div class="funcs-chain">
          <template v-for="(func, fi) in rule.functions" :key="fi">
            <template v-if="fi < MAX_VISIBLE_FUNCS || rule.functions.length <= MAX_VISIBLE_FUNCS + 1">
              <div class="func-group">
                <template v-if="fi > 0">
                  <div class="logic-badge">&&</div>
                </template>
                <div class="func-box">
                  <div class="func-head">
                    <label class="not-toggle" title="Negate this condition">
                      <input type="checkbox" :checked="func.not" @change="updateFunction(ri, fi, 'not', $event.target.checked)" />
                      <span class="not-label">!</span>
                    </label>
                    <select class="input func-select" :value="func.name"
                      @change="updateFunction(ri, fi, 'name', $event.target.value)">
                      <option v-for="f in schema.functions" :key="f.name" :value="f.name">{{ f.label || f.name }}</option>
                    </select>
                    <button class="btn-remove-func" @click="removeFunction(ri, fi)" title="Remove condition">✕</button>
                  </div>
                  <div class="func-params">
                    <template v-for="(param, pi) in func.params" :key="pi">
                      <div v-if="pi < MAX_VISIBLE_PARAMS || isParamsExpanded(ri, fi) || func.params.length <= MAX_VISIBLE_PARAMS" class="param-line">
                        <!-- Key dropdown (if function has paramKeys) -->
                        <select v-if="findFuncDef(func.name)?.paramKeys" class="input param-key-select"
                          :value="param.key"
                          @change="updateFuncParamKey(ri, fi, pi, $event.target.value)">
                          <option v-for="pk in findFuncDef(func.name).paramKeys" :key="pk.value" :value="pk.value">
                            {{ pk.label }}
                          </option>
                        </select>
                        <!-- Value: dropdown for enum, input for text -->
                        <select v-if="findFuncDef(func.name)?.params[pi]?.options" class="input param-val-select"
                          :value="param.value"
                          @change="updateFuncParamValue(ri, fi, pi, $event.target.value)">
                          <option value=""></option>
                          <option v-for="opt in findFuncDef(func.name).params[pi]?.options" :key="opt" :value="opt">{{ opt }}</option>
                        </select>
                        <input v-else class="input param-val-input" type="text"
                          :value="param.value"
                          :placeholder="getParamPlaceholder(findFuncDef(func.name), param)"
                          @input="updateFuncParamValue(ri, fi, pi, $event.target.value)" />
                        <button v-if="func.params.length > 1" class="btn-remove-param" @click="removeFuncParam(ri, fi, pi)" title="Remove param">✕</button>
                      </div>
                    </template>
                    <div v-if="func.params.length > MAX_VISIBLE_PARAMS && !isParamsExpanded(ri, fi)" class="param-overflow" @click="toggleParams(ri, fi)">
                      … +{{ func.params.length - MAX_VISIBLE_PARAMS }} more params
                    </div>
                    <div v-if="func.params.length > MAX_VISIBLE_PARAMS && isParamsExpanded(ri, fi)" class="param-overflow" @click="toggleParams(ri, fi)">
                      … show less
                    </div>
                    <button class="btn-add-param" @click="addFuncParam(ri, fi)" title="Add another param">+ param</button>
                  </div>
                </div>
              </div>
            </template>
            <template v-else-if="fi === MAX_VISIBLE_FUNCS">
              <span class="overflow-hint">… +{{ rule.functions.length - MAX_VISIBLE_FUNCS }} more</span>
            </template>
          </template>
          <button class="btn-and-add" @click="addFunction(ri)" title="Add AND condition">+ AND</button>
        </div>

        <!-- Outbound -->
        <div class="outbound-row">
          <span class="arrow">→</span>
          <select class="input outbound-select" :value="rule.outbound" @change="updateOutbound(ri, $event.target.value)">
            <option v-for="ob in allOutbounds" :key="ob" :value="ob">{{ ob }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Fallback -->
    <div class="fallback-row">
      <label class="field-label">Fallback</label>
      <span class="fallback-hint">(default when no rule matches)</span>
      <select class="input fallback-select" :value="modelValue.fallback"
        @change="updateFallback($event.target.value)">
        <option v-for="ob in allOutbounds" :key="ob" :value="ob">{{ ob }}</option>
      </select>
    </div>
  </div>
</template>

<style scoped>
.routing-section { display: flex; flex-direction: column; gap: 8px; }

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.list-header span { color: #8b949e; font-size: 13px; }

.header-btns { display: flex; align-items: center; gap: 12px; }
.hint { font-size: 11px; color: #484f58; }

.btn-add {
  padding: 5px 14px;
  background: #21262d;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}
.btn-add:hover { background: #30363d; }

.empty { color: #484f58; font-size: 13px; padding: 16px 0; }

.rule-card {
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #0d1117;
  overflow: hidden;
}
.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 10px;
  background: #161b22;
  gap: 8px;
}
.rule-header-left { display: flex; align-items: center; gap: 4px; }
.rule-header-right { display: flex; align-items: center; gap: 8px; overflow: hidden; }
.rule-num { font-size: 12px; color: #8b949e; font-weight: 600; white-space: nowrap; }
.rule-summary {
  font-size: 11px;
  color: #484f58;
  font-family: 'SF Mono', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 350px;
}

.btn-order {
  background: transparent;
  border: 1px solid #30363d;
  color: #8b949e;
  cursor: pointer;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  line-height: 1;
}
.btn-order:hover:not(:disabled) { color: #c9d1d9; background: #21262d; }
.btn-order:disabled { opacity: 0.3; cursor: default; }

.btn-insert {
  background: transparent;
  border: 1px solid #30363d;
  color: #58a6ff;
  cursor: pointer;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  line-height: 1;
}
.btn-insert:hover { background: #1f6feb20; }

.btn-remove {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 6px;
  flex-shrink: 0;
}
.btn-remove:hover { color: #ff7b72; }

.btn-remove-func {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 12px;
  padding: 2px 4px;
  flex-shrink: 0;
  line-height: 1;
  margin-left: auto;
}
.btn-remove-func:hover { color: #ff7b72; }

.btn-remove-param {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 10px;
  padding: 1px 3px;
  flex-shrink: 0;
  line-height: 1;
}
.btn-remove-param:hover { color: #ff7b72; }

.btn-add-param {
  background: none;
  border: 1px dashed #30363d;
  color: #58a6ff;
  cursor: pointer;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  margin-top: 2px;
}
.btn-add-param:hover { background: #1f6feb10; }

.rule-body { padding: 8px 10px; }

.funcs-chain {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
}

.func-group { display: flex; align-items: center; gap: 4px; }

.func-box {
  border: 1px solid #30363d;
  border-radius: 4px;
  background: #161b22;
}
.func-head {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 3px 4px;
  border-bottom: 1px solid #21262d;
}
.func-params {
  padding: 2px 4px 4px;
}
.param-line {
  display: flex;
  align-items: center;
  gap: 3px;
  margin-top: 2px;
}

.param-overflow {
  font-size: 10px;
  color: #58a6ff;
  padding: 1px 4px;
  margin-top: 2px;
  cursor: pointer;
  user-select: none;
}
.param-overflow:hover { text-decoration: underline; }

.logic-badge {
  padding: 3px 5px;
  background: #23863620;
  color: #3fb950;
  border: 1px solid #23863640;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 700;
  font-family: monospace;
}

.not-toggle { display: flex; align-items: center; cursor: pointer; }
.not-toggle input { accent-color: #f85149; margin: 0; }
.not-label { font-size: 10px; font-weight: 700; color: #f85149; padding: 0 1px; }

.func-select { width: 95px; }
.param-key-select { width: 70px; }
.param-val-select { width: 70px; }
.param-val-input { width: 130px; }

.input {
  padding: 3px 5px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 3px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
}
.input:focus { border-color: #1f6feb; }
select.input { cursor: pointer; }

.overflow-hint {
  font-size: 11px;
  color: #8b949e;
  padding: 3px 6px;
  background: #21262d;
  border-radius: 3px;
}

.btn-and-add {
  background: none;
  border: 1px dashed #30363d;
  color: #58a6ff;
  cursor: pointer;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 3px;
}
.btn-and-add:hover { background: #1f6feb10; }

.outbound-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding-top: 6px;
  border-top: 1px solid #21262d;
}
.arrow { color: #3fb950; font-size: 14px; font-weight: 700; }
.outbound-select { min-width: 140px; }

.fallback-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 6px;
}
.field-label { font-size: 12px; color: #8b949e; font-weight: 600; }
.fallback-hint { font-size: 11px; color: #484f58; }
.fallback-select { min-width: 140px; }
</style>
