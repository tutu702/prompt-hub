<script setup lang="ts">
/**
 * 调优页左侧「原始提示词」面板。
 *
 * 相比通用只读卡片（PromptDetailReadonly），这里只展示正文 + 顶部/底部元信息，
 * 与右侧「调优结果」形成对称结构。代码块用 <pre> + 等宽字体，简化行号。
 */
import { computed } from 'vue'
import type { Prompt } from '@/types'

const props = defineProps<{
  prompt: Prompt
  categoryName?: string
}>()

const charCount = computed(() => props.prompt.content.length)
const lineCount = computed(() => {
  // 空字符串视为 0 行；否则按换行符 +1
  if (!props.prompt.content) return 0
  return props.prompt.content.split('\n').length
})

const formattedUpdatedAt = computed(() => {
  try {
    return new Date(props.prompt.updated_at).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return props.prompt.updated_at
  }
})
</script>

<template>
  <section class="original-pane">
    <header class="original-pane__header">
      <div class="original-pane__title-group">
        <h3 class="original-pane__title">原始提示词</h3>
        <span class="original-pane__readonly-tag">只读</span>
      </div>
      <span class="original-pane__meta">
        {{ charCount }} 字 · {{ lineCount }} 行
      </span>
    </header>

    <pre class="original-pane__body">{{ prompt.content }}</pre>

    <footer class="original-pane__footer">
      <span v-if="categoryName" class="original-pane__footer-item">
        分类: {{ categoryName }}
      </span>
      <span v-else class="original-pane__footer-item">分类: 未分类</span>
      <span class="original-pane__footer-item">
        最后调优: {{ formattedUpdatedAt }}
      </span>
    </footer>
  </section>
</template>

<style scoped>
.original-pane {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 18px 22px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 480px;
}

.original-pane__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.original-pane__title-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.original-pane__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.original-pane__readonly-tag {
  font-size: 11px;
  color: #94a3b8;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  padding: 1px 8px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.original-pane__meta {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #94a3b8;
}

.original-pane__body {
  flex: 1;
  margin: 0;
  padding: 16px 18px;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.75;
  color: #0f172a;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 60vh;
  overflow: auto;
}

.original-pane__footer {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: #94a3b8;
  padding-top: 8px;
  border-top: 1px dashed #eef2f7;
}

.original-pane__footer-item {
  color: #64748b;
}
</style>
