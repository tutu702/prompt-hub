<script setup lang="ts">
/**
 * 提示词 AI 调优页（PRD P07）。
 *
 * 布局（与设计图保持一致）：
 *   - 顶部一行：返回按钮 / 标题 / 副标题 ... 右侧「并排对比 / Diff 高亮」开关
 *   - 中部两栏：左 PromptOriginalPanel（只读正文） + 右 TuneResultPanel（调优结果）
 *   - 底部一行：TuneRequestForm（诉求输入 + 一键建议 + 模型 + 开始调优）
 *
 * 行为：通过 POST /api/v1/llm/tune（SSE）拉取真实调优；记录 elapsedMs 在完成后展示；
 * 完成后可保存为新版本 / 应用到当前 / 重新生成 / 复制结果。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft as ArrowLeftIcon,
  Grid as SplitIcon,
  Document as DiffIcon,
} from '@element-plus/icons-vue'

import PromptOriginalPanel from '@/components/tune/PromptOriginalPanel.vue'
import TuneRequestForm from '@/components/tune/TuneRequestForm.vue'
import TuneResultPanel from '@/components/tune/TuneResultPanel.vue'
import { tuneApi } from '@/api'
import type { SSEConsumer } from '@/utils/sse'
import { promptApi } from '@/api'
import { useCategoryStore } from '@/stores/category'
import { usePromptStore } from '@/stores/prompt'
import { useTuneStore } from '@/stores/tune'
import { copyToClipboard } from '@/utils'
import type { Prompt } from '@/types'

const route = useRoute()
const router = useRouter()
const categoryStore = useCategoryStore()
const promptStore = usePromptStore()
const tuneStore = useTuneStore()

const promptId = computed(() => Number(route.params.id))
const prompt = ref<Prompt | null>(null)
const loading = ref(false)
const notFound = ref(false)

let streamConsumer: SSEConsumer | null = null
let streamStartTs = 0
const elapsedMs = ref<number | null>(null)

const categoryName = computed(() =>
  prompt.value ? categoryStore.getName(prompt.value.category_id) : '',
)

/** 上次调优时间（updated_at）。 */
const lastTunedLabel = computed(() => {
  if (!prompt.value) return ''
  try {
    return new Date(prompt.value.updated_at).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return prompt.value.updated_at
  }
})

/** 调优面板显示的模型标签：调优成功后从 history.model_config_id 反查；未发起时显示「—」。 */
const modelLabel = ref<string>('—')

onMounted(async () => {
  await categoryStore.loadIfNeed()
  await fetchDetail()
  // 进入页面：清空旧流式状态 + 预拉历史
  tuneStore.reset()
  elapsedMs.value = null
  try {
    await tuneStore.loadHistories(promptId.value)
  } catch {
    /* 静默 */
  }
})

onBeforeUnmount(() => {
  streamConsumer?.abort()
})

