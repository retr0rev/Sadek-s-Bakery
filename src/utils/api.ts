export async function apiFetch(url: string, options?: RequestInit): Promise<Response> {
  const res = await fetch(url, {
    ...options,
    headers: {
      ...(options?.headers || {}),
      Accept: 'application/json',
    },
    credentials: 'include',
  })
  return res
}

export async function apiJson<T = any>(url: string, options?: RequestInit): Promise<T> {
  const res = await apiFetch(url, options)

  if (!res.ok) {
    let message = 'Request failed'
    try {
      const body = await res.json()
      message = body.message || message
    } catch {
      const statusMessages: Record<number, string> = {
        502: 'الخادم الخلفي غير متصل. شغّل الأمر: npm run server',
        401: 'غير مخول. يرجى تسجيل الدخول أولاً',
        404: 'الرابط غير موجود',
        500: 'خطأ داخلي في الخادم',
      }
      message = statusMessages[res.status] || `خطأ في الخادم (${res.status})`
    }
    throw new Error(message)
  }

  const text = await res.text()
  if (!text) return {} as T
  try {
    return JSON.parse(text) as T
  } catch {
    throw new Error('Invalid JSON response from server')
  }
}
