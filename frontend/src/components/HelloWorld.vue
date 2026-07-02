<script lang="ts" setup>
import {reactive} from 'vue'
import {Greet} from '../../wailsjs/go/main/App'

const data = reactive({
  name: "",
  resultText: "Please enter your name below 👇",
})

function greet() {
  Greet(data.name).then(result => {
    data.resultText = result
  })
}

</script>

<template>
  <main>
    <div id="result" class="result">{{ data.resultText }}</div>
    <div id="input" class="input-box">
      <t-input
        id="name"
        v-model="data.name"
        autocomplete="off"
        class="name-input"
        placeholder="Please enter your name"
      />
      <t-button theme="primary" @click="greet">Greet</t-button>
    </div>
  </main>
</template>

<style scoped>
.result {
  min-height: 20px;
  line-height: 20px;
  margin: 1.5rem auto;
}

.input-box {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
}

.name-input {
  width: 240px;
}
</style>
