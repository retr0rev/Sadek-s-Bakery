<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { apiFetch, apiJson } from '@/utils/api'

interface Product {
  id: number
  name: string
  description: string
  price: number
  ingredients: string
  image: string | null
  category: string
  created_at: string
}

const router = useRouter()
const authStore = useAuthStore()
const products = ref<Product[]>([])
const loading = ref(true)

const showForm = ref(false)
const editingId = ref<number | null>(null)
const formData = ref({
  name: '',
  description: '',
  price: '',
  ingredients: '',
  category: 'general',
})
const selectedFile = ref<File | null>(null)
const saving = ref(false)

async function fetchProducts() {
  try {
    products.value = await apiJson<Product[]>('/api/products')
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  formData.value = { name: '', description: '', price: '', ingredients: '', category: 'general' }
  selectedFile.value = null
  showForm.value = true
}

function openEdit(product: Product) {
  editingId.value = product.id
  formData.value = {
    name: product.name,
    description: product.description,
    price: String(product.price),
    ingredients: product.ingredients,
    category: product.category,
  }
  selectedFile.value = null
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
  editingId.value = null
}

function onFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0]
  }
}

async function saveProduct() {
  saving.value = true
  try {
    const form = new FormData()
    form.append('name', formData.value.name)
    form.append('description', formData.value.description)
    form.append('price', formData.value.price)
    form.append('ingredients', formData.value.ingredients)
    form.append('category', formData.value.category)
    if (selectedFile.value) {
      form.append('image', selectedFile.value)
    }

    const url = editingId.value
      ? `/api/products/${editingId.value}`
      : `/api/products`

    const res = await apiFetch(url, {
      method: editingId.value ? 'PUT' : 'POST',
      body: form,
    })

    if (!res.ok) {
      const text = await res.text()
      let message = 'فشل الحفظ'
      try {
        const data = JSON.parse(text)
        message = data.message || message
      } catch {}
      throw new Error(message)
    }

    showForm.value = false
    await fetchProducts()
  } catch (e: any) {
    alert(e.message)
  } finally {
    saving.value = false
  }
}

async function deleteProduct(id: number) {
  console.log('deleteProduct called with id:', id)
  try {
    if (typeof confirm !== 'function' || !confirm('هل أنت متأكد من حذف هذا المنتج؟')) {
      console.log('deleteProduct: cancelled or confirm not available')
      return
    }
  } catch (e) {
    console.error('deleteProduct: confirm error:', e)
    return
  }

  try {
    console.log('deleteProduct: sending DELETE request for id:', id)
    const res = await apiFetch(`/api/products/${id}`, { method: 'DELETE' })
    console.log('deleteProduct: response status:', res.status)
    if (!res.ok) {
      let msg = 'فشل الحذف'
      try { const d = await res.json(); msg = d.message || msg } catch {}
      throw new Error(msg)
    }
    console.log('deleteProduct: success, refetching products')
    await fetchProducts()
  } catch (e: any) {
    console.error('deleteProduct: error:', e)
    alert(e.message)
  }
}

function getImageUrl(image: string | null): string {
  if (!image) return ''
  return image
}

async function handleLogout() {
  await authStore.logout()
  router.push('/admin')
}

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    try {
      await authStore.checkAuth()
    } catch {
      router.push('/admin')
      return
    }
  }
  await fetchProducts()
})
</script>

