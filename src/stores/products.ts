import { ref } from 'vue'
import { defineStore } from 'pinia'

const API_BASE = ''

export interface Product {
  id: number
  name: string
  description: string
  price: number
  ingredients: string
  image: string | null
  category: string
  created_at: string
}

export const useProductStore = defineStore('products', () => {
  const products = ref<Product[]>([])
  const loading = ref(false)

  async function fetchProducts() {
    loading.value = true
    try {
      const res = await fetch(`${API_BASE}/api/products`)
      products.value = await res.json()
    } catch (e) {
      console.error('Failed to fetch products:', e)
    } finally {
      loading.value = false
    }
  }

  return { products, loading, fetchProducts }
})
