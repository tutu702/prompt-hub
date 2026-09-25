<script setup lang="ts">
/**
 * 提示词库 P03（PRD §三 P03）。
 *
 * 顶部问候 + 工具栏 + 卡片网格 + 分页 + 空态。
 * 支持：搜索（防抖 300ms）/ 筛选 / 排序 / 收藏切换 / 复制 / 删除 / 批量删除。
 * 不支持（占位 toast）：编辑 / AI 调优 / 详情 / 批量导出。
 */
import { computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus as PlusIcon,
  Collection as CollectionIcon,
} from '@element-plus/icons-vue'

import PromptList from '@/components/prompts/PromptList.vue'
import Toolbar from '@/components/prompts/Toolbar.vue'
import { promptApi } from '@/api'
import { usePromptStore } from '@/stores/prompt'
import { useCategoryStore } from '@/stores/category'
import { useUserStore } from '@/stores/user'
import type { Prompt as PromptT } from '@/types'
import { copyToClipboard } from '@/utils'

const promptStore = usePromptStore()
const categoryStore = useCategoryStore()
const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const {
  list,
  total,
  page,
  pageSize,
  loading,
  batchMode,
  selectedIds,
  viewMode,
} = storeToRefs(promptStore)

/** 问候语（按本地小时数简单切换） */
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

/** 显示用用户名（昵称 > 邮箱 > "用户"） */
const displayName = computed(
  () =>
    userStore.userInfo?.nickname?.trim() ||
    (userStore.userInfo?.email ? userStore.userInfo.email.split('@')[0] : '') ||
    '用户',
)

/** 今日新增条数（基于当前列表 created_at 计算，仅为占位展示） */
const todayNewCount = computed(() => {
  const today = new Date().toDateString()
  return list.value.filter(
    (p) => new Date(p.created_at).toDateString() === today,
  ).length
})

/* --------------------------------------------------------------
 * 数据拉取与登录态联动
 * ------------------------------------------------------------*/

onMounted(async () => {
  // 若 URL 带 q=xxx（从顶部全局搜索跳转过来），先填到 store
  const incoming = route.query.q
  if (typeof incoming === 'string' && incoming) {
    promptStore.setKeyword(incoming)
  }

  await Promise.all([
    categoryStore.loadIfNeed(),
    promptStore.fetchList(),
  ])
})

watch(
  () => userStore.isLoggedIn(),
  (logged) => {
    if (logged) {
      categoryStore.loadIfNeed(true)
      promptStore.fetchList()
    }
  },
)

const showEmpty = computed(() => !loading.value && list.value.length === 0)

/* --------------------------------------------------------------
 * 卡片交互
 * ------------------------------------------------------------*/

async function handleCopy(p: PromptT) {
  const ok = await copyToClipboard(p.content)
  if (ok) {
    ElMessage.success('已复制到剪贴板')
  } else {
    ElMessage.error('复制失败，请手动复制')
  }
}

async function handleFavorite(p: PromptT) {
  const target = !p.is_favorite
  // 乐观更新
  const idx = promptStore.list.findIndex((x) => x.id === p.id)
  const original = idx >= 0 ? { ...promptStore.list[idx] } : null
  if (idx >= 0) {
    promptStore.list[idx].is_favorite = target
  }
  try {
    await promptApi.toggleFavorite(p.id, { is_favorite: target })
  } catch {
    if (original && idx >= 0) {
      promptStore.list.splice(idx, 1, original)
    }
    ElMessage.error('操作失败，请稍后再试')
  }
}

async function handleDelete(p: PromptT) {
  try {
    await ElMessageBox.confirm(
      `确定删除【${p.title}】吗？删除后可在 30 天内恢复`,
      '提示',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
      },
    )
  } catch {
    return
  }
  try {
    await promptApi.deletePrompt(p.id)
    promptStore.removeFromList([p.id])
    ElMessage.success('已删除')
  } catch {
    ElMessage.error('删除失败，请稍后再试')
  }
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确定删除已选中的 ${selectedIds.value.length} 项提示词吗？删除后可在 30 天内恢复`,
      '批量删除',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
      },
    )
  } catch {
    return
  }
  try {
    const res = await promptApi.batchDeletePrompts({ ids: [...selectedIds.value] })
    ElMessage.success(`已删除 ${res.deleted} 项`)
    await promptStore.fetchList()
  } catch {
    ElMessage.error('批量删除失败，请稍后再试')
  }
}

function handleBatchExport() {
  ElMessage.info('批量导出将在后续版本开放')
}

function handleCreate() {
  router.push('/prompts/new')
}

function handleAiCreate() {
  router.push({ name: 'PromptCreate', query: { tab: 'ai' } })
}

function handleEdit(p: PromptT) {
  router.push({ path: `/prompts/${p.id}`, query: { mode: 'edit' } })
}

function handleTune(p: PromptT) {
  router.push(`/prompts/${p.id}/tune`)
}

function handleDetail(p: PromptT) {
  router.push(`/prompts/${p.id}`)
}

function onPageChange(p: number) {
  promptStore.goPage(p)
  promptStore.fetchList()
}
</script>

<template>
  <div class="prompts">
    <!-- 问候 -->
    <section class="prompts__hello">
      <div class="prompts__hello-text">
        <h1 class="prompts__hello-title">
          {{ greeting }}，{{ displayName }} <span class="prompts__hello-emoji">👋</span>
        </h1>
        <p class="prompts__hello-sub">
          今日新增
          <strong>{{ todayNewCount }}</strong>
          条提示词 ·
          共 <strong>{{ total }}</strong> 条资产
        </p>
      </div>
    </section>

    <!-- 工具栏 -->
    <Toolbar
      class="prompts__toolbar"
      @create="handleCreate"
      @ai-create="handleAiCreate"
      @batch-delete="handleBatchDelete"
      @batch-export="handleBatchExport"
    />

    <!-- 列表 -->
    <main class="prompts__main">
      <PromptList
        :list="list"
        :loading="loading"
        :batch-mode="batchMode"
        :view-mode="viewMode"
        :selected-ids="selectedIds"
        @copy="handleCopy"
        @edit="handleEdit"
        @tune="handleTune"
        @delete="handleDelete"
        @detail="handleDetail"
        @favorite="handleFavorite"
        @select="(p: PromptT) => promptStore.toggleSelect(p.id)"
      >
        <template #empty>
          <div v-if="showEmpty" class="prompts__empty">
            <div class="prompts__empty-emoji">📝</div>
            <h3 class="prompts__empty-title">还没有任何提示词</h3>
            <p class="prompts__empty-desc">从手动创建开始，或用 AI 帮你生成第一段</p>
            <div class="prompts__empty-actions">
              <el-button type="primary" :icon="PlusIcon" @click="handleCreate">
                新建提示词
              </el-button>
              <el-button type="success" plain :icon="CollectionIcon" @click="handleAiCreate">
                AI 生成提示词
              </el-button>
            </div>
          </div>
        </template>
      </PromptList>
    </main>

    <!-- 分页 -->
    <footer v-if="!showEmpty" class="prompts__pagination">
      <el-pagination
        background
        layout="prev, pager, next"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="onPageChange"
      />
    </footer>
  </div>
</template>

<style scoped>
.prompts {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ---------- 问候区 ---------- */
.prompts__hello {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
}

.prompts__hello-text { min-width: 0; }

.prompts__hello-title {
  margin: 0 0 8px;
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: 0.2px;
}

.prompts__hello-emoji {
  display: inline-block;
  margin-left: 4px;
  animation: prompts-wave 1.4s ease-in-out infinite;
  transform-origin: 70% 70%;
}

@keyframes prompts-wave {
  0%, 60%, 100% { transform: rotate(0deg); }
  10% { transform: rotate(14deg); }
  20% { transform: rotate(-8deg); }
  30% { transform: rotate(14deg); }
  40% { transform: rotate(-4deg); }
  50% { transform: rotate(10deg); }
}

.prompts__hello-sub {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.prompts__hello-sub strong {
  color: #334155;
  font-weight: 600;
  margin: 0 2px;
}

/* ---------- 主列表 ---------- */
.prompts__main {
  min-height: 240px;
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
}

/* ---------- 分页 ---------- */
.prompts__pagination {
  display: flex;
  justify-content: center;
  padding: 8px 0 24px;
}

/* ---------- 空态 ---------- */
.prompts__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px dashed #e2e8f0;
  text-align: center;
}

.prompts__empty-emoji {
  font-size: 64px;
  line-height: 1;
}

.prompts__empty-title {
  margin: 16px 0 6px;
  font-size: 17px;
  font-weight: 600;
  color: #1f2937;
}

.prompts__empty-desc {
  margin: 0 0 20px;
  color: #94a3b8;
  font-size: 13px;
}

.prompts__empty-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: center;
}

@media (max-width: 600px) {
  .prompts__hello-title { font-size: 22px; }
}
</style>