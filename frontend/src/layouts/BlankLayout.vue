<script setup lang="ts">
/**
 * 空白布局：用于登录 / 注册等无导航条页面。
 * 蓝色渐变背景 + 居中卡片，遵循 PRD §3.1 主色调。
 */
</script>

<template>
  <div class="auth-layout">
    <div class="auth-card">
      <header class="auth-card__header">
        <div class="auth-card__logo">PromptHub</div>
        <div class="auth-card__slogan">AI 提示词资产管理工具</div>
      </header>
      <main class="auth-card__body">
        <router-view />
      </main>
      <footer class="auth-card__footer">© 2026 PromptHub</footer>
    </div>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1677ff 0%, #4096ff 100%);
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.16);
  padding: 36px 32px 24px;
}

.auth-card__header {
  text-align: center;
  margin-bottom: 24px;
}

.auth-card__logo {
  font-size: 26px;
  font-weight: 700;
  color: #1677ff;
  letter-spacing: 0.5px;
}

.auth-card__slogan {
  margin-top: 6px;
  font-size: 13px;
  color: #94a3b8;
}

.auth-card__body {
  margin-bottom: 16px;
}

.auth-card__footer {
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
}

@media (max-width: 480px) {
  .auth-card {
    padding: 28px 20px 18px;
  }
  .auth-card__logo {
    font-size: 22px;
  }
}
</style>