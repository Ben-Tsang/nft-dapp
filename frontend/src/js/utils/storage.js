// 封装 localStorage，处理 JSON 序列化
export const Storage = {
  set(key, value) {
    localStorage.setItem(key, JSON.stringify(value))
  },

  get(key, defaultValue = null) {
    const data = localStorage.getItem(key)
    return data ? JSON.parse(data) : defaultValue
  },

  remove(key) {
    localStorage.removeItem(key)
  },

  clear() {
    localStorage.clear()
  }
}