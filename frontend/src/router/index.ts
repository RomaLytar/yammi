import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      component: () => import('@/pages/auth/LoginPage.vue'),
      meta: { layout: 'auth', guest: true },
    },
    {
      path: '/register',
      component: () => import('@/pages/auth/RegisterPage.vue'),
      meta: { layout: 'auth', guest: true },
    },
    {
      path: '/boards',
      component: () => import('@/pages/boards/BoardListPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/boards/:boardId',
      component: () => import('@/pages/boards/BoardPage.vue'),
      meta: { layout: 'board', requiresAuth: true },
    },
    {
      path: '/profile',
      component: () => import('@/pages/profile/ProfilePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/',
      redirect: '/boards',
    },
    {
      path: '/:pathMatch(.*)*',
      component: () => import('@/pages/NotFoundPage.vue'),
      meta: { layout: 'auth' },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  console.log('[ROUTER] beforeEach:', to.path, 'isAuthenticated =', auth.isAuthenticated, 'isHydrating =', auth.isHydrating)

  // ВАЖНО: Не делаем редирект пока идет восстановление сессии
  if (auth.isHydrating) {
    console.log('[ROUTER] isHydrating = true, allowing navigation')
    return // пропускаем, дадим hydrate() завершиться
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    console.log('[ROUTER] requiresAuth but not authenticated, redirect to login')
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    console.log('[ROUTER] guest route but authenticated, redirect to boards')
    return '/boards'
  }
})
