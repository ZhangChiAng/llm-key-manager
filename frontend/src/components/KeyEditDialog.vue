<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import type { main } from '../../wailsjs/go/models'
import { useKeyManager } from '../composables/useKeyManager'
import KeyForm from './KeyForm.vue'

/** The dialog receives metadata only; a blank key value preserves its ciphertext. */
const props = defineProps<{
  record: main.KeyRecord
}>()
const emit = defineEmits<{
  close: []
}>()

const { providerPresets, updateKey } = useKeyManager()

const form = reactive({
  provider: props.record.provider,
  name: props.record.name,
  value: '',
})
const isVisible = ref(true)
const isUpdating = ref(false)
const errorMessage = ref('')
const isUnchanged = computed(
  () =>
    form.provider.trim() === props.record.provider.trim() &&
    form.name.trim() === props.record.name.trim() &&
    form.value.trim() === '',
)

/** Keeps the draft open on write failure and closes after a successful update. */
async function save() {
  errorMessage.value = ''

  isUpdating.value = true

  try {
    await updateKey(props.record.id, form.provider, form.name, form.value)
    isVisible.value = false
  } catch (error) {
    errorMessage.value = (error as Error).message
  } finally {
    isUpdating.value = false
  }
}
</script>

<template>
  <t-dialog
    v-model:visible="isVisible"
    header="编辑 Key"
    width="520px"
    :footer="false"
    :close-on-overlay-click="!isUpdating"
    :close-on-esc-keydown="!isUpdating"
    @closed="emit('close')"
  >
    <KeyForm
      v-model:provider="form.provider"
      v-model:name="form.name"
      v-model:value="form.value"
      :presets="providerPresets"
      :require-value="false"
      layout="stacked"
      value-placeholder="留空表示不修改"
      @submit="save"
      @invalid="errorMessage = $event"
    >
      <template #actions>
        <div class="dialog-actions">
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <span v-else></span>
          <div class="dialog-buttons">
            <t-button variant="outline" :disabled="isUpdating" @click="isVisible = false">
              取消
            </t-button>
            <t-button
              theme="primary"
              type="submit"
              :disabled="isUpdating || isUnchanged"
              :loading="isUpdating"
            >
              保存
            </t-button>
          </div>
        </div>
      </template>
    </KeyForm>
  </t-dialog>
</template>

<style scoped>
.dialog-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 40px;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  white-space: nowrap;
}

@media (max-width: 760px) {
  .dialog-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .dialog-buttons {
    justify-content: stretch;
  }
}
</style>
