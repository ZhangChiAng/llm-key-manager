<script lang="ts" setup>
import { reactive } from 'vue'
import { useKeyManager } from '../composables/useKeyManager'
import KeyForm from './KeyForm.vue'

const {
  providerPresets,
  isProbing,
  errorMessage,
  successMessage,
  isSaving,
  isLoading,
  clearMessages,
  refreshKeys,
  saveKey,
} = useKeyManager()

const form = reactive({
  provider: '',
  name: '',
  value: '',
})

function showValidationError(message: string) {
  if (isProbing.value) return
  clearMessages()
  errorMessage.value = message
}

/** Clears the draft after a successful write, even if reloading the list fails. */
async function submitKey() {
  if (isProbing.value) return
  clearMessages()

  if (await saveKey(form.provider, form.name, form.value)) {
    form.provider = ''
    form.name = ''
    form.value = ''
  }
}
</script>

<template>
  <section class="panel">
    <div class="panel-header">
      <div>
        <h1>LLM Key Manager</h1>
        <p>本地加密保存 API Key。</p>
      </div>
      <t-button
        theme="default"
        variant="outline"
        :loading="isLoading"
        :disabled="isProbing"
        @click="refreshKeys"
      >
        刷新
      </t-button>
    </div>

    <KeyForm
      v-model:provider="form.provider"
      v-model:name="form.name"
      v-model:value="form.value"
      :presets="providerPresets"
      :require-value="true"
      layout="inline"
      value-placeholder="sk-..."
      :disabled="isProbing"
      @submit="submitKey"
      @invalid="showValidationError"
    >
      <template #actions>
        <div class="form-actions">
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <p v-else-if="successMessage" class="success-message">{{ successMessage }}</p>
          <span v-else></span>
          <t-button theme="primary" type="submit" :loading="isSaving" :disabled="isProbing"
            >保存</t-button
          >
        </div>
      </template>
    </KeyForm>
  </section>
</template>

<style scoped>
.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  border-bottom: 1px solid #e8edf3;
}

.panel-header p {
  margin-top: 8px;
  color: #667085;
}

h1 {
  font-size: 28px;
  line-height: 1.2;
  color: #111827;
}

.form-actions {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 40px;
}

@media (max-width: 760px) {
  .panel-header,
  .form-actions {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
