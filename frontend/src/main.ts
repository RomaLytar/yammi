import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { setupCalendar } from 'v-calendar'
import 'v-calendar/style.css'
import App from './App.vue'
import { router } from './router'
import { useAuthStore } from './stores/auth'
import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(setupCalendar, { locales: { ru: { firstDayOfWeek: 2, masks: { title: 'MMMM YYYY' } } } })

app.config.errorHandler = (err, _instance, info) => {
  console.error(`[Global Error] ${info}:`, err)
}

// Восстанавливаем сессию из localStorage до первого рендера
const authStore = useAuthStore()
authStore.hydrate().then(() => {
  app.mount('#app')
})
