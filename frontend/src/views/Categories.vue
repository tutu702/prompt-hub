<script setup lang="ts">
/**
 * 分类管理页（PRD P06）。
 *
 * 列表：拖拽手柄 / 颜色色块 / 名称 / HEX+复制 / 提示词数量 / 排序 / 创建时间 / 更新时间 / 行内操作。
 * 操作：编辑 / 上移 / 下移 / 删除（图标按钮）。
 * 删除约束：分类下还有提示词时后端返回 409，前端用 ElMessageBox.alert 引导去 /prompts 迁移。
 */
import { computed, onMounted, ref } from 'vue'
import {
  ElMessage,
  ElMessageBox,
} from 'element-plus'
import {
  Plus as PlusIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Top as TopIcon,
  Bottom as BottomIcon,
  Rank as DragIcon,
  CopyDocument as CopyIcon,
  WarningFilled as WarnIcon,
  InfoFilled as InfoIcon,
} from '@element-plus/icons-vue'

import CategoryFormDialog from '@/components/categories/CategoryFormDialog.vue'
import { useCategoryStore } from '@/stores/category'
import type { Category, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'
import { formatRelativeTime } from '@/utils'

const categoryStore = useCategoryStore()

const dialogOpen = ref(false)
const editing = ref<Category | null>(null)
const submitting = ref(false)

onMounted(async () => {
  await categoryStore.reload()
})

const total = computed(() => categoryStore.list.length)
const totalPrompts = computed(() =>
  categoryStore.list.reduce((sum, c) => sum + (c.prompt_count ?? 0), 0),
)

function fullTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

function colorOf(c: Category): string {
  return c.color || '#cbd5e1'
}

function openCreate() {
  editing.value = null
  dialogOpen.value = true
}

function openEdit(c: Category) {
  editing.value = c
  dialogOpen.value = true
}

async function onDialogSubmit(payload: {
  id: number | null
  data: CreateCategoryPayload | UpdateCategoryPayload
}) {
  submitting.value = true
  try {
    if (payload.id) {
      await categoryStore.updateCategory(
        payload.id,
        payload.data as UpdateCategoryPayload,
      )
      ElMessage.success('已更新')
    } else {
      await categoryStore.createCategory(payload.data as CreateCategoryPayload)
      ElMessage.success('已创建')
    }
    dialogOpen.value = false
  } catch (err) {
    ElMessage.error((err as Error)?.message || '保存失败，请稍后再试')
  } finally {
    submitting.value = false
  }
}

async function handleMoveUp(c: Category) {
  try {
    await categoryStore.moveUp(c.id)
  } catch (err) {
    ElMessage.error((err as Error)?.message || '上移失败')
  }
}

async function handleMoveDown(c: Category) {
  try {
    await categoryStore.moveDown(c.id)
  } catch (err) {
    ElMessage.error((err as Error)?.message || '下移失败')
  }
}

async function handleDelete(c: Category) {
  try {
    await ElMessageBox.confirm(
      `确定删除分类【${c.name}】吗？`,
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
    await categoryStore.deleteCategory(c.id)
    ElMessage.success('已删除')
  } catch (err) {
    // 409（分类下有提示词）时由 http 拦截器弹过 message；额外用 alert 引导
    const e = err as { response?: { status?: number; data?: { message?: string } } }
    const status = e?.response?.status
    if (status === 409) {
      await ElMessageBox.alert(
        `分类下还有 ${c.prompt_count ?? 0} 条提示词，无法删除。请先到「提示词库」将该分类下的提示词移至其他分类后再试。`,
        '无法删除',
        { type: 'warning', confirmButtonText: '我知道了' },
      )
    } else {
      ElMessage.error(
        e?.response?.data?.message || (err as Error)?.message || '删除失败',
      )
    }
  }
}

function rowIndex(c: Category): number {
  return categoryStore.list.findIndex((x) => x.id === c.id)
}

async function copyColor(c: Category) {
  if (!c.color) return
  try {
    await navigator.clipboard.writeText(c.color.toUpperCase())
    ElMessage.success(`已复制 ${c.color.toUpperCase()}`)
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}
</script>

<template>
  <div class="categories">
    <header class="categories__header">
      <div class="categories__title-row">
        <div>
          <h2 class="categories__title">分类管理</h2>
          <p class="categories__sub">将你的提示词有序组织起来</p>
        </div>
        <el-button
          type="primary"
          size="default"
          class="categories__create-btn"
          @click="openCreate"
        >
          <el-icon><PlusIcon /></el-icon>
          <span>新建分类</span>
        </el-button>
      </div>
    </header>

    <el-table
      v-loading="categoryStore.loading"
      :data="categoryStore.list"
      class="categories__table"
      empty-text="还没有分类，点击右上「新建分类」开始"
      :row-style="{ height: '64px' }"
    >
      <el-table-column label="分类名称" min-width="200">
        <template #default="{ row }">
          <div class="categories__name">
            <span class="categories__drag" :title="'拖拽以调整顺序（开发中）'">
              <el-icon><DragIcon /></el-icon>
            </span>
            <span
              class="categories__color-dot"
              :style="{ background: colorOf(row as Category) }"
            />
            <span class="categories__name-text">{{ (row as Category).name }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="颜色" min-width="140">
        <template #default="{ row }">
          <div v-if="(row as Category).color" class="categories__color">
            <span
              class="categories__color-swatch"
              :style="{ background: (row as Category).color }"
            />
            <span class="categories__color-text">
              {{ (row as Category).color!.toUpperCase() }}
            </span>
            <button
              type="button"
              class="categories__copy-btn"
              :title="'复制颜色'"
              @click="copyColor(row as Category)"
            >
              <el-icon><CopyIcon /></el-icon>
            </button>
          </div>
          <span v-else class="categories__color-empty">默认灰</span>
        </template>
      </el-table-column>

      <el-table-column label="提示词数量" width="120" align="left">
        <template #default="{ row }">
          <span class="categories__count">
            {{ (row as Category).prompt_count ?? 0 }}
            <span class="categories__count-unit">条</span>
          </span>
        </template>
      </el-table-column>

      <el-table-column label="排序" width="80" align="center">
        <template #default="{ row }">
          <span class="categories__sort">{{ (row as Category).sort }}</span>
        </template>
      </el-table-column>

      <el-table-column label="创建时间" min-width="140">
        <template #default="{ row }">
          <span class="categories__time" :title="fullTime((row as Category).created_at)">
            {{ formatRelativeTime((row as Category).created_at) }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="更新时间" min-width="140">
        <template #default="{ row }">
          <span class="categories__time" :title="fullTime((row as Category).updated_at)">
            {{ formatRelativeTime((row as Category).updated_at) }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <div class="categories__actions">
            <button
              type="button"
              class="categories__action-btn"
              title="上移"
              :disabled="rowIndex(row as Category) <= 0"
              @click="handleMoveUp(row as Category)"
            >
              <el-icon><TopIcon /></el-icon>
            </button>
            <button
              type="button"
              class="categories__action-btn"
              title="下移"
              :disabled="rowIndex(row as Category) >= categoryStore.list.length - 1"
              @click="handleMoveDown(row as Category)"
            >
              <el-icon><BottomIcon /></el-icon>
            </button>
            <button
              type="button"
              class="categories__action-btn"
              title="编辑"
              @click="openEdit(row as Category)"
            >
              <el-icon><EditIcon /></el-icon>
            </button>
            <button
              type="button"
              class="categories__action-btn categories__action-btn--danger"
              title="删除"
              @click="handleDelete(row as Category)"
            >
              <el-icon><DeleteIcon /></el-icon>
            </button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <footer class="categories__footer">
      <div class="categories__footer-left">
        共 <strong>{{ total }}</strong> 个分类 ·
        <strong>{{ totalPrompts }}</strong> 条提示词已归类
      </div>
      <div class="categories__footer-right">
        <el-icon class="categories__footer-icon"><InfoIcon /></el-icon>
        <span>分类顺序会影响提示词列表中的筛选顺序，可拖拽左侧图标调整</span>
      </div>
    </footer>

    <div class="categories__notice">
      <el-icon class="categories__notice-icon"><WarnIcon /></el-icon>
      <span class="categories__notice-text">
        删除分类前请先迁移其下的提示词，否则将无法删除。
      </span>
    </div>

    <CategoryFormDialog
      v-model="dialogOpen"
      :initial="editing"
      :submitting="submitting"
      @submit="onDialogSubmit"
    />
  </div>
</template>

<style scoped>
.categories {
  max-width: 1280px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.categories__header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.categories__title-row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.categories__title {
  margin: 0 0 6px;
  font-size: 24px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.2px;
}

.categories__sub {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.categories__create-btn {
  font-weight: 500;
  padding: 10px 18px;
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.18);
}

.categories__table {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eef2f7;
  overflow: hidden;
}

.categories__table :deep(.el-table__header-wrapper th) {
  background: #fafbfc;
  color: #64748b;
  font-weight: 500;
  font-size: 13px;
  border-bottom: 1px solid #eef2f7;
}

.categories__table :deep(.el-table__row td) {
  border-bottom: 1px solid #f1f5f9;
}

.categories__table :deep(.el-table__row:last-child td) {
  border-bottom: none;
}

/* 分类名称 */
.categories__name {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.categories__drag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  color: #cbd5e1;
  cursor: grab;
  flex-shrink: 0;
  font-size: 14px;
  transition: color 0.15s;
}

.categories__drag:hover {
  color: #64748b;
}

.categories__color-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 1px solid rgba(15, 23, 42, 0.08);
  flex-shrink: 0;
}

.categories__name-text {
  font-weight: 500;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 颜色列 */
.categories__color {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.categories__color-swatch {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  flex-shrink: 0;
}

.categories__color-text {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #475569;
  letter-spacing: 0.2px;
}

.categories__copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #94a3b8;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: color 0.15s, background 0.15s;
}

.categories__copy-btn:hover {
  color: #2563eb;
  background: #eef4ff;
}

.categories__color-empty {
  font-size: 12px;
  color: #94a3b8;
}

/* 提示词数量 */
.categories__count {
  font-weight: 600;
  color: #0f172a;
  font-size: 14px;
}

.categories__count-unit {
  font-weight: 400;
  color: #64748b;
  font-size: 12px;
  margin-left: 2px;
}

/* 排序 */
.categories__sort {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: #94a3b8;
  font-size: 13px;
}

/* 时间 */
.categories__time {
  color: #64748b;
  font-size: 13px;
}

/* 操作按钮 */
.categories__actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.categories__action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #94a3b8;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: color 0.15s, background 0.15s;
}

.categories__action-btn:hover:not(:disabled) {
  background: #f1f5f9;
  color: #2563eb;
}

.categories__action-btn:disabled {
  color: #e2e8f0;
  cursor: not-allowed;
}

.categories__action-btn--danger:hover:not(:disabled) {
  background: #fef2f2;
  color: #ef4444;
}

/* 底部统计行 */
.categories__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 4px;
  font-size: 13px;
  color: #64748b;
}

.categories__footer-left strong {
  color: #0f172a;
  font-weight: 600;
  margin: 0 2px;
}

.categories__footer-right {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #94a3b8;
  font-size: 12px;
}

.categories__footer-icon {
  font-size: 14px;
}

/* 警告提示 */
.categories__notice {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 10px;
  color: #b45309;
  font-size: 13px;
}

.categories__notice-icon {
  font-size: 16px;
  color: #f59e0b;
  flex-shrink: 0;
}

.categories__notice-text {
  color: #b45309;
}

@media (max-width: 768px) {
  .categories__title-row {
    flex-direction: column;
    align-items: stretch;
  }
  .categories__footer {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
}
</style>
