<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import DefaultLayout from './layouts/DefaultLayout.vue'
import AuthLayout from './layouts/AuthLayout.vue'
import BoardLayout from './layouts/BoardLayout.vue'

const route = useRoute()

const layoutComponents = {
  default: DefaultLayout,
  auth: AuthLayout,
  board: BoardLayout,
} as const

type LayoutName = keyof typeof layoutComponents

const layout = computed(() => {
  const name = (route.meta.layout as LayoutName) || 'default'
  return layoutComponents[name] || DefaultLayout
})
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>
