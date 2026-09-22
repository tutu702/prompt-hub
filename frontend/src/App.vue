<script setup lang="ts">
/**
 * 应用根组件
 * 通过 Vue Router 渲染当前路由对应的视图。
 */
</script>

<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<style scoped>
/* 根组件样式 - 全局基础样式请放入 src/styles */
</style>
