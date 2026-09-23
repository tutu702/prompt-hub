/**
 * 调优历史相关 API（P07）。
 *
 * 路径约定：`http.ts` baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
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