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

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    return '/boards'
  }
})
