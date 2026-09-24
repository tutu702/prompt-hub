<script setup lang="ts">
/**
 * 统一设置页（PRD P09 重构版）。
 *
 * 布局（与设计图保持一致）：
 *   - 左侧侧栏：导航（个人资料 / 修改密码 / 大模型配置 / 数据导出）+ 账号安全建议
 *   - 右侧内容：根据 activeSection 切换四个区块
 *
 * 区块：
 *   - profile：个人资料卡 + 基本资料表单（昵称 / 邮箱 / 注册时间 / 最近登录）
 *   - password：修改密码表单（UI 占位）
 *   - export：数据导出占位
 *
 * 大模型配置已移出此侧栏，统一通过顶部「模型配置」入口进入 /settings/llm。
 */
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from 'element-plus'
import {
  User as UserNavIcon,
  Lock as LockNavIcon,
  Download as ExportNavIcon,
  Edit as EditIcon,
  CircleCheck as ShieldIcon,
  Promotion as RightArrowIcon,
} from '@element-plus/icons-vue'

import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

/* ---------- 侧栏导航 ---------- */
type Section = 'profile' | 'password' | 'export'
const validSections: Section[] = ['profile', 'password', 'export']

function parseSection(q: unknown): Section {
  const s = String(q || 'profile')
  return (validSections as string[]).includes(s) ? (s as Section) : 'profile'
}
const activeSection = ref<Section>(parseSection(route.query.section))

watch(
  () => route.query.section,
  (v) => {
    activeSection.value = parseSection(v)
  },
)

function switchSection(s: Section) {
  activeSection.value = s
  router.replace({ path: '/settings', query: { section: s } })
}

const navItems = computed(() => [
  { key: 'profile' as const, label: '个人资料', icon: UserNavIcon, soon: false },
  { key: 'password' as const, label: '修改密码', icon: LockNavIcon, soon: false },
  { key: 'export' as const, label: '数据导出', icon: ExportNavIcon, soon: true },
])

/* ---------- 数据加载 ---------- */
const loadingProfile = ref(false)

async function loadProfile(force = false) {
  if (!force && userStore.userInfo?.created_at) return
  loadingProfile.value = true
  try {
    await userStore.fetchProfile()
  } catch (err) {
    ElMessage.error((err as Error).message || '资料加载失败')
  } finally {
    loadingProfile.value = false
  }
}

onMounted(() => {
  loadProfile()
})

/* ---------- 头像 / 文本辅助 ---------- */
const avatarInitial = computed(() => {
  const name = userStore.userInfo?.nickname || userStore.userInfo?.email || '?'
  return name.trim().charAt(0).toUpperCase()
})
const currentNickname = computed(() => userStore.userInfo?.nickname ?? '')
const currentEmail = computed(() => userStore.userInfo?.email ?? '—')

const createdAtText = computed(() => {
  const raw = userStore.userInfo?.created_at
  if (!raw) return '—'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
})

/* ---------- profile：昵称内联编辑 ---------- */
const nicknameEditing = ref(false)
const nicknameDraft = ref('')
const formRef = ref<FormInstance>()
const submitting = ref(false)
const rules: FormRules = {
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 1, max: 20, message: '昵称长度需为 1-20 字符', trigger: 'blur' },
  ],
}

watch(currentNickname, (v) => {
  if (!nicknameEditing.value) nicknameDraft.value = v
})

function openNicknameEdit() {
  nicknameDraft.value = currentNickname.value
  nicknameEditing.value = true
  formRef.value?.clearValidate()
}

function cancelNicknameEdit() {
  nicknameEditing.value = false
  nicknameDraft.value = currentNickname.value
  formRef.value?.clearValidate()
}

async function saveNickname() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  const nickname = nicknameDraft.value.trim()
  if (nickname === currentNickname.value) {
    nicknameEditing.value = false
    ElMessage.info('昵称未修改')
    return
  }
  submitting.value = true
  try {
    await userStore.updateProfile({ nickname })
    ElMessage.success('昵称已更新')
    nicknameEditing.value = false
  } catch (err) {
    ElMessage.error((err as Error).message || '保存失败')
  } finally {
    submitting.value = false
  }
}

/* ---------- password：修改密码表单（占位） ---------- */
const pwdFormRef = ref<FormInstance>()
const pwdForm = ref({ old: '', next: '', confirm: '' })
const pwdSubmitting = ref(false)

const pwdRules: FormRules = {
  old: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  next: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, max: 64, message: '密码长度需为 8-64 字符', trigger: 'blur' },
  ],
  confirm: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator(_, value: string, cb: (msg?: string) => void) {
        if (value !== pwdForm.value.next) return cb('两次输入的密码不一致')
        cb()
      },
      trigger: 'blur',
    },
  ],
}

async function onChangePassword() {
  if (!pwdFormRef.value) return
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return
  pwdSubmitting.value = true
  try {
    // V1 后端未提供改密接口；占位提示
    await new Promise((r) => setTimeout(r, 400))
    ElMessage.warning('改密功能即将开放')
    pwdForm.value = { old: '', next: '', confirm: '' }
  } finally {
    pwdSubmitting.value = false
  }
}

