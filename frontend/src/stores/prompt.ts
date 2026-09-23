/**
 * 提示词列表 / 筛选 / 排序 / 选中 管理（不进入 localStorage）。
 *
 * 每次进入 /prompts 视图时由 onMounted 主动调用 fetchList()；
 * 离开页面时无需主动清空，store 在路由切换后保留以便用户回退时保留筛选态。
 */
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { promptApi } from '@/api'
import type { CreatePromptPayload, Prompt, PromptListQuery, UpdatePromptPayload } from '@/types'

export type SortKey = 'updated_desc' | 'created_desc' | 'title_asc'
export type QInKey = 'title' | 'all'
export type ViewMode = 'card' | 'list'

export const usePromptStore = defineStore('prompt', () => {
  // 列表
  const list = ref<Prompt[]>([])
  const total = ref<number>(0)
  const page = ref<number>(1)
  const pageSize = ref<number>(20)
  const loading = ref<boolean>(false)

  // 过滤
  const keyword = ref<string>('')
  const qIn = ref<QInKey>('all')
  const sort = ref<SortKey>('updated_desc')
  const categoryId = ref<number | null>(null)
  const isFavorite = ref<'all' | 'fav'>('all')

  // 视图与批量
  const viewMode = ref<ViewMode>('card')
  const batchMode = ref<boolean>(false)
  const selectedIds = ref<number[]>([])

  /** 拉取当前过滤条件下的列表。 */
  async function fetchList(): Promise<void> {
    loading.value = true
    try {
      const q: PromptListQuery = {
        page: page.value,
        page_size: pageSize.value,
        keyword: keyword.value || undefined,
        q_in: qIn.value,
        sort: sort.value,
        category_id: categoryId.value,
        is_favorite: isFavorite.value === 'fav' ? true : undefined,
      }
      const res = await promptApi.listPrompts(q)
      list.value = res.list
      total.value = res.total
      page.value = res.page
      pageSize.value = res.page_size
      // 清理 selectedIds 中已被删除的项
      cleanupSelection()
    } finally {
      loading.value = false
    }
  }

  /** 关键词变化（debounce 由调用方控制） */
  function setKeyword(k: string) {
    keyword.value = k
    page.value = 1
  }

  function setQIn(v: QInKey) {
    qIn.value = v
    page.value = 1
  }

  function setSort(v: SortKey) {
    sort.value = v
    page.value = 1
  }

  function setCategoryId(id: number | null) {
    categoryId.value = id
    page.value = 1
  }

  function setIsFavorite(v: 'all' | 'fav') {
    isFavorite.value = v
    page.value = 1
  }

  function goPage(p: number) {
    page.value = p
  }

  function setViewMode(v: ViewMode) {
    viewMode.value = v
  }

  function setBatchMode(v: boolean) {
    batchMode.value = v
    if (!v) selectedIds.value = []
  }

  function toggleSelect(id: number) {
    const idx = selectedIds.value.indexOf(id)
    if (idx >= 0) {
      selectedIds.value.splice(idx, 1)
    } else {
      selectedIds.value.push(id)
    }
  }

  function selectAll() {
    selectedIds.value = list.value.map((p) => p.id)
  }

  function clearSelection() {
    selectedIds.value = []
  }

  function cleanupSelection() {
    const validIds = new Set(list.value.map((p) => p.id))
    selectedIds.value = selectedIds.value.filter((id) => validIds.has(id))
  }

  /** 从列表中移除指定 id（无需重拉）。 */
  function removeFromList(ids: number[]) {
    if (!ids.length) return
    const idSet = new Set(ids)
    list.value = list.value.filter((p) => !idSet.has(p.id))
    total.value = Math.max(0, total.value - ids.length)
    cleanupSelection()
  }

  /** 局部更新一条（收藏切换、详情更新等）。 */
  function upsertOne(p: Prompt) {
    const idx = list.value.findIndex((x) => x.id === p.id)
    if (idx >= 0) {
      list.value.splice(idx, 1, p)
    } else {
      list.value.unshift(p)
      total.value += 1
    }
  }

  /** 创建一条提示词（用于 /prompts/new 提交）。成功后 upsert 到列表顶部。 */
  async function createPrompt(payload: CreatePromptPayload): Promise<Prompt> {
    const created = await promptApi.createPrompt(payload)
    list.value.unshift(created)
    total.value += 1
    cleanupSelection()
    return created
  }

  /** 更新一条提示词（详情页编辑后 / 列表收藏切换后）。成功后 upsert 同步列表。 */
  async function updatePrompt(id: number, payload: UpdatePromptPayload): Promise<Prompt> {
    const updated = await promptApi.updatePrompt(id, payload)
    upsertOne(updated)
    return updated
  }

  /** 重置过滤（保留页码为 1）。 */
  function resetFilters() {
    keyword.value = ''
    qIn.value = 'all'
    sort.value = 'updated_desc'
    categoryId.value = null
    isFavorite.value = 'all'
    page.value = 1
  }

  return {
    // state
    list,
    total,
    page,
    pageSize,
    loading,
    keyword,
    qIn,
    sort,
    categoryId,
    isFavorite,
    viewMode,
    batchMode,
    selectedIds,
    // actions
    fetchList,
    setKeyword,
    setQIn,
    setSort,
    setCategoryId,
    setIsFavorite,
    goPage,
    setViewMode,
    setBatchMode,
    toggleSelect,
    selectAll,
    clearSelection,
    cleanupSelection,
    removeFromList,
    upsertOne,
    createPrompt,
    updatePrompt,
    resetFilters,
  }
})