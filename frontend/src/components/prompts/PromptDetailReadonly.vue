<script setup lang="ts">
/**
 * 提示词只读详情卡片（PRD P05）。
 *
 * 字段：分类 / 描述 / 标签 / 内容（带行号渲染）。
 * 内容行号实现：把 content 按 \n 切分到 <ol><li> 里，左右两栏 flex 布局；
 * 这种写法比 <pre> + CSS counter 更可靠（CSS counter 无法对原始多行文本逐行计数）。
 */
import { computed } from 'vue'
import {
  Calendar as CalendarIcon,
  Clock as ClockIcon,
  DocumentCopy as CopyIcon,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { Prompt } from '@/types'
import { copyToClipboard, formatRelativeTime } from '@/utils'

const props = defineProps<{
  prompt: Prompt
  categoryName?: string
}>()

const emit = defineEmits<{
  (e: 'copy'): void
  (e: 'toggle-favorite'): void
}>()

const lines = computed(() => props.prompt.content.split('\n'))
const hasDescription = computed(() => !!props.prompt.description?.trim())
const hasTags = computed(() => Array.isArray(props.prompt.tags) && props.prompt.tags.length > 0)

function fullTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

async function onCopy() {
  const ok = await copyToClipboard(props.prompt.content)
  if (ok) {
    ElMessage.success('已复制到剪贴板')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
  emit('copy')
}
</script>

<template>
  <section class="readonly">
    <!-- 顶部元信息条 -->
    <header class="readonly__meta">
      <el-tag v-if="categoryName" type="primary" plain effect="plain" round>
        {{ categoryName }}
      </el-tag>
      <span v-else class="readonly__meta-tag readonly__meta-tag--empty">未分类</span>

      <span class="readonly__meta-version">v{{ prompt.version_no }}</span>

      <span class="readonly__meta-dot" />

      <span class="readonly__meta-time" :title="fullTime(prompt.updated_at)">
        <el-icon><ClockIcon /></el-icon>
        {{ formatRelativeTime(prompt.updated_at) }}
      </span>

      <span class="readonly__meta-dot" />

      <span class="readonly__meta-time" :title="fullTime(prompt.created_at)">
        <el-icon><CalendarIcon /></el-icon>
        创建于 {{ formatRelativeTime(prompt.created_at) }}
      </span>
    </header>

    <!-- 描述 -->
    <section v-if="hasDescription" class="readonly__desc">
      <h4 class="readonly__label">描述</h4>
      <p class="readonly__desc-text">{{ prompt.description }}</p>
    </section>

    <!-- 标签 -->
    <section v-if="hasTags" class="readonly__tags">
      <h4 class="readonly__label">标签</h4>
      <div class="readonly__tags-list">
        <el-tag
          v-for="t in prompt.tags"
          :key="t"
          type="info"
          plain
          effect="plain"
        >
          {{ t }}
        </el-tag>
      </div>
    </section>

    <!-- 正文 + 行号 + 复制 -->
    <section class="readonly__content-section">
      <header class="readonly__content-header">
        <h4 class="readonly__label">内容</h4>
        <el-button text :icon="CopyIcon" @click="onCopy">复制全文</el-button>
      </header>
      <ol class="readonly-content">
        <li v-for="(line, i) in lines" :key="i">
          <span class="readonly-content__ln">{{ i + 1 }}</span>
          <span class="readonly-content__txt">{{ line || ' ' }}</span>
        </li>
      </ol>
    </section>
  </section>
</template>

<style scoped>
.readonly {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 元信息条 */
.readonly__meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: #475569;
  flex-wrap: wrap;
}

.readonly__meta-tag--empty {
  color: #94a3b8;
  background: #f1f5f9;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.readonly__meta-version {
  font-size: 12px;
  color: #94a3b8;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  padding: 2px 8px;
  background: #f8fafc;
  border-radius: 6px;
}

.readonly__meta-time {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #94a3b8;
}

.readonly__meta-dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #cbd5e1;
}

/* 描述 / 标签 / 内容 区块 */
.readonly__label {
  margin: 0 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  letter-spacing: 0.2px;
}

.readonly__desc-text {
  margin: 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.readonly__tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

/* 内容块 */
.readonly__content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.readonly__content-header .readonly__label {
  margin: 0;
}

.readonly-content {
  list-style: none;
  margin: 0;
  padding: 14px 14px 14px 0;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  max-height: 60vh;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Courier New', monospace;
  line-height: 1.7;
  color: #0f172a;
}

.readonly-content li {
  display: flex;
  align-items: flex-start;
}

.readonly-content__ln {
  flex-shrink: 0;
  width: 48px;
  padding: 0 12px 0 14px;
  text-align: right;
  color: #94a3b8;
  user-select: none;
  border-right: 1px solid #e2e8f0;
  margin-right: 12px;
  font-size: 12px;
  line-height: 1.7;
}

.readonly-content__txt {
  white-space: pre-wrap;
  word-break: break-word;
  flex: 1;
  padding-right: 12px;
}
</style>