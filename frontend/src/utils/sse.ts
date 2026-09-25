/**
 * SSE（Server-Sent Events）消费工具。
 *
 * 使用 fetch + ReadableStream 读取 `text/event-stream` 响应，按 SSE 规范解析：
 *   - 一行 `event: <name>`：下一个 data 行的归属事件
 *   - 一行 `data: <payload>`：事件数据
 *   - 空行：本帧结束，触发回调
 *
 * 设计要点：
 *   - 返回 AbortController；调用方可通过 controller.abort() 主动取消（关闭页面、点取消按钮）。
 *   - 业务错误（HTTP 4xx/5xx）作为 Error 抛给 onError；SSE 协议内的 `event: error` 也走 onEvent。
 *   - 不依赖第三方库；axios 拦截器对 SSE 流不解包，本工具直接走 fetch。
 *
 * 后端约定（见 handler/llm_tune.go）：
 *   event: chunk  data: {"delta":"..."}
 *   event: done   data: {"history":{...}}
 *   event: error  data: {"message":"..."}
 *
 * Token 从 localStorage('ph_access') 取，与 user store 写入位置保持一致（store/user.ts STORAGE.access）。
 */

export interface SSEEvent {
  /** 事件名；缺省时为 "message" */
  name: string
  /** 解析后的 data（JSON 字符串经 JSON.parse，失败则保留原始字符串） */
  payload: unknown
}

export interface SSEConsumer {
  abort(): void
}

export interface SSEHandlers {
  onEvent?: (e: SSEEvent) => void
  onError?: (err: Error) => void
}

/** http baseURL：与 src/api/http.ts 保持一致 */
function baseURL(): string {
  // import.meta.env 在 Vite 构建时静态替换
  const raw = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api'
  return raw
}

/**
 * 打开一个 SSE 连接：POST JSON + Accept text/event-stream。
 */
export function openSSE(
  url: string,
  body: unknown,
  handlers: SSEHandlers,
): SSEConsumer {
  const controller = new AbortController()

  ;(async () => {
    try {
      const fullURL = url.startsWith('http') ? url : `${baseURL()}${url}`

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
      }
      const token = localStorage.getItem('ph_access')
      if (token) headers['Authorization'] = `Bearer ${token}`

      const resp = await fetch(fullURL, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
        signal: controller.signal,
      })

      if (!resp.ok) {
        // 业务错误（4xx/5xx）：尝试读 {code,message}
        let serverMsg = `HTTP ${resp.status}`
        try {
          const j = (await resp.json()) as { message?: string }
          if (j?.message) serverMsg = j.message
        } catch {
          /* noop */
        }
        handlers.onError?.(new Error(serverMsg))
        return
      }

      if (!resp.body) {
        handlers.onError?.(new Error('SSE 响应无 body'))
        return
      }

      const reader = resp.body.getReader()
      const decoder = new TextDecoder('utf-8')
      let buffer = ''
      let eventName = 'message'
      let dataLines: string[] = []

      const flushFrame = () => {
        if (dataLines.length === 0) {
          eventName = 'message'
          return
        }
        const raw = dataLines.join('\n')
        dataLines = []
        let payload: unknown = raw
        try {
          payload = JSON.parse(raw)
        } catch {
          /* 保留 raw 字符串 */
        }
        handlers.onEvent?.({ name: eventName, payload })
        eventName = 'message'
      }

      while (true) {
        const { value, done } = await reader.read()
        if (done) {
          // 收尾：处理剩余 buffer（可能是不带尾部 \n\n 的最后一帧）
          if (buffer.length > 0) {
            buffer = buffer.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
            for (const line of buffer.split('\n')) {
              if (line.startsWith(':')) {
                /* 注释行 */
              } else if (line.startsWith('event:')) {
                eventName = line.slice(6).trim() || 'message'
              } else if (line.startsWith('data:')) {
                dataLines.push(line.slice(5).trimStart())
              }
            }
            // buffer 末尾未必有 \n\n 收尾，这里手动 flush 一次避免漏掉最后一帧
            flushFrame()
          }
          break
        }
        buffer += decoder.decode(value, { stream: true })
        // 拆 frame（连续两个换行 = 一帧结束）
        // 注意：frame = buffer.slice(0, nlAt) 不包含末尾的 \n\n，
        // 因此 frame.split('\n') 不会有空字符串行，flushFrame 必须放在循环体外显式调用。
        let nlAt = buffer.indexOf('\n\n')
        while (nlAt >= 0) {
          const frame = buffer.slice(0, nlAt)
          buffer = buffer.slice(nlAt + 2)
          for (const line of frame.split('\n')) {
            if (line.startsWith(':')) {
              /* 注释行 */
            } else if (line.startsWith('event:')) {
              eventName = line.slice(6).trim() || 'message'
            } else if (line.startsWith('data:')) {
              dataLines.push(line.slice(5).trimStart())
            }
            // SSE 帧内的空行不会再出现（已被 \n\n 切走），
            // 故这里不再依赖 line === '' 触发 flush。
          }
          // 帧解析完，flush 一次
          flushFrame()
          nlAt = buffer.indexOf('\n\n')
        }
      }
    } catch (e) {
      if ((e as { name?: string })?.name === 'AbortError') {
        return
      }
      handlers.onError?.(e as Error)
    }
  })()

  return {
    abort: () => controller.abort(),
  }
}