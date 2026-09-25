<script setup lang="ts">
/**
 * 提示词创建页（PRD P04）。
 *
 * 顶部 Tabs 切换【手动创建】/【AI 生成】。
 * - 手动 Tab：渲染 <PromptForm>，底部固定【取消】【保存并新建】【保存】三按钮。
 * - AI Tab：渲染 <AIGeneratePanel>，本期为占位（仅 toast 提示去配置大模型）。
 *
 * 离开守卫：表单脏态时路由离开 / 取消按钮 → 二次确认弹窗。
 *
 * 保存成功后跳 /prompts 列表（依赖 store.upsertOne 把新条目插到顶部）。
 */
import { computed, onMounted, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import PromptForm from '@/components/prompts/PromptForm.vue'
import AIGeneratePanel from '@/components/prompts/AIGeneratePanel.vue'
import { usePromptStore } from '@/stores/prompt'
import { useCategoryStore } from '@/stores/category'
import type { CreatePromptPayload } from '@/types'

const router = useRouter()
const route = useRoute()
const promptStore = usePromptStore()
const categoryStore = useCategoryStore()

type TabKey = 'manual' | 'ai'
const tab = ref<TabKey>(
  route.query.tab === 'ai' ? 'ai' : 'manual',
)

// dirty 状态：手动表单 / AI 表单任意一方变脏
const formDirty = ref(false)
const aiDirty = ref(false)
const isDirty = computed(() => formDirty.value || aiDirty.value)

// PromptForm ref + submitting 状态（透传用于按钮 loading）
const manualFormRef = ref<InstanceType<typeof PromptForm> | null>(null)
const submitting = ref(false)

// 页面进入时拉取分类下拉
onMounted(async () => {
  await categoryStore.loadIfNeed()
})

async function confirmLeave(): Promise<boolean> {
  try {
    await ElMessageBox.confirm(
      '当前表单有未保存的修改，确定离开吗？',
      '提示',
      {
        type: 'warning',
        confirmButtonText: '离开',
        cancelButtonText: '留在页面',
      },
    )
    return true
  } catch {
    return false
  }
}

// 路由离开守卫
onBeforeRouteLeave(async () => {
  if (!isDirty.value) return true
  return await confirmLeave()
})

async function handleCancel() {
  if (isDirty.value) {
    const allow = await confirmLeave()
    if (!allow) return
  }
  router.push('/prompts')
}

async function handleManualSave(payload: CreatePromptPayload) {
  submitting.value = true
  try {
    const created = await promptStore.createPrompt(payload)
    ElMessage.success(`已创建【${created.title}】`)
    // 离开守卫要先放行：清脏再 push
    formDirty.value = false
    router.push('/prompts')
  } catch (err) {
    ElMessage.error((err as Error).message || '创建失败')
  } finally {
    submitting.value = false
  }
}

async function handleManualSaveAndNew(payload: CreatePromptPayload) {
  submitting.value = true
  try {
    const created = await promptStore.createPrompt(payload)
    ElMessage.success(`已创建【${created.title}】，请继续创建下一条`)
    // 保存并新建：清脏 + 重置表单（不离开页面）
    formDirty.value = false
    manualFormRef.value?.reset()
  } catch (err) {
    ElMessage.error((err as Error).message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function onFormDirty(v: boolean) {
  formDirty.value = v
}

function onAiDirty(v: boolean) {
  aiDirty.value = v
}

function handleAiApply(payload: {
  title?: string
  content: string
  description?: string
  tags?: string[]
}) {
  // 切换到手动 Tab 并把 AI 输出套到表单上（占位逻辑未真正调用）
  tab.value = 'manual'
  manualFormRef.value?.applyAiResult(payload)
  aiDirty.value = false
  formDirty.value = true
}

// 暴露给底部按钮的快捷调用
function onSaveAndNewClick() {
  manualFormRef.value?.saveAndNew()
}
</script>

<template>
  <div class="prompt-create">
    <header class="prompt-create__header">
      <div>
        <h2 class="prompt-create__title">新建提示词</h2>
        <p class="prompt-create__sub">选择手动创建，或用 AI 帮你生成</p>
      </div>
    </header>

    <el-tabs v-model="tab" class="prompt-create__tabs">
      <el-tab-pane label="手动创建" name="manual">
        <PromptForm
          ref="manualFormRef"
          @dirty="onFormDirty"
          @submit="handleManualSave"
          @save-and-new="handleManualSaveAndNew"
        />
      </el-tab-pane>
      <el-tab-pane label="AI 生成" name="ai">
        <AIGeneratePanel @dirty="onAiDirty" @apply="handleAiApply" />
      </el-tab-pane>
    </el-tabs>

    <footer v-if="tab === 'manual'" class="prompt-create__footer">
      <el-button @click="handleCancel">取消</el-button>
      <el-button :loading="submitting" @click="onSaveAndNewClick">保存</el-button>
    </footer>
  </div>
</template>

<style scoped>
.prompt-create {
  max-width: 960px;
  margin: 0 auto;
  padding: 16px 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.prompt-create__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 0 8px;
}

.prompt-create__title {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.prompt-create__sub {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

.prompt-create__tabs {
  background: transparent;
}

.prompt-create__tabs :deep(.el-tabs__nav-wrap)::after {
  background-color: #e2e8f0;
}

.prompt-create__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eef2f7;
  position: sticky;
  bottom: 16px;
}

@media (max-width: 768px) {
  .prompt-create {
    padding: 12px 14px 24px;
  }
  .prompt-create__footer {
    flex-wrap: wrap;
    gap: 8px;
  }
  .prompt-create__header {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>