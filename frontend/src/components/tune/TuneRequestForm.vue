<script setup lang="ts">
/**
 * AI 调优诉求输入区（P07 重构版）。
 *
 * 布局（与设计图保持一致）：
 *   - 左侧：调优诉求 textarea + 下方一行建议「试试 ...」
 *   - 右侧：模型下拉 + 「开始调优」主按钮（占满右栏宽度）
 *
 * 字段：
 *   - 调优诉求 textarea（必填 1-2000）
 *   - 模型下拉：本期 disabled 占位（P08 接真实 LLM 时启用）
 *   - 行内按钮：开始调优 / 取消（流式中切换）
 */
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  MagicStick as MagicIcon,
  CircleClose as CancelIcon,
  ChatLineRound as SuggestionIcon,
} from '@element-plus/icons-vue'
import { useLLMStore } from '@/stores/llm'

defineProps<{
  originalContent: string
  loading: boolean
}>()

const emit = defineEmits<{
  (
    e: 'submit',
    payload: { requestText: string; useContext: boolean; modelConfigId?: number },
  ): void
  (e: 'cancel'): void
}>()

const requestText = ref<string>('')
const useContext = ref<boolean>(false)

const llmStore = useLLMStore()
const modelConfigId = ref<number | null>(null)

/** 当前是否选好了一个可用模型（用于控制提交按钮 + 顶部告警）。 */
const hasUsableModel = computed(
  () => modelConfigId.value !== null && llmStore.enabledList.some((c) => c.id === modelConfigId.value),
)

onMounted(async () => {
  await llmStore.loadIfNeed()
  if (llmStore.defaultConfig) {
    modelConfigId.value = llmStore.defaultConfig.id
  } else if (llmStore.enabledList.length > 0) {
    modelConfigId.value = llmStore.enabledList[0].id
  } else {
    ElMessage.warning('请先去【设置 → 大模型配置】启用至少一个模型')
  }
})

/** 调优诉求内置的几个一键建议。 */
const suggestions = [
  '按重要性排序，并给出可执行的修改建议',
  '让输出更结构化、增加示例、语气更克制',
  '精简到 200 字以内，去除冗余',
]

function applySuggestion(text: string) {
  requestText.value = text
}

function onSubmit() {
  const req = requestText.value.trim()
  if (!req || !modelConfigId.value) return
  emit('submit', {
    requestText: req,
    useContext: useContext.value,
    modelConfigId: modelConfigId.value,
  })
}

function onCancel() {
  emit('cancel')
}
</script>

<template>
  <section class="tune-form">
    <!-- 未配置模型的引导：与 AIGeneratePanel 风格一致，给出明确跳转 -->
    <el-alert
      v-if="llmStore.loaded && !hasUsableModel"
      title="尚未配置任何启用的大模型"
      type="warning"
      :closable="false"
      show-icon
      class="tune-form__alert"
    >
      <template #default>
        请先在【系统设置 → 大模型配置】中添加并启用至少一个模型，否则 AI 调优无法调用大模型。
        <el-link type="primary" :underline="false" @click="$router.push('/settings/llm')">
          前往配置 →
        </el-link>
      </template>
    </el-alert>

    <div class="tune-form__main">
      <!-- 左侧：诉求 textarea + 一键建议 -->
      <div class="tune-form__request">
        <el-input
          v-model="requestText"
          type="textarea"
          :rows="5"
          placeholder="描述你的调优诉求，例如：让输出更结构化、增加示例、语气更克制…"
          maxlength="2000"
          show-word-limit
          :disabled="loading"
        />
        <div class="tune-form__suggestion">
          <el-icon class="tune-form__suggestion-icon">
            <SuggestionIcon />
          </el-icon>
          <span class="tune-form__suggestion-label">试试</span>
          <div class="tune-form__suggestion-chips">
            <button
              v-for="(s, idx) in suggestions"
              :key="idx"
              type="button"
              class="tune-form__chip"
              :disabled="loading"
              @click="applySuggestion(s)"
            >
              「{{ s }}」
            </button>
          </div>
        </div>
      </div>

      <!-- 右侧：模型 + 主操作 -->
      <div class="tune-form__actions">
        <el-select
          v-model="modelConfigId"
          :loading="llmStore.loading"
          :disabled="llmStore.enabledList.length === 0"
          placeholder="请选择模型"
          class="tune-form__model-select"
          size="default"
        >
          <template #prefix>
            <el-icon class="tune-form__model-icon"><MagicIcon /></el-icon>
          </template>
          <el-option
            v-for="c in llmStore.enabledList"
            :key="c.id"
            :label="`${c.name}（${c.model}）`"
            :value="c.id"
          />
        </el-select>

        <el-button
          v-if="!loading"
          type="primary"
          size="large"
          class="tune-form__submit"
          :icon="MagicIcon"
          :disabled="!requestText.trim() || !hasUsableModel"
          @click="onSubmit"
        >
          开始调优
        </el-button>
        <el-button v-else plain size="large" class="tune-form__submit" :icon="CancelIcon" @click="onCancel">
          取消
        </el-button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.tune-form {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 18px 22px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tune-form__main {
  display: grid;
  grid-template-columns: 1fr 240px;
  gap: 20px;
  align-items: stretch;
}

.tune-form__request {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

/* 一键建议行 */
.tune-form__suggestion {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
  flex-wrap: wrap;
}

.tune-form__suggestion-icon {
  color: #6366f1;
  font-size: 14px;
  margin-top: 1px;
  flex-shrink: 0;
}

.tune-form__suggestion-label {
  color: #94a3b8;
  flex-shrink: 0;
  margin-top: 1px;
}

.tune-form__suggestion-chips {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tune-form__chip {
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #475569;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.tune-form__chip:hover:not(:disabled) {
  border-color: #93c5fd;
  color: #2563eb;
  background: #eef4ff;
}

.tune-form__chip:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* 右侧操作栏 */
.tune-form__actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tune-form__model-select {
  width: 100%;
}

.tune-form__model-select :deep(.el-input__wrapper) {
  background: #f8fafc;
  border-radius: 8px;
}

.tune-form__alert {
  margin-bottom: 4px;
}

.tune-form__alert :deep(.el-alert__content) {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.tune-form__model-icon {
  color: #6366f1;
  font-size: 14px;
}

.tune-form__submit {
  width: 100%;
  height: 44px;
  font-weight: 500;
  letter-spacing: 0.5px;
  border-radius: 10px;
}

.tune-form__submit.tune-form__submit--primary {
  box-shadow: 0 6px 14px rgba(37, 99, 235, 0.22);
}

@media (max-width: 900px) {
  .tune-form__main {
    grid-template-columns: 1fr;
  }
}
</style>
