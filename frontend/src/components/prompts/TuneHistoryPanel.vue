<script setup lang="ts">
/**
 * 调优历史面板（PRD P05 / F10）。
 *
 * 数据来源：tuneStore.loadHistories(promptId) → GET /api/v1/prompts/:id/tune-histories。
 * 操作：
 *   - 展开查看 tuned_content 全文（默认折叠）
 *   - 应用到当前提示词（POST /api/v1/tune-histories/:id/apply）
 *   - 复制调优后内容
 */
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Clock as ClockIcon,
  ArrowDown as ApplyIcon,
  DocumentCopy as CopyIcon,
  CircleCheck as CheckIcon,
} from '@element-plus/icons-vue'

import { useTuneStore } from '@/stores/tune'
import { copyToClipboard, formatDate } from '@/utils'
import type { TuneHistory } from '@/types'

const props = defineProps<{
  promptId: number
}>()

const tuneStore = useTuneStore()

const loading = ref(false)
const loadError = ref('')

/** 当前展开的历史 ID（同时只展开一条，避免长 diff 占太多屏幕）。 */
const expandedId = ref<number | null>(null)
/** 应用中的 history.id，用于 disable 按钮显示 loading */
const applyingId = ref<number | null>(null)

const list = computed<TuneHistory[]>(() =>
  tuneStore.histories[props.promptId] || [],
)

onMounted(async () => {
  loading.value = true
  loadError.value = ''
  try {
    await tuneStore.loadHistories(props.promptId)
  } catch (err) {
    loadError.value = (err as Error).message || '加载历史失败'
  } finally {
    loading.value = false
  }
})

/** 简单的行级 diff：标记"原文有但调优无"和"调优有但原文无"的行。 */
function lineDiff(original: string, tuned: string) {
  const a = (original || '').split('\n')
  const b = (tuned || '').split('\n')
  const setA = new Set(a)
  const setB = new Set(b)
  return {
    added: b.filter((l) => !setA.has(l)),
    removed: a.filter((l) => !setB.has(l)),
  }
}

function toggleExpand(id: number) {
  expandedId.value = expandedId.value === id ? null : id
}

