import { ref } from 'vue'
import { CopyKey, DeleteKey, ListKeys, SaveKey, UpdateKey } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

/** Manages saved key records, backend operations, and page-level feedback. */
export function useKeyManager() {
  const keys = ref<main.KeyRecord[]>([])
  const errorMessage = ref('')
  const successMessage = ref('')
  const isSaving = ref(false)
  const isLoading = ref(false)
  const copyingKeyId = ref('')
  const deletingKeyId = ref('')

  function clearMessages() {
    errorMessage.value = ''
    successMessage.value = ''
  }

  /** Reloads records and reports whether the displayed list is current. */
  async function refreshKeys(): Promise<boolean> {
    clearMessages()
    isLoading.value = true

    try {
      keys.value = await ListKeys()
      return true
    } catch (error) {
      errorMessage.value = getErrorMessage(error, '读取 Key 列表失败')
      return false
    } finally {
      isLoading.value = false
    }
  }

  /** Distinguishes a successful write from a failure to reload its result. */
  async function refreshAfterWrite(operation: string) {
    if (await refreshKeys()) {
      successMessage.value = `Key 已${operation}`
    } else {
      errorMessage.value = `Key 已${operation}，但列表刷新失败，请重试刷新`
    }
  }

  /** Returns write success so the caller can clear a persisted draft. */
  async function saveKey(provider: string, name: string, value: string): Promise<boolean> {
    clearMessages()
    isSaving.value = true

    try {
      await SaveKey(provider, name, value)
      await refreshAfterWrite('保存')
      return true
    } catch (error) {
      errorMessage.value = getErrorMessage(error, '保存 Key 失败')
      return false
    } finally {
      isSaving.value = false
    }
  }

  /** Rejects only failed writes with a localized error for the edit dialog. */
  async function updateKey(
    id: string,
    provider: string,
    name: string,
    value: string,
  ): Promise<void> {
    clearMessages()

    try {
      await UpdateKey(id, provider, name, value)
    } catch (error) {
      throw new Error(getErrorMessage(error, '更新 Key 失败'), { cause: error })
    }

    await refreshAfterWrite('更新')
  }

  /** Copies through the backend without returning plaintext key material. */
  async function copyKey(record: main.KeyRecord) {
    clearMessages()
    copyingKeyId.value = record.id

    try {
      await CopyKey(record.id)
      successMessage.value = 'Key 已复制到剪贴板'
    } catch (error) {
      errorMessage.value = getErrorMessage(error, '复制 Key 失败')
    } finally {
      copyingKeyId.value = ''
    }
  }

  /** Deletes the confirmed record and reloads the saved list. */
  async function deleteKey(record: main.KeyRecord) {
    clearMessages()
    deletingKeyId.value = record.id

    try {
      await DeleteKey(record.id)
      await refreshAfterWrite('删除')
    } catch (error) {
      errorMessage.value = getErrorMessage(error, '删除 Key 失败')
    } finally {
      deletingKeyId.value = ''
    }
  }

  return {
    keys,
    errorMessage,
    successMessage,
    isSaving,
    isLoading,
    copyingKeyId,
    deletingKeyId,
    clearMessages,
    refreshKeys,
    saveKey,
    updateKey,
    copyKey,
    deleteKey,
  }
}

/** Translates backend error codes into Chinese messages for the current operation. */
function getErrorMessage(error: unknown, fallback: string) {
  const message = error instanceof Error ? error.message : typeof error === 'string' ? error : ''

  if (message.includes('ERR_KEY_DUPLICATE')) {
    return '已存在相同提供商、名称和内容的 Key'
  }
  if (message.includes('ERR_KEY_STORE_INVALID')) {
    return '本地 Key 存储文件格式异常'
  }

  return fallback
}
