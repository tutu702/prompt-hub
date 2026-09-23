<script setup lang="ts">
/**
 * 提示词手动创建 / 编辑表单（PRD P04 模式 A）。
 *
 * 字段：title (1-100) / description (0-200) / category_id (选填) / content (1-10000) / tags[]
 * 暴露方法（父组件通过 ref 调用）：
 *   - submit()        触发 el-form 校验 → emit submit(payload)
 *   - saveAndNew()    同上 → emit save-and-new(payload)
 *   - reset()         清空表单 + emit dirty=false
 *   - applyAiResult() 把 AI 生成结果填入表单
 *   - submitting      当前是否在提交中
 *
 * 受父页面（P04 PromptCreate.vue 或未来的 P05 编辑页）调度，校验由前端 + 后端双保险。
 */
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { storeToRefs } from 'pinia'
import { useCategoryStore } from '@/stores/category'
import type { CreatePromptPayload } from '@/types'

const props = defineProps<{
  initial?: Partial<CreatePromptPayload>
}>()

const emit = defineEmits<{
  (e: 'dirty', value: boolean): void
  (e: 'submit', payload: CreatePromptPayload): void
  (e: 'save-and-new', payload: CreatePromptPayload): void
}>()

defineExpose({
  submit,
  saveAndNew,
  reset,
  applyAiResult,
  get submitting() {
    return submitting.value
  },
})

const categoryStore = useCategoryStore()
const { list: categoryList } = storeToRefs(categoryStore)

const formRef = ref<FormInstance>()
const submitting = ref(false)

function makeEmptyForm(): {
  title: string
  description: string
  category_id: number | null
  content: string
  tags: string[]
} {
  return {
    title: '',
    description: '',
    category_id: null,
    content: '',
    tags: [],
  }
}

const form = reactive(makeEmptyForm())

// 初始值：仅在 mount 时合并一次 props.initial
if (props.initial) {
  Object.assign(form, {
    title: props.initial.title || '',
    description: props.initial.description || '',
    category_id: props.initial.category_id ?? null,
    content: props.initial.content || '',
    tags: props.initial.tags || [],
  })
}

// category 下拉选项
const categoryOptions = computed(() =>
  categoryList.value.map((c) => ({ id: c.id, name: c.name })),
)

// 字符计数（>9000 变橙）
const contentLen = computed(() => form.content.length)
const contentOver = computed(() => contentLen.value > 9000)

// 校验规则
const rules: FormRules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度需为 1-100 字符', trigger: 'blur' },
  ],
  description: [
    { max: 200, message: '描述不超过 200 字符', trigger: 'blur' },
  ],
  content: [
    { required: true, message: '请输入内容', trigger: 'blur' },
    {
      validator(_, value: string, cb: (msg?: string) => void) {
        if (!value || value.length === 0) return cb('请输入内容')
        if (value.length > 10000) return cb('内容不超过 10000 字符')
        cb()
      },
      trigger: 'blur',
    },
  ],
}

// 标签输入时 sanitize（trim + 限长）
watch(
  () => form.tags,
  (next) => {
    const cleaned = (next || [])
      .filter((t) => typeof t === 'string')
      .map((t) => t.trim())
      .filter((t) => t.length > 0)
      .slice(0, 20)
    const overLong = cleaned.filter((t) => t.length > 32)
    const truncated = cleaned.map((t) => (t.length > 32 ? t.slice(0, 32) : t))
    if (
      overLong.length > 0 ||
      cleaned.length !== (next.length || 0)
    ) {
      form.tags = truncated
    }
    markDirty()
  },
  { deep: true },
)

// 通用 dirty 触发（除 tags 外其它字段通过 input handler 触发）
function markDirty() {
  emit('dirty', true)
}

function onInput() {
  markDirty()
}

function buildPayload(): CreatePromptPayload | null {
  const cleanTags = form.tags
    .map((t) => t.trim())
    .filter((t) => t.length > 0)
    .filter((t) => t.length <= 32)

  const description = form.description.trim()
  return {
    title: form.title.trim(),
    description: description || undefined,
    content: form.content,
    category_id: form.category_id,
    tags: cleanTags.length > 0 ? cleanTags : undefined,
  }
}

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    ElMessage.error('表单填写有误，请检查后重试')
    return
  }
  const payload = buildPayload()
  if (!payload) return
  submitting.value = true
  try {
    emit('submit', payload)
  } finally {
    submitting.value = false
  }
}

async function saveAndNew() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    ElMessage.error('表单填写有误，请检查后重试')
    return
  }
  const payload = buildPayload()
  if (!payload) return
  submitting.value = true
  try {
    emit('save-and-new', payload)
  } finally {
    submitting.value = false
  }
}

function reset() {
  Object.assign(form, makeEmptyForm())
  formRef.value?.clearValidate()
  emit('dirty', false)
}

function applyAiResult(payload: {
  title?: string
  content?: string
  description?: string
  tags?: string[]
}) {
  if (payload.title !== undefined) form.title = payload.title
  if (payload.content !== undefined) form.content = payload.content
  if (payload.description !== undefined) form.description = payload.description
  if (payload.tags !== undefined) form.tags = [...payload.tags]
  formRef.value?.clearValidate()
  emit('dirty', true)
}
</script>

<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-position="top"
    size="default"
    class="prompt-form"
    @submit.prevent
  >
    <el-form-item label="标题" prop="title" required>
      <el-input
        v-model="form.title"
        placeholder="给这条提示词起一个名字"
        maxlength="100"
        show-word-limit
        clearable
        @input="onInput"
      />
    </el-form-item>

    <el-form-item label="描述" prop="description">
      <el-input
        v-model="form.description"
        type="textarea"
        :rows="2"
        placeholder="一句话说明这条提示词的用途（选填）"
        maxlength="200"
        show-word-limit
        @input="onInput"
      />
    </el-form-item>

    <el-form-item label="分类">
      <el-select
        v-model="form.category_id"
        placeholder="不分类"
        clearable
        style="width: 100%"
        @change="onInput"
      >
        <el-option
          v-for="c in categoryOptions"
          :key="c.id"
          :label="c.name"
          :value="c.id"
        />
      </el-select>
    </el-form-item>

    <el-form-item label="标签">
      <el-select
        v-model="form.tags"
        multiple
        filterable
        allow-create
        default-first-option
        placeholder="回车添加标签（每个 ≤32 字符）"
        style="width: 100%"
      />
    </el-form-item>

    <el-form-item label="内容" prop="content" required>
      <div class="prompt-form__content-wrap">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="12"
          placeholder="写下提示词正文（1-10000 字符）"
          resize="vertical"
          @input="onInput"
        />
        <div
          class="prompt-form__counter"
          :class="{ 'prompt-form__counter--warn': contentOver }"
        >
          {{ contentLen }} / 10000
        </div>
      </div>
    </el-form-item>
  </el-form>
</template>

<style scoped>
.prompt-form {
  background: #ffffff;
  padding: 20px 24px;
  border-radius: 12px;
  border: 1px solid #eef2f7;
}

.prompt-form__content-wrap {
  position: relative;
  width: 100%;
}

.prompt-form__content-wrap :deep(.el-textarea__inner) {
  width: 100%;
}

.prompt-form__counter {
  position: absolute;
  right: 8px;
  bottom: 6px;
  font-size: 12px;
  color: #94a3b8;
  pointer-events: none;
  background: rgba(255, 255, 255, 0.85);
  padding: 0 4px;
  border-radius: 4px;
}

.prompt-form__counter--warn {
  color: #f59e0b;
  font-weight: 600;
}
</style>