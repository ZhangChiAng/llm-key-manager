<script lang="ts" setup>
import { AutoComplete } from 'tdesign-vue-next'
import type { AutoCompleteProps } from 'tdesign-vue-next'

/** Custom provider names stay editable without being added to the fixed suggestions. */
const provider = defineModel<string>({ required: true })
const presets = ['DeepSeek', 'MiniMax', 'OfoxAI']

/** Selecting a suggestion with Enter must not also submit the surrounding form. */
const selectProvider: NonNullable<AutoCompleteProps['onSelect']> = (_value, { e }) => {
  e.preventDefault()
}
</script>

<template>
  <AutoComplete
    v-model="provider"
    :options="presets"
    placeholder="选择或输入提供商"
    empty="可直接输入自定义提供商"
    filterable
    clearable
    @select="selectProvider"
  />
</template>
