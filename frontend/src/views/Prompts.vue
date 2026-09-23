<script setup lang="ts">
/**
 * 提示词库 P03（PRD §三 P03）。
 *
 * 顶部工具栏 + 卡片网格 + 分页 + 空态。
 * 支持：搜索（防抖 300ms）/ 筛选 / 排序 / 收藏切换 / 复制 / 删除 / 批量删除。
 * 不支持（占位 toast）：新建 / AI 生成 / 详情 / 编辑 / AI 调优 / 批量导出。
 */
import { computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus as PlusIcon, MagicStick as MagicIcon } from '@element-plus/icons-vue'

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

const {
  list,
  total,
  page,
  pageSize,
  loading,
  batchMode,
  selectedIds,
} = storeToRefs(promptStore)

onMounted(async () => {
  // 拉取分类 + 拉取列表（并行）
  await Promise.all([
    categoryStore.loadIfNeed(),
    promptStore.fetchList(),
  ])
})

// 登录态变化时重置 store 并重新拉取
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
    // 回滚
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
  ElMessage.info('提示词创建页即将开放')
}

function handleAiCreate() {
  ElMessage.info('AI 生成提示词即将开放')
}

function handleEdit(_p: PromptT) {
  ElMessage.info('提示词编辑页即将开放')
}

function handleTune(_p: PromptT) {
  ElMessage.info('AI 调优即将开放')
}

function handleDetail(_p: PromptT) {
  ElMessage.info('提示词详情页即将开放')
}

function onPageChange(p: number) {
  promptStore.goPage(p)
  promptStore.fetchList()
}

function handleLogout() {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    type: 'warning',
    confirmButtonText: '退出',
    cancelButtonText: '取消',
  })
    .then(async () => {
      await userStore.logout()
      ElMessage.success('已退出登录')
    })
    .catch(() => {})
}
</script>

<template>
  <div class="prompts">
    <header class="prompts__header">
      <div>
        <h2 class="prompts__title">提示词库</h2>
        <p class="prompts__sub">
          欢迎回来，
          <strong>{{ userStore.userInfo?.nickname || userStore.userInfo?.email || '用户' }}</strong>
        </p>
      </div>
      <el-button @click="handleLogout">退出登录</el-button>
    </header>

    <Toolbar
      class="prompts__toolbar"
      @create="handleCreate"
      @ai-create="handleAiCreate"
      @batch-delete="handleBatchDelete"
      @batch-export="handleBatchExport"
    />

    <main class="prompts__main">
      <PromptList
        :list="list"
        :loading="loading"
        :batch-mode="batchMode"
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
              <el-button type="success" plain :icon="MagicIcon" @click="handleAiCreate">
                AI 生成提示词
              </el-button>
            </div>
          </div>
        </template>
      </PromptList>
    </main>

    <footer v-if="!showEmpty" class="prompts__pagination">
      <el-pagination
        background
        layout="prev, pager, next, jumper, total"
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
  max-width: 1280px;
  margin: 0 auto;
  padding: 16px 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.prompts__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 0 8px;
}

.prompts__title {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.prompts__sub {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

.prompts__main {
  min-height: 240px;
}

.prompts__pagination {
  display: flex;
  justify-content: center;
  padding: 16px 0 0;
}

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

@media (max-width: 768px) {
  .prompts {
    padding: 12px 14px 24px;
  }
  .prompts__header {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>