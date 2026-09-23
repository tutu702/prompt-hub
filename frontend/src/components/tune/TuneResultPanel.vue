<script setup lang="ts">
/**
 * AI 调优结果展示面板（P07）。
 *
 * 与左侧「原始提示词」对称：顶部标题 + 模型标签 + 字数行数；
 * 流式/完成态展示文本；底部用时 + 4 个操作按钮。
 *
 * 完成态操作（顺序：复制结果 / 保存为新版本 / 应用到当前 / 重新生成）：
 *   - 复制结果（emit copy）
 *   - 保存为新版本（emit save-version）
 *   - 应用到当前提示词（emit apply，二次确认在父级）
 *   - 重新生成（emit regenerate）
 */
import { computed } from 'vue'
import {
  DocumentCopy as CopyIcon,
  Refresh as RefreshIcon,
  MagicStick as ApplyIcon,
  Histogram as SaveIcon,
} from '@element-plus/icons-vue'

const props = defineProps<{
  streaming: boolean
  error: string
  text: string
  modelLabel?: string
  elapsedMs?: number
  canApply: boolean
  canSave: boolean
}>()

const emit = defineEmits<{
  (e: 'copy'): void
  (e: 'save-version'): void
  (e: 'apply'): void
  (e: 'regenerate'): void
}>()

const isEmpty = computed(() => !props.text && !props.streaming && !props.error)
const isError = computed(() => !!props.error && !props.streaming)

const charCount = computed(() => props.text.length)
const lineCount = computed(() => {
  if (!props.text) return 0
  return props.text.split('\n').length
})

const elapsedLabel = computed(() => {
  if (props.elapsedMs == null) return ''
  const s = props.elapsedMs / 1000
  // 1 位小数：1.8s / 0.3s
  return `用时 ${s.toFixed(1)}s`
})
</script>

<template>
  <section class="tune-result">
    <header class="tune-result__header">
      <div class="tune-result__title-group">
        <h3 class="tune-result__title">调优结果</h3>
        <span v-if="modelLabel" class="tune-result__model-tag">
          {{ modelLabel }}
        </span>
        <el-tag v-else-if="streaming" type="info" size="small" effect="plain">
          生成中…
        </el-tag>
      </div>
      <span v-if="!isEmpty && !isError" class="tune-result__meta">
        {{ charCount }} 字 · {{ lineCount }} 行
      </span>
    </header>

    <!-- 空状态 -->
    <div v-if="isEmpty" class="tune-result__body tune-result__body--empty">
      <el-empty description="调优结果将在这里展示" :image-size="80" />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="isError" class="tune-result__body tune-result__body--error">
      <el-alert
        :title="error"
        type="error"
        :closable="false"
        show-icon
      />
      <el-button type="primary" plain :icon="RefreshIcon" @click="emit('regenerate')">
        重试
      </el-button>
    </div>

    <!-- 流式 / 完成 -->
    <div v-else class="tune-result__body">
      <pre
        class="tune-result__text"
        :class="{ 'tune-result__text--streaming': streaming }"
      >{{ text }}</pre>

      <footer v-if="!streaming" class="tune-result__actions">
        <div class="tune-result__actions-left">
          <span v-if="elapsedLabel" class="tune-result__elapsed">{{ elapsedLabel }}</span>
        </div>
        <div class="tune-result__actions-right">
          <el-button
            :icon="CopyIcon"
            plain
            @click="emit('copy')"
          >
            复制结果
          </el-button>
          <el-button
            :icon="SaveIcon"
            plain
            :disabled="!canSave"
            @click="emit('save-version')"
          >
            保存为新版本
          </el-button>
          <el-button
            type="primary"
            :icon="ApplyIcon"
            :disabled="!canApply"
            @click="emit('apply')"
          >
            应用到当前
          </el-button>
          <el-button
            :icon="RefreshIcon"
            @click="emit('regenerate')"
          >
            重新生成
          </el-button>
        </div>
      </footer>
    </div>
  </section>
</template>

<style scoped>
.tune-result {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 18px 22px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 480px;
}

.tune-result__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.tune-result__title-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.tune-result__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.tune-result__model-tag {
  font-size: 11px;
  color: #2563eb;
  background: #eef4ff;
  border: 1px solid #dbeafe;
  padding: 1px 8px;
  border-radius: 4px;
  font-weight: 500;
  letter-spacing: 0.3px;
}

.tune-result__meta {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #94a3b8;
}

.tune-result__body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
}

.tune-result__body--empty {
  align-items: center;
  justify-content: center;
}

.tune-result__body--error {
  align-items: stretch;
}

.tune-result__text {
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

.tune-result__text--streaming::after {
  content: '▍';
  display: inline-block;
  margin-left: 2px;
  color: #1677ff;
  animation: tune-caret 1s steps(2) infinite;
}

@keyframes tune-caret {
  0%, 50% { opacity: 1; }
  50.01%, 100% { opacity: 0; }
}

.tune-result__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 10px;
  border-top: 1px dashed #eef2f7;
  flex-wrap: wrap;
}

.tune-result__actions-left {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #94a3b8;
}

.tune-result__elapsed {
  letter-spacing: 0.3px;
}

.tune-result__actions-right {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>