/* ---------- 退出登录 ---------- */
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
  <div class="settings">
    <!-- 左侧导航 -->
    <aside class="settings__sidebar">
      <nav class="settings__nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="settings__nav-item"
          :class="{ 'settings__nav-item--active': activeSection === item.key }"
          :disabled="item.soon"
          @click="switchSection(item.key)"
        >
          <el-icon class="settings__nav-icon">
            <component :is="item.icon" />
          </el-icon>
          <span class="settings__nav-label">{{ item.label }}</span>
          <el-tag v-if="item.soon" size="small" effect="plain" type="info" class="settings__nav-tag">
            敬请期待
          </el-tag>
        </button>
      </nav>

      <!-- 安全建议 -->
      <div class="settings__safety">
        <el-icon class="settings__safety-icon"><ShieldIcon /></el-icon>
        <span class="settings__safety-text">
          账号安全建议：每 90 天更新一次密码。
        </span>
      </div>
    </aside>

    <!-- 右侧内容 -->
    <main class="settings__main">
      <!-- 个人资料 -->
      <section v-if="activeSection === 'profile'" class="settings__pane">
        <header class="settings__pane-header">
          <h2 class="settings__pane-title">个人设置</h2>
          <p class="settings__pane-sub">管理你的账号资料与安全设置</p>
        </header>

        <!-- 资料卡 -->
        <article class="settings__profile-card" v-loading="loadingProfile">
          <div class="settings__avatar">{{ avatarInitial }}</div>
          <div class="settings__profile-info">
            <div class="settings__profile-name">{{ currentNickname || '未设置昵称' }}</div>
            <div class="settings__profile-email">{{ currentEmail }}</div>
            <div class="settings__profile-meta">
              <span class="settings__meta-pill">加入于 {{ createdAtText.split(' ')[0] }}</span>
              <span class="settings__meta-pill settings__meta-pill--pro">Pro 试用中</span>
            </div>
          </div>
          <el-button :icon="EditIcon" plain class="settings__profile-edit" @click="openNicknameEdit">
            编辑资料
          </el-button>
        </article>

        <!-- 基本资料表单 -->
        <article class="settings__block">
          <el-form
            ref="formRef"
            :model="{ nickname: nicknameDraft }"
            :rules="rules"
            label-position="left"
            label-width="84px"
            class="settings__form"
          >
            <el-form-item label="昵称">
              <div v-if="!nicknameEditing" class="settings__form-row">
                <span class="settings__form-value">{{ currentNickname || '—' }}</span>
                <el-button text :icon="EditIcon" @click="openNicknameEdit">编辑</el-button>
              </div>
              <div v-else class="settings__form-row settings__form-row--editing">
                <el-input
                  v-model="nicknameDraft"
                  placeholder="请输入昵称"
                  maxlength="20"
                  show-word-limit
                  class="settings__form-input"
                />
                <el-button @click="cancelNicknameEdit" :disabled="submitting">取消</el-button>
                <el-button
                  type="primary"
                  :icon="EditIcon"
                  :loading="submitting"
                  @click="saveNickname"
                >
                  保存
                </el-button>
              </div>
            </el-form-item>

            <el-form-item label="邮箱">
              <span class="settings__form-value">{{ currentEmail }}</span>
              <span class="settings__form-tag">只读</span>
            </el-form-item>

            <el-form-item label="注册时间">
              <span class="settings__form-value">{{ createdAtText }}</span>
              <span class="settings__form-tag">只读</span>
            </el-form-item>

            <el-form-item label="最近登录">
              <span class="settings__form-value">2026-09-21 09:12 · 上海</span>
              <span class="settings__form-tag">只读</span>
            </el-form-item>
          </el-form>
        </article>

        <!-- 登录密码 -->
        <article class="settings__block settings__block--password">
          <div class="settings__block-info">
            <h3 class="settings__block-title">登录密码</h3>
            <p class="settings__block-sub">上次更新于 2026-08-02，建议定期更换以保障账号安全。</p>
          </div>
          <el-button :icon="LockNavIcon" plain @click="switchSection('password')">
            修改密码
          </el-button>
        </article>

        <!-- 退出登录 -->
        <footer class="settings__logout-row">
          <span class="settings__logout-hint">退出当前账号后需重新登录。</span>
          <el-button type="danger" :icon="RightArrowIcon" @click="handleLogout">
            退出登录
          </el-button>
        </footer>
      </section>

      <!-- 修改密码 -->
      <section v-else-if="activeSection === 'password'" class="settings__pane">
        <header class="settings__pane-header">
          <h2 class="settings__pane-title">修改密码</h2>
          <p class="settings__pane-sub">定期更换密码可以提升账号安全性</p>
        </header>

        <article class="settings__block">
          <el-form
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-position="top"
            class="settings__form settings__form--vertical"
          >
            <el-form-item label="当前密码" prop="old">
              <el-input
                v-model="pwdForm.old"
                type="password"
                show-password
                placeholder="请输入当前密码"
                autocomplete="current-password"
              />
            </el-form-item>
            <el-form-item label="新密码" prop="next">
              <el-input
                v-model="pwdForm.next"
                type="password"
                show-password
                placeholder="8-64 位，建议字母 + 数字 + 符号"
                autocomplete="new-password"
              />
            </el-form-item>
            <el-form-item label="确认新密码" prop="confirm">
              <el-input
                v-model="pwdForm.confirm"
                type="password"
                show-password
                placeholder="请再次输入新密码"
                autocomplete="new-password"
              />
            </el-form-item>
            <div class="settings__form-actions">
              <el-button type="primary" :loading="pwdSubmitting" @click="onChangePassword">
                保存新密码
              </el-button>
            </div>
          </el-form>
        </article>
      </section>

      <!-- 数据导出 -->
      <section v-else-if="activeSection === 'export'" class="settings__pane">
        <header class="settings__pane-header">
          <h2 class="settings__pane-title">数据导出</h2>
          <p class="settings__pane-sub">导出提示词、分类、调优历史等数据</p>
        </header>

        <article class="settings__block settings__block--center">
          <el-icon class="settings__block-icon"><ExportNavIcon /></el-icon>
          <h3 class="settings__block-title">敬请期待</h3>
          <p class="settings__block-sub">导出功能将在后续版本开放。</p>
        </article>
      </section>
    </main>
  </div>
</template>

<style scoped>
.settings {
  max-width: 1280px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 24px;
  align-items: flex-start;
}

/* ===== 左侧导航 ===== */
.settings__sidebar {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 12px;
  position: sticky;
  top: 76px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.settings__nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings__nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 0;
  background: transparent;
  color: #475569;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  text-align: left;
  transition: background 0.15s, color 0.15s;
}

.settings__nav-item:hover:not(:disabled) {
  background: #f1f5f9;
  color: #0f172a;
}

.settings__nav-item--active {
  background: #eef4ff;
  color: #2563eb;
  font-weight: 500;
}

.settings__nav-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.settings__nav-label {
  flex: 1;
  min-width: 0;
}

.settings__nav-tag {
  margin-left: auto;
  flex-shrink: 0;
}

.settings__nav-item:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* 安全建议 */
.settings__safety {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 8px;
  font-size: 12px;
  color: #64748b;
  line-height: 1.6;
}

.settings__safety-icon {
  color: #10b981;
  font-size: 14px;
  margin-top: 2px;
  flex-shrink: 0;
}

.settings__safety-text {
  color: #64748b;
}

/* ===== 右侧内容 ===== */
.settings__main {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.settings__pane {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings__pane-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings__pane-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.2px;
}

.settings__pane-sub {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}

/* 资料卡 */
.settings__profile-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
}

.settings__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: 16px;
  background: linear-gradient(135deg, #dbeafe, #c7d2fe);
  color: #2563eb;
  font-size: 32px;
  font-weight: 700;
  flex-shrink: 0;
}

.settings__profile-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings__profile-name {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

.settings__profile-email {
  font-size: 13px;
  color: #475569;
}

.settings__profile-meta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 4px;
}

.settings__meta-pill {
  font-size: 11px;
  color: #475569;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  padding: 2px 8px;
  border-radius: 999px;
}

.settings__meta-pill--pro {
  color: #047857;
  background: #ecfdf5;
  border-color: #a7f3d0;
}

.settings__profile-edit {
  flex-shrink: 0;
  border-radius: 8px;
}

/* 通用 block */
.settings__block {
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 18px 24px;
}

.settings__block-title {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.settings__block-sub {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
}

.settings__block-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings__block--password {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.settings__block--center {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 36px 24px;
  gap: 8px;
}

.settings__block-icon {
  font-size: 32px;
  color: #6366f1;
  margin-bottom: 4px;
}

/* 表单 */
.settings__form {
  width: 100%;
}

.settings__form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.settings__form-row--editing {
  align-items: center;
}

.settings__form-value {
  color: #0f172a;
  font-size: 14px;
  flex: 1;
  min-width: 0;
}

.settings__form-input {
  flex: 1;
  max-width: 320px;
}

.settings__form-tag {
  font-size: 11px;
  color: #94a3b8;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  padding: 1px 8px;
  border-radius: 4px;
  margin-left: auto;
}

.settings__form :deep(.el-form-item) {
  border-bottom: 1px solid #f1f5f9;
  padding: 12px 0;
  margin: 0;
}

.settings__form :deep(.el-form-item:last-child) {
  border-bottom: none;
}

.settings__form :deep(.el-form-item__label) {
  color: #475569;
  font-weight: 500;
  font-size: 13px;
}

.settings__form--vertical {
  max-width: 480px;
}

.settings__form-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}

/* 退出登录 */
.settings__logout-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 4px 8px;
  margin-top: 8px;
}

.settings__logout-hint {
  font-size: 13px;
  color: #94a3b8;
}

@media (max-width: 960px) {
  .settings {
    grid-template-columns: 1fr;
  }
  .settings__sidebar {
    position: static;
  }
  .settings__nav {
    flex-direction: row;
    flex-wrap: wrap;
  }
  .settings__nav-item {
    flex: 1 1 calc(50% - 2px);
  }
}
</style>