<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { apiJson } from '@/utils/api'

interface Product {
  id: number
  name: string
  description: string
  price: number
  ingredients: string
  image: string | null
  category: string
}

const products = ref<Product[]>([])
const loading = ref(true)
const selectedCategory = ref('all')
const searchQuery = ref('')

const categories = computed(() => {
  const cats = new Set(products.value.map((p) => p.category))
  return ['all', ...Array.from(cats)]
})

const filteredProducts = computed(() => {
  let result = products.value
  if (selectedCategory.value !== 'all') {
    result = result.filter((p) => p.category === selectedCategory.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    result = result.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q) ||
        p.ingredients.toLowerCase().includes(q),
    )
  }
  return result
})

async function fetchProducts() {
  try {
    products.value = await apiJson<Product[]>('/api/products')
  } catch (e) {
    console.error('Failed to fetch products:', e)
  } finally {
    loading.value = false
  }
}

function getImageUrl(image: string | null): string {
  if (!image) return ''
  return image
}

onMounted(fetchProducts)
</script>

<template>
  <div class="app-container">
    <header class="header">
      <div class="header-top">
        <div class="container">
          <div class="contact-info">
            <a
              href="https://wa.me/963535747523?text=%D9%85%D8%B1%D8%AD%D8%A8%D8%A7%D8%8C%20%D8%A3%D9%86%D8%A7%20%D8%A3%D8%AA%D9%81%D9%82%D8%AF%20%D8%A7%D9%84%D9%85%D9%86%D8%AA%D8%AC%D8%A7%D8%AA"
              target="_blank"
              class="contact-link"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                <path
                  d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"
                />
              </svg>
              <span dir="ltr">+963 935 747 523</span>
            </a>
            <span class="contact-sep">|</span>
            <span class="contact-link">sadekbakery11@gmail.com</span>
          </div>
        </div>
      </div>
      <div class="header-main">
        <div class="container">
          <div class="brand">
            <h1 class="brand-name">حلويات ومعجنات الصادق</h1>
            <p class="brand-sub">Sadek's Sweets & Pastries</p>
          </div>
          <nav class="nav-links">
            <a href="#products" class="nav-link">المنتجات</a>
            <a href="#about" class="nav-link">عن الفرن</a>
            <a href="#contact" class="nav-link">اتصل بنا</a>
          </nav>
        </div>
      </div>
    </header>

    <section class="hero">
      <div class="container">
        <div class="hero-content">
          <h2 class="hero-title">أشهى الحلويات والمعجنات</h2>
          <p class="hero-desc">
            نقدم لكم أجود أنواع الكعك والحلويات الفرنسية والمعجنات الطازجة يومياً
          </p>
        </div>
      </div>
    </section>

    <section id="products" class="products-section">
      <div class="container">
        <div class="section-header">
          <h2 class="section-title">منتجاتنا</h2>
          <p class="section-desc">تصفح أشهى منتجاتنا الطازجة</p>
        </div>

        <div class="filters">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="ابحث عن منتج..."
            class="search-input"
          />
          <div class="category-filters">
            <button
              v-for="cat in categories"
              :key="cat"
              :class="['category-btn', { active: selectedCategory === cat }]"
              @click="selectedCategory = cat"
            >
              {{ cat === 'all' ? 'الكل' : cat }}
            </button>
          </div>
        </div>

        <div v-if="loading" class="loading">جاري تحميل المنتجات...</div>

        <div v-else-if="filteredProducts.length === 0" class="empty">
          لا توجد منتجات متاحة حالياً
        </div>

        <div v-else class="products-grid">
          <div v-for="product in filteredProducts" :key="product.id" class="product-card">
            <div class="product-image">
              <img v-if="product.image" :src="getImageUrl(product.image)" :alt="product.name" />
              <div v-else class="no-image">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="#ccc">
                  <path
                    d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"
                  />
                </svg>
              </div>
            </div>
            <div class="product-info">
              <h3 class="product-name">{{ product.name }}</h3>
              <p class="product-desc">{{ product.description }}</p>
              <div v-if="product.ingredients" class="product-ingredients">
                <span class="ingredients-label">المكونات:</span>
                {{ product.ingredients }}
              </div>
              <div class="product-price">{{ product.price.toLocaleString() }} ل.س</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section id="about" class="about-section">
      <div class="container">
        <div class="about-content">
          <h2 class="section-title">عن الفرن</h2>
          <p class="about-text">
            تأسس حلويات ومعجنات الصادق عام 1995 على يد الأخوين جميل وخالد صادق باسم حلويات ومعجنات
            لبنان، لينقلوا ثقافة الكعك والحلويات الفرنسية من لبنان إلى مدينة إدلب.
          </p>
          <p class="about-text">
            منذ ذلك الحين ونحن نلتزم بتقديم أجود أنواع المعجنات والكعك والحلويات باستخدام أفضل
            المكونات الطازجة، لرضا عملائنا الكرام.
          </p>
          <div class="about-meta">
            <div class="meta-item">
              <span class="meta-label">الإدارة</span>
              <span class="meta-value">جميل صادق</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">تأسس عام</span>
              <span class="meta-value">1995</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section id="contact" class="contact-section">
      <div class="container">
        <h2 class="section-title">اتصل بنا</h2>
        <div class="contact-cards">
          <a
            href="https://wa.me/963535747523?text=%D9%85%D8%B1%D8%AD%D8%A8%D8%A7%D8%8C%20%D8%A3%D9%86%D8%A7%20%D8%A3%D8%AA%D9%81%D9%82%D8%AF%20%D8%A7%D9%84%D9%85%D9%86%D8%AA%D8%AC%D8%A7%D8%AA"
            target="_blank"
            class="contact-card"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"
              />
            </svg>
            <span>واتساب</span>
            <span dir="ltr">+963 935 747 523</span>
          </a>
          <div class="contact-card">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"
              />
            </svg>
            <span>البريد الإلكتروني</span>
            <span>sadekbakery11@gmail.com</span>
          </div>
        </div>
      </div>
    </section>

    <footer class="footer">
      <div class="container">
        <p>جميع الحقوق محفوظة &copy; {{ new Date().getFullYear() }} حلويات ومعجنات الصادق</p>
        <p class="dev-credit">بتطوير retr0rev</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  background: #faf7f2;
  font-family:
    system-ui,
    -apple-system,
    sans-serif;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.header-top {
  background: #2c1810;
  color: #e8d5c4;
  padding: 8px 0;
  font-size: 14px;
}

.contact-info {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  direction: ltr;
}

.contact-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #e8d5c4;
  text-decoration: none;
}

.contact-link:hover {
  color: #d4a574;
}

.contact-sep {
  color: #5a3e2e;
}

.header-main {
  background: #3d2317;
  color: #fff;
  padding: 16px 0;
  border-bottom: 3px solid #d4a574;
}

.header-main .container {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand-name {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 1px;
}

.brand-sub {
  font-size: 13px;
  color: #d4a574;
  margin: 2px 0 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.nav-links {
  display: flex;
  gap: 32px;
}

.nav-link {
  color: #e8d5c4;
  text-decoration: none;
  font-size: 16px;
  transition: color 0.2s;
}

.nav-link:hover {
  color: #d4a574;
}

.hero {
  background: linear-gradient(135deg, #3d2317 0%, #5a3e2e 50%, #3d2317 100%);
  color: #fff;
  padding: 80px 0;
  text-align: center;
  position: relative;
  overflow: hidden;
}

.hero::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at 30% 50%, rgba(212, 165, 116, 0.1) 0%, transparent 50%);
}

.hero-content {
  position: relative;
  z-index: 1;
}

.hero-title {
  font-size: 42px;
  font-weight: 300;
  margin: 0 0 16px;
  letter-spacing: 2px;
}

.hero-desc {
  font-size: 18px;
  color: #d4a574;
  max-width: 600px;
  margin: 0 auto;
  line-height: 1.8;
}

.products-section,
.about-section,
.contact-section {
  padding: 80px 0;
}

.section-header {
  text-align: center;
  margin-bottom: 48px;
}

.section-title {
  font-size: 32px;
  font-weight: 300;
  color: #2c1810;
  margin: 0 0 8px;
  letter-spacing: 1px;
}

.section-desc {
  color: #8a7a6e;
  font-size: 16px;
}

.filters {
  margin-bottom: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.search-input {
  width: 100%;
  max-width: 400px;
  padding: 12px 20px;
  border: 1px solid #d4c5b8;
  border-radius: 8px;
  font-size: 15px;
  background: #fff;
  text-align: right;
  outline: none;
  transition: border-color 0.2s;
}

.search-input:focus {
  border-color: #d4a574;
}

.category-filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.category-btn {
  padding: 8px 20px;
  border: 1px solid #d4c5b8;
  border-radius: 20px;
  background: #fff;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.category-btn:hover {
  border-color: #d4a574;
  color: #d4a574;
}

.category-btn.active {
  background: #3d2317;
  color: #fff;
  border-color: #3d2317;
}

.loading,
.empty {
  text-align: center;
  padding: 60px;
  color: #8a7a6e;
  font-size: 18px;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
}

.product-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  transition:
    transform 0.2s,
    box-shadow 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.product-image {
  width: 100%;
  height: 220px;
  overflow: hidden;
  background: #f5f0eb;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-info {
  padding: 20px;
}

.product-name {
  font-size: 18px;
  font-weight: 600;
  color: #2c1810;
  margin: 0 0 8px;
}

.product-desc {
  font-size: 14px;
  color: #6b5a4e;
  margin: 0 0 12px;
  line-height: 1.6;
}

.product-ingredients {
  font-size: 13px;
  color: #8a7a6e;
  margin-bottom: 12px;
  line-height: 1.5;
}

.ingredients-label {
  color: #5a3e2e;
  font-weight: 500;
}

.product-price {
  font-size: 20px;
  font-weight: 600;
  color: #3d2317;
}

.about-section {
  background: #fff;
}

.about-content {
  max-width: 800px;
  margin: 0 auto;
  text-align: center;
}

.about-text {
  font-size: 17px;
  line-height: 2;
  color: #4a3a2e;
  margin: 0 0 16px;
}

.about-meta {
  display: flex;
  justify-content: center;
  gap: 48px;
  margin-top: 40px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.meta-label {
  font-size: 13px;
  color: #8a7a6e;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.meta-value {
  font-size: 18px;
  font-weight: 600;
  color: #3d2317;
}

.contact-section {
  background: #f5f0eb;
}

.contact-cards {
  display: flex;
  justify-content: center;
  gap: 24px;
  flex-wrap: wrap;
}

.contact-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 32px 48px;
  background: #fff;
  border-radius: 12px;
  text-decoration: none;
  color: #2c1810;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  transition: transform 0.2s;
}

.contact-card:hover {
  transform: translateY(-4px);
}

.contact-card svg {
  color: #d4a574;
}

.contact-card span:first-of-type {
  font-size: 14px;
  color: #8a7a6e;
}

.contact-card span:last-of-type {
  font-size: 16px;
  font-weight: 600;
}

.footer {
  background: #2c1810;
  color: #8a7a6e;
  text-align: center;
  padding: 24px 0;
  font-size: 14px;
}

.footer p {
  margin: 0;
}

.dev-credit {
  margin-top: 4px !important;
  font-size: 12px;
  color: #5a3e2e;
}

@media (max-width: 768px) {
  .header-main .container {
    flex-direction: column;
    gap: 16px;
  }

  .hero-title {
    font-size: 28px;
  }

  .products-grid {
    grid-template-columns: 1fr;
  }

  .about-meta {
    flex-direction: column;
    gap: 24px;
  }

  .contact-cards {
    flex-direction: column;
    align-items: center;
  }
}
</style>
