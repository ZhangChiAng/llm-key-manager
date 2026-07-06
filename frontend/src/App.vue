<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { CopyKey, DeleteKey, ListKeys, SaveKey, UpdateKey } from '../wailsjs/go/main/App'
import type { main } from '../wailsjs/go/models'

const form = reactive({
  provider: '',
  name: '',
  value: '',
})

const editForm = reactive({
  provider: '',
  name: '',
  value: '',
})

const keys = ref<main.KeyRecord[]>([])
const errorMessage = ref('')
const successMessage = ref('')
const editErrorMessage = ref('')
const isSaving = ref(false)
const isLoading = ref(false)
const isEditDialogVisible = ref(false)
const isUpdating = ref(false)
const copyingKeyId = ref('')
const deletingKeyId = ref('')
const selectedEditRecord = ref<main.KeyRecord | null>(null)
const originalEditProvider = ref('')
const originalEditName = ref('')
const expandedProviders = reactive(new Map<string, boolean>())

type ProviderGroup = {
  provider: string
  latestUpdatedAt: number
  records: main.KeyRecord[]
}

const isEditUnchanged = computed(
  () =>
    editForm.provider.trim() === originalEditProvider.value.trim() &&
    editForm.name.trim() === originalEditName.value.trim() &&
    editForm.value.trim() === '',
)

const groupedKeys = computed(() => {
  const groups = new Map<string, ProviderGroup>()

  for (const record of keys.value) {
    const provider = record.provider.trim()
    const updatedAt = getDateTime(record.updatedAt)
    const group = groups.get(provider)

    if (group) {
      group.records.push(record)
      group.latestUpdatedAt = Math.max(group.latestUpdatedAt, updatedAt)
      continue
    }

    groups.set(provider, {
      provider,
      latestUpdatedAt: updatedAt,
      records: [record],
    })
  }

  return Array.from(groups.values())
    .map((group) => ({
      ...group,
      records: [...group.records].sort((left, right) => {
        return getDateTime(right.updatedAt) - getDateTime(left.updatedAt)
      }),
    }))
    .sort((left, right) => right.latestUpdatedAt - left.latestUpdatedAt)
})

/**
 * Reloads key records from the Wails backend so the table reflects the local
 * encrypted key store on disk.
 */
async function refreshKeys() {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    keys.value = await ListKeys()
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '读取 Key 列表失败')
  } finally {
    isLoading.value = false
  }
}

/**
 * Persists the current form values through the Wails backend and clears the
 * form only after the local key store write succeeds.
 */
async function saveKey() {
  errorMessage.value = ''
  successMessage.value = ''

  if (!form.provider.trim() || !form.name.trim() || !form.value.trim()) {
    errorMessage.value = '请填写提供商、Key 名称和 Key 内容'
    return
  }

  isSaving.value = true

  try {
    await SaveKey(form.provider, form.name, form.value)
    form.provider = ''
    form.name = ''
    form.value = ''
    await refreshKeys()
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存 Key 失败')
  } finally {
    isSaving.value = false
  }
}

/**
 * Opens the edit dialog with editable metadata only; key content stays blank so
 * plaintext key material is not requested from the backend.
 */
function openEditDialog(record: main.KeyRecord) {
  errorMessage.value = ''
  successMessage.value = ''
  editErrorMessage.value = ''
  selectedEditRecord.value = record
  editForm.provider = record.provider
  editForm.name = record.name
  editForm.value = ''
  originalEditProvider.value = record.provider
  originalEditName.value = record.name
  isEditDialogVisible.value = true
}

/**
 * Clears edit state after the dialog has closed.
 */
function resetEditForm() {
  selectedEditRecord.value = null
  editForm.provider = ''
  editForm.name = ''
  editForm.value = ''
  originalEditProvider.value = ''
  originalEditName.value = ''
  editErrorMessage.value = ''
}

/**
 * Updates metadata and replaces the encrypted key value only when the user
 * entered new key content.
 */
async function updateKey() {
  editErrorMessage.value = ''

  if (!editForm.provider.trim() || !editForm.name.trim()) {
    editErrorMessage.value = '请填写提供商和 Key 名称'
    return
  }

  const record = selectedEditRecord.value
  if (!record) {
    editErrorMessage.value = '更新 Key 失败'
    return
  }

  isUpdating.value = true

  try {
    await UpdateKey(record.id, editForm.provider, editForm.name, editForm.value)
    isEditDialogVisible.value = false
    await refreshKeys()
    successMessage.value = 'Key 已更新'
  } catch (error) {
    editErrorMessage.value = getErrorMessage(error, '更新 Key 失败')
  } finally {
    isUpdating.value = false
  }
}

/**
 * Requests a backend clipboard copy so plaintext key material is never returned
 * to the frontend.
 */
