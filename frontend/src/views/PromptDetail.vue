<script setup lang="ts">
/**
 * 提示词详情/编辑页（PRD P05）。
 *
 * 顶部：返回链接 + 标题 + 收藏星标。
 * 中部：模式切换 —— 只读详情卡 vs 编辑模式（复用 PromptForm）。
 * 底部操作区：复制 / 编辑切换 / AI 调优（占位 toast） / 收藏 / 删除。
 * 调优历史：折叠面板，本期为占位 empty state。
 *
 * 编辑态离开守卫：脏态 → ElMessageBox 二次确认。
 * 编辑保存后留在当前页（不跳走），同步列表局部刷新。
 */
import { computed, onMounted, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft as ArrowLeftIcon,
  Star as StarIcon,
  StarFilled as StarFilledIcon,
  DocumentCopy as CopyIcon,
  Edit as EditIcon,
  MagicStick as TuneIcon,
  Delete as DeleteIcon,
  Histogram as HistogramIcon,
} from '@element-plus/icons-vue'

import PromptDetailReadonly from '@/components/prompts/PromptDetailReadonly.vue'
import PromptForm from '@/components/prompts/PromptForm.vue'
import TuneHistoryPanel from '@/components/prompts/TuneHistoryPanel.vue'
import { promptApi } from '@/api'
import { usePromptStore } from '@/stores/prompt'
import { useCategoryStore } from '@/stores/category'
import type { Prompt as PromptT, UpdatePromptPayload } from '@/types'
import { copyToClipboard } from '@/utils'

const route = useRoute()
const router = useRouter()
const promptStore = usePromptStore()
const categoryStore = useCategoryStore()

const prompt = ref<PromptT | null>(null)
const loading = ref(false)
const notFound = ref(false)
const mode = ref<'view' | 'edit'>('view')
const formDirty = ref(false)
const submitting = ref(false)
const formRef = ref<InstanceType<typeof PromptForm> | null>(null)

const id = computed(() => Number(route.params.id))
const categoryName = computed(() =>
  prompt.value ? categoryStore.getName(prompt.value.category_id) : '',
)
const historyExpanded = ref<string[]>([])

onMounted(async () => {
  await categoryStore.loadIfNeed()
  await fetchDetail()
  // 可选增强：URL 带 ?mode=edit 时直接进入编辑态
  if (route.query.mode === 'edit') {
    mode.value = 'edit'
  }
})

async function fetchDetail() {
  loading.value = true
  notFound.value = false
  try {
    const p = await promptApi.getPrompt(id.value)
    prompt.value = p
    promptStore.upsertOne(p)
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

/* ---------- 只读操作 ---------- */

async function handleCopy() {
  if (!prompt.value) return
  const ok = await copyToClipboard(prompt.value.content)
  ElMessage[ok ? 'success' : 'error'](ok ? '已复制到剪贴板' : '复制失败，请手动复制')
}

async function handleToggleFavorite() {
  if (!prompt.value) return
  const target = !prompt.value.is_favorite
  // 乐观更新（详情卡片 + 列表同步）
  const original = { ...prompt.value }
  prompt.value = { ...prompt.value, is_favorite: target }
  try {
    await promptApi.toggleFavorite(prompt.value.id, { is_favorite: target })
    promptStore.upsertOne({ ...prompt.value })
  } catch {
    prompt.value = original
    ElMessage.error('操作失败，请稍后再试')
  }
}

function handleEdit() {
  if (!prompt.value) return
  mode.value = 'edit'
  formDirty.value = false
}

function handleCancelEdit() {
  if (!formDirty.value) {
    mode.value = 'view'
    return
  }
  ElMessageBox.confirm('当前修改未保存，确定放弃吗？', '提示', {
    type: 'warning',
    confirmButtonText: '放弃修改',
    cancelButtonText: '留在页面',
  })
    .then(() => {
      mode.value = 'view'
      formDirty.value = false
    })
    .catch(() => {
      /* 留在编辑 */
    })
}

async function handleDelete() {
  if (!prompt.value) return
  try {
    await ElMessageBox.confirm(
      `确定删除【${prompt.value.title}】吗？删除后可在 30 天内恢复`,
      '提示',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await promptApi.deletePrompt(prompt.value.id)
    promptStore.removeFromList([prompt.value.id])
    ElMessage.success('已删除')
    router.push('/prompts')
  } catch {
    ElMessage.error('删除失败，请稍后再试')
  }
}

function handleTune() {
  ElMessage.info('AI 调优即将开放')
}

/* ---------- 编辑模式 ---------- */

const formInitial = computed<Partial<{
  title: string
  description: string
  content: string
  category_id: number | null
  tags: string[]
}> | undefined>(() =>
  prompt.value
    ? {
        title: prompt.value.title,
        description: prompt.value.description,
        content: prompt.value.content,
        category_id: prompt.value.category_id ?? null,
        tags: prompt.value.tags,
      }
    : undefined,
)

async function handleSubmit(payload: UpdatePromptPayload) {
  if (!prompt.value) return
  submitting.value = true
  try {
    const updated = await promptStore.updatePrompt(prompt.value.id, payload)
    prompt.value = updated
    ElMessage.success('已更新')
    // 先清脏再切回只读，避免触发离开守卫
    formDirty.value = false
    mode.value = 'view'
  } catch (err) {
    ElMessage.error((err as Error)?.message || '更新失败，请稍后再试')
  } finally {
    submitting.value = false
  }
}

function onFormDirty(v: boolean) {
  formDirty.value = v
}

function onSaveClick() {
  formRef.value?.submit()
}

/* ---------- 离开守卫 ---------- */

async function confirmLeave(): Promise<boolean> {
  try {
    await ElMessageBox.confirm(
      '当前修改未保存，确定离开吗？',
      '提示',
      {
        type: 'warning',
        confirmButtonText: '离开',
        cancelButtonText: '留在页面',
      },
    )
    return true
  } catch {
    return false
  }
}

onBeforeRouteLeave(async () => {
  if (mode.value !== 'edit' || !formDirty.value) return true
  return await confirmLeave()
})
</script>

<template>
  <div class="prompt-detail">
    <el-empty v-if="notFound" description="提示词不存在或已被删除" />
    <div v-else-if="loading" class="prompt-detail__loading">
      <el-skeleton :rows="6" animated />
    </div>

    <template v-else-if="prompt">
      <header class="prompt-detail__header">
        <el-button link @click="router.push('/prompts')">
          <el-icon><ArrowLeftIcon /></el-icon>
          <span style="margin-left: 4px">返回列表</span>
        </el-button>
        <div class="prompt-detail__title-row">
          <h2 class="prompt-detail__title">{{ prompt.title }}</h2>
          <button
            type="button"
            class="prompt-detail__star"
            :title="prompt.is_favorite ? '取消收藏' : '收藏'"
            @click="handleToggleFavorite"
          >
            <el-icon v-if="prompt.is_favorite" class="is-fav"><StarFilledIcon /></el-icon>
            <el-icon v-else><StarIcon /></el-icon>
          </button>
        </div>
      </header>

      <!-- 模式切换：只读 / 编辑 -->
      <PromptDetailReadonly
        v-if="mode === 'view'"
        :prompt="prompt"
        :category-name="categoryName"
        @copy="handleCopy"
        @toggle-favorite="handleToggleFavorite"
      />

      <section v-else class="prompt-detail__edit">
        <PromptForm
          ref="formRef"
          :initial="formInitial"
          @dirty="onFormDirty"
          @submit="handleSubmit"
          @save-and-new="handleSubmit"
        />
        <footer class="prompt-detail__edit-footer">
          <el-button @click="handleCancelEdit">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="onSaveClick">
            保存
          </el-button>
        </footer>
      </section>

      <!-- 操作按钮组 -->
      <section class="prompt-detail__actions">
        <el-button :icon="CopyIcon" @click="handleCopy">复制</el-button>
        <el-button v-if="mode === 'view'" :icon="EditIcon" @click="handleEdit">编辑</el-button>
        <el-button :icon="TuneIcon" @click="handleTune">AI 调优</el-button>
        <el-button
          :icon="prompt.is_favorite ? StarFilledIcon : StarIcon"
          @click="handleToggleFavorite"
        >
          {{ prompt.is_favorite ? '取消收藏' : '收藏' }}
        </el-button>
        <el-button :icon="DeleteIcon" type="danger" plain @click="handleDelete">
          删除
        </el-button>
      </section>

      <!-- 调优历史折叠面板 -->
      <section class="prompt-detail__history">
        <el-collapse v-model="historyExpanded">
          <el-collapse-item name="history">
            <template #title>
              <span class="prompt-detail__history-title">
                <el-icon><HistogramIcon /></el-icon>
                调优历史
              </span>
            </template>
            <TuneHistoryPanel :prompt-id="prompt.id" />
          </el-collapse-item>
        </el-collapse>
      </section>
    </template>
  </div>
</template>

<style scoped>
.prompt-detail {
  max-width: 960px;
  margin: 0 auto;
  padding: 16px 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.prompt-detail__loading {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 24px;
}

.prompt-detail__header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.prompt-detail__title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.prompt-detail__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
  letter-spacing: 0.2px;
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.prompt-detail__star {
  flex-shrink: 0;
  border: 0;
  background: transparent;
  padding: 6px;
  cursor: pointer;
  color: #cbd5e1;
  font-size: 22px;
  border-radius: 6px;
  transition: color 0.15s, background 0.15s, transform 0.15s;
}
.prompt-detail__star:hover {
  color: #f59e0b;
  background: #fffbeb;
  transform: scale(1.1);
}
.prompt-detail__star .is-fav {
  color: #f59e0b;
}

/* 编辑模式容器 */
.prompt-detail__edit {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.prompt-detail__edit-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eef2f7;
  position: sticky;
  bottom: 16px;
}

/* 操作按钮组 */
.prompt-detail__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 16px 24px;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
}

/* 调优历史折叠 */
.prompt-detail__history {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 0 24px;
}

.prompt-detail__history-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: #1f2937;
}

@media (max-width: 768px) {
  .prompt-detail {
    padding: 12px 14px 24px;
  }
  .prompt-detail__actions {
    flex-direction: column;
    align-items: stretch;
  }
  .prompt-detail__edit-footer {
    flex-direction: column-reverse;
  }
}
</style>