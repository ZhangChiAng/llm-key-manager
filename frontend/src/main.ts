import { createApp } from 'vue'
import { Button, Dialog, Input, Popconfirm } from 'tdesign-vue-next'
import 'tdesign-vue-next/es/style/index.css'
import App from './App.vue'
import './style.css'

createApp(App).use(Button).use(Dialog).use(Input).use(Popconfirm).mount('#app')
