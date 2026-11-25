import { computed, ref } from 'vue'
import {
  CopyKey,
  DeleteKey,
  ListKeys,
  ListProviderPresets,
  ProbeKey,
  SaveKey,
  UpdateKey,
} from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

import type { ProbeDisplayState, ProbeStatus } from '../probeStatus'

let shared: ReturnType<typeof createKeyManager> | undefined

/** Returns the single key-manager state shared by this application's window. */
export function useKeyManager() {
  return (shared ??= createKeyManager())
}

/** Manages saved key records, backend operations, and page-level feedback. */
function createKeyManager() {
  const providerPresets = ref<string[]>([])
  const presetsLoaded = ref(false)
  const isProbing = ref(false)
  const probeStates = ref(new Map<string, ProbeStatus | 'probing'>())
  const canProbe = computed(() => keys.value.some(isProbeSupported))

  function isProbeSupported(record: main.KeyRecord) {
    return providerPresets.value.includes(record.provider.trim())
  }

  /** Keeps transient conclusions in memory; custom providers never send requests. */
  async function probeKeys() {
    if (
      isProbing.value ||
      isLoading.value ||
      isSaving.value ||
      deletingKeyId.value ||
      !canProbe.value
    )
      return
    clearMessages()
    isProbing.value = true
    const records = keys.value.filter(isProbeSupported)
    for (const record of records) probeStates.value.set(record.id, 'probing')
    try {
      await Promise.allSettled(
        records.map(async (record) => {
          try {
            const result = await ProbeKey(record.id)
            /** The backend returns only the closed set of internal/probe Status values. */
            probeStates.value.set(record.id, result.status as ProbeStatus)
          } catch (error) {
            probeStates.value.set(record.id, 'failed')
            errorMessage.value = getErrorMessage(error, '读取或解密 Key 失败')
          }
        }),
      )
    } finally {
      isProbing.value = false
    }
  }

  function probeState(record: main.KeyRecord): ProbeDisplayState {
    if (presetsLoaded.value && !isProbeSupported(record)) return 'unsupported'
    return probeStates.value.get(record.id) ?? ''
  }

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
      const [records, presets] = await Promise.all([ListKeys(), ListProviderPresets()])
      /** Keeps a probe result only if its record is identical to the one shown before. */
      const previous = new Map(keys.value.map((record) => [record.id, JSON.stringify(record)]))
      probeStates.value = new Map(
        records.flatMap((record) => {
          const state = probeStates.value.get(record.id)
          return state && previous.get(record.id) === JSON.stringify(record)
            ? [[record.id, state]]
            : []
        }),
      )
      keys.value = records
      providerPresets.value = presets
      presetsLoaded.value = true
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
    if (isProbing.value) return false
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
    if (isProbing.value) throw new Error('检测期间无法修改记录')
    clearMessages()

    try {
      await UpdateKey(id, provider, name, value)
      probeStates.value.delete(id)
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
    if (isProbing.value) return
    clearMessages()
    deletingKeyId.value = record.id

    try {
      await DeleteKey(record.id)
      probeStates.value.delete(record.id)
      await refreshAfterWrite('删除')
    } catch (error) {
      errorMessage.value = getErrorMessage(error, '删除 Key 失败')
    } finally {
      deletingKeyId.value = ''
    }
  }

  return {
    providerPresets,
    isProbing,
    canProbe,
    probeKeys,
    probeState,
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
