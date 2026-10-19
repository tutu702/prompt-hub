/**
 * 调优页临时状态（PRD P07）。
 *
 * 流式结果 + 调优诉求：仅本次会话有效，切走页面丢弃。
 * 历史缓存：按 promptId 分桶，避免重复请求。
 */
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { tuneApi } from '@/api'
import { usePromptStore } from './prompt'
import type { TuneHistory } from '@/types'

export const useTuneStore = defineStore('tune', () => {
  // 当前流式/最终结果
  const resultText = ref<string>('')
  const streaming = ref<boolean>(false)
  const error = ref<string>('')

  // 调优诉求与上下文开关（页面级）
  const requestText = ref<string>('')
  const useContext = ref<boolean>(false)
  const modelConfigId = ref<number | null>(null)

  // 历史缓存：key = promptId
  const histories = ref<Record<number, TuneHistory[]>>({})
  const historiesLoaded = ref<Record<number, boolean>>({})

  function reset() {
    resultText.value = ''
    streaming.value = false
    error.value = ''
    requestText.value = ''
    modelConfigId.value = null
  }

  async function loadHistories(
    promptId: number,
    force = false,
  ): Promise<TuneHistory[]> {
    if (!force && historiesLoaded.value[promptId]) {
      return histories.value[promptId] || []
    }
    const res = await tuneApi.listTuneHistories(promptId, {
      page: 1,
      page_size: 100,
    })
    histories.value[promptId] = res.list
    historiesLoaded.value[promptId] = true
    return res.list
  }

  /** 把当前调优结果写入历史；写入后立即更新缓存。 */
  async function saveCurrentResult(
    promptId: number,
    payload: {
      request_text: string
      original_content: string
      tuned_content: string
    },
  ): Promise<TuneHistory> {
    const created = await tuneApi.createTuneHistory(promptId, payload)
    const arr = histories.value[promptId] || []
    histories.value[promptId] = [created, ...arr]
    historiesLoaded.value[promptId] = true
    return created
  }

  /** 应用历史到当前 prompt；成功后 promptStore.upsertOne 同步列表。 */
  async function applyHistory(historyId: number): Promise<void> {
    const updated = await tuneApi.applyTuneHistory(historyId)
    const promptStore = usePromptStore()
    promptStore.upsertOne(updated)
  }

  return {
    resultText,
    streaming,
    error,
    requestText,
    useContext,
    modelConfigId,
    histories,
    historiesLoaded,
    reset,
    loadHistories,
    saveCurrentResult,
    applyHistory,
  }
})