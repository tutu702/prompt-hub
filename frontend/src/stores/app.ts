/**
 * 全局应用状态：站点标题、侧边栏折叠等基础 UI 状态。
 * 业务 store（如用户、提示词、分类）按需在 stores/ 下新增。
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const title = ref<string>(import.meta.env.VITE_APP_TITLE || 'PromptHub')
  const sidebarCollapsed = ref<boolean>(false)
  const theme = ref<'light' | 'dark'>('light')

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setTheme(t: 'light' | 'dark') {
    theme.value = t
  }

  return { title, sidebarCollapsed, theme, toggleSidebar, setTheme }
})
