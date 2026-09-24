<script setup lang="ts">
/**
 * 个人设置页（PRD P09 /settings/profile）。
 *
 * 顶部：资料卡（头像占位 + 昵称 + 邮箱 + 注册时间）
 * 中部：昵称编辑（邮箱 / 注册时间只读）
 * 下方：跳转入口（修改密码 V1 可选 / 模型配置 / 退出登录）
 */
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  ArrowLeft as ArrowLeftIcon,
  Edit as EditIcon,
  Lock as LockIcon,
  Connection as ConnectionIcon,
  SwitchButton as LogoutIcon,
} from '@element-plus/icons-vue'

import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

/** 资料加载态（首次进入时显示 skeleton） */
const loadingProfile = ref(false)

/** 昵称编辑表单 */
const formRef = ref<FormInstance>()
const nicknameDraft = ref('')
const submitting = ref(false)
/** 标记表单是否有改动；用于控制"保存"按钮可用性 */
const dirty = ref(false)

/** 头像首字母（与 DefaultLayout 保持一致） */
const avatarInitial = computed(() => {
  const name = userStore.userInfo?.nickname || userStore.userInfo?.email || '?'
  return name.trim().charAt(0).toUpperCase()
})

/** 当前昵称 */
const currentNickname = computed(() => userStore.userInfo?.nickname ?? '')
/** 当前邮箱 */
const currentEmail = computed(() => userStore.userInfo?.email ?? '—')
/** 注册时间（已格式化为中文） */
const createdAtText = computed(() => {
  const raw = userStore.userInfo?.created_at
  if (!raw) return '—'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
})

/** 昵称校验：1-20 字符（与后端 binding 保持一致） */
const rules: FormRules = {
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 1, max: 20, message: '昵称长度需为 1-20 字符', trigger: 'blur' },
  ],
}

/** 进入页面时拉取最新资料；本地已有则跳过网络请求以加快首屏 */
async function loadProfile(force = false) {
  if (!force && userStore.userInfo?.created_at) {
    nicknameDraft.value = currentNickname.value
    dirty.value = false
    return
  }
  loadingProfile.value = true
  try {
    await userStore.fetchProfile()
    nicknameDraft.value = currentNickname.value
    dirty.value = false
  } catch (err) {
    ElMessage.error((err as Error).message || '资料加载失败')
  } finally {
    loadingProfile.value = false
  }
}

onMounted(() => {
  loadProfile()
})

/** 草稿改动时打标记；store 外部更新昵称时也同步进表单 */
watch(currentNickname, (v) => {
  // 仅当用户没有在编辑中（避免编辑过程中被覆盖）
  if (!dirty.value) nicknameDraft.value = v
})
watch(nicknameDraft, () => {
  dirty.value = nicknameDraft.value.trim() !== currentNickname.value
})

/** 保存昵称 */
async function onSave() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  const nickname = nicknameDraft.value.trim()
  if (nickname === currentNickname.value) {
    ElMessage.info('昵称未修改')
    return
  }
  submitting.value = true
  try {
    await userStore.updateProfile({ nickname })
    ElMessage.success('昵称已更新')
    dirty.value = false
  } catch (err) {
    ElMessage.error((err as Error).message || '保存失败')
  } finally {
    submitting.value = false
  }
}

/** 取消编辑（恢复原值） */
function onCancel() {
  nicknameDraft.value = currentNickname.value
  dirty.value = false
  formRef.value?.clearValidate()
}

function onBack() {
  router.push('/prompts')
}

/** 跳到模型配置 */
function goLLM() {
  router.push('/settings/llm')
}

/** 跳到修改密码（统一设置页中的 password 子模块） */
function goPassword() {
  router.push({ path: '/settings', query: { section: 'password' } })
}

/**
 * 退出登录：先二次确认，再清 token，最后跳登录页。
 * 复用 DefaultLayout 中已有的退出逻辑体验，保持一致。
 */
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
    ElMessage.success('已退出登录')
    router.push('/login')
  } catch (err) {
    ElMessage.error((err as Error).message || '退出失败')
  }
}
</script>

<template>
  <div class="profile-settings">
    <header class="profile-settings__header">
      <el-button link @click="onBack">
        <el-icon><ArrowLeftIcon /></el-icon>
        <span style="margin-left: 4px">返回提示词库</span>
      </el-button>
      <h2 class="profile-settings__title">个人设置</h2>
      <p class="profile-settings__sub">查看账户信息、修改昵称，或跳转到其他设置</p>
    </header>

    <!-- 资料卡 -->
    <section class="profile-card" v-loading="loadingProfile">
      <div class="profile-card__avatar">{{ avatarInitial }}</div>
      <div class="profile-card__info">
        <div class="profile-card__name">{{ currentNickname || '未设置昵称' }}</div>
        <div class="profile-card__email">{{ currentEmail }}</div>
        <div class="profile-card__meta">注册时间：{{ createdAtText }}</div>
      </div>
    </section>

    <!-- 昵称编辑 -->
    <section class="profile-block">
      <h3 class="profile-block__title">基本资料</h3>
      <el-form
        ref="formRef"
        :model="{ nickname: nicknameDraft }"
        :rules="rules"
        label-position="top"
        class="profile-form"
      >
        <el-form-item label="昵称" prop="nickname">
          <el-input
            v-model="nicknameDraft"
            placeholder="请输入昵称"
            maxlength="20"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input :model-value="currentEmail" disabled />
        </el-form-item>
        <el-form-item label="注册时间">
          <el-input :model-value="createdAtText" disabled />
        </el-form-item>
        <div class="profile-form__actions">
          <el-button @click="onCancel" :disabled="!dirty || submitting">取消</el-button>
          <el-button
            type="primary"
            :icon="EditIcon"
            :loading="submitting"
            :disabled="!dirty"
            @click="onSave"
          >
            保存
          </el-button>
        </div>
      </el-form>
    </section>

    <!-- 快捷入口 -->
    <section class="profile-block">
      <h3 class="profile-block__title">快捷入口</h3>
      <ul class="profile-links">
        <li class="profile-link" @click="goLLM">
          <el-icon class="profile-link__icon"><ConnectionIcon /></el-icon>
          <div class="profile-link__body">
            <div class="profile-link__title">模型配置</div>
            <div class="profile-link__desc">管理供应商、API Key 与默认模型</div>
          </div>
          <span class="profile-link__chevron">›</span>
        </li>
        <li class="profile-link" @click="goPassword">
          <el-icon class="profile-link__icon"><LockIcon /></el-icon>
          <div class="profile-link__body">
            <div class="profile-link__title">修改密码</div>
            <div class="profile-link__desc">定期更换密码可以提升账号安全</div>
          </div>
          <span class="profile-link__chevron">›</span>
        </li>
        <li class="profile-link profile-link--danger" @click="handleLogout">
          <el-icon class="profile-link__icon"><LogoutIcon /></el-icon>
          <div class="profile-link__body">
            <div class="profile-link__title">退出登录</div>
            <div class="profile-link__desc">清除本地凭证并跳转登录页</div>
          </div>
          <span class="profile-link__chevron">›</span>
        </li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.profile-settings {
  max-width: 720px;
  margin: 0 auto;
  padding: 16px 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.profile-settings__header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.profile-settings__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.profile-settings__sub {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

/* ---- 资料卡 ---- */
.profile-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04);
}

.profile-card__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #ffffff;
  font-size: 26px;
  font-weight: 700;
  flex-shrink: 0;
}

.profile-card__info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.profile-card__name {
  font-size: 18px;
  font-weight: 600;
  color: #0f172a;
}

.profile-card__email {
  font-size: 13px;
  color: #475569;
}

.profile-card__meta {
  font-size: 12px;
  color: #94a3b8;
}

/* ---- 通用区块 ---- */
.profile-block {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04);
}

.profile-block__title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

/* ---- 表单 ---- */
.profile-form {
  max-width: 480px;
}

.profile-form__actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 4px;
}

/* ---- 快捷入口 ---- */
.profile-links {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.profile-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.profile-link:hover {
  background: #eef4ff;
  border-color: #c7d7fe;
}

.profile-link--disabled {
  cursor: not-allowed;
  opacity: 0.7;
}
.profile-link--disabled:hover {
  background: #f8fafc;
  border-color: transparent;
}

.profile-link--danger:hover {
  background: #fef2f2;
  border-color: #fecaca;
}
.profile-link--danger:hover .profile-link__icon {
  color: #ef4444;
}

.profile-link__icon {
  font-size: 20px;
  color: #2563eb;
  flex-shrink: 0;
}

.profile-link__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.profile-link__title {
  font-size: 14px;
  font-weight: 500;
  color: #1f2937;
}

.profile-link__desc {
  font-size: 12px;
  color: #64748b;
}

.profile-link__chevron {
  font-size: 22px;
  color: #cbd5e1;
  line-height: 1;
}
</style>