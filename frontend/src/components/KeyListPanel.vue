<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import { Loading } from 'tdesign-vue-next'
import type { main } from '../../wailsjs/go/models'
import { useKeyManager } from '../composables/useKeyManager'
import { probeStatusLabel } from '../probeStatus'
import KeyEditDialog from './KeyEditDialog.vue'

const {
  isProbing,
  canProbe,
  probeKeys,
  probeState,
  keys,
  isSaving,
  isLoading,
  copyingKeyId,
  deletingKeyId,
  clearMessages,
  copyKey,
  deleteKey,
} = useKeyManager()

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
 * Opens the edit dialog with editable metadata only; key content stays blank so
 * plaintext key material is not requested from the backend.
 */
function openEditDialog(record: main.KeyRecord) {
  if (isProbing.value) return
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
</script>

<template>
  <section class="panel table-panel">
    <div class="table-header">
      <h2>已保存 Key</h2>
      <t-button
        theme="default"
        variant="outline"
        :loading="isProbing"
        :disabled="
          isProbing || !canProbe || isLoading || isSaving || !!deletingKeyId || !!selectedEditRecord
        "
        @click="probeKeys"
        >过期检测</t-button
      >
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
                <td class="key-value">
                  {{ record.maskedValue }}
                  <span class="probe-result" role="status">
                    <Loading
                      v-if="probeState(record) === 'probing'"
                      size="small"
                      aria-label="检测中"
                    />
                    <span v-else>{{ probeStatusLabel(probeState(record)) }}</span>
                  </span>
                </td>
                <td>{{ formatDate(record.createdAt) }}</td>
                <td>{{ formatDate(record.updatedAt) }}</td>
                <td class="table-actions">
                  <t-button
                    theme="default"
                    variant="text"
                    size="small"
                    :disabled="isProbing"
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
                    :confirm-btn="{ disabled: isProbing }"
                    content="确认删除这条 Key 记录？"
                    theme="danger"
                    @confirm="deleteKey(record)"
                  >
                    <t-button
                      theme="danger"
                      variant="text"
                      size="small"
                      :disabled="isProbing"
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
    @close="selectedEditRecord = null"
  />
</template>

<style scoped>
.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  border-bottom: 1px solid #e8edf3;
}

h2 {
  font-size: 18px;
  line-height: 1.3;
  color: #111827;
}

.table-panel {
  overflow: hidden;
}

.provider-groups {
  display: flex;
  flex-direction: column;
}

.provider-group {
  border-bottom: 1px solid #edf1f6;
}

.provider-group:last-child {
  border-bottom: 0;
}

.provider-header {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 12px;
  padding: 16px 24px;
  border: 0;
  color: #1f2937;
  background: #ffffff;
  cursor: pointer;
  text-align: left;
}

.provider-header:hover {
  background: #f8fafc;
}

.provider-toggle {
  width: 18px;
  flex: 0 0 18px;
  color: #667085;
  line-height: 1;
}

.provider-name {
  min-width: 0;
  flex: 1;
  overflow-wrap: anywhere;
  font-weight: 700;
}

.provider-count {
  border: 1px solid #d9e1ec;
  border-radius: 999px;
  padding: 2px 10px;
  color: #475467;
  background: #f8fafc;
  font-size: 13px;
  white-space: nowrap;
}

.provider-state {
  color: #667085;
  font-size: 13px;
  white-space: nowrap;
}

.table-wrap {
  overflow-x: auto;
  border-top: 1px solid #edf1f6;
}

table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  text-align: left;
}

th,
td {
  padding: 14px 18px;
  border-bottom: 1px solid #edf1f6;
  vertical-align: top;
}

th {
  font-size: 13px;
  font-weight: 700;
  color: #475467;
  background: #f8fafc;
}

td {
  color: #1f2937;
}

tbody tr:last-child td {
  border-bottom: 0;
}

.key-value {
  overflow-wrap: anywhere;
  font-family: Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
}

.table-actions {
  white-space: nowrap;
}

.empty-state {
  padding: 48px 24px;
  text-align: center;
  color: #667085;
}

.probe-result {
  font-size: 12px;
  color: #667085;
  display: inline-flex;
  margin-left: 8px;
  vertical-align: middle;
  font-family: 'Segoe UI', sans-serif;
}

@media (max-width: 760px) {
  th,
  td {
    padding: 12px;
  }
}
</style>