async function copyKey(record: main.KeyRecord) {
  errorMessage.value = ''
  successMessage.value = ''
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

/**
 * Removes a saved key record after the user confirms the row-level delete
 * prompt, then reloads records from the backend store.
 */
async function deleteKey(record: main.KeyRecord) {
  errorMessage.value = ''
  successMessage.value = ''
  deletingKeyId.value = record.id

  try {
    await DeleteKey(record.id)
    await refreshKeys()
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除 Key 失败')
  } finally {
    deletingKeyId.value = ''
  }
}

/**
 * Translates stable backend error codes and hides internal error details from
 * the UI behind operation-specific Chinese fallbacks.
 */
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

function getDateTime(value: string) {
  const time = new Date(value).getTime()

  return Number.isNaN(time) ? 0 : time
}

function formatDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

function isProviderExpanded(provider: string) {
  return expandedProviders.get(provider) === true
}

function toggleProvider(provider: string) {
  expandedProviders.set(provider, !isProviderExpanded(provider))
}

onMounted(refreshKeys)
</script>

<template>
  <main class="app-shell">
    <section class="panel">
      <div class="panel-header">
        <div>
          <h1>LLM Key Manager</h1>
          <p>本地加密保存 API Key。</p>
        </div>
        <t-button theme="default" variant="outline" :loading="isLoading" @click="refreshKeys">
          刷新
        </t-button>
      </div>

      <form class="key-form" @submit.prevent="saveKey">
        <label class="field">
          <span>提供商</span>
          <t-input v-model="form.provider" placeholder="DeepSeek / OfoxAI" clearable />
        </label>

        <label class="field">
          <span>Key 名称</span>
          <t-input v-model="form.name" placeholder="Chatbox / OpenCode" clearable />
        </label>

        <label class="field field-wide">
          <span>Key 内容</span>
          <t-input v-model="form.value" placeholder="sk-..." clearable />
        </label>

        <div class="form-actions">
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <p v-else-if="successMessage" class="success-message">{{ successMessage }}</p>
          <span v-else></span>
          <t-button theme="primary" type="submit" :loading="isSaving">保存</t-button>
        </div>
      </form>
    </section>

    <section class="panel table-panel">
      <div class="table-header">
        <h2>已保存 Key</h2>
        <span>{{ keys.length }} 条记录</span>
      </div>

      <div v-if="keys.length === 0" class="empty-state">暂无记录</div>

      <div v-else class="provider-groups">
        <section v-for="group in groupedKeys" :key="group.provider" class="provider-group">
          <button
            class="provider-header"
            type="button"
            :aria-expanded="isProviderExpanded(group.provider)"
            @click="toggleProvider(group.provider)"
          >
            <span class="provider-toggle" aria-hidden="true">
              {{ isProviderExpanded(group.provider) ? '▾' : '▸' }}
            </span>
            <span class="provider-name">{{ group.provider }}</span>
            <span class="provider-count">{{ group.records.length }} 条</span>
            <span class="provider-state">
              {{ isProviderExpanded(group.provider) ? '收起' : '展开' }}
            </span>
          </button>

          <div v-if="isProviderExpanded(group.provider)" class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>名称</th>
                  <th>Key 内容</th>
                  <th>创建时间</th>
                  <th>更新时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="record in group.records" :key="record.id">
                  <td>{{ record.name }}</td>
                  <td class="key-value">{{ record.maskedValue }}</td>
                  <td>{{ formatDate(record.createdAt) }}</td>
                  <td>{{ formatDate(record.updatedAt) }}</td>
                  <td class="table-actions">
                    <t-button
                      theme="default"
                      variant="text"
                      size="small"
                      @click="openEditDialog(record)"
                    >
                      编辑
                    </t-button>
                    <t-button
                      theme="primary"
                      variant="text"
                      size="small"
                      :loading="copyingKeyId === record.id"
                      @click="copyKey(record)"
                    >
                      复制
                    </t-button>
                    <t-popconfirm
                      content="确认删除这条 Key 记录？"
                      theme="danger"
                      @confirm="deleteKey(record)"
                    >
                      <t-button
                        theme="danger"
                        variant="text"
                        size="small"
                        :loading="deletingKeyId === record.id"
                      >
                        删除
                      </t-button>
                    </t-popconfirm>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </section>

    <t-dialog
      v-model:visible="isEditDialogVisible"
      header="编辑 Key"
      width="520px"
      :footer="false"
      :close-on-overlay-click="!isUpdating"
      :close-on-esc-keydown="!isUpdating"
      @closed="resetEditForm"
    >
      <form class="edit-form" @submit.prevent="updateKey">
        <label class="field">
          <span>提供商</span>
          <t-input v-model="editForm.provider" placeholder="DeepSeek / OfoxAI" clearable />
        </label>

        <label class="field">
          <span>Key 名称</span>
          <t-input v-model="editForm.name" placeholder="Chatbox / OpenCode" clearable />
        </label>

        <label class="field">
          <span>Key 内容</span>
          <t-input v-model="editForm.value" placeholder="留空表示不修改" clearable />
        </label>

        <div class="dialog-actions">
          <p v-if="editErrorMessage" class="error-message">{{ editErrorMessage }}</p>
          <span v-else></span>
          <div class="dialog-buttons">
            <t-button variant="outline" :disabled="isUpdating" @click="isEditDialogVisible = false">
              取消
            </t-button>
            <t-button
              theme="primary"
              type="submit"
              :disabled="isUpdating || isEditUnchanged"
              :loading="isUpdating"
            >
              保存
            </t-button>
          </div>
        </div>
      </form>
    </t-dialog>
  </main>
</template>
