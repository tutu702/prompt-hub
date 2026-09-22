<script setup lang="ts">
/**
 * 提示词库占位页（已登录可见）。
 * 真实业务（CRUD / 搜索 / 筛选 / 导出 / 调优）按 PRD §三 在后续迭代补齐。
 */
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await userStore.logout()
  } catch (e) {
    // 用户已清空本地 token 视为完成
  }
  ElMessage.success('已退出登录')
  router.replace({ name: 'Login' })
}
</script>

<template>
  <div class="prompts">
    <header class="prompts__header">
      <div>
        <h2 class="prompts__title">提示词库</h2>
        <p class="prompts__sub">
          欢迎回来，
          <strong>{{ userStore.userInfo?.nickname || userStore.userInfo?.email || '用户' }}</strong>
        </p>
      </div>
      <el-button @click="handleLogout">退出登录</el-button>
    </header>

    <el-card class="prompts__card" shadow="never">
      <el-empty
        description="提示词业务尚未开放，敬请期待"
        :image-size="120"
      >
        <template #image>
          <div class="prompts__emoji">📝</div>
        </template>
        <el-button type="primary" plain @click="router.push('/')">返回首页</el-button>
      </el-empty>
    </el-card>
  </div>
</template>

<style scoped>
.prompts {
  max-width: 960px;
  margin: 0 auto;
  padding: 8px 4px;
}

.prompts__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.prompts__title {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.prompts__sub {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

.prompts__card {
  border-radius: 12px;
  padding: 24px 0;
}

.prompts__emoji {
  font-size: 64px;
  line-height: 1;
}
</style>