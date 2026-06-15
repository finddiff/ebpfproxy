<script setup>
import { ref, reactive } from 'vue'

const props = defineProps({
  schema: { type: Object, required: true },
  modelValue: { type: Array, required: true }
})

const emit = defineEmits(['update:modelValue'])

function addGroup() {
  const groups = [...props.modelValue]
  groups.push({
    name: '',
    policy: 'random',
    filters: [],
    tcp_check_url: '',
    tcp_check_http_method: '',
    udp_check_dns: '',
    check_interval: '',
    check_tolerance: '',
  })
  emit('update:modelValue', groups)
}

function updateGroup(index, field, value) {
  const groups = [...props.modelValue]
  groups[index] = { ...groups[index], [field]: value }
  emit('update:modelValue', groups)
}

function removeGroup(index) {
  const groups = [...props.modelValue]
  groups.splice(index, 1)
  emit('update:modelValue', groups)
}

// --- Filter structured editor ---

function addFilterLine(groupIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  filters.push('name()')
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

function removeFilterLine(groupIndex, filterIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  filters.splice(filterIndex, 1)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

// Parse filter string "name(keyword:hk) && !name(regex:HK)" into structured expressions
function parseFilterExpr(filterStr) {
  if (!filterStr) return [{ name: 'name', not: false, params: [{ key: '', value: '' }] }]
  const parts = filterStr.split('&&').map(s => s.trim())
  return parts.map(part => {
    let not = false
    let p = part
    if (p.startsWith('!')) { not = true; p = p.substring(1) }
    const m = p.match(/^(\w+)\((.*)\)/)
    if (!m) return { name: 'name', not, params: [{ key: '', value: '' }] }
    const funcName = m[1]
    const paramsStr = m[2]
    const params = []
    if (paramsStr) {
      const paramParts = splitParamParts(paramsStr)
      for (const pp of paramParts) {
        if (!pp) continue
        const kv = pp.match(/^(\w+)\s*:\s*(.*)/)
        if (kv) {
          let val = kv[2].trim()
          if ((val.startsWith("'") && val.endsWith("'")) || (val.startsWith('"') && val.endsWith('"'))) {
            val = val.slice(1, -1)
          }
          params.push({ key: kv[1], value: val })
        } else {
          let bare = pp.trim()
          if ((bare.startsWith("'") && bare.endsWith("'")) || (bare.startsWith('"') && bare.endsWith('"'))) {
            bare = bare.slice(1, -1)
          }
          params.push({ key: '', value: bare })
        }
      }
    }
    if (params.length === 0) params.push({ key: '', value: '' })
    return { name: funcName, not, params }
  })
}

function splitParamParts(str) {
  const result = []
  let depth = 0
  let inQuote = false
  let current = ''
  for (const ch of str) {
    if (inQuote) { current += ch; if (ch === "'" || ch === '"') inQuote = false; continue }
    if (ch === "'" || ch === '"') { inQuote = true; current += ch; continue }
    if (ch === '(') { depth++; current += ch; continue }
    if (ch === ')') { depth--; current += ch; continue }
    if (ch === ',' && depth === 0) { result.push(current.trim()); current = ''; continue }
    current += ch
  }
  if (current.trim()) result.push(current.trim())
  return result
}

// Serialize structured expressions back to filter string
function serializeFilterExpr(exprs) {
  return exprs.map(e => {
    const prefix = e.not ? '!' : ''
    const params = e.params
      .filter(p => p.key || p.value)
      .map(p => p.key ? `${p.key}:${p.value}` : p.value)
      .join(', ')
    return `${prefix}${e.name}(${params})`
  }).join(' && ')
}

function updateFilterExpr(groupIndex, filterIndex, exprIndex, field, value, paramIndex, paramField, paramValue) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  const exprs = parseFilterExpr(filters[filterIndex])

  if (field === 'name') {
    // Reset params when function changes
    exprs[exprIndex] = { ...exprs[exprIndex], name: value, params: [{ key: '', value: '' }] }
  } else if (field) {
    exprs[exprIndex] = { ...exprs[exprIndex], [field]: value }
  }

  if (paramIndex !== undefined && paramField !== undefined) {
    const params = [...exprs[exprIndex].params]
    if (paramField === 'key') {
      params[paramIndex] = { ...params[paramIndex], key: paramValue }
    } else {
      params[paramIndex] = { ...params[paramIndex], value: paramValue }
    }
    exprs[exprIndex] = { ...exprs[exprIndex], params }
  }

  filters[filterIndex] = serializeFilterExpr(exprs)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

function addFilterExpr(groupIndex, filterIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  const exprs = parseFilterExpr(filters[filterIndex])
  exprs.push({ name: 'name', not: false, params: [{ key: '', value: '' }] })
  filters[filterIndex] = serializeFilterExpr(exprs)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

function removeFilterExpr(groupIndex, filterIndex, exprIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  const exprs = parseFilterExpr(filters[filterIndex])
  exprs.splice(exprIndex, 1)
  if (exprs.length === 0) exprs.push({ name: 'name', not: false, params: [{ key: '', value: '' }] })
  filters[filterIndex] = serializeFilterExpr(exprs)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

function addFilterParam(groupIndex, filterIndex, exprIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  const exprs = parseFilterExpr(filters[filterIndex])
  const params = [...exprs[exprIndex].params]
  const funcDef = findFilterFuncDef(exprs[exprIndex].name)
  let defaultKey = ''
  if (funcDef && funcDef.paramKeys && funcDef.paramKeys.length > 1) {
    defaultKey = funcDef.paramKeys[1].value
  }
  params.push({ key: defaultKey, value: '' })
  exprs[exprIndex] = { ...exprs[exprIndex], params }
  filters[filterIndex] = serializeFilterExpr(exprs)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

function removeFilterParam(groupIndex, filterIndex, exprIndex, paramIndex) {
  const groups = [...props.modelValue]
  const filters = [...groups[groupIndex].filters]
  const exprs = parseFilterExpr(filters[filterIndex])
  const params = [...exprs[exprIndex].params]
  params.splice(paramIndex, 1)
  exprs[exprIndex] = { ...exprs[exprIndex], params }
  filters[filterIndex] = serializeFilterExpr(exprs)
  groups[groupIndex] = { ...groups[groupIndex], filters }
  emit('update:modelValue', groups)
}

const filterParamExpanded = reactive({})
function filterExpandKey(groupIndex, filterIndex, exprIndex) {
  return `${groupIndex}-${filterIndex}-${exprIndex}`
}
function isFilterExpanded(groupIndex, filterIndex, exprIndex) {
  return !!filterParamExpanded[filterExpandKey(groupIndex, filterIndex, exprIndex)]
}
function toggleFilterExpanded(groupIndex, filterIndex, exprIndex) {
  const key = filterExpandKey(groupIndex, filterIndex, exprIndex)
  filterParamExpanded[key] = !filterParamExpanded[key]
}

function findFilterFuncDef(name) {
  return (props.schema.filterFunctions || []).find(f => f.name === name) || null
}

const expandedGroups = ref({})
function toggleGroup(index) {
  expandedGroups.value[index] = !expandedGroups.value[index]
}
</script>

<template>
  <div class="group-section">
    <div class="list-header">
      <span>Proxy Groups ({{ modelValue.length }})</span>
      <button class="btn-add" @click="addGroup">+ Add Group</button>
    </div>

    <div v-if="modelValue.length === 0" class="empty">
      No groups defined. Add a group to use in routing rules.
    </div>

    <div v-for="(group, gi) in modelValue" :key="gi" class="group-card">
      <div class="group-header" @click="toggleGroup(gi)">
        <span class="group-name">{{ group.name || '(unnamed)' }}</span>
        <span class="group-meta">{{ group.policy || 'no policy' }}</span>
        <span class="group-meta">filters: {{ group.filters.length }}</span>
        <button class="btn-remove" @click.stop="removeGroup(gi)">✕</button>
      </div>

      <div v-if="expandedGroups[gi]" class="group-body">
        <div class="field-row">
          <label class="field-label">Group Name</label>
          <input class="input" type="text" :value="group.name"
            placeholder="my_group"
            @input="updateGroup(gi, 'name', $event.target.value)" />
        </div>

        <div class="field-row">
          <label class="field-label">Policy</label>
          <select class="input" :value="group.policy" @change="updateGroup(gi, 'policy', $event.target.value)">
            <option v-for="po in schema.policyOptions" :key="po.value" :value="po.value">
              {{ po.label }}
            </option>
          </select>
        </div>

        <!-- Structured Filter Builder -->
        <div class="field-row" style="flex-direction: column; align-items: stretch; gap: 8px">
          <div style="display: flex; justify-content: space-between; align-items: center">
            <label class="field-label">Filters</label>
            <button class="btn-add" style="font-size:11px;padding:3px 10px" @click="addFilterLine(gi)">+ Add OR line</button>
          </div>

          <div v-if="group.filters.length === 0" class="empty" style="padding:4px 0">No filters - all nodes match</div>

          <div v-for="(filterStr, fi) in group.filters" :key="fi" class="filter-line-card">
            <div class="filter-line-header">
              <span class="filter-or-badge" v-if="fi > 0">OR</span>
              <span v-else class="filter-or-badge" style="background: transparent; color: #484f58">#{{ fi + 1 }}</span>
              <button class="btn-remove-sm" @click="removeFilterLine(gi, fi)" style="margin-left: auto">✕</button>
            </div>
            <div class="filter-exprs">
              <template v-for="(expr, ei) in parseFilterExpr(filterStr)" :key="ei">
                <template v-if="ei > 0">
                  <select class="logic-select" disabled><option>&&</option></select>
                </template>
                <div class="filter-expr-box">
                  <div class="filter-expr-top">
                    <label class="not-toggle">
                      <input type="checkbox" :checked="expr.not"
                        @change="updateFilterExpr(gi, fi, ei, 'not', $event.target.checked)" />
                      <span class="not-label">NOT</span>
                    </label>
                    <select class="input filter-func-select" :value="expr.name"
                      @change="updateFilterExpr(gi, fi, ei, 'name', $event.target.value)">
                      <option v-for="ff in (schema.filterFunctions || [])" :key="ff.name" :value="ff.name">{{ ff.label }}</option>
                    </select>
                    <button v-if="parseFilterExpr(filterStr).length > 1" class="btn-remove-sm" @click="removeFilterExpr(gi, fi, ei)">✕</button>
                  </div>
                  <div class="filter-expr-params">
                    <template v-for="(param, pi) in expr.params" :key="pi">
                      <div v-if="pi < 3 || isFilterExpanded(gi, fi, ei) || expr.params.length <= 3" class="param-row">
                        <select v-if="findFilterFuncDef(expr.name)?.paramKeys" class="input param-key-select"
                          :value="param.key || ''"
                          @change="updateFilterExpr(gi, fi, ei, undefined, undefined, pi, 'key', $event.target.value)">
                          <option v-for="pk in findFilterFuncDef(expr.name).paramKeys" :key="pk.value" :value="pk.value">{{ pk.label }}</option>
                        </select>
                        <input class="input param-input" type="text" :value="param.value"
                          placeholder="value"
                          @input="updateFilterExpr(gi, fi, ei, undefined, undefined, pi, 'value', $event.target.value)" />
                        <button v-if="expr.params.length > 1" class="btn-remove-param" @click="removeFilterParam(gi, fi, ei, pi)">✕</button>
                      </div>
                    </template>
                    <div v-if="expr.params.length > 3 && !isFilterExpanded(gi, fi, ei)" class="param-overflow" @click="toggleFilterExpanded(gi, fi, ei)">
                      … +{{ expr.params.length - 3 }} more
                    </div>
                    <div v-if="expr.params.length > 3 && isFilterExpanded(gi, fi, ei)" class="param-overflow" @click="toggleFilterExpanded(gi, fi, ei)">
                      … show less
                    </div>
                    <button class="btn-add-param" @click="addFilterParam(gi, fi, ei)">+ param</button>
                  </div>
                </div>
              </template>
              <button class="btn-link" @click="addFilterExpr(gi, fi)">+ ADD</button>
            </div>
          </div>
        </div>

        <div class="field-row" v-for="extra of ['tcp_check_url','tcp_check_http_method','udp_check_dns','check_interval','check_tolerance']" :key="extra">
          <label class="field-label">{{ extra }}</label>
          <input class="input" type="text" :value="group[extra] || ''"
            :placeholder="extra"
            @input="updateGroup(gi, extra, $event.target.value)" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.group-section { display: flex; flex-direction: column; gap: 8px; }

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.list-header span { color: #8b949e; font-size: 13px; }

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

.group-card {
  border: 1px solid #30363d;
  border-radius: 6px;
  overflow: hidden;
}
.group-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: #0d1117;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s;
}
.group-header:hover { background: #161b22; }
.group-name { font-weight: 600; color: #c9d1d9; font-size: 14px; }
.group-meta { font-size: 12px; color: #484f58; }

.group-body {
  padding: 12px 14px;
  border-top: 1px solid #30363d;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row { display: flex; align-items: center; gap: 12px; }
.field-label { font-size: 12px; color: #8b949e; min-width: 100px; }

.input {
  flex: 1;
  padding: 5px 10px;
  background: #161b22;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  max-width: 400px;
}
.input:focus { border-color: #1f6feb; }
select.input { cursor: pointer; }

/* Filter builder styles */
.filter-line-card {
  border: 1px solid #21262d;
  border-radius: 4px;
  background: #0d1117;
  overflow: hidden;
}
.filter-line-header {
  display: flex;
  align-items: center;
  padding: 3px 8px;
  background: #161b22;
  border-bottom: 1px solid #21262d;
}
.filter-or-badge {
  font-size: 10px;
  font-weight: 700;
  background: #f0883e30;
  color: #f0883e;
  padding: 1px 6px;
  border-radius: 3px;
}
.filter-exprs {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 4px;
  padding: 6px 8px;
}
.filter-expr-box {
  border: 1px solid #30363d;
  border-radius: 4px;
  padding: 4px 6px;
  background: #161b22;
}
.filter-expr-top {
  display: flex;
  align-items: center;
  gap: 4px;
}
.filter-func-select { max-width: 80px; }
.filter-expr-params { margin-top: 2px; }

.not-toggle { display: flex; align-items: center; gap: 1px; cursor: pointer; }
.not-toggle input { accent-color: #f85149; }
.not-label { font-size: 9px; font-weight: 700; color: #f85149; text-transform: uppercase; }

.logic-select {
  padding: 3px 5px;
  background: #23863630;
  color: #3fb950;
  border: 1px solid #238636;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  font-family: monospace;
  margin-top: 5px;
}

.param-row { display: flex; gap: 3px; margin-top: 2px; }
.param-key-select { max-width: 75px; flex-shrink: 0; padding: 3px 5px; font-size: 12px; }
.param-input { flex: 1; min-width: 90px; padding: 3px 6px; font-size: 12px; }

.param-overflow {
  font-size: 10px;
  color: #58a6ff;
  padding: 1px 4px;
  margin-top: 2px;
  cursor: pointer;
  user-select: none;
}
.param-overflow:hover { text-decoration: underline; }

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

.btn-remove {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 16px;
  padding: 4px 8px;
  margin-left: auto;
}
.btn-remove:hover { color: #ff7b72; }
.btn-remove-sm {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 12px;
  padding: 2px 4px;
}
.btn-link {
  background: none;
  border: none;
  color: #58a6ff;
  cursor: pointer;
  font-size: 12px;
  padding: 2px 0;
  margin-top: 3px;
}
.btn-link:hover { text-decoration: underline; }
</style>
