<script setup lang="ts">
/**
 * 提示词卡片：标题（1 行）/ 描述（2 行）/ 分类 chip + 时间 / 操作图标。
 *
 * 设计要点（对齐设计稿）：
 * - 卡片更大、内边距宽松
 * - 描述最大 2 行
 * - 分类为浅蓝底色小 chip（el-tag type="primary" plain）
 * - 操作图标：复制 / 编辑 / AI 调优 / 删除 4 个，hover 显色
 */
import { computed } from 'vue'
import {
  Star as StarIcon,
  StarFilled as StarFilledIcon,
  DocumentCopy as CopyIcon,
  Edit as EditIcon,
  MagicStick as TuneIcon,
  Delete as DeleteIcon,
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
const categoryName = computed(() =>
  categoryStore.getName(props.prompt.category_id),
)
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
    <!-- 批量多选框（hover 时也显示，便于快速多选） -->
    <el-checkbox
      v-if="batchMode"
      class="prompt-card__check"
      :model-value="selected"
      @change="onSelect"
    />

    <!-- 头部：标题 + 收藏 -->
    <header class="prompt-card__header">
      <h3 class="prompt-card__title" :title="prompt.title">
        {{ prompt.title }}
      </h3>
      <button
        type="button"
        class="prompt-card__star"
        :title="prompt.is_favorite ? '取消收藏' : '收藏'"
        @click.stop="handleFavorite"
      >
        <el-icon v-if="prompt.is_favorite" class="is-fav">
          <StarFilledIcon />
        </el-icon>
        <el-icon v-else><StarIcon /></el-icon>
      </button>
    </header>

    <!-- 描述 -->
    <p class="prompt-card__desc" :title="prompt.description || '无描述'">
      {{ prompt.description || '无描述' }}
    </p>

    <!-- 底部：分类 + 时间 + 操作 -->
    <footer class="prompt-card__footer">
      <div class="prompt-card__meta">
        <span
          v-if="hasCategory"
          class="prompt-card__tag"
          :title="categoryName"
        >
          {{ categoryName }}
        </span>
        <span class="prompt-card__time">{{ formatRelativeTime(prompt.updated_at) }}</span>
      </div>

      <div class="prompt-card__actions">
        <button
          type="button"
          class="prompt-card__action"
          title="复制"
          @click.stop="emit('copy', prompt)"
        >
          <el-icon><CopyIcon /></el-icon>
        </button>
        <button
          type="button"
          class="prompt-card__action"
          title="编辑"
          @click.stop="emit('edit', prompt)"
        >
          <el-icon><EditIcon /></el-icon>
        </button>
        <button
          type="button"
          class="prompt-card__action"
          title="AI 调优"
          @click.stop="emit('tune', prompt)"
        >
          <el-icon><TuneIcon /></el-icon>
        </button>
        <button
          type="button"
          class="prompt-card__action prompt-card__action--danger"
          title="删除"
          @click.stop="emit('delete', prompt)"
        >
          <el-icon><DeleteIcon /></el-icon>
        </button>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.prompt-card {
  position: relative;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 14px 16px 12px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 132px;
  transition: border-color 0.18s, box-shadow 0.18s, transform 0.18s;
}

.prompt-card:hover {
  border-color: #c7d2fe;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
  transform: translateY(-2px);
}

.prompt-card--selected {
  border-color: #2563eb;
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.12);
}

/* 复选框绝对定位（左上角） */
.prompt-card__check {
  position: absolute;
  top: 14px;
  left: 14px;
  z-index: 2;
}

/* 头部：标题 + 收藏 */
.prompt-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 22px; /* 给复选框让位 */
}

.prompt-card__title {
  flex: 1;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0.2px;
}

/* 收藏星 */
.prompt-card__star {
  flex-shrink: 0;
  border: 0;
  background: transparent;
  padding: 3px;
  cursor: pointer;
  color: #cbd5e1;
  font-size: 16px;
  border-radius: 6px;
  transition: color 0.15s, background 0.15s, transform 0.15s;
}
.prompt-card__star:hover {
  color: #f59e0b;
  background: #fffbeb;
  transform: scale(1.12);
}
.prompt-card__star .is-fav { color: #f59e0b; }

/* 描述 */
.prompt-card__desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.55;
  color: #64748b;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 36px;
}

/* 底部：分类 + 时间在左，操作在右 */
.prompt-card__footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 4px;
}

.prompt-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.prompt-card__tag {
  font-size: 12px;
  font-weight: 500;
  color: #2563eb;
  background: #eef4ff;
  padding: 2px 10px;
  border-radius: 999px;
  max-width: 110px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.prompt-card__time {
  font-size: 12px;
  color: #94a3b8;
  white-space: nowrap;
}

/* 操作图标组 */
.prompt-card__actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.prompt-card__action {
  border: 0;
  background: transparent;
  width: 24px;
  height: 24px;
  border-radius: 5px;
  cursor: pointer;
  color: #94a3b8;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s;
}

.prompt-card__action:hover {
  background: #f1f5f9;
  color: #2563eb;
}

.prompt-card__action--danger:hover {
  color: #ef4444;
  background: #fef2f2;
}
</style>