async function onApply(h: TuneHistory) {
  if (h.is_applied) {
    ElMessage.info('该版本已应用到当前提示词')
    return
  }
  try {
    await ElMessageBox.confirm(
      '将覆盖当前提示词的正文，并升级一档 version_no。该操作不可撤销（调优历史会保留原内容）。',
      '应用此版本到当前提示词',
      { type: 'warning', confirmButtonText: '应用', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  applyingId.value = h.id
  try {
    await tuneStore.applyHistory(h.id)
    ElMessage.success(`v${h.version_no} 已应用到当前提示词`)
    // apply 成功后该 history.is_applied 已在服务端更新；强制刷新历史缓存
    await tuneStore.loadHistories(props.promptId, true)
    // prompt 内容已在 tuneStore.applyHistory → promptStore.upsertOne 同步
  } catch (err) {
    ElMessage.error((err as Error).message || '应用失败')
  } finally {
    applyingId.value = null
  }
}

async function onCopy(h: TuneHistory) {
  const ok = await copyToClipboard(h.tuned_content)
  ElMessage[ok ? 'success' : 'error'](ok ? '已复制到剪贴板' : '复制失败')
}
</script>

<template>
  <div class="tune-history">
    <!-- 加载中 -->
    <div v-if="loading" class="tune-history__loading">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 加载失败 -->
    <el-alert
      v-else-if="loadError"
      :title="loadError"
      type="error"
      :closable="false"
      show-icon
    />

    <!-- 空态 -->
    <el-empty
      v-else-if="list.length === 0"
      :image-size="80"
      description="还没有调优历史"
    />

    <!-- 列表 -->
    <ul v-else class="tune-history__list">
      <li v-for="h in list" :key="h.id" class="tune-history__item">
        <header class="tune-history__item-head">
          <div class="tune-history__meta">
            <span class="tune-history__version">v{{ h.version_no }}</span>
            <span class="tune-history__time">
              <el-icon><ClockIcon /></el-icon>
              {{ formatDate(h.created_at, true) }}
            </span>
            <el-tag
              v-if="h.is_applied"
              size="small"
              type="success"
              effect="plain"
              class="tune-history__applied-tag"
            >
              <el-icon><CheckIcon /></el-icon>
              已应用
            </el-tag>
          </div>
          <div class="tune-history__actions">
            <el-button
              size="small"
              :icon="ApplyIcon"
              type="primary"
              :loading="applyingId === h.id"
              :disabled="h.is_applied"
              @click="onApply(h)"
            >
              应用
            </el-button>
            <el-button size="small" :icon="CopyIcon" plain @click="onCopy(h)">
              复制
            </el-button>
            <el-button size="small" plain @click="toggleExpand(h.id)">
              {{ expandedId === h.id ? '收起' : '展开' }}
            </el-button>
          </div>
        </header>

        <p class="tune-history__request">
          <span class="tune-history__label">调优诉求：</span>
          {{ h.request_text || '（无）' }}
        </p>

        <!-- 展开：行级 diff 摘要 + 全文 -->
        <div v-if="expandedId === h.id" class="tune-history__detail">
          <div class="tune-history__diff">
            <div class="tune-history__diff-col">
              <div class="tune-history__diff-title">新增行</div>
              <ul v-if="lineDiff(h.original_content, h.tuned_content).added.length" class="tune-history__diff-list">
                <li
                  v-for="(l, i) in lineDiff(h.original_content, h.tuned_content).added"
                  :key="`a-${i}`"
                  class="tune-history__diff-line tune-history__diff-line--add"
                >
                  + {{ l }}
                </li>
              </ul>
              <p v-else class="tune-history__diff-empty">（无新增）</p>
            </div>
            <div class="tune-history__diff-col">
              <div class="tune-history__diff-title">删除行</div>
              <ul v-if="lineDiff(h.original_content, h.tuned_content).removed.length" class="tune-history__diff-list">
                <li
                  v-for="(l, i) in lineDiff(h.original_content, h.tuned_content).removed"
                  :key="`r-${i}`"
                  class="tune-history__diff-line tune-history__diff-line--del"
                >
                  - {{ l }}
                </li>
              </ul>
              <p v-else class="tune-history__diff-empty">（无删除）</p>
            </div>
          </div>

          <details class="tune-history__raw">
            <summary>查看调优后全文</summary>
            <pre>{{ h.tuned_content }}</pre>
          </details>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.tune-history {
  padding: 8px 0;
}

.tune-history__loading {
  padding: 12px 0;
}

.tune-history__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tune-history__item {
  border: 1px solid #eef2f7;
  border-radius: 10px;
  padding: 12px 14px;
  background: #fafbfc;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tune-history__item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.tune-history__meta {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  font-size: 12px;
  color: #64748b;
}

.tune-history__version {
  display: inline-flex;
  align-items: center;
  padding: 1px 8px;
  background: #eef4ff;
  color: #2563eb;
  border: 1px solid #dbeafe;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  font-weight: 600;
}

.tune-history__time {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.tune-history__applied-tag {
  margin-left: 4px;
}

.tune-history__actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.tune-history__request {
  margin: 0;
  font-size: 13px;
  color: #334155;
  line-height: 1.6;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  padding: 8px 10px;
}

.tune-history__label {
  color: #64748b;
  font-weight: 500;
}

.tune-history__detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 4px;
}

.tune-history__diff {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.tune-history__diff-col {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  padding: 8px 10px;
}

.tune-history__diff-title {
  font-size: 11px;
  color: #94a3b8;
  font-weight: 600;
  margin-bottom: 4px;
  letter-spacing: 0.5px;
}

.tune-history__diff-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 180px;
  overflow: auto;
}

.tune-history__diff-line {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  padding: 2px 4px;
  border-radius: 4px;
}

.tune-history__diff-line--add {
  background: #ecfdf5;
  color: #065f46;
}

.tune-history__diff-line--del {
  background: #fef2f2;
  color: #991b1b;
}

.tune-history__diff-empty {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
}

.tune-history__raw {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 8px;
  padding: 8px 10px;
}

.tune-history__raw summary {
  cursor: pointer;
  font-size: 12px;
  color: #475569;
}

.tune-history__raw pre {
  margin: 8px 0 0;
  padding: 10px;
  background: #f8fafc;
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 360px;
  overflow: auto;
}

@media (max-width: 720px) {
  .tune-history__diff {
    grid-template-columns: 1fr;
  }
}
</style>