import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { useAuthStore } from './stores/auth'
import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

app.config.errorHandler = (err, _instance, info) => {
  console.error(`[Global Error] ${info}:`, err)
}

// Восстанавливаем сессию из localStorage до первого рендера
const authStore = useAuthStore()
authStore.hydrate().then(() => {
  app.mount('#app')
})
