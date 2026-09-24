<script setup lang="ts">
/**
 * 模型配置新建/编辑弹窗（P08）。
 *
 * 字段：name / provider / model / base_url / api_key / is_enabled / is_default
 * 校验：name 1-20、provider 枚举、model 1-100、base_url URL、api_key 编辑可留空
 * 体验：选择 provider 时自动填 BaseURL 默认值与模型名 placeholder
 */
import { computed, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import type {
  CreateLLMConfigPayload,
  LLMConfig,
  LLMProvider,
  UpdateLLMConfigPayload,
} from '@/types'

const props = defineProps<{
  modelValue: boolean
  initial?: LLMConfig | null
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', payload: {
    id: number | null
    data: CreateLLMConfigPayload | UpdateLLMConfigPayload
  }): void
}>()

interface FormState {
  name: string
  provider: LLMProvider
  model: string
  base_url: string
  api_key: string
  is_enabled: boolean
  is_default: boolean
}

const providerOptions: { value: LLMProvider; label: string }[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'qwen', label: '通义千问' },
  { value: 'zhipu', label: '智谱' },
  { value: 'moonshot', label: 'Moonshot' },
  { value: 'custom', label: '自定义' },
]

/** 供应商 → 默认 BaseURL + 模型名 placeholder（OpenAI 兼容协议） */
const providerPresets: Record<LLMProvider, { baseURL: string; modelExample: string }> = {
  openai:    { baseURL: 'https://api.openai.com/v1',                  modelExample: 'gpt-4o-mini' },
  anthropic: { baseURL: 'https://api.anthropic.com/v1',              modelExample: 'claude-3-5-sonnet-latest' },
  deepseek:  { baseURL: 'https://api.deepseek.com/v1',                modelExample: 'deepseek-chat' },
  qwen:      { baseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1', modelExample: 'qwen-plus' },
  zhipu:     { baseURL: 'https://open.bigmodel.cn/api/paas/v4',       modelExample: 'glm-4-plus' },
  moonshot:  { baseURL: 'https://api.moonshot.cn/v1',                 modelExample: 'moonshot-v1-8k' },
  custom:    { baseURL: '',                                            modelExample: '' },
}

/** 当前选择供应商的 BaseURL 示例（用于 placeholder） */
const baseUrlPlaceholder = computed(() => providerPresets[form.value.provider]?.baseURL || 'https://...')
/** 当前选择供应商的模型名示例（用于 placeholder） */
const modelPlaceholder = computed(() => providerPresets[form.value.provider]?.modelExample || '模型名称')

const formRef = ref<FormInstance>()
const form = ref<FormState>({
  name: '',
  provider: 'openai',
  model: '',
  base_url: 'https://api.openai.com/v1',
  api_key: '',
  is_enabled: true,
  is_default: false,
})

const isEdit = computed(() => !!props.initial)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      if (props.initial) {
        form.value = {
          name: props.initial.name,
          provider: props.initial.provider,
          model: props.initial.model,
          base_url: props.initial.base_url,
          api_key: '',
          is_enabled: props.initial.is_enabled,
          is_default: props.initial.is_default,
        }
      } else {
        form.value = {
          name: '',
          provider: 'openai',
          model: '',
          base_url: providerPresets.openai.baseURL,
          api_key: '',
          is_enabled: true,
          is_default: false,
        }
      }
      formRef.value?.clearValidate()
    }
  },
  { immediate: true },
)

/** 切换供应商时自动填 BaseURL 默认值（model 仅 placeholder，不自动写入）
 *  规则：仅当当前 base_url 为空、或与上一个 provider 的默认值相同时才覆盖
 *  → 避免覆盖用户已手填的 BaseURL
 */
watch(
  () => form.value.provider,
  (newProv, oldProv) => {
    if (isEdit.value) return
    if (!newProv) return
    const preset = providerPresets[newProv]
    const oldPreset = oldProv ? providerPresets[oldProv] : undefined
    const isDefaultOrEmpty =
      !form.value.base_url ||
      form.value.base_url === '' ||
      (oldPreset && form.value.base_url === oldPreset.baseURL)
    if (isDefaultOrEmpty) {
      form.value.base_url = preset.baseURL
    }
  },
)

const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { min: 1, max: 20, message: '名称长度需为 1-20 字符', trigger: 'blur' },
  ],
  provider: [
    { required: true, message: '请选择供应商', trigger: 'change' },
  ],
  model: [
    { required: true, message: '请输入模型名称', trigger: 'blur' },
    { min: 1, max: 100, message: '模型名称长度需为 1-100 字符', trigger: 'blur' },
  ],
  base_url: [
    { required: true, message: '请输入 BaseURL', trigger: 'blur' },
    {
      validator(_, value: string, cb: (msg?: string) => void) {
        if (!value) return cb()
        if (!/^https?:\/\/.+/.test(value)) {
          return cb('BaseURL 必须以 http:// 或 https:// 开头')
        }
        cb()
      },
      trigger: 'blur',
    },
  ],
  api_key: [
    {
      validator(_: unknown, value: string, cb: (msg?: string) => void) {
        if (isEdit.value) {
          // 编辑时留空表示不改
          cb()
          return
        }
        if (!value || value.length < 1) {
          cb('API Key 必填')
          return
        }
        cb()
      },
      trigger: 'blur',
    },
  ],
}

const canSubmit = computed(() => {
  if (!form.value.name.trim() || !form.value.model.trim() || !form.value.base_url.trim()) {
    return false
  }
  if (!isEdit.value && !form.value.api_key) return false
  return true
})

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
  if (form.value.is_default && !form.value.is_enabled) {
    ElMessage.error('默认模型必须同时启用')
    return
  }
  const id = props.initial?.id ?? null
  if (id) {
    const initial = props.initial!
    const payload: UpdateLLMConfigPayload = {}
    if (form.value.name !== initial.name) payload.name = form.value.name.trim()
    if (form.value.provider !== initial.provider) payload.provider = form.value.provider
    if (form.value.model !== initial.model) payload.model = form.value.model.trim()
    if (form.value.base_url !== initial.base_url) payload.base_url = form.value.base_url.trim()
    if (form.value.api_key) payload.api_key = form.value.api_key
    if (form.value.is_enabled !== initial.is_enabled) payload.is_enabled = form.value.is_enabled
    if (form.value.is_default !== initial.is_default) payload.is_default = form.value.is_default
    emit('submit', { id, data: payload })
  } else {
    const payload: CreateLLMConfigPayload = {
      name: form.value.name.trim(),
      provider: form.value.provider,
      model: form.value.model.trim(),
      base_url: form.value.base_url.trim(),
      api_key: form.value.api_key,
      is_enabled: form.value.is_enabled,
      is_default: form.value.is_default,
    }
    emit('submit', { id: null, data: payload })
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="isEdit ? '编辑模型配置' : '新建模型配置'"
    width="560px"
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
      <el-form-item label="名称（别名）" prop="name" required>
        <el-input
          v-model="form.name"
          placeholder="例如：主力 GPT-4o"
          maxlength="20"
          show-word-limit
          clearable
        />
      </el-form-item>

      <el-form-item label="供应商" prop="provider" required>
        <el-select v-model="form.provider" placeholder="选择供应商" style="width: 100%">
          <el-option
            v-for="o in providerOptions"
            :key="o.value"
            :label="o.label"
            :value="o.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="模型名称" prop="model" required>
        <el-input
          v-model="form.model"
          placeholder="例如：gpt-4o-mini"
          maxlength="100"
          clearable
        />
      </el-form-item>

      <el-form-item label="BaseURL" prop="base_url" required>
        <el-input
          v-model="form.base_url"
          placeholder="https://api.openai.com/v1"
          clearable
        />
      </el-form-item>

      <el-form-item label="API Key" prop="api_key" :required="!isEdit">
        <el-input
          v-model="form.api_key"
          type="password"
          show-password
          :placeholder="isEdit ? '留空表示不修改' : '请输入 API Key（前端永不保存明文到日志）'"
          clearable
        />
      </el-form-item>

      <el-form-item>
        <div class="llm-dialog__switches">
          <el-switch v-model="form.is_enabled" />
          <span class="llm-dialog__switch-label">启用</span>

          <el-switch v-model="form.is_default" style="margin-left: 24px" />
          <span class="llm-dialog__switch-label">设为默认模型</span>
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
.llm-dialog__switches {
  display: flex;
  align-items: center;
  gap: 6px;
}

.llm-dialog__switch-label {
  font-size: 13px;
  color: #475569;
}
</style>