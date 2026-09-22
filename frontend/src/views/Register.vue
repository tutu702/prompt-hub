<script setup lang="ts">
/**
 * 注册页 P02
 *
 * 字段：邮箱、密码（含强度提示）、确认密码、昵称（可选）、协议勾选。
 * 注册成功自动登录并跳转 /prompts。
 */
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  email: '',
  password: '',
  confirmPassword: '',
  nickname: '',
  agreement: false,
})

// 实时密码强度提示文案
const strengthHint = ref('')
const passwordStrong = ref(false)

function checkStrength(pwd: string) {
  passwordStrong.value = false
  if (!pwd) {
    strengthHint.value = ''
    return
  }
  const tips: string[] = []
  if (pwd.length < 8) tips.push('至少 8 位')
  if (pwd.length > 64) tips.push('不超过 64 位')
  if (!/[A-Za-z]/.test(pwd)) tips.push('需包含字母')
  if (!/[0-9]/.test(pwd)) tips.push('需包含数字')
  strengthHint.value = tips.length ? tips.join('，') : '密码强度良好'
  passwordStrong.value = tips.length === 0
}

watch(
  () => form.password,
  (v) => checkStrength(v),
)

const validatePassword = (_: unknown, value: string, callback: (err?: Error) => void) => {
  if (!value) {
    callback(new Error('请输入密码'))
    return
  }
  if (value.length < 8 || value.length > 64) {
    callback(new Error('密码长度需为 8-64 位'))
    return
  }
  if (!/[A-Za-z]/.test(value) || !/[0-9]/.test(value)) {
    callback(new Error('密码必须同时包含字母与数字'))
    return
  }
  callback()
}

const validateConfirm = (_: unknown, value: string, callback: (err?: Error) => void) => {
  if (!value) {
    callback(new Error('请再次输入密码'))
    return
  }
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
    return
  }
  callback()
}

const validateAgreement = (_: unknown, value: boolean, callback: (err?: Error) => void) => {
  if (!value) {
    callback(new Error('请阅读并同意用户协议与隐私政策'))
    return
  }
  callback()
}

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [{ validator: validatePassword, trigger: 'blur' }],
  confirmPassword: [{ validator: validateConfirm, trigger: 'blur' }],
  nickname: [{ max: 32, message: '昵称不超过 32 个字符', trigger: 'blur' }],
  agreement: [{ validator: validateAgreement, trigger: 'change' }],
}

async function onSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.register({
      email: form.email.trim(),
      password: form.password,
      confirmPassword: form.confirmPassword,
      nickname: form.nickname.trim() || undefined,
      agreement: form.agreement,
    })
    ElMessage.success('注册成功，已自动登录')
    router.replace('/prompts')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '注册失败'
    ElMessage.error(msg || '注册失败')
  } finally {
    loading.value = false
  }
}

function goLogin() {
  router.push({ name: 'Login' })
}
</script>

<template>
  <div class="register">
    <h2 class="register__title">注册</h2>
    <p class="register__sub">创建 PromptHub 账号，开始管理你的提示词资产</p>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      size="large"
      @submit.prevent="onSubmit"
    >
      <el-form-item label="邮箱" prop="email">
        <el-input
          v-model="form.email"
          placeholder="请输入邮箱"
          autocomplete="username"
          clearable
        />
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          placeholder="8-64 位，包含字母与数字"
          autocomplete="new-password"
          show-password
          clearable
        />
        <div v-if="form.password" class="register__hint" :class="{ ok: passwordStrong }">
          {{ strengthHint }}
        </div>
      </el-form-item>

      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="再次输入密码"
          autocomplete="new-password"
          show-password
          clearable
        />
      </el-form-item>

      <el-form-item label="昵称（可选）" prop="nickname">
        <el-input
          v-model="form.nickname"
          placeholder="留空将使用邮箱前缀"
          maxlength="32"
          clearable
        />
      </el-form-item>

      <el-form-item prop="agreement">
        <el-checkbox v-model="form.agreement">
          我已阅读并同意
          <el-link type="primary" :underline="false" @click.stop>《用户协议》</el-link>
          与
          <el-link type="primary" :underline="false" @click.stop>《隐私政策》</el-link>
        </el-checkbox>
      </el-form-item>

      <el-button
        type="primary"
        :loading="loading"
        :disabled="!form.agreement"
        class="register__submit"
        @click="onSubmit"
      >
        注册并登录
      </el-button>

      <div class="register__footer">
        已有账号？
        <el-link type="primary" :underline="false" @click="goLogin">直接登录</el-link>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.register {
  width: 100%;
}

.register__title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 4px;
}

.register__sub {
  margin: 0 0 20px;
  font-size: 13px;
  color: #94a3b8;
}

.register__hint {
  margin-top: 6px;
  font-size: 12px;
  color: #f59e0b;
  line-height: 1.4;
}

.register__hint.ok {
  color: #10b981;
}

.register__submit {
  width: 100%;
  height: 40px;
  font-size: 15px;
  margin-top: 4px;
}

.register__footer {
  margin-top: 16px;
  text-align: center;
  font-size: 13px;
  color: #475569;
}
</style>