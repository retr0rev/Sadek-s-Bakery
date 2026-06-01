import { ref } from 'vue'
import { defineStore } from 'pinia'
import { apiFetch, apiJson } from '@/utils/api'

export const useAuthStore = defineStore('auth', () => {
  const admin = ref<{ id: number; username: string } | null>(null)
  const isAuthenticated = ref(false)

  async function checkAuth() {
    try {
      const data = await apiJson<{ admin: { id: number; username: string } }>('/api/auth/me')
      admin.value = data.admin
      isAuthenticated.value = true
    } catch (e) {
      admin.value = null
      isAuthenticated.value = false
      throw e
    }
  }

  async function login(username: string, password: string) {
    const data = await apiJson<{ admin: { id: number; username: string } }>('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    admin.value = data.admin
    isAuthenticated.value = true
  }

  async function logout() {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
      })
    } catch {
      // ignore
    }
    admin.value = null
    isAuthenticated.value = false
  }

  return { admin, isAuthenticated, checkAuth, login, logout }
})
