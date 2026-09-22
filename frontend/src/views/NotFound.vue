<script setup lang="ts">
/**
 * 404 兜底视图
 */
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="notfound">
    <h1>404</h1>
    <p>页面不存在</p>
    <el-button type="primary" @click="router.push('/')">返回首页</el-button>
  </div>
</template>

<style scoped>
.notfound {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 16px;
}
.notfound h1 {
  font-size: 72px;
  margin: 0;
  color: var(--el-color-primary);
}
</style>
