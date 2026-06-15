<script setup>
const props = defineProps({
  schema: { type: Object, required: true },
  modelValue: { type: Array, required: true }
})

const emit = defineEmits(['update:modelValue'])

function addItem() {
  const list = [...props.modelValue, '']
  emit('update:modelValue', list)
}

function updateItem(index, value) {
  const list = [...props.modelValue]
  list[index] = value
  emit('update:modelValue', list)
}

function removeItem(index) {
  const list = [...props.modelValue]
  list.splice(index, 1)
  emit('update:modelValue', list)
}
</script>

<template>
  <div class="list-section">
    <div class="list-header">
      <span>{{ schema.label }} Entries ({{ modelValue.length }})</span>
      <button class="btn-add" @click="addItem">+ Add {{ schema.label }}</button>
    </div>

    <div v-if="modelValue.length === 0" class="empty">
      No {{ schema.label.toLowerCase() }} entries defined
    </div>

    <div v-for="(item, i) in modelValue" :key="i" class="list-row">
      <span class="row-num">{{ i + 1 }}</span>
      <input class="input" type="text"
        :value="item"
        :placeholder="schema.placeholder || ''"
        @input="updateItem(i, $event.target.value)" />
      <button class="btn-remove" @click="removeItem(i)" title="Remove">✕</button>
    </div>
  </div>
</template>

<style scoped>
.list-section { display: flex; flex-direction: column; gap: 8px; }

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

.list-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.row-num {
  font-size: 12px;
  color: #484f58;
  min-width: 20px;
  text-align: right;
}

.input {
  flex: 1;
  padding: 6px 10px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 4px;
  font-size: 13px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  outline: none;
}
.input:focus { border-color: #1f6feb; }

.btn-remove {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  font-size: 14px;
  padding: 4px 8px;
}
.btn-remove:hover { color: #ff7b72; }
</style>
