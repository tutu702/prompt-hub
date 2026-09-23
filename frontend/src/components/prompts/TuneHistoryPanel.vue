<script setup lang="ts">
/**
 * 调优历史面板（PRD P05 占位）。
 *
 * 本期不接入后端接口（P07 范围），渲染 el-empty 给用户明确引导。
 */
import { Clock as ClockIcon } from '@element-plus/icons-vue'

defineProps<{
  promptId: number
}>()
</script>

<template>
  <div class="tune-history-empty">
    <el-empty :image-size="80" description="调优历史即将开放">
      <template #image>
        <el-icon :size="56" class="tune-history-empty__icon">
          <ClockIcon />
        </el-icon>
      </template>
      <template #default>
        <p class="tune-history-empty__hint">
          配置 AI 模型后可查看这条提示词的调优时间线与 diff 对比。
        </p>
        <p class="tune-history-empty__id">提示词 ID: {{ promptId }}</p>
      </template>
    </el-empty>
  </div>
</template>

<style scoped>
.tune-history-empty {
  padding: 12px 0;
}

.tune-history-empty__icon {
  color: #cbd5e1;
}

.tune-history-empty__hint {
  margin: 4px 0 8px;
  color: #64748b;
  font-size: 13px;
}

.tune-history-empty__id {
  margin: 0;
  color: #94a3b8;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>