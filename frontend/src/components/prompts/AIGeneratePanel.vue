<script setup lang="ts">
/**
 * AI 生成模式占位面板（PRD P04 模式 B）。
 *
 * 本期不接入 LLM，仅渲染 UI 并在「AI 生成」按钮被点击时弹出引导 toast，
 * 告知用户去系统设置 → 大模型配置 启用至少一个模型（P08 之后接入）。
 *
 * 面板不持有任何生成结果状态，避免误导；fill-out 由父页面在 apply 时切换到手动 Tab。
 */
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { useLLMStore } from '@/stores/llm'

const emit = defineEmits<{
  (e: 'dirty', value: boolean): void
  (e: 'apply', payload: { title?: string; content: string; description?: string; tags?: string[] }): void
}>()

const llmStore = useLLMStore()

const form = reactive({
  purpose: '',
  model: '' as number | string,
  style: '',
  language: '',
})

onMounted(async () => {
  await llmStore.loadIfNeed()
  if (llmStore.defaultConfig) {
    form.model = llmStore.defaultConfig.id
  } else if (llmStore.enabledList.length > 0) {
    form.model = llmStore.enabledList[0].id
  }
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

const isDirty = ref(false)
watch(
  form,
  () => {
    isDirty.value = true
    emit('dirty', true)
  },
  { deep: true },
)

const canGenerate = computed(() => form.purpose.trim().length > 0)

function onGenerate() {
  if (!canGenerate.value) {
    ElMessageBox.alert('请先填写用途描述', '提示', { type: 'warning' })
    return
  }
  ElMessageBox.alert(
    'AI 生成功能需要先在【系统设置 → 大模型配置】中添加至少一个启用的模型配置。配置完成后即可一键生成。',
    'AI 生成暂未启用',
    { type: 'info', confirmButtonText: '我知道了' },
  )
  // 故意不更新 isDirty / emit dirty — 避免用户误以为已生成
  // 故意不调用 emit('apply') —— 没有真实生成结果
}
</script>

<template>
  <div class="ai-panel">
    <el-alert
      title="AI 生成（占位）"
      type="info"
      :closable="false"
      show-icon
      description="后端大模型配置与生成接口尚未上线，本期仅提供 UI 占位。提交 AI 表单不会写入数据库。"
    />

    <el-form label-position="top" size="default" class="ai-panel__form">
      <el-form-item label="用途描述" required>
        <el-input
          v-model="form.purpose"
          type="textarea"
          :rows="4"
          placeholder="用自然语言描述你想生成的提示词用途，如「帮我写一个 SQL 优化的提示词」"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="选择模型">
        <el-select
          v-model="form.model"
          :loading="llmStore.loading"
          :disabled="llmStore.enabledList.length === 0"
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
          <el-select v-model="form.style" placeholder="不限" clearable style="width: 100%">
            <el-option
              v-for="o in styleOptions"
              :key="o.value"
              :label="o.label"
              :value="o.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="输出语言">
          <el-select v-model="form.language" placeholder="不限" clearable style="width: 100%">
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
        <el-button type="primary" :disabled="!canGenerate" @click="onGenerate">
          AI 生成
        </el-button>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.ai-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
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

@media (max-width: 640px) {
  .ai-panel__row2 {
    grid-template-columns: 1fr;
  }
}
</style>