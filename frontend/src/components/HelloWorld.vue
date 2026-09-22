<script setup lang="ts">
/**
 * 占位组件：默认 Layout 内可放置通用业务组件
 * 例如：顶部导航、侧边栏、面包屑等。
 */
defineProps<{ name?: string }>()
</script>

<template>
  <div class="placeholder">
    <span>{{ name || '占位组件' }}</span>
  </div>
</template>

<style scoped>
.placeholder {
  padding: 16px;
  border: 1px dashed var(--el-border-color);
  border-radius: 8px;
  color: var(--el-text-color-secondary);
  text-align: center;
}
</style>
