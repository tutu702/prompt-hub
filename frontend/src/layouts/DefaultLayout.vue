<script setup lang="ts">
/**
 * 默认布局：顶部导航 + 内容区
 * 后续可扩展为带侧边栏的 MainLayout、空白布局 BlankLayout 等。
 */
</script>

<template>
  <div class="layout">
    <header class="layout__header">
      <div class="layout__logo">PromptHub</div>
      <nav class="layout__nav">
        <router-link to="/home">首页</router-link>
      </nav>
    </header>
    <main class="layout__main">
      <slot />
    </main>
    <footer class="layout__footer">
      <span>© 2026 PromptHub</span>
    </footer>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.layout__header {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 0 32px;
  height: 64px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}
.layout__logo {
  font-weight: 700;
  font-size: 18px;
  color: #1677ff;
}
.layout__nav a {
  color: #475569;
}
.layout__nav a.router-link-active {
  color: #1677ff;
}
.layout__main {
  flex: 1;
  padding: 24px 32px;
}
.layout__footer {
  text-align: center;
  padding: 16px;
  color: #94a3b8;
  font-size: 12px;
  border-top: 1px solid #e2e8f0;
}
</style>
