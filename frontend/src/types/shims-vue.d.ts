/**
 * Vue SFC 模块声明
 * 让 TypeScript 能识别 *.vue 文件。
 */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}
