/**
 * 分类 store：缓存用户的分类列表，跨页面复用。
 *
 * `loaded` 标记避免重复拉取；登录态变化时由调用方调 `reset()` 触发重载。
 */
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { categoryApi } from '@/api'
import type { Category } from '@/types'

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

  return {
    list,
    loaded,
    loading,
    loadIfNeed,
    reset,
    getName,
  }
})