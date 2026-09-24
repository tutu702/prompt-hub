/**
 * 大模型配置相关 API（P08）。
 *
 * URL 约定：`http.ts` baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
import type {
  CreateLLMConfigPayload,
  LLMConfig,
  TestConnectionResult,
  UpdateLLMConfigPayload,
} from '@/types'

/** GET /api/v1/llm-configs */
export function listLLMConfigs(): Promise<{ list: LLMConfig[] }> {
  return http.get('/v1/llm-configs') as unknown as Promise<{ list: LLMConfig[] }>
}

/** GET /api/v1/llm-configs/:id */
export function getLLMConfig(id: number): Promise<LLMConfig> {
  return http.get(`/v1/llm-configs/${id}`) as unknown as Promise<LLMConfig>
}

/** POST /api/v1/llm-configs */
export function createLLMConfig(
  payload: CreateLLMConfigPayload,
): Promise<LLMConfig> {
  return http.post('/v1/llm-configs', payload) as unknown as Promise<LLMConfig>
}

/** PUT /api/v1/llm-configs/:id */
export function updateLLMConfig(
  id: number,
  payload: UpdateLLMConfigPayload,
): Promise<LLMConfig> {
  return http.put(`/v1/llm-configs/${id}`, payload) as unknown as Promise<LLMConfig>
}

/** DELETE /api/v1/llm-configs/:id */
export function deleteLLMConfig(id: number): Promise<void> {
  return http.delete(`/v1/llm-configs/${id}`) as unknown as Promise<void>
}

/** POST /api/v1/llm-configs/:id/default */
export function setDefaultLLMConfig(id: number): Promise<LLMConfig> {
  return http.post(`/v1/llm-configs/${id}/default`) as unknown as Promise<LLMConfig>
}

/** POST /api/v1/llm-configs/:id/test-connection */
export function testConnectionLLMConfig(id: number): Promise<TestConnectionResult> {
  return http.post(
    `/v1/llm-configs/${id}/test-connection`,
  ) as unknown as Promise<TestConnectionResult>
}