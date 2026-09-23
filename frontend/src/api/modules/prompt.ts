/**
 * 提示词相关 API
 *
 * URL 前缀约定：`http.ts` 已设 baseURL='/api'，因此这里只写 '/v1/...'。
 */
import http from '../http'
import type {
  BatchDeletePayload,
  CreatePromptPayload,
  Prompt,
  PromptListQuery,
  PromptListResult,
  ToggleFavoritePayload,
  UpdatePromptPayload,
} from '@/types'

/** GET /api/v1/prompts */
export function listPrompts(query: PromptListQuery = {}): Promise<PromptListResult> {
  // 把 undefined 字段过滤掉；null 会保留（用于 category_id=null 含义）
  const params: Record<string, unknown> = {}
  Object.entries(query).forEach(([k, v]) => {
    if (v !== undefined) params[k] = v
  })
  return http.get('/v1/prompts', { params }) as unknown as Promise<PromptListResult>
}

/** GET /api/v1/prompts/:id */
export function getPrompt(id: number): Promise<Prompt> {
  return http.get(`/v1/prompts/${id}`) as unknown as Promise<Prompt>
}

/** POST /api/v1/prompts */
export function createPrompt(payload: CreatePromptPayload): Promise<Prompt> {
  return http.post('/v1/prompts', payload) as unknown as Promise<Prompt>
}

/** PUT /api/v1/prompts/:id */
export function updatePrompt(
  id: number,
  payload: UpdatePromptPayload,
): Promise<Prompt> {
  return http.put(`/v1/prompts/${id}`, payload) as unknown as Promise<Prompt>
}

/** DELETE /api/v1/prompts/:id */
export function deletePrompt(id: number): Promise<void> {
  return http.delete(`/v1/prompts/${id}`) as unknown as Promise<void>
}

/** POST /api/v1/prompts/batch-delete */
export function batchDeletePrompts(
  payload: BatchDeletePayload,
): Promise<{ deleted: number }> {
  return http.post(
    '/v1/prompts/batch-delete',
    payload,
  ) as unknown as Promise<{ deleted: number }>
}

/** POST /api/v1/prompts/:id/favorite */
export function toggleFavorite(
  id: number,
  payload: ToggleFavoritePayload,
): Promise<{ is_favorite: boolean }> {
  return http.post(
    `/v1/prompts/${id}/favorite`,
    payload,
  ) as unknown as Promise<{ is_favorite: boolean }>
}