<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import type { main } from '../wailsjs/go/models'
import KeyEditDialog from './components/KeyEditDialog.vue'
import { useKeyManager } from './composables/useKeyManager'

const {
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
} = useKeyManager()

const form = reactive({
  provider: '',
  name: '',
  value: '',
})

const selectedEditRecord = ref<main.KeyRecord | null>(null)
const expandedProviders = reactive(new Map<string, boolean>())

type ProviderGroup = {
  provider: string
  latestUpdatedAt: number
  records: main.KeyRecord[]
}

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
 * Validates required input and clears the draft after the store write succeeds,
 * including when reloading the saved list fails.
 */
async function submitKey() {
  clearMessages()

  if (!form.provider.trim() || !form.name.trim() || !form.value.trim()) {
    errorMessage.value = '请填写提供商、Key 名称和 Key 内容'
    return
  }

  if (await saveKey(form.provider, form.name, form.value)) {
    form.provider = ''
    form.name = ''
    form.value = ''
  }
}

/**
 * Opens the edit dialog with editable metadata only; key content stays blank so
 * plaintext key material is not requested from the backend.
 */
function openEditDialog(record: main.KeyRecord) {
  clearMessages()
  selectedEditRecord.value = record
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

      <form class="key-form" @submit.prevent="submitKey">
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
              <colgroup>
                <col style="width: 18%" />
                <col style="width: 30%" />
                <col style="width: 20%" />
                <col style="width: 20%" />
                <col style="width: 180px" />
              </colgroup>
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

    <KeyEditDialog
      v-if="selectedEditRecord"
      :record="selectedEditRecord"
      :update-key="updateKey"
      @close="selectedEditRecord = null"
    />
  </main>
</template>
