<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import type { main } from '../../wailsjs/go/models'

/** The dialog receives metadata only; a blank key value preserves its ciphertext. */
const props = defineProps<{
  record: main.KeyRecord
  updateKey: (id: string, provider: string, name: string, value: string) => Promise<void>
}>()
const emit = defineEmits<{
  close: []
}>()

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

  if (!form.provider.trim() || !form.name.trim()) {
    errorMessage.value = '请填写提供商和 Key 名称'
    return
  }

  isUpdating.value = true

  try {
    await props.updateKey(props.record.id, form.provider, form.name, form.value)
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
    <form class="edit-form" @submit.prevent="save">
      <label class="field">
        <span>提供商</span>
        <t-input v-model="form.provider" placeholder="DeepSeek / OfoxAI" clearable />
      </label>

      <label class="field">
        <span>Key 名称</span>
        <t-input v-model="form.name" placeholder="Chatbox / OpenCode" clearable />
      </label>

      <label class="field">
        <span>Key 内容</span>
        <t-input v-model="form.value" placeholder="留空表示不修改" clearable />
      </label>

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
    </form>
  </t-dialog>
</template>
