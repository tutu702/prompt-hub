<script setup lang="ts">
/**
 * 默认布局：顶部导航 + 内容区
 *
 * 顶部结构（从左到右）：
 * - 品牌：图标 + 名称
 * - 主导航：首页 / 分类 / AI 调优 / 设置
 * - 顶部全局搜索框（⌘K / Ctrl+K）
 * - 通知铃铛
 * - 用户头像下拉
 *
 * - 默认布局: 顶部导航 + 内容区
 * - nav 项当前激活态通过 router-link-active 自动高亮
 * - 用户菜单（登出、设置入口等）通过 el-dropdown 触发
 */
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  MagicStick as BrandIcon,
  Bell as BellIcon,
  Search as SearchIcon,
  SwitchButton as LogoutIcon,
  User as UserIcon,
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { usePromptStore } from '@/stores/prompt'
import { debounce } from '@/utils'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const promptStore = usePromptStore()

/** 用户名首字母（用于头像占位） */
const userInitial = computed(() => {
  const name =
    userStore.userInfo?.nickname || userStore.userInfo?.email || '?'
  return name.trim().charAt(0).toUpperCase()
})

/** 主导航菜单（高亮由 router-link-active 自动处理） */
const navItems = [
  { label: '首页', to: '/prompts' },
  { label: '分类', to: '/categories' },
  { label: '模型配置', to: '/settings/llm' },
  { label: '设置', to: '/settings' },
] as const

/** 搜索输入（与 prompt store 双向同步，但带防抖） */
const headerSearchInput = ref('')
const debouncedRouteSearch = debounce((...args: unknown[]) => {
  const kw = args[0] as string
  // 仅在当前位于 /prompts 时同步到列表 store，避免误触发其他页逻辑
  if (route.path.startsWith('/prompts')) {
    promptStore.setKeyword(kw)
    promptStore.goPage(1)
    promptStore.fetchList()
  }
}, 300)

function onHeaderSearchInput(v: string) {
  headerSearchInput.value = v
  if (route.path !== '/prompts') {
    // 非提示词页面：跳转并带 query 让 /prompts 接管
    router.push({ path: '/prompts', query: v ? { q: v } : undefined })
    return
  }
  debouncedRouteSearch(v)
}

function clearHeaderSearch() {
  headerSearchInput.value = ''
  onHeaderSearchInput('')
}

/** 顶部搜索框 ⌘K / Ctrl+K 快捷键 */
function onSearchKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    const el = document.getElementById('layout-global-search') as HTMLInputElement | null
    el?.focus()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onSearchKeydown)
})
onUnmounted(() => {
  window.removeEventListener('keydown', onSearchKeydown)
})

/** 用户菜单 */
function handleMenuCommand(cmd: string) {
  if (cmd === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消',
    })
      .then(async () => {
        await userStore.logout()
        ElMessage.success('已退出登录')
        // 当前已位于受保护路由，store 清空后页面不会自动跳转；
        // 显式 push 到 /login 让路由守卫把用户带到正确位置
        router.push('/login')
      })
      .catch(() => {})
  }
}

function onBellClick() {
  ElMessage.info('暂无新通知')
}
</script>

<template>
  <div class="layout">
    <header class="layout__header">
      <!-- 品牌 -->
      <router-link to="/prompts" class="layout__brand">
        <span class="layout__brand-mark">
          <el-icon><BrandIcon /></el-icon>
        </span>
        <span class="layout__brand-name">PromptHub</span>
      </router-link>

      <!-- 主导航 -->
      <nav class="layout__nav">
        <router-link
          v-for="item in navItems"
          :key="item.label"
          :to="item.to"
          class="layout__nav-item"
        >
          {{ item.label }}
        </router-link>
      </nav>

      <div class="layout__spacer" />

      <!-- 顶部全局搜索 -->
      <div class="layout__search">
        <el-input
          id="layout-global-search"
          v-model="headerSearchInput"
          placeholder="搜索提示词..."
          clearable
          :prefix-icon="SearchIcon"
          @input="onHeaderSearchInput"
          @clear="clearHeaderSearch"
        />
        <span class="layout__search-kbd" aria-hidden="true">⌘K</span>
      </div>

      <!-- 通知铃铛 -->
      <button
        type="button"
        class="layout__icon-btn"
        title="通知"
        @click="onBellClick"
      >
        <el-icon><BellIcon /></el-icon>
      </button>

      <!-- 用户头像下拉 -->
      <el-dropdown trigger="click" @command="handleMenuCommand">
        <button class="layout__avatar" type="button" :title="userStore.userInfo?.nickname || '用户'">
          {{ userInitial }}
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><UserIcon /></el-icon>
              <span style="margin-left: 6px">
                {{ userStore.userInfo?.nickname || userStore.userInfo?.email || '未登录' }}
              </span>
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <el-icon><LogoutIcon /></el-icon>
              <span style="margin-left: 6px">退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </header>

    <main class="layout__main">
      <slot />
    </main>

    <footer class="layout__footer">
      <span>© 2026 PromptHub · 让提示词成为可复用的资产</span>
    </footer>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f8fafc;
}

.layout__header {
  position: sticky;
  top: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 24px;
  height: 60px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: saturate(180%) blur(10px);
  border-bottom: 1px solid #eef2f7;
}

.layout__brand {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 12px;
  flex-shrink: 0;
}

.layout__brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #ffffff;
  font-size: 16px;
}

.layout__brand-name {
  font-weight: 700;
  font-size: 16px;
  color: #1e293b;
  letter-spacing: 0.2px;
}

.layout__nav {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  margin-left: 48px;
}

.layout__nav-item {
  position: relative;
  padding: 6px 12px;
  font-size: 14px;
  color: #475569;
  border-radius: 8px;
  transition: background 0.15s, color 0.15s;
}

.layout__nav-item:hover {
  background: #f1f5f9;
  color: #1e293b;
}

.layout__nav-item.router-link-active {
  background: #eef4ff;
  color: #2563eb;
  font-weight: 600;
}

.layout__spacer {
  flex: 1;
}

/* 顶部全局搜索：带 ⌘K 角标 */
.layout__search {
  position: relative;
  width: 280px;
  max-width: 28vw;
}

.layout__search :deep(.el-input__wrapper) {
  background: #f1f5f9;
  border-radius: 10px;
  box-shadow: none;
  padding: 2px 8px;
}

.layout__search :deep(.el-input__inner) {
  height: 34px;
  font-size: 13px;
  color: #334155;
}

.layout__search-kbd {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  font-size: 11px;
  font-weight: 500;
  color: #64748b;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 1px 6px;
}

/* 铃铛按钮 */
.layout__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 0;
  background: transparent;
  color: #64748b;
  border-radius: 8px;
  cursor: pointer;
  font-size: 18px;
  transition: background 0.15s, color 0.15s;
}

.layout__icon-btn:hover {
  background: #f1f5f9;
  color: #2563eb;
}

/* 用户头像 */
.layout__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  transition: transform 0.15s, box-shadow 0.15s;
}

.layout__avatar:hover {
  transform: scale(1.04);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.35);
}

.layout__main {
  flex: 1;
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px;
}

.layout__footer {
  text-align: center;
  padding: 16px;
  color: #94a3b8;
  font-size: 12px;
}

@media (max-width: 960px) {
  .layout__nav { display: none; }
  .layout__search { width: 180px; max-width: 40vw; }
  .layout__search-kbd { display: none; }
  .layout__main { padding: 16px; }
}
</style>
