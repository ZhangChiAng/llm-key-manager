<script lang="ts" setup>
import { onMounted } from 'vue'
import AddKeyPanel from './components/AddKeyPanel.vue'
import KeyListPanel from './components/KeyListPanel.vue'
import { useKeyManager } from './composables/useKeyManager'

const { refreshKeys } = useKeyManager()

onMounted(refreshKeys)
</script>

<template>
  <main class="app-shell">
    <AddKeyPanel />
    <KeyListPanel />
  </main>
</template>
