<script setup lang="ts">
/**
 * 默认首页占位视图
 * 后续按 PRD 添加：提示词展示页 / 登录 / 注册等真实页面。
 */
</script>

<template>
  <div class="home">
    <header class="home__header">
      <h1>PromptHub</h1>
      <p class="home__sub">AI 提示词资产管理工具 · 项目骨架已就绪</p>
    </header>

    <el-card class="home__card" shadow="hover">
      <template #header>
        <span>当前进度</span>
      </template>
      <ul class="home__list">
        <li>✅ Vue3 + Vite + TypeScript 已就绪</li>
        <li>✅ Element Plus + Pinia + Axios 已集成</li>
        <li>✅ 路由 / 状态管理 / HTTP 客户端已配置</li>
        <li>⏳ 业务页面按 PRD 后续实现</li>
      </ul>
    </el-card>
  </div>
</template>

<style scoped>
.home {
  max-width: 960px;
  margin: 0 auto;
  padding: 48px 24px;
}
.home__header {
  text-align: center;
  margin-bottom: 32px;
}
.home__header h1 {
  font-size: 36px;
  font-weight: 700;
  color: var(--el-color-primary);
  margin: 0;
}
.home__sub {
  margin-top: 8px;
  color: var(--el-text-color-secondary);
}
.home__card {
  border-radius: 12px;
}
.home__list {
  margin: 0;
  padding-left: 20px;
  line-height: 2;
  color: var(--el-text-color-regular);
}
</style>
