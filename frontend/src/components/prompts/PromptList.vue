<script setup lang="ts">
/**
 * 卡片网格列表。
 *
 * 列数自适应：
 *   - ≥1280px → 4 列（对齐设计稿）
 *   - ≥960px  → 3 列
 *   - ≥640px  → 2 列
 *   - <640px  → 1 列（移动端）
 *
 * 视图模式：
 *   - 'card'（默认）：网格布局
 *   - 'list'：单列紧凑列表
 *
 * 加载态显示骨架屏；空态由父组件通过 slot 渲染。
 */
import type { Prompt } from '@/types'
import PromptCard from './PromptCard.vue'

defineProps<{
  list: Prompt[]
  loading: boolean
  batchMode: boolean
  selectedIds: number[]
  viewMode?: 'card' | 'list'
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
    <div
      v-if="loading"
      class="prompt-grid"
      :class="[
        list.length === 0 ? 'prompt-grid--cards' : `prompt-grid--cards`,
      ]"
    >
      <el-skeleton
        v-for="i in 8"
        :key="i"
        :rows="3"
        animated
        class="prompt-grid__skeleton"
      />
    </div>

    <div
      v-else-if="list.length > 0"
      class="prompt-grid"
      :class="[
        viewMode === 'list' ? 'prompt-grid--list' : 'prompt-grid--cards',
      ]"
    >
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
/* 卡片网格：自适应列数，向下渐变；最大 3 列以保证每张卡片不会过宽 */
.prompt-grid--cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 14px;
}

@media (min-width: 960px) {
  .prompt-grid--cards {
    grid-template-columns: repeat(3, 1fr);
    gap: 14px;
  }
}

@media (min-width: 640px) and (max-width: 959px) {
  .prompt-grid--cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* 列表视图（占位） */
.prompt-grid--list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 骨架屏 */
.prompt-grid__skeleton {
  background: #ffffff;
  border-radius: 12px;
  padding: 14px 16px;
  min-height: 132px;
  border: 1px solid #eef2f7;
}

.prompt-grid__skeleton :deep(.el-skeleton__item) {
  height: 14px;
  margin-bottom: 10px;
  border-radius: 4px;
}
</style>