<template>
  <div class="dashboard">
    <header class="dash-header">
      <div class="dash-header-content">
        <h1>لوحة التحكم</h1>
        <div class="dash-actions">
          <button class="btn btn-primary" @click="openCreate">إضافة منتج جديد</button>
          <button class="btn btn-outline" @click="handleLogout">تسجيل خروج</button>
        </div>
      </div>
    </header>

    <main class="dash-main">
      <div v-if="showForm" class="form-overlay">
        <div class="product-form">
          <h2>{{ editingId ? 'تعديل منتج' : 'إضافة منتج جديد' }}</h2>
          <form @submit.prevent="saveProduct">
            <div class="form-grid">
              <div class="form-group">
                <label>اسم المنتج *</label>
                <input v-model="formData.name" type="text" required placeholder="اسم المنتج" />
              </div>
              <div class="form-group">
                <label>السعر (ل.س) *</label>
                <input v-model="formData.price" type="number" required placeholder="0" />
              </div>
              <div class="form-group">
                <label>التصنيف</label>
                <input v-model="formData.category" type="text" placeholder="حلويات, معجنات, كعك..." />
              </div>
              <div class="form-group">
                <label>صورة المنتج</label>
                <input type="file" accept="image/*" @change="onFileSelect" />
              </div>
              <div class="form-group full-width">
                <label>الوصف</label>
                <textarea v-model="formData.description" rows="3" placeholder="وصف المنتج"></textarea>
              </div>
              <div class="form-group full-width">
                <label>المكونات</label>
                <textarea v-model="formData.ingredients" rows="3" placeholder="المكونات"></textarea>
              </div>
            </div>
            <div class="form-buttons">
              <button type="submit" class="btn btn-primary" :disabled="saving">
                {{ saving ? 'جاري الحفظ...' : 'حفظ' }}
              </button>
              <button type="button" class="btn btn-outline" @click="cancelForm">إلغاء</button>
            </div>
          </form>
        </div>
      </div>

      <div v-if="loading" class="loading-state">جاري تحميل المنتجات...</div>

      <div v-else class="products-table">
        <div class="table-header">
          <span class="col-image">الصورة</span>
          <span class="col-name">الاسم</span>
          <span class="col-price">السعر</span>
          <span class="col-category">التصنيف</span>
          <span class="col-actions">إجراءات</span>
        </div>

        <div v-if="products.length === 0" class="empty-state">
          لا توجد منتجات. قم بإضافة أول منتج الآن.
        </div>

        <div
          v-for="product in products"
          :key="product.id"
          class="table-row"
        >
          <div class="col-image">
            <img
              v-if="product.image"
              :src="getImageUrl(product.image)"
              :alt="product.name"
              class="thumb"
            />
            <div v-else class="thumb-placeholder"></div>
          </div>
          <div class="col-name">
            <span class="product-name">{{ product.name }}</span>
          </div>
          <div class="col-price">{{ product.price.toLocaleString() }} ل.س</div>
          <div class="col-category">{{ product.category }}</div>
          <div class="col-actions">
            <button class="btn btn-sm btn-edit" @click="openEdit(product)">تعديل</button>
            <button class="btn btn-sm btn-delete" @click="deleteProduct(product.id)">حذف</button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: #f5f0eb;
  font-family: system-ui, -apple-system, sans-serif;
  direction: rtl;
}

.dash-header {
  background: #2c1810;
  color: #fff;
  padding: 16px 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.dash-header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dash-header h1 {
  font-size: 20px;
  font-weight: 300;
  margin: 0;
  letter-spacing: 1px;
}

.dash-actions {
  display: flex;
  gap: 12px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #d4a574;
  color: #2c1810;
}

.btn-primary:hover:not(:disabled) {
  background: #c49464;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-outline {
  background: transparent;
  color: #e8d5c4;
  border: 1px solid #5a3e2e;
}

.btn-outline:hover {
  border-color: #d4a574;
  color: #d4a574;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.btn-edit {
  background: #d4a574;
  color: #2c1810;
}

.btn-edit:hover {
  background: #c49464;
}

.btn-delete {
  background: #e74c3c;
  color: #fff;
}

.btn-delete:hover {
  background: #c0392b;
}

.dash-main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.form-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 20px;
}

.product-form {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.product-form h2 {
  font-size: 22px;
  font-weight: 300;
  color: #2c1810;
  margin: 0 0 24px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  font-size: 13px;
  color: #5a3e2e;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  padding: 10px 14px;
  border: 1px solid #d4c5b8;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  text-align: right;
}

.form-group input:focus,
.form-group textarea:focus {
  border-color: #d4a574;
}

.form-buttons {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  justify-content: flex-start;
}

.table-header {
  display: grid;
  grid-template-columns: 60px 2fr 120px 120px 140px;
  gap: 12px;
  padding: 12px 16px;
  background: #fff;
  border-radius: 8px 8px 0 0;
  font-size: 13px;
  font-weight: 600;
  color: #8a7a6e;
  text-transform: uppercase;
  letter-spacing: 1px;
  border-bottom: 2px solid #f5f0eb;
}

.table-row {
  display: grid;
  grid-template-columns: 60px 2fr 120px 120px 140px;
  gap: 12px;
  padding: 12px 16px;
  background: #fff;
  align-items: center;
  border-bottom: 1px solid #f5f0eb;
  transition: background 0.2s;
}

.table-row:hover {
  background: #faf7f2;
}

.table-row:last-child {
  border-radius: 0 0 8px 8px;
}

.thumb {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
}

.thumb-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: #f5f0eb;
}

.product-name {
  font-weight: 500;
  color: #2c1810;
}

.col-actions {
  display: flex;
  gap: 8px;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 60px 24px;
  background: #fff;
  border-radius: 8px;
  color: #8a7a6e;
  font-size: 16px;
}
</style>