async function fetchDetail() {
  loading.value = true
  notFound.value = false
  try {
    const p = await promptApi.getPrompt(promptId.value)
    prompt.value = p
    promptStore.upsertOne(p)
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

/* ---------- 调优流程 ---------- */

function onTuneStart(payload: {
  requestText: string
  useContext: boolean
  modelConfigId?: number
}) {
  if (!prompt.value) return

  // 1. 清空旧结果 + 取消旧流
  tuneStore.resultText = ''
  tuneStore.error = ''
  elapsedMs.value = null
  streamConsumer?.abort()

  // 2. 把诉求暂存（供保存/重新生成复用）
  tuneStore.requestText = payload.requestText
  tuneStore.useContext = payload.useContext
  tuneStore.streaming = true
  streamStartTs = performance.now()

  // 3. 打开 SSE 连接 —— 始终走真实 LLM；后端在缺模型 / 未启用时会返回明确错误。
  // （之前 `mock: !payload.modelConfigId` 会把"没选模型"静默降级为 mock，
  //  导致用户看到的是 mock 拼接出来的"（已针对「...」优化）"假结果，误以为没调 LLM。）
  streamConsumer = tuneApi.tuneStream(
    {
      prompt_id: prompt.value.id,
      request_text: payload.requestText,
      use_context: payload.useContext,
      model_config_id: payload.modelConfigId,
      mock: false,
    },
    {
      onChunk: (delta) => {
        tuneStore.resultText += delta
      },
      onDone: (history) => {
        tuneStore.streaming = false
        elapsedMs.value = performance.now() - streamStartTs
        modelLabel.value = history.model_config_id ? `配置 #${history.model_config_id}` : '—'
        // 把刚写入的历史塞进 store 缓存（无需再 GET 一次）
        const arr = tuneStore.histories[prompt.value!.id] || []
        if (!arr.find((h) => h.id === history.id)) {
          tuneStore.histories[prompt.value!.id] = [history, ...arr]
        }
      },
      onError: (msg) => {
        tuneStore.streaming = false
        tuneStore.error = msg
        ElMessage.error(msg)
      },
    },
  )
}

function onTuneCancel() {
  streamConsumer?.abort()
  streamConsumer = null
  tuneStore.streaming = false
  ElMessage.info('已取消调优')
}

async function onCopy() {
  const ok = await copyToClipboard(tuneStore.resultText)
  ElMessage[ok ? 'success' : 'error'](
    ok ? '已复制到剪贴板' : '复制失败，请手动复制',
  )
}

async function onSaveVersion() {
  if (!prompt.value || !tuneStore.resultText) return
  try {
    const created = await tuneStore.saveCurrentResult(prompt.value.id, {
      request_text: tuneStore.requestText,
      original_content: prompt.value.content,
      tuned_content: tuneStore.resultText,
    })
    ElMessage.success(`已保存为 v${created.version_no}`)
  } catch (err) {
    ElMessage.error((err as Error).message || '保存失败')
  }
}

async function onApply() {
  if (!prompt.value || !tuneStore.resultText) return
  try {
    await ElMessageBox.confirm(
      `将覆盖当前提示词【${prompt.value.title}】的正文，并把 version 升级一档。该操作不可撤销（但调优历史会保留原内容）。`,
      '应用到当前提示词',
      {
        type: 'warning',
        confirmButtonText: '应用',
        cancelButtonText: '取消',
      },
    )
  } catch {
    return
  }
  try {
    // 先保存为历史（拿到 history.id），再调 apply
    const history = await tuneStore.saveCurrentResult(prompt.value.id, {
      request_text: tuneStore.requestText,
      original_content: prompt.value.content,
      tuned_content: tuneStore.resultText,
    })
    await tuneStore.applyHistory(history.id)
    ElMessage.success('已应用到当前提示词')
    await fetchDetail()
  } catch (err) {
    ElMessage.error((err as Error).message || '应用失败')
  }
}

function onRegenerate() {
  if (!tuneStore.requestText.trim()) {
    ElMessage.warning('请先填写调优诉求')
    return
  }
  onTuneStart({
    requestText: tuneStore.requestText,
    useContext: tuneStore.useContext,
  })
}

function onBack() {
  router.push(`/prompts/${promptId.value}`)
}
</script>

<template>
  <div class="prompt-tune">
    <header class="prompt-tune__header">
      <div class="prompt-tune__header-left">
        <button
          type="button"
          class="prompt-tune__back"
          title="返回详情"
          @click="onBack"
        >
          <el-icon><ArrowLeftIcon /></el-icon>
        </button>
        <div v-if="prompt" class="prompt-tune__titles">
          <h2 class="prompt-tune__title">AI 调优 · {{ prompt.title }}</h2>
          <p class="prompt-tune__sub">
            <span>分类：{{ categoryName || '未分类' }}</span>
            <span class="prompt-tune__sub-dot">·</span>
            <span>上次调优 {{ lastTunedLabel }}</span>
          </p>
        </div>
      </div>

      <div class="prompt-tune__header-right">
        <el-button
          type="primary"
          plain
          :icon="SplitIcon"
          class="prompt-tune__toggle"
        >
          并排对比
        </el-button>
        <el-button :icon="DiffIcon" class="prompt-tune__toggle">
          Diff 高亮
        </el-button>
      </div>
    </header>

    <el-empty v-if="notFound" description="提示词不存在或已被删除" />
    <div v-else-if="loading" class="prompt-tune__loading">
      <el-skeleton :rows="6" animated />
    </div>

    <template v-else-if="prompt">
      <section class="prompt-tune__panels">
        <div class="prompt-tune__pane">
          <PromptOriginalPanel
            :prompt="prompt"
            :category-name="categoryName"
          />
        </div>

        <div class="prompt-tune__pane">
          <TuneResultPanel
            :streaming="tuneStore.streaming"
            :error="tuneStore.error"
            :text="tuneStore.resultText"
            :model-label="modelLabel"
            :elapsed-ms="elapsedMs ?? undefined"
            :can-apply="!tuneStore.streaming && !!tuneStore.resultText"
            :can-save="!tuneStore.streaming && !!tuneStore.resultText"
            @copy="onCopy"
            @save-version="onSaveVersion"
            @apply="onApply"
            @regenerate="onRegenerate"
          />
        </div>
      </section>

      <footer class="prompt-tune__footer">
        <TuneRequestForm
          :original-content="prompt.content"
          :loading="tuneStore.streaming"
          @submit="onTuneStart"
          @cancel="onTuneCancel"
        />
      </footer>
    </template>
  </div>
</template>

<style scoped>
.prompt-tune {
  max-width: 1440px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* 顶部 header */
.prompt-tune__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.prompt-tune__header-left {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.prompt-tune__back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: 1px solid #eef2f7;
  background: #ffffff;
  color: #475569;
  border-radius: 10px;
  cursor: pointer;
  font-size: 16px;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.prompt-tune__back:hover {
  background: #f1f5f9;
  color: #2563eb;
  border-color: #dbeafe;
}

.prompt-tune__titles {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.prompt-tune__title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.prompt-tune__sub {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.prompt-tune__sub-dot {
  color: #cbd5e1;
}

.prompt-tune__header-right {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.prompt-tune__toggle {
  border-radius: 8px;
}

/* 加载 / 空态 */
.prompt-tune__loading {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 24px;
}

/* 双栏 */
.prompt-tune__panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  align-items: stretch;
}

.prompt-tune__pane {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.prompt-tune__footer {
  display: flex;
  flex-direction: column;
}

@media (max-width: 900px) {
  .prompt-tune__panels {
    grid-template-columns: 1fr;
  }
  .prompt-tune__header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
