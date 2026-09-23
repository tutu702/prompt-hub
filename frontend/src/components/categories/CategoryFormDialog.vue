<script setup lang="ts">
/**
 * 分类新建/编辑弹窗（PRD P06）。
 *
 * 字段：name (1-20 必填) / color (HEX #RRGGBB，可空)。
 * 颜色：el-color-picker + 旁挂 el-input（用户可直接粘贴 HEX），本地正则双保险。
 * 提交按钮 disabled：name 空 / color 非法 / submitting。
 */
import { computed, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import type { Category, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'

const props = defineProps<{
  modelValue: boolean
  initial?: Category | null
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', payload: { id: number | null; data: CreateCategoryPayload | UpdateCategoryPayload }): void
}>()

const formRef = ref<FormInstance>()

const HEX = /^#[0-9A-Fa-f]{6}$/

function makeEmpty() {
  return { name: '', color: '' }
}
const form = ref(makeEmpty())
const editingId = ref<number | null>(null)
const isEdit = computed(() => editingId.value !== null)

// 同步 modelValue：true 时填充，false 时重置
watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      if (props.initial) {
        editingId.value = props.initial.id
        form.value = {
          name: props.initial.name,
          color: props.initial.color || '',
        }
      } else {
        editingId.value = null
        form.value = makeEmpty()
      }
      // 清除校验状态
      formRef.value?.clearValidate()
    }
  },
  { immediate: true },
)

const colorValid = computed(() => {
  if (!form.value.color) return true
  return HEX.test(form.value.color)
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 1, max: 20, message: '名称长度需为 1-20 字符', trigger: 'blur' },
  ],
  color: [
    {
      validator(_, value: string, cb: (msg?: string) => void) {
        if (!value) return cb()
        if (!HEX.test(value)) return cb('颜色格式不正确，应为 #RRGGBB')
        cb()
      },
      trigger: 'blur',
    },
  ],
}

const canSubmit = computed(
  () => !!form.value.name.trim() && colorValid.value && !props.submitting,
)

function close() {
  emit('update:modelValue', false)
}

async function onSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    ElMessage.error('表单填写有误，请检查后重试')
    return
  }
  if (!colorValid.value) {
    ElMessage.error('颜色格式不正确')
    return
  }
  const data: CreateCategoryPayload = {
    name: form.value.name.trim(),
    color: form.value.color || undefined,
  }
  emit('submit', { id: editingId.value, data })
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="isEdit ? '编辑分类' : '新建分类'"
    width="480px"
    :close-on-click-modal="false"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @close="close"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      size="default"
      @submit.prevent="onSubmit"
    >
      <el-form-item label="分类名称" prop="name" required>
        <el-input
          v-model="form.name"
          placeholder="例如：编程 / 写作 / 研究"
          maxlength="20"
          show-word-limit
          clearable
        />
      </el-form-item>

      <el-form-item label="标签颜色" prop="color">
        <div class="category-dialog__color-row">
          <el-color-picker
            v-model="form.color"
            :show-alpha="false"
            color-format="hex"
            size="default"
          />
          <el-input
            v-model="form.color"
            placeholder="可空，例：#1677FF"
            class="category-dialog__color-input"
            clearable
          />
        </div>
        <div
          v-if="form.color && !colorValid"
          class="category-dialog__color-warn"
        >
          颜色格式不正确，应为 #RRGGBB（6 位十六进制）
        </div>
        <div v-else class="category-dialog__color-hint">
          不填则使用默认灰色
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="close">取消</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="!canSubmit"
        @click="onSubmit"
      >
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.category-dialog__color-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.category-dialog__color-input {
  flex: 1;
}

.category-dialog__color-hint {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
}

.category-dialog__color-warn {
  font-size: 12px;
  color: #ef4444;
  margin-top: 4px;
}
</style>