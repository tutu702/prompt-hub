<script setup lang="ts">
/**
 * 登录页 P01
 *
 * 字段：邮箱、密码、记住我；登录成功跳 /prompts 或 redirect 指定页面。
 */
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  email: '',
  password: '',
  remember: false,
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, max: 64, message: '密码长度需为 8-64 位', trigger: 'blur' },
  ],
}

function getRedirect(): string {
  const raw = route.query.redirect
  if (typeof raw === 'string' && /^\/(?!\/)/.test(raw)) {
    return raw
  }
  return '/prompts'
}

async function onSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.login({
      email: form.email.trim(),
      password: form.password,
      remember: form.remember,
    })
    ElMessage.success('登录成功')
    router.replace(getRedirect())
  } catch (err) {
    const msg = err instanceof Error ? err.message : '邮箱或密码错误'
    ElMessage.error(msg || '邮箱或密码错误')
  } finally {
    loading.value = false
  }
}

function goRegister() {
  router.push({ name: 'Register' })
}
</script>

<template>
  <div class="login">
    <h2 class="login__title">登录</h2>
    <p class="login__sub">使用注册时的邮箱与密码登录</p>

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
          placeholder="请输入密码"
          autocomplete="current-password"
          show-password
          clearable
        />
      </el-form-item>

      <div class="login__row">
        <el-checkbox v-model="form.remember">记住我</el-checkbox>
        <el-link type="primary" :underline="false" @click="ElMessage.info('请联系管理员重置密码')">
          忘记密码？
        </el-link>
      </div>

      <el-button
        type="primary"
        :loading="loading"
        class="login__submit"
        @click="onSubmit"
      >
        登录
      </el-button>

      <div class="login__footer">
        还没有账号？
        <el-link type="primary" :underline="false" @click="goRegister">立即注册</el-link>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.login {
  width: 100%;
}

.login__title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 4px;
}

.login__sub {
  margin: 0 0 20px;
  font-size: 13px;
  color: #94a3b8;
}

.login__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.login__submit {
  width: 100%;
  height: 40px;
  font-size: 15px;
}

.login__footer {
  margin-top: 16px;
  text-align: center;
  font-size: 13px;
  color: #475569;
}
</style>