<script setup lang="ts">
/**
 * 提示词卡片：标题（1行）、描述（2行）、分类、更新时间、操作图标组。
 */
import { computed } from 'vue'
import {
  Star as StarIcon,
  StarFilled as StarFilledIcon,
  DocumentCopy as CopyIcon,
  Edit as EditIcon,
  MagicStick as TuneIcon,
  Delete as DeleteIcon,
  View as DetailIcon,
} from '@element-plus/icons-vue'
import type { Prompt } from '@/types'
import { useCategoryStore } from '@/stores/category'
import { formatRelativeTime } from '@/utils'

const props = defineProps<{
  prompt: Prompt
  selected?: boolean
  batchMode?: boolean
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

const categoryStore = useCategoryStore()
const categoryName = computed(() => categoryStore.getName(props.prompt.category_id))
const hasCategory = computed(() => props.prompt.category_id != null)

function onSelect(checked: string | number | boolean) {
  emit('select', props.prompt, Boolean(checked))
}

function handleFavorite() {
  emit('favorite', props.prompt)
}
</script>

<template>
  <div class="prompt-card" :class="{ 'prompt-card--selected': selected }">
    <header class="prompt-card__header">
      <el-checkbox
        v-if="batchMode"
        :model-value="selected"
        class="prompt-card__check"
        @change="onSelect"
      />
      <h3 class="prompt-card__title" :title="prompt.title">{{ prompt.title }}</h3>
      <button
        type="button"
        class="prompt-card__star"
        :title="prompt.is_favorite ? '取消收藏' : '收藏'"
        @click.stop="handleFavorite"
      >
        <el-icon v-if="prompt.is_favorite" class="is-fav">
          <StarFilledIcon />
        </el-icon>
        <el-icon v-else>
          <StarIcon />
        </el-icon>
      </button>
    </header>

    <p class="prompt-card__desc" :title="prompt.description || '无描述'">
      {{ prompt.description || '无描述' }}
    </p>

    <footer class="prompt-card__meta">
      <el-tag
        v-if="hasCategory"
        size="small"
        type="info"
        effect="plain"
        round
      >
        {{ categoryName }}
      </el-tag>
      <span v-else class="prompt-card__no-cat">{{ categoryName }}</span>

      <span class="prompt-card__time">{{ formatRelativeTime(prompt.updated_at) }}</span>

      <div class="prompt-card__actions">
        <el-tooltip content="复制" placement="top">
          <el-button text :icon="CopyIcon" @click.stop="emit('copy', prompt)" />
        </el-tooltip>
        <el-tooltip content="查看详情" placement="top">
          <el-button text :icon="DetailIcon" @click.stop="emit('detail', prompt)" />
        </el-tooltip>
        <el-tooltip content="编辑" placement="top">
          <el-button text :icon="EditIcon" @click.stop="emit('edit', prompt)" />
        </el-tooltip>
        <el-tooltip content="AI 调优" placement="top">
          <el-button text :icon="TuneIcon" @click.stop="emit('tune', prompt)" />
        </el-tooltip>
        <el-tooltip content="删除" placement="top">
          <el-button text :icon="DeleteIcon" class="is-danger" @click.stop="emit('delete', prompt)" />
        </el-tooltip>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.prompt-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px 18px;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.05);
  border: 1px solid #eef2f7;
  transition: box-shadow 0.18s, border-color 0.18s, transform 0.18s;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 156px;
}

.prompt-card:hover {
  box-shadow: 0 6px 20px rgba(15, 23, 42, 0.08);
  border-color: #c7d2fe;
  transform: translateY(-1px);
}

.prompt-card--selected {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.12);
}

.prompt-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt-card__check {
  margin-right: 4px;
}

.prompt-card__title {
  flex: 1;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.prompt-card__star {
  border: 0;
  background: transparent;
  padding: 4px;
  cursor: pointer;
  color: #cbd5e1;
  font-size: 18px;
  border-radius: 6px;
  transition: color 0.15s, background 0.15s;
}

.prompt-card__star:hover {
  color: #f59e0b;
  background: #fffbeb;
}

.prompt-card__star .is-fav {
  color: #f59e0b;
}

.prompt-card__desc {
  margin: 0;
  font-size: 13px;
  color: #475569;
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 40px;
}

.prompt-card__meta {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.prompt-card__no-cat {
  font-size: 12px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 999px;
}

.prompt-card__time {
  font-size: 12px;
  color: #94a3b8;
  margin-left: auto;
}

.prompt-card__actions {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: 8px;
}

.prompt-card__actions .is-danger :deep(svg) {
  color: #ef4444;
}

.prompt-card__actions :deep(.el-button.is-text) {
  padding: 4px 6px;
}
</style>