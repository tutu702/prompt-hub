/**
 * 大模型配置 store（P08）。
 *
 * - 列表缓存（按用户）
 * - 默认置顶 → 启用优先 → 名称 排序
 * - 乐观更新（create / update / setDefault / remove / toggleEnabled）
 */
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { llmApi } from '@/api'
import type {
  CreateLLMConfigPayload,
  LLMConfig,
  TestConnectionResult,
  UpdateLLMConfigPayload,
} from '@/types'

export const useLLMStore = defineStore('llm', () => {
  const list = ref<LLMConfig[]>([])
  const loaded = ref(false)
  const loading = ref(false)

  const enabledList = computed(() =>
    list.value.filter((c) => c.is_enabled),
  )
  const defaultConfig = computed<LLMConfig | null>(() => {
    const d = list.value.find((c) => c.is_default)
    return d || null
  })

  async function loadIfNeed(force = false): Promise<void> {
    if (loaded.value && !force) return
    loading.value = true
    try {
      const res = await llmApi.listLLMConfigs()
      list.value = res.list || []
      sortList()
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  function sortList() {
    list.value.sort((a, b) => {
      if (a.is_default !== b.is_default) return a.is_default ? -1 : 1
      if (a.is_enabled !== b.is_enabled) return a.is_enabled ? -1 : 1
      return a.name < b.name ? -1 : 1
    })
  }

  async function create(payload: CreateLLMConfigPayload): Promise<LLMConfig> {
    const created = await llmApi.createLLMConfig(payload)
    list.value.push(created)
    if (created.is_default) clearOthersDefault(created.id)
    sortList()
    loaded.value = true
    return created
  }

  async function update(
    id: number,
    payload: UpdateLLMConfigPayload,
  ): Promise<LLMConfig> {
    const updated = await llmApi.updateLLMConfig(id, payload)
    const idx = list.value.findIndex((x) => x.id === id)
    if (idx >= 0) list.value.splice(idx, 1, updated)
    if (updated.is_default) clearOthersDefault(updated.id)
    sortList()
    return updated
  }

  async function remove(id: number): Promise<void> {
    await llmApi.deleteLLMConfig(id)
    list.value = list.value.filter((x) => x.id !== id)
  }

  async function setDefault(id: number): Promise<LLMConfig> {
    const updated = await llmApi.setDefaultLLMConfig(id)
    clearOthersDefault(updated.id)
    sortList()
    return updated
  }

  async function toggleEnabled(c: LLMConfig): Promise<LLMConfig> {
    return update(c.id, { is_enabled: !c.is_enabled })
  }

  async function testConnection(id: number): Promise<TestConnectionResult> {
    return llmApi.testConnectionLLMConfig(id)
  }

  function clearOthersDefault(keepId: number) {
    list.value = list.value.map((x) => ({
      ...x,
      is_default: x.id === keepId,
    }))
  }

  function reset() {
    list.value = []
    loaded.value = false
  }

  return {
    list,
    loaded,
    loading,
    enabledList,
    defaultConfig,
    loadIfNeed,
    create,
    update,
    remove,
    setDefault,
    toggleEnabled,
    testConnection,
    reset,
  }
})