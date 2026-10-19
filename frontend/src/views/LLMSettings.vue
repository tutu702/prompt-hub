<script setup lang="ts">
/**
 * 模型配置页（PRD P08 /settings/llm）。
 *
 * 列表：名称 / 供应商 / 模型 / BaseURL / API Key（脱敏） / 启用 / 默认 / 操作
 * 行内操作：测试连接 / 设为默认 / 启用停用 / 编辑 / 删除
 */
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus as PlusIcon,
  Connection as ConnectionIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Star as StarIcon,
  StarFilled as StarFilledIcon,
} from '@element-plus/icons-vue'

import LLMConfigFormDialog from '@/components/llm/LLMConfigFormDialog.vue'
import { useLLMStore } from '@/stores/llm'
import type { LLMConfig } from '@/types'

const llmStore = useLLMStore()

const dialogOpen = ref(false)
const editing = ref<LLMConfig | null>(null)
const submitting = ref(false)

onMounted(async () => {
  await llmStore.loadIfNeed()
})

function openCreate() {
  editing.value = null
  dialogOpen.value = true
}

function openEdit(c: LLMConfig) {
  editing.value = c
  dialogOpen.value = true
}

async function onDialogSubmit(payload: {
  id: number | null
  data: Parameters<typeof llmStore.create>[0] | Parameters<typeof llmStore.update>[1]
}) {
  submitting.value = true
  try {
    if (payload.id) {
      await llmStore.update(payload.id, payload.data as Parameters<typeof llmStore.update>[1])
      ElMessage.success('已更新')
    } else {
      await llmStore.create(payload.data as Parameters<typeof llmStore.create>[0])
      ElMessage.success('已创建')
    }
    dialogOpen.value = false
  } catch (err) {
    ElMessage.error((err as Error).message || '保存失败')
  } finally {
    submitting.value = false
  }
}

async function handleToggleEnabled(c: LLMConfig) {
  try {
    await llmStore.toggleEnabled(c)
  } catch (err) {
    ElMessage.error((err as Error).message || '切换启用状态失败')
  }
}

async function handleSetDefault(c: LLMConfig) {
  if (c.is_default) return
  try {
    await llmStore.setDefault(c.id)
    ElMessage.success(`已将「${c.name}」设为默认模型`)
  } catch (err) {
    ElMessage.error((err as Error).message || '设置默认失败')
  }
}

async function handleTestConnection(c: LLMConfig) {
  try {
    const res = await llmStore.testConnection(c.id)
    if (res.ok) {
      await ElMessageBox.alert(
        `连接成功，用时 ${res.latency_ms} ms`,
        '测试连接',
        { type: 'success', confirmButtonText: '好的' },
      )
    } else {
      const messages: Record<string, string> = {
        AUTH_FAILED: '鉴权失败，请检查 API Key',
        TIMEOUT: '连接超时',
        MODEL_NOT_FOUND: '模型不存在或无权访问',
        QUOTA_EXHAUSTED: '配额已用尽',
        NETWORK_ERROR: '网络异常',
      }
      const friendly = (res.error_code && messages[res.error_code]) || '未知错误'
      await ElMessageBox.alert(
        `${friendly}\n\n${res.error_message || ''}`,
        '测试连接失败',
        { type: 'error', confirmButtonText: '我知道了' },
      )
    }
  } catch (err) {
    ElMessage.error((err as Error).message || '测试连接请求失败')
  }
}

async function handleDelete(c: LLMConfig) {
  try {
    await ElMessageBox.confirm(
      `确定删除配置【${c.name}】吗？`,
      '提示',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
      },
    )
  } catch {
    return
  }
  try {
    await llmStore.remove(c.id)
    ElMessage.success('已删除')
  } catch (err) {
    ElMessage.error((err as Error).message || '删除失败')
  }
}
</script>

<template>
  <div class="llm-settings">
    <header class="llm-settings__header">
      <div class="llm-settings__title-row">
        <div>
          <h2 class="llm-settings__title">模型配置</h2>
          <p class="llm-settings__sub">管理 AI 调优与生成所用的供应商、API Key 与默认模型</p>
        </div>
        <el-button type="primary" :icon="PlusIcon" @click="openCreate">
          新建配置
        </el-button>
      </div>
    </header>

    <el-table
      v-loading="llmStore.loading"
      :data="llmStore.list"
      class="llm-settings__table"
      empty-text="还没有配置，点击右上「新建配置」开始"
      :row-style="{ height: '64px' }"
    >
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <div class="llm-settings__name">
            <el-icon v-if="row.is_default" class="llm-settings__star" :title="'默认模型'">
              <StarFilledIcon />
            </el-icon>
            <el-icon v-else class="llm-settings__star--off" :title="'非默认'">
              <StarIcon />
            </el-icon>
            <span class="llm-settings__name-text">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="供应商" min-width="110" prop="provider_label" />

      <el-table-column label="模型" min-width="160" prop="model" />

      <el-table-column label="BaseURL" min-width="200">
        <template #default="{ row }">
          <span class="llm-settings__url" :title="row.base_url">{{ row.base_url }}</span>
        </template>
      </el-table-column>

      <el-table-column label="API Key" min-width="160">
        <template #default="{ row }">
          <code class="llm-settings__key">{{ row.api_key_masked }}</code>
        </template>
      </el-table-column>

      <el-table-column label="启用" width="80" align="center">
        <template #default="{ row }">
          <el-switch
            :model-value="row.is_enabled"
            :disabled="row.is_default"
            @change="() => handleToggleEnabled(row as unknown as LLMConfig)"
          />
        </template>
      </el-table-column>

      <el-table-column label="默认" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="primary" size="small" effect="plain">默认</el-tag>
          <span v-else class="llm-settings__default-empty">—</span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button text :icon="ConnectionIcon" @click="handleTestConnection(row as unknown as LLMConfig)">
            测试
          </el-button>
          <el-button
            text
            :icon="StarIcon"
            :disabled="row.is_default"
            @click="handleSetDefault(row as unknown as LLMConfig)"
          >
            设为默认
          </el-button>
          <el-button text :icon="EditIcon" @click="openEdit(row as unknown as LLMConfig)">编辑</el-button>
          <el-button text :icon="DeleteIcon" class="llm-settings__action-danger" @click="handleDelete(row as unknown as LLMConfig)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <LLMConfigFormDialog
      v-model="dialogOpen"
      :initial="editing"
      :submitting="submitting"
      @submit="onDialogSubmit"
    />
  </div>
</template>

<style scoped>
.llm-settings {
  max-width: 1280px;
  margin: 0 auto;
  padding: 16px 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.llm-settings__header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.llm-settings__title-row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.llm-settings__title {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.llm-settings__sub {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

.llm-settings__table {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eef2f7;
  padding: 8px 4px;
}

.llm-settings__name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.llm-settings__star,
.llm-settings__star--off {
  font-size: 14px;
}
.llm-settings__star { color: #f59e0b; }
.llm-settings__star--off { color: #cbd5e1; }

.llm-settings__name-text {
  font-weight: 500;
  color: #0f172a;
}

.llm-settings__url {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #475569;
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.llm-settings__key {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #2563eb;
  background: #eef4ff;
  padding: 1px 8px;
  border-radius: 4px;
}

.llm-settings__default-empty {
  color: #cbd5e1;
}

.llm-settings__action-danger :deep(svg) {
  color: #ef4444;
}

@media (max-width: 768px) {
  .llm-settings__title-row {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>