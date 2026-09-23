/**
 * Mock 流式输出工具（P07 占位实现）。
 *
 * 把整段文本切成 1-3 字一段，用 setTimeout 25-40ms / 段推到 onChunk。
 * 返回 cancel() 用于中止（用户点"取消"或组件卸载时调用）。
 *
 * 不污染全局 utils：仅作为 P07 私有工具放在 components/tune/。
 */

export interface MockStreamHandle {
  cancel(): void
}

/**
 * mockStream 模拟大模型逐段流式输出。
 *
 * @param text   完整输出文本
 * @param onChunk 每段推到的回调（同步触发）
 */
export function mockStream(
  text: string,
  onChunk: (chunk: string) => void,
): MockStreamHandle {
  let cancelled = false
  const tokens = text.match(/[\s\S]{1,3}/g) || []
  let i = 0

  function next() {
    if (cancelled || i >= tokens.length) return
    onChunk(tokens[i])
    i += 1
    setTimeout(next, 25 + Math.floor(Math.random() * 15))
  }
  // 首次延迟稍大，让 UI 先切到流式态
  setTimeout(next, 50)

  return {
    cancel: () => {
      cancelled = true
    },
  }
}