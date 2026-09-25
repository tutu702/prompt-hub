<script setup lang="ts">
/**
 * AI 生成模式面板（PRD P04 模式 B）。
 *
 * 行为：
 *   - 表单：用途描述 / 选择模型 / 期望风格 / 输出语言
 *   - 点击「AI 生成」调 POST /api/v1/llm/generate SSE 流式输出
 *   - 流式结果渲染到下方"预览"区（带流式光标）
 *   - 完成时按"第一行标题 + 空行 + 正文"解析为 {title, content, description}
 *   - 用户点「填入手动表单」→ emit('apply', ...) → 父页面切回手动 Tab 并把结果套入
 *
 * 校验：
 *   - 至少 1 个启用模型（store 加载后判断；缺则禁用按钮 + 顶部提示）
 *   - 用途必填、1-1000
 */
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick as MagicIcon, CircleClose as CancelIcon } from '@element-plus/icons-vue'

import { useLLMStore } from '@/stores/llm'
import { generateStream, type GenerateResult, type SSEConsumer } from '@/api/modules/tune'

const emit = defineEmits<{
  (e: 'dirty', value: boolean): void
  (e: 'apply', payload: { title?: string; content: string; description?: string; tags?: string[] }): void
}>()

const llmStore = useLLMStore()

const form = reactive({
  purpose: '',
  model: '' as number | string,
  style: '' as '' | 'professional' | 'friendly' | 'concise' | 'detailed',
  language: '' as '' | 'zh' | 'en' | 'auto',
})

const styleOptions = [
  { label: '专业', value: 'professional' },
  { label: '友好', value: 'friendly' },
  { label: '简洁', value: 'concise' },
  { label: '详细', value: 'detailed' },
]

const languageOptions = [
  { label: '中文', value: 'zh' },
  { label: '英文', value: 'en' },
  { label: '自动', value: 'auto' },
]

onMounted(async () => {
  await llmStore.loadIfNeed()
  if (llmStore.defaultConfig) {
    form.model = llmStore.defaultConfig.id
  } else if (llmStore.enabledList.length > 0) {
    form.model = llmStore.enabledList[0].id
  }
})

/* ---------- 流式状态 ---------- */
const streaming = ref(false)
const streamText = ref('')
const streamError = ref('')
const lastResult = ref<GenerateResult | null>(null)
let consumer: SSEConsumer | null = null

onBeforeUnmount(() => {
  consumer?.abort()
})

const isDirty = ref(false)
watch(
  () => ({ ...form }),
  () => {
    // 改了表单 = 脏；用户进入手动模式后再回来 dirty 状态会被父级清掉
    isDirty.value = true
    emit('dirty', true)
  },
  { deep: true },
)

const canGenerate = computed(() => {
  return (
    form.purpose.trim().length > 0 &&
    form.purpose.trim().length <= 1000 &&
    !!form.model
  )
})

const hasEnabledModel = computed(() => llmStore.enabledList.length > 0)

function onGenerate() {
  if (!canGenerate.value) {
    ElMessageBox.alert('请先填写用途描述', '提示', { type: 'warning' })
    return
  }
  // 重置
  streaming.value = true
  streamText.value = ''
  streamError.value = ''
  lastResult.value = null
  consumer?.abort()

  const modelId = typeof form.model === 'number' ? form.model : Number(form.model)
  consumer = generateStream(
    {
      purpose: form.purpose.trim(),
      style: form.style || undefined,
      language: form.language || undefined,
      model_config_id: Number.isFinite(modelId) && modelId > 0 ? modelId : undefined,
      // 没传模型时后端会 503；前端已禁用按钮，正常情况下不会触发
      mock: false,
    },
    {
      onChunk: (delta) => {
        streamText.value += delta
      },
      onDone: (result) => {
        streaming.value = false
        lastResult.value = result
        // 把流式累积文本用解析后的结构化结果覆盖（避免解析与显示不一致）
        streamText.value = (result.title ? result.title + '\n\n' : '') + result.content
      },
      onError: (msg) => {
        streaming.value = false
        streamError.value = msg
        ElMessage.error(msg)
      },
    },
  )
}

function onCancel() {
  consumer?.abort()
  consumer = null
  streaming.value = false
  ElMessage.info('已取消生成')
}

/** 把生成结果填入手动表单（emit 让父页面切 Tab 并套用）。 */
function onApply() {
  if (!lastResult.value) {
    ElMessage.warning('暂无可应用的结果')
    return
  }
  emit('apply', {
    title: lastResult.value.title,
    content: lastResult.value.content,
    description: lastResult.value.description,
    tags: [],
  })
  ElMessage.success('已填入手动表单，可继续编辑后保存')
  // 清脏：手动 Tab 的 formDirty 由父页面管理
  isDirty.value = false
  emit('dirty', false)
}

function onRegenerate() {
  onGenerate()
}
</script>

<template>
  <div class="ai-panel">
    <!-- 未配置模型的引导（仅在真的没模型时显示，避免之前误报"暂未启用"） -->
    <el-alert
      v-if="!hasEnabledModel"
      title="尚未配置任何启用的大模型"
      type="warning"
      :closable="false"
      show-icon
      class="ai-panel__alert"
    >
      <template #default>
        请先在【系统设置 → 大模型配置】中添加并启用至少一个模型。
        <el-link type="primary" :underline="false" @click="$router.push('/settings/llm')">
          前往配置 →
        </el-link>
      </template>
    </el-alert>

    <el-form label-position="top" size="default" class="ai-panel__form">
      <el-form-item label="用途描述" required>
        <el-input
          v-model="form.purpose"
          type="textarea"
          :rows="4"
          placeholder="用自然语言描述你想生成的提示词用途，如「帮我写一个 SQL 优化的提示词」"
          maxlength="1000"
          show-word-limit
          :disabled="streaming"
        />
      </el-form-item>

      <el-form-item label="选择模型" required>
        <el-select
          v-model="form.model"
          :loading="llmStore.loading"
          :disabled="llmStore.enabledList.length === 0 || streaming"
          placeholder="请选择模型"
          style="width: 100%"
        >
          <el-option
            v-for="c in llmStore.enabledList"
            :key="c.id"
            :label="`${c.name}（${c.model}）`"
            :value="c.id"
          />
        </el-select>
      </el-form-item>

      <div class="ai-panel__row2">
        <el-form-item label="期望风格">
          <el-select
            v-model="form.style"
            placeholder="不限"
            clearable
            :disabled="streaming"
            style="width: 100%"
          >
            <el-option
              v-for="o in styleOptions"
              :key="o.value"
              :label="o.label"
              :value="o.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="输出语言">
          <el-select
            v-model="form.language"
            placeholder="不限"
            clearable
            :disabled="streaming"
            style="width: 100%"
          >
            <el-option
              v-for="o in languageOptions"
              :key="o.value"
              :label="o.label"
              :value="o.value"
            />
          </el-select>
        </el-form-item>
      </div>

      <div class="ai-panel__actions">
        <el-button
          v-if="!streaming"
          type="primary"
          :icon="MagicIcon"
          :disabled="!canGenerate"
          @click="onGenerate"
        >
          AI 生成
        </el-button>
        <el-button v-else :icon="CancelIcon" plain @click="onCancel">
          取消
        </el-button>
      </div>
    </el-form>

    <!-- 流式预览 / 结果 -->
    <section v-if="streaming || streamText || streamError" class="ai-panel__result">
      <header class="ai-panel__result-header">
        <span class="ai-panel__result-title">
          {{ streaming ? '生成中…' : streamError ? '生成失败' : '生成结果' }}
        </span>
        <span v-if="!streaming && streamText && !streamError" class="ai-panel__result-meta">
          {{ streamText.length }} 字
        </span>
      </header>

      <el-alert
        v-if="streamError"
        :title="streamError"
        type="error"
        :closable="false"
        show-icon
      />
      <pre
        v-else
        class="ai-panel__result-text"
        :class="{ 'ai-panel__result-text--streaming': streaming }"
      >{{ streamText }}</pre>

      <footer v-if="!streaming && lastResult" class="ai-panel__result-actions">
        <el-button plain @click="onRegenerate">重新生成</el-button>
        <el-button type="primary" @click="onApply">填入手动表单</el-button>
      </footer>
    </section>
  </div>
</template>

<style scoped>
.ai-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ai-panel__alert :deep(.el-alert__content) {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ai-panel__form {
  background: #ffffff;
  padding: 20px 24px;
  border-radius: 12px;
  border: 1px solid #eef2f7;
}

.ai-panel__row2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.ai-panel__actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}

/* 结果预览 */
.ai-panel__result {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-panel__result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.ai-panel__result-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.ai-panel__result-meta {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #94a3b8;
}

.ai-panel__result-text {
  margin: 0;
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.75;
  color: #0f172a;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 360px;
  overflow: auto;
}

.ai-panel__result-text--streaming::after {
  content: '▍';
  display: inline-block;
  margin-left: 2px;
  color: #1677ff;
  animation: ai-caret 1s steps(2) infinite;
}

@keyframes ai-caret {
  0%, 50% { opacity: 1; }
  50.01%, 100% { opacity: 0; }
}

.ai-panel__result-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px dashed #eef2f7;
}

@media (max-width: 640px) {
  .ai-panel__row2 {
    grid-template-columns: 1fr;
  }
}
</style>