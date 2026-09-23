<script setup lang="ts">
/**
 * 工具栏（单行版）：搜索 / 收藏 tab / 分类 / 排序 / 视图 / 新建 / AI 生成。
 *
 * 搜索框同时覆盖标题/描述/内容（由 store 默认 q_in='all' 控制）。
 * 二级浮层用于展示批量操作，避免主行被撑高。
 */
import { computed, ref, watch } from 'vue'
import {
  Search as SearchIcon,
  Plus as PlusIcon,
  MagicStick as MagicIcon,
  Grid as GridIcon,
  List as ListIcon,
} from '@element-plus/icons-vue'
import { storeToRefs } from 'pinia'
import { useCategoryStore } from '@/stores/category'
import { usePromptStore } from '@/stores/prompt'
import { debounce } from '@/utils'

const promptStore = usePromptStore()
const categoryStore = useCategoryStore()

const {
  keyword,
  isFavorite,
  sort,
  categoryId,
  viewMode,
  batchMode,
  selectedIds,
} = storeToRefs(promptStore)

const searchInput = ref<string>(keyword.value)

// 外部 store 中 keyword 变化时（如分页后被清空）同步回 input
watch(keyword, (v) => {
  if (searchInput.value !== v) searchInput.value = v
})

const debouncedFetch = debounce(() => {
  promptStore.setKeyword(searchInput.value)
  promptStore.fetchList()
}, 300)

function onSearchInput(v: string) {
  searchInput.value = v
  debouncedFetch()
}

function onFavChange(v: string | number | boolean | undefined) {
  if (v === 'all' || v === 'fav') {
    promptStore.setIsFavorite(v)
    promptStore.fetchList()
  }
}

function onCategoryChange(v: string | number | boolean | null | undefined) {
  if (v === '' || v === undefined || v === null) {
    promptStore.setCategoryId(null)
  } else {
    const n = Number(v)
    promptStore.setCategoryId(Number.isNaN(n) ? null : n)
  }
  promptStore.fetchList()
}

function onSortChange(v: string | number | boolean | undefined) {
  if (
    v === 'updated_desc' ||
    v === 'created_desc' ||
    v === 'title_asc'
  ) {
    promptStore.setSort(v as 'updated_desc' | 'created_desc' | 'title_asc')
    promptStore.fetchList()
  }
}

function onViewChange(v: string | number | boolean | undefined) {
  if (v === 'list') {
    // 列表视图占位（图中 grid/list 切换按钮）
    promptStore.setViewMode('list')
    return
  }
  if (v === 'card') {
    promptStore.setViewMode('card')
  }
}

function toggleBatchMode() {
  promptStore.setBatchMode(!batchMode.value)
}

function onBatchDelete() {
  emit('batch-delete')
}

function onBatchExport() {
  emit('batch-export')
}

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'ai-create'): void
  (e: 'batch-delete'): void
  (e: 'batch-export'): void
}>()

const categoryOptions = computed(() => [
  { id: null as number | null, name: '全部分类' },
  ...categoryStore.list.map((c) => ({ id: c.id, name: c.name })),
])

const selectedCount = computed(() => selectedIds.value.length)
</script>

<template>
  <div class="toolbar">
    <!-- 主行：搜索 + 全部/仅收藏 + 分类 + 排序 + 视图 + 主操作 -->
    <div class="toolbar__main">
      <el-input
        v-model="searchInput"
        placeholder="搜索标题或内容..."
        clearable
        :prefix-icon="SearchIcon"
        class="toolbar__search"
        @input="onSearchInput"
        @clear="onSearchInput('')"
      />

      <el-radio-group
        :model-value="isFavorite"
        class="toolbar__fav"
        @update:model-value="onFavChange"
      >
        <el-radio-button value="all">全部</el-radio-button>
        <el-radio-button value="fav">仅收藏</el-radio-button>
      </el-radio-group>

      <el-select
        :model-value="categoryId ?? ''"
        class="toolbar__cat"
        placeholder="全部分类"
        clearable
        @update:model-value="onCategoryChange"
        @clear="onCategoryChange(null)"
      >
        <el-option
          v-for="c in categoryOptions"
          :key="c.id ?? 'all'"
          :label="c.name"
          :value="c.id ?? ''"
        />
      </el-select>

      <el-select
        :model-value="sort"
        class="toolbar__sort"
        @update:model-value="onSortChange"
      >
        <el-option label="最近更新" value="updated_desc" />
        <el-option label="最近创建" value="created_desc" />
        <el-option label="标题 A-Z" value="title_asc" />
      </el-select>

      <el-radio-group
        :model-value="viewMode"
        class="toolbar__view"
        @update:model-value="onViewChange"
      >
        <el-radio-button value="card">
          <el-icon><GridIcon /></el-icon>
        </el-radio-button>
        <el-radio-button value="list">
          <el-icon><ListIcon /></el-icon>
        </el-radio-button>
      </el-radio-group>

      <div class="toolbar__spacer" />

      <el-button type="primary" :icon="PlusIcon" @click="emit('create')">
        新建提示词
      </el-button>
      <el-button type="primary" plain :icon="MagicIcon" @click="emit('ai-create')">
        AI 一键生成
      </el-button>
    </div>

    <!-- 批量操作浮条 -->
    <transition name="toolbar-batch">
      <div v-if="batchMode" class="toolbar__batch">
        <span class="toolbar__count">已选中 {{ selectedCount }} 项</span>
        <el-button :disabled="selectedCount === 0" @click="onBatchExport">
          批量导出
        </el-button>
        <el-button
          type="danger"
          :disabled="selectedCount === 0"
          @click="onBatchDelete"
        >
          批量删除
        </el-button>
        <el-button text @click="toggleBatchMode">退出批量</el-button>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.toolbar {
  background: #ffffff;
  border-radius: 14px;
  padding: 14px 18px;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
  border: 1px solid #eef2f7;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.toolbar__main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* 搜索框：占据主要剩余宽度 */
.toolbar__search {
  flex: 1 1 280px;
  min-width: 220px;
  max-width: 460px;
}
.toolbar__search :deep(.el-input__wrapper) {
  border-radius: 10px;
  background: #f8fafc;
}

/* 全部/仅收藏：胶囊 tab */
.toolbar__fav :deep(.el-radio-button__inner) {
  border-radius: 8px !important;
  padding: 6px 14px;
  font-size: 13px;
  border: 1px solid #e2e8f0;
}
.toolbar__fav :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: 8px 0 0 8px !important;
}
.toolbar__fav :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 0 8px 8px 0 !important;
}
.toolbar__fav :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #eef4ff;
  color: #2563eb;
  border-color: #bfd6ff;
  box-shadow: -1px 0 0 0 #bfd6ff;
}

.toolbar__cat,
.toolbar__sort {
  width: 140px;
  flex-shrink: 0;
}
.toolbar__cat :deep(.el-input__wrapper),
.toolbar__sort :deep(.el-input__wrapper) {
  border-radius: 10px;
}

/* 视图切换：方形图标按钮组 */
.toolbar__view :deep(.el-radio-button__inner) {
  padding: 6px 10px;
  border-radius: 8px !important;
  font-size: 13px;
  border: 1px solid #e2e8f0;
}
.toolbar__view :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: 8px 0 0 8px !important;
}
.toolbar__view :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 0 8px 8px 0 !important;
}

.toolbar__spacer { flex: 1; }

/* 批量操作条 */
.toolbar__batch {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 10px;
  flex-wrap: wrap;
}

.toolbar__count {
  font-size: 13px;
  color: #9a3412;
  margin-right: 4px;
}

/* 批量条进出动画 */
.toolbar-batch-enter-from,
.toolbar-batch-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
.toolbar-batch-enter-active,
.toolbar-batch-leave-active {
  transition: opacity 0.18s, transform 0.18s;
}

@media (max-width: 960px) {
  .toolbar__cat,
  .toolbar__sort { width: 110px; }
  .toolbar__search { min-width: 180px; }
}

@media (max-width: 720px) {
  .toolbar__main { gap: 8px; }
  .toolbar__cat,
  .toolbar__sort { width: 100%; }
}
</style>
