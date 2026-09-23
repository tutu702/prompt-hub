<script setup lang="ts">
/**
 * 卡片网格列表（CSS Grid 自适应 1/2/3 列）。
 * 加载态显示 skeleton；空态由父组件判断渲染。
 */
import type { Prompt } from '@/types'
import PromptCard from './PromptCard.vue'

defineProps<{
  list: Prompt[]
  loading: boolean
  batchMode: boolean
  selectedIds: number[]
}>()

const emit = defineEmits<{
  (e: 'copy', p: Prompt): void
  (e: 'edit', p: Prompt): void
  (e: 'tune', p: Prompt): void
  (e: 'delete', p: Prompt): void
  (e: 'detail', p: Prompt): void
  (e: 'favorite', p: Prompt): void
  (e: 'select', p: Prompt, checked: boolean): void
}>()
</script>

<template>
  <div>
    <div v-if="loading" class="prompt-grid">
      <el-skeleton
        v-for="i in 6"
        :key="i"
        :rows="3"
        animated
        class="prompt-grid__skeleton"
      />
    </div>

    <div v-else-if="list.length > 0" class="prompt-grid">
      <PromptCard
        v-for="p in list"
        :key="p.id"
        :prompt="p"
        :selected="selectedIds.includes(p.id)"
        :batch-mode="batchMode"
        @copy="(item: Prompt) => emit('copy', item)"
        @edit="(item: Prompt) => emit('edit', item)"
        @tune="(item: Prompt) => emit('tune', item)"
        @delete="(item: Prompt) => emit('delete', item)"
        @detail="(item: Prompt) => emit('detail', item)"
        @favorite="(item: Prompt) => emit('favorite', item)"
        @select="(item: Prompt, checked: boolean) => emit('select', item, checked)"
      />
    </div>

    <slot v-else name="empty" />
  </div>
</template>

<style scoped>
.prompt-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.prompt-grid__skeleton {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px 18px;
  min-height: 156px;
  border: 1px solid #eef2f7;
}

.prompt-grid__skeleton :deep(.el-skeleton__item) {
  height: 14px;
  margin-bottom: 10px;
}
</style>