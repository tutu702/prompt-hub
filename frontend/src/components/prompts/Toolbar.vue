<script setup lang="ts">
/**
 * 顶部工具栏：搜索/筛选/排序/视图/批量/新建/AI 生成。
 */
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { storeToRefs } from 'pinia'
import { useCategoryStore } from '@/stores/category'
import { usePromptStore } from '@/stores/prompt'
import { debounce } from '@/utils'

const promptStore = usePromptStore()
const categoryStore = useCategoryStore()

const {
  keyword,
  qIn,
  sort,
  categoryId,
  isFavorite,
  viewMode,
  batchMode,
  selectedIds,
} = storeToRefs(promptStore)

const searchInput = ref<string>(keyword.value)

// 防抖 fetch（输入即触发）
const debouncedFetch = debounce(() => {
  promptStore.setKeyword(searchInput.value)
  promptStore.fetchList()
}, 300)

function onSearchInput(v: string) {
  searchInput.value = v
  debouncedFetch()
}

function onQInChange(v: string | number | boolean | undefined) {
  if (v === 'title' || v === 'all') {
    promptStore.setQIn(v)
    promptStore.fetchList()
  }
}

function onIsFavoriteChange(v: string | number | boolean | undefined) {
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
  if (v === 'updated_desc' || v === 'created_desc' || v === 'title_asc') {
    promptStore.setSort(v)
    promptStore.fetchList()
  }
}

function onViewChange(v: string | number | boolean | undefined) {
  if (v === 'list') {
    ElMessage.info('列表视图将在后续版本开放')
    return
  }
  if (v === 'card') {
    promptStore.setViewMode(v)
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
  { id: null, name: '全部分类' },
  ...categoryStore.list.map((c) => ({ id: c.id, name: c.name })),
])

const selectedCount = computed(() => selectedIds.value.length)
</script>

<template>
  <div class="toolbar">
    <div class="toolbar__row">
      <el-input
        v-model="searchInput"
        placeholder="搜索标题、描述或内容"
        clearable
        :prefix-icon="Search"
        class="toolbar__search"
        @input="onSearchInput"
        @clear="onSearchInput('')"
      />

      <el-select
        :model-value="qIn"
        class="toolbar__qin"
        @update:model-value="onQInChange"
      >
        <el-option label="仅标题" value="title" />
        <el-option label="全部字段" value="all" />
      </el-select>

      <el-select
        :model-value="isFavorite"
        class="toolbar__fav"
        @update:model-value="onIsFavoriteChange"
      >
        <el-option label="全部" value="all" />
        <el-option label="仅收藏" value="fav" />
      </el-select>

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
    </div>

    <div class="toolbar__row toolbar__row--second">
      <el-radio-group :model-value="sort" @update:model-value="onSortChange">
        <el-radio-button value="updated_desc">最新更新</el-radio-button>
        <el-radio-button value="created_desc">最新创建</el-radio-button>
        <el-radio-button value="title_asc">标题 A-Z</el-radio-button>
      </el-radio-group>

      <el-radio-group :model-value="viewMode" @update:model-value="onViewChange">
        <el-radio-button value="card">卡片</el-radio-button>
        <el-tooltip content="列表视图将在后续版本开放" placement="top">
          <el-radio-button value="list" disabled>列表</el-radio-button>
        </el-tooltip>
      </el-radio-group>

      <div class="toolbar__spacer" />

      <el-button @click="toggleBatchMode">
        {{ batchMode ? '退出批量' : '批量管理' }}
      </el-button>
      <el-button type="primary" @click="emit('create')">新建提示词</el-button>
      <el-button type="success" plain @click="emit('ai-create')">AI 生成</el-button>
    </div>

    <div v-if="batchMode" class="toolbar__row toolbar__batch">
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
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.toolbar__row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.toolbar__row--second {
  border-top: 1px dashed #f1f5f9;
  padding-top: 12px;
}

.toolbar__search {
  flex: 1 1 280px;
  min-width: 220px;
  max-width: 480px;
}

.toolbar__qin,
.toolbar__fav,
.toolbar__cat {
  width: 140px;
  flex-shrink: 0;
}

.toolbar__spacer {
  flex: 1;
}

.toolbar__batch {
  background: #fffbe6;
  border-radius: 8px;
  padding: 8px 12px;
  border: 1px solid #ffe58f;
}

.toolbar__count {
  font-size: 13px;
  color: #475569;
  margin-right: 8px;
}

@media (max-width: 768px) {
  .toolbar__qin,
  .toolbar__fav,
  .toolbar__cat {
    width: 110px;
  }
}
</style>