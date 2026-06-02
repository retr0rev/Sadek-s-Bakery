import { ref } from 'vue'
import { defineStore } from 'pinia'
import { apiJson } from '@/utils/api'

export const useAuthStore = defineStore('auth', () => {
  const admin = ref<{ id: number; username: string } | null>(null)
  const isAuthenticated = ref(false)
  const token = ref<string | null>(null)

  function loadToken() {
    const stored = localStorage.getItem('auth_token')
    if (stored) {
      token.value = stored
    }
  }

  function saveToken(t: string) {
    token.value = t
    localStorage.setItem('auth_token', t)
  }

  function clearToken() {
    token.value = null
    localStorage.removeItem('auth_token')
  }

  async function checkAuth() {
    try {
      const data = await apiJson<{ admin: { id: number; username: string } }>('/api/auth/me')
      admin.value = data.admin
      isAuthenticated.value = true
    } catch (e) {
      admin.value = null
      isAuthenticated.value = false
      clearToken()
      throw e
    }
  }

  async function login(username: string, password: string) {
    const data = await apiJson<{ admin: { id: number; username: string }; token: string }>('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    admin.value = data.admin
    token.value = data.token
    isAuthenticated.value = true
    saveToken(data.token)
  }

  async function logout() {
    try {
      const headers: Record<string, string> = {}
      if (token.value) {
        headers['Authorization'] = `Bearer ${token.value}`
      }
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers,
      })
    } catch {
      // ignore
    }
    admin.value = null
    isAuthenticated.value = false
    clearToken()
  }

  loadToken()

  return { admin, isAuthenticated, token, checkAuth, login, logout }
})
