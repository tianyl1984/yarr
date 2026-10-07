import { reactive } from 'vue'
import { api } from '@/api/api.js'
import { state, showSettings } from './store.js'

const storageKey = 'yarr.opml.last-file'
export const opml = reactive({ file: null, loading: false, error: '', storageError: '' })

try {
  const saved = JSON.parse(localStorage.getItem(storageKey) || 'null')
  if (saved) {
    const bytes = Uint8Array.from(atob(saved.data), (char) => char.charCodeAt(0))
    opml.file = new File([bytes], saved.name, { type: 'application/xml' })
  }
} catch {
  opml.storageError = '无法读取上次文件，请重新选择。'
}

export async function compareOPML(file = opml.file) {
  if (!file || opml.loading) return
  opml.loading = true
  opml.error = ''
  try {
    const result = await api.compare_opml(file)
    opml.file = file
    state.opmlCompareResult = result || []
    try {
      const bytes = new Uint8Array(await file.arrayBuffer())
      let binary = ''
      for (const byte of bytes) binary += String.fromCharCode(byte)
      localStorage.setItem(storageKey, JSON.stringify({ name: file.name, data: btoa(binary) }))
      opml.storageError = ''
    } catch {
      opml.storageError = '文件已保留在当前页面，但无法保存到浏览器；重新加载页面后需重新选择。'
    }
  } catch {
    opml.error = '对比失败，请检查文件或网络后重试。'
  } finally {
    opml.loading = false
  }
}

export function addComparedFeed(result) {
  showSettings('create', result.feedUrl)
  state.feedCreateReturn = 'compare-opml'
}
