/**
 * 分类 store：缓存用户的分类列表，跨页面复用。
 *
 * `loaded` 标记避免重复拉取；登录态变化时由调用方调 `reset()` 触发重载。
 */
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { categoryApi } from '@/api'
import type { Category, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'

export const useCategoryStore = defineStore('category', () => {
  const list = ref<Category[]>([])
  const loaded = ref<boolean>(false)
  const loading = ref<boolean>(false)

  async function loadIfNeed(force = false): Promise<void> {
    if (loaded.value && !force) return
    loading.value = true
    try {
      const res = await categoryApi.listCategories()
      list.value = res.list || []
      sortList()
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  function reset() {
    list.value = []
    loaded.value = false
  }

  function getName(id: number | null | undefined): string {
    if (id == null) return '未分类'
    const c = list.value.find((x) => x.id === id)
    return c ? c.name : '未分类'
  }

  /** 排序 list 按 (sort ASC, created_at ASC)，与 gormrepo.List 一致。 */
  function sortList() {
    list.value.sort((a, b) => {
      if (a.sort !== b.sort) return a.sort - b.sort
      return a.created_at < b.created_at ? -1 : 1
    })
  }

  /** 创建分类；成功后 push 到 list，重排。 */
  async function createCategory(payload: CreateCategoryPayload): Promise<Category> {
    const created = await categoryApi.createCategory(payload)
    list.value.push(created)
    sortList()
    loaded.value = true
    return created
  }

  /** 更新分类；成功后替换条目并重排。 */
  async function updateCategory(
    id: number,
    payload: UpdateCategoryPayload,
  ): Promise<Category> {
    const updated = await categoryApi.updateCategory(id, payload)
    const idx = list.value.findIndex((x) => x.id === id)
    if (idx >= 0) list.value.splice(idx, 1, updated)
    sortList()
    return updated
  }

  /** 删除分类；成功后从 list 中移除。 */
  async function deleteCategory(id: number): Promise<void> {
    await categoryApi.deleteCategory(id)
    list.value = list.value.filter((x) => x.id !== id)
  }

  /** 上移：与上一行交换 sort，本地先换避免视觉跳动，失败回滚 + reload。 */
  async function moveUp(id: number): Promise<void> {
    const idx = list.value.findIndex((x) => x.id === id)
    if (idx <= 0) return
    const cur = list.value[idx]
    const prev = list.value[idx - 1]
    const curSort = cur.sort
    cur.sort = prev.sort
    prev.sort = curSort
    try {
      await Promise.all([
        categoryApi.updateCategory(cur.id, { sort: cur.sort }),
        categoryApi.updateCategory(prev.id, { sort: prev.sort }),
      ])
    } catch (e) {
      prev.sort = cur.sort
      cur.sort = curSort
      await loadIfNeed(true)
      throw e instanceof Error ? e : new Error('排序更新失败，已回滚')
    }
    sortList()
  }

  /** 下移：与下一行交换 sort。 */
  async function moveDown(id: number): Promise<void> {
    const idx = list.value.findIndex((x) => x.id === id)
    if (idx < 0 || idx >= list.value.length - 1) return
    const cur = list.value[idx]
    const next = list.value[idx + 1]
    const curSort = cur.sort
    cur.sort = next.sort
    next.sort = curSort
    try {
      await Promise.all([
        categoryApi.updateCategory(cur.id, { sort: cur.sort }),
        categoryApi.updateCategory(next.id, { sort: next.sort }),
      ])
    } catch (e) {
      next.sort = cur.sort
      cur.sort = curSort
      await loadIfNeed(true)
      throw e instanceof Error ? e : new Error('排序更新失败，已回滚')
    }
    sortList()
  }

  /** reload 别名（语义化）。 */
  const reload = (force = true) => loadIfNeed(force)

  return {
    list,
    loaded,
    loading,
    loadIfNeed,
    reload,
    reset,
    getName,
    createCategory,
    updateCategory,
    deleteCategory,
    moveUp,
    moveDown,
  }
})