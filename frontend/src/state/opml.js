import { reactive } from 'vue'
import { api } from '@/api/api.js'
import { state, showSettings } from './store.js'

// Remove the file stored by older versions; file contents now live on the server.
try { localStorage.removeItem('yarr.opml.last-file') } catch {}
export const opml = reactive({ filename: '', loading: false, error: '' })

export async function compareOPML(file) {
  if (opml.loading) return
  opml.loading = true
  opml.error = ''
  try {
    const result = await api.compare_opml(file)
    opml.filename = result.filename
    state.opmlCompareResult = result.results || []
  } catch (error) {
    if (error.status === 404) {
      opml.filename = ''
      state.opmlCompareResult = []
    }
    opml.error = error.message || '对比失败，请检查文件或网络后重试。'
  } finally {
    opml.loading = false
  }
}

export function addComparedFeed(result) {
  showSettings('create', result.feedUrl)
  state.feedCreateReturn = 'compare-opml'
}
