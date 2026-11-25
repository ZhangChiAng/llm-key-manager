<script lang="ts" setup>
import ProviderInput from './ProviderInput.vue'

const props = defineProps<{
  presets: string[]
  requireValue: boolean
  layout: 'inline' | 'stacked'
  valuePlaceholder: string
  disabled?: boolean
}>()
const provider = defineModel<string>('provider', { required: true })
const name = defineModel<string>('name', { required: true })
const value = defineModel<string>('value', { required: true })
const emit = defineEmits<{
  submit: []
  invalid: [message: string]
}>()

/** Validates trimmed drafts without changing the values sent to the backend. */
function submit() {
  if (!provider.value.trim() || !name.value.trim() || (props.requireValue && !value.value.trim())) {
    emit(
      'invalid',
      props.requireValue ? '请填写提供商、Key 名称和 Key 内容' : '请填写提供商和 Key 名称',
    )
    return
  }

  emit('submit')
}
</script>

<template>
  <form :class="layout === 'inline' ? 'key-form' : 'edit-form'" @submit.prevent="submit">
    <label class="field">
      <span>提供商</span>
      <ProviderInput v-model="provider" :presets="presets" :disabled="disabled" />
    </label>

    <label class="field">
      <span>Key 名称</span>
      <t-input v-model="name" :disabled="disabled" placeholder="Chatbox / OpenCode" clearable />
    </label>

    <label class="field" :class="{ 'field-wide': layout === 'inline' }">
      <span>Key 内容</span>
      <t-input v-model="value" :disabled="disabled" :placeholder="valuePlaceholder" clearable />
    </label>

    <slot name="actions"></slot>
  </form>
</template>

<style scoped>
.key-form {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  padding: 24px;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.field > span {
  font-size: 14px;
  font-weight: 600;
  color: #344054;
}

.field-wide {
  grid-column: 1 / -1;
}

@media (max-width: 760px) {
  .key-form {
    grid-template-columns: 1fr;
    padding: 18px;
  }
}
</style>
