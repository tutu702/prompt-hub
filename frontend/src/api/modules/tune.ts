/**
 * 调优历史相关 API（P07 / F09 / F10）。
 *
 * 路径约定：`http.ts` baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
import { openSSE, type SSEConsumer } from '@/utils/sse'
import type {
  CreateTuneHistoryPayload,
  Prompt,
  TuneHistory,
  TuneHistoryListQuery,
  TuneHistoryListResult,
} from '@/types'

/** GET /api/v1/prompts/:id/tune-histories */
export function listTuneHistories(
  promptId: number,
  query: TuneHistoryListQuery = {},
): Promise<TuneHistoryListResult> {
  const params: Record<string, unknown> = {}
  Object.entries(query).forEach(([k, v]) => {
    if (v !== undefined) params[k] = v
  })
  return http.get(`/v1/prompts/${promptId}/tune-histories`, { params }) as unknown as
    Promise<TuneHistoryListResult>
}

/** POST /api/v1/prompts/:id/tune-histories */
export function createTuneHistory(
  promptId: number,
  payload: CreateTuneHistoryPayload,
): Promise<TuneHistory> {
  return http.post(`/v1/prompts/${promptId}/tune-histories`, payload) as unknown as
    Promise<TuneHistory>
}

/** GET /api/v1/tune-histories/:id */
export function getTuneHistory(id: number): Promise<TuneHistory> {
  return http.get(`/v1/tune-histories/${id}`) as unknown as Promise<TuneHistory>
}

/** POST /api/v1/tune-histories/:id/apply —— 返回更新后的 prompt */
export function applyTuneHistory(id: number): Promise<Prompt> {
  return http.post(`/v1/tune-histories/${id}/apply`) as unknown as Promise<Prompt>
}

/* ---------- SSE 流式调优 ---------- */

/** tuneStream 入参（POST /api/v1/llm/tune 请求体）。 */
export interface TuneStreamPayload {
  prompt_id: number
  request_text: string
  use_context?: boolean
  model_config_id?: number
  /** true = 跳过真实 LLM，本地/CI 用 */
  mock?: boolean
}

/** SSE 事件回调（按后端协议）。 */
export interface TuneStreamCallbacks {
  /** 流式片段回调 */
  onChunk: (delta: string) => void
  /** 完成事件回调（拿到写入后的 history 记录） */
  onDone: (history: TuneHistory) => void
  /** 业务错误（来自 SSE `event: error` 或网络/HTTP 失败） */
  onError: (message: string) => void
}

/**
 * tuneStream 打开 SSE 调优连接。
 *
 * 返回 SSEConsumer：调用方在组件卸载 / 用户点"取消"时调用 abort()。
 */
export function tuneStream(
  payload: TuneStreamPayload,
  callbacks: TuneStreamCallbacks,
): SSEConsumer {
  return openSSE('/v1/llm/tune', payload, {
    onEvent: (e) => {
      if (e.name === 'chunk') {
        const p = e.payload as { delta?: string }
        if (p?.delta) callbacks.onChunk(p.delta)
      } else if (e.name === 'done') {
        const p = e.payload as { history?: TuneHistory }
        if (p?.history) callbacks.onDone(p.history)
        else callbacks.onError('调优完成但未返回历史记录')
      } else if (e.name === 'error') {
        const p = e.payload as { message?: string }
        callbacks.onError(p?.message || '调优失败')
      }
    },
    onError: (err) => {
      callbacks.onError(err.message || '网络异常，请稍后再试')
    },
  })
}

/* ---------- SSE 流式生成 ---------- */

/** generateStream 入参（POST /api/v1/llm/generate 请求体）。 */
export interface GenerateStreamPayload {
  purpose: string
  style?: 'professional' | 'friendly' | 'concise' | 'detailed' | ''
  language?: 'zh' | 'en' | 'auto' | ''
  model_config_id?: number
  mock?: boolean
}

/** 生成的最终结构化结果（与后端 dto 对齐）。 */
export interface GenerateResult {
  title: string
  content: string
  description: string
}

/** SSE 事件回调。 */
export interface GenerateStreamCallbacks {
  onChunk: (delta: string) => void
  onDone: (result: GenerateResult) => void
  onError: (message: string) => void
}

/**
 * generateStream 打开 SSE 生成连接。
 */
export function generateStream(
  payload: GenerateStreamPayload,
  callbacks: GenerateStreamCallbacks,
): SSEConsumer {
  return openSSE('/v1/llm/generate', payload, {
    onEvent: (e) => {
      if (e.name === 'chunk') {
        const p = e.payload as { delta?: string }
        if (p?.delta) callbacks.onChunk(p.delta)
      } else if (e.name === 'done') {
        const p = e.payload as { result?: GenerateResult }
        if (p?.result) callbacks.onDone(p.result)
        else callbacks.onError('生成完成但未返回结果')
      } else if (e.name === 'error') {
        const p = e.payload as { message?: string }
        callbacks.onError(p?.message || '生成失败')
      }
    },
    onError: (err) => {
      callbacks.onError(err.message || '网络异常，请稍后再试')
    },
  })
}