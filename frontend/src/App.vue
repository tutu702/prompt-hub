<script setup lang="ts">
/**
 * 应用根组件
 * 根据 route.meta.layout 动态选择 BlankLayout 或 DefaultLayout 包裹 <router-view />。
 */
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import DefaultLayout from '@/layouts/DefaultLayout.vue'
import BlankLayout from '@/layouts/BlankLayout.vue'

const route = useRoute()

const layoutComponent = computed(() => {
  const layout = (route.meta?.layout as string | undefined) || 'default'
  return layout === 'blank' ? BlankLayout : DefaultLayout
})
</script>

<template>
  <el-config-provider>
    <component :is="layoutComponent">
      <router-view />
    </component>
  </el-config-provider>
</template>

<style scoped>
/* 根组件样式 - 全局基础样式请放入 src/styles */
</style>