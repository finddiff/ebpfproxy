<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  schema: { type: Object, required: true },
  modelValue: { type: Object, required: true }
})

const emit = defineEmits(['update:modelValue'])
const showAdvanced = ref(false)

const visibleFields = computed(() => {
  const fields = props.schema.fields || []
  if (showAdvanced.value) return fields
  return fields.filter(f => !f.advanced)
})

function setField(name, value) {
  const newVal = { ...props.modelValue, [name]: value }
  emit('update:modelValue', newVal)
}

function isFieldVisible(field) {
  if (!field.depends) return true
  const depVal = props.modelValue[field.depends.field]
  if (field.depends.value === true) return depVal === true
  if (field.depends.value === false) return depVal === false
  if (field.depends.value === 'true') return depVal === true || depVal === 'true'
  if (field.depends.value === 'false') return depVal === false || depVal === 'false'
  return depVal === field.depends.value
}

function getValue(name) {
  return props.modelValue[name] ?? ''
}
</script>

<template>
  <div class="global-section">
    <div class="fields-grid">
      <div v-for="field in visibleFields" :key="field.name" class="field-row" v-show="isFieldVisible(field)">
        <label class="field-label">
          {{ field.label }}
          <span v-if="field.advanced" class="badge">adv</span>
        </label>

        <div class="field-input-area">
          <!-- Enum / Select -->
          <select v-if="field.type === 'enum'" class="input"
            :value="getValue(field.name)"
            @change="setField(field.name, $event.target.value === 'false' ? false : $event.target.value === 'true' ? true : $event.target.value)">
            <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
          </select>

          <!-- Boolean -->
          <label v-else-if="field.type === 'boolean'" class="toggle-label">
            <input type="checkbox" class="toggle"
              :checked="getValue(field.name) === true"
              @change="setField(field.name, $event.target.checked)" />
            <span class="toggle-text">{{ getValue(field.name) === true ? 'On' : 'Off' }}</span>
          </label>

          <!-- Number -->
          <input v-else-if="field.type === 'number'" class="input" type="text"
            :value="getValue(field.name)"
            :placeholder="String(field.default ?? '')"
            @input="setField(field.name, $event.target.value === '' ? '' : isNaN(Number($event.target.value)) ? $event.target.value : Number($event.target.value))" />

          <!-- Duration -->
          <input v-else-if="field.type === 'duration'" class="input" type="text"
            :value="getValue(field.name)"
            :placeholder="field.default || '30s'"
            @input="setField(field.name, $event.target.value)" />

          <!-- String list -->
          <input v-else-if="field.type === 'string_list'" class="input" type="text"
            :value="getValue(field.name)"
            :placeholder="field.default || ''"
            @input="setField(field.name, $event.target.value)" />

          <!-- Default: string -->
          <input v-else class="input" type="text"
            :value="getValue(field.name)"
            :placeholder="String(field.default ?? '')"
            @input="setField(field.name, $event.target.value)" />
        </div>

        <span v-if="field.desc" class="field-desc">{{ field.desc }}</span>
      </div>
    </div>

    <button v-if="!showAdvanced" class="btn-advanced" @click="showAdvanced = true">
      Show advanced settings...
    </button>
    <button v-else class="btn-advanced" @click="showAdvanced = false">
      Hide advanced settings
    </button>
  </div>
</template>

<style scoped>
.global-section { display: flex; flex-direction: column; gap: 0; }

.fields-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
}

.field-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-bottom: 1px solid #21262d;
}
.field-row:nth-child(odd) { border-right: 1px solid #21262d; }

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: #8b949e;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 6px;
}

.badge {
  font-size: 9px;
  background: #1f6feb30;
  color: #58a6ff;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 400;
}

.field-input-area { display: flex; align-items: center; }

.input {
  width: 100%;
  padding: 6px 10px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
}
.input:focus { border-color: #1f6feb; }
select.input { cursor: pointer; }

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.toggle { accent-color: #1f6feb; }
.toggle-text { font-size: 12px; color: #8b949e; }

.field-desc {
  font-size: 11px;
  color: #484f58;
  line-height: 1.3;
}

.btn-advanced {
  margin-top: 12px;
  background: transparent;
  border: none;
  color: #58a6ff;
  cursor: pointer;
  font-size: 12px;
  padding: 4px 0;
  text-align: left;
}
.btn-advanced:hover { text-decoration: underline; }
</style>
