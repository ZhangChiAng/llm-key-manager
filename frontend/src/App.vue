<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue'
import { ListKeys, SaveKey } from '../wailsjs/go/main/App'
import type { main } from '../wailsjs/go/models'

const form = reactive({
  provider: '',
  name: '',
  value: '',
})

const keys = ref<main.KeyRecord[]>([])
const errorMessage = ref('')
const isSaving = ref(false)
const isLoading = ref(false)

/**
 * Reloads key records from the Wails backend so the table reflects the local
 * plaintext key store on disk.
 */
async function refreshKeys() {
  isLoading.value = true
  errorMessage.value = ''

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
 * Normalizes Wails bridge errors and frontend exceptions into text that can be
 * displayed in the form error area.
 */
function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) {
    return error.message
  }
  if (typeof error === 'string' && error) {
    return error
  }

  return fallback
}

function formatDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

onMounted(refreshKeys)
</script>

<template>
  <main class="app-shell">
    <section class="panel">
      <div class="panel-header">
        <div>
          <h1>LLM Key Manager</h1>
          <p>本地明文保存和查看 API Key。</p>
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

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>提供商</th>
              <th>名称</th>
              <th>Key 内容</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in keys" :key="record.id">
              <td>{{ record.provider }}</td>
              <td>{{ record.name }}</td>
              <td class="key-value">{{ record.value }}</td>
              <td>{{ formatDate(record.createdAt) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </main>
</template>
