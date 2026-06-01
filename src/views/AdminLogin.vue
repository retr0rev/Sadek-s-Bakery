<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/admin/dashboard')
  } catch (e: any) {
    error.value = e.message || 'فشل تسجيل الدخول'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>لوحة التحكم</h1>
        <p>حلويات ومعجنات الصادق</p>
      </div>
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label for="username">اسم المستخدم</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="أدخل اسم المستخدم"
            required
            dir="auto"
          />
        </div>
        <div class="form-group">
          <label for="password">كلمة المرور</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="أدخل كلمة المرور"
            required
          />
        </div>
        <p v-if="error" class="error-message">{{ error }}</p>
        <button type="submit" class="login-btn" :disabled="loading">
          {{ loading ? 'جاري تسجيل الدخول...' : 'تسجيل الدخول' }}
        </button>
      </form>
      <router-link to="/" class="back-link">العودة للصفحة الرئيسية</router-link>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #3d2317 0%, #5a3e2e 100%);
  font-family: system-ui, -apple-system, sans-serif;
}

.login-card {
  background: #fff;
  border-radius: 16px;
  padding: 48px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.2);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 24px;
  font-weight: 300;
  color: #2c1810;
  margin: 0 0 4px;
  letter-spacing: 1px;
}

.login-header p {
  font-size: 14px;
  color: #8a7a6e;
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 14px;
  color: #5a3e2e;
  font-weight: 500;
}

.form-group input {
  padding: 12px 16px;
  border: 1px solid #d4c5b8;
  border-radius: 8px;
  font-size: 15px;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus {
  border-color: #d4a574;
}

.error-message {
  color: #c0392b;
  font-size: 14px;
  text-align: center;
  margin: 0;
}

.login-btn {
  padding: 14px;
  background: #3d2317;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.login-btn:hover:not(:disabled) {
  background: #2c1810;
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.back-link {
  display: block;
  text-align: center;
  margin-top: 20px;
  color: #8a7a6e;
  text-decoration: none;
  font-size: 14px;
}

.back-link:hover {
  color: #3d2317;
}
</style>
