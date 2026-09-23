/**
 * 分类相关 API
 *
 * URL 前缀约定：`http.ts` 已设 baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
import type { Category, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'

/** GET /api/v1/categories */
export function listCategories(): Promise<{ list: Category[] }> {
  return http.get('/v1/categories') as unknown as Promise<{ list: Category[] }>
}

/** POST /api/v1/categories */
export function createCategory(payload: CreateCategoryPayload): Promise<Category> {
  return http.post('/v1/categories', payload) as unknown as Promise<Category>
}

/** PUT /api/v1/categories/:id */
export function updateCategory(
  id: number,
  payload: UpdateCategoryPayload,
): Promise<Category> {
  return http.put(`/v1/categories/${id}`, payload) as unknown as Promise<Category>
}

/** DELETE /api/v1/categories/:id */
export function deleteCategory(id: number): Promise<void> {
  return http.delete(`/v1/categories/${id}`) as unknown as Promise<void>
}