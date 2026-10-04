<script setup>
import { computed } from 'vue'
import { state } from '@/state/store.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'

const missingSubscriptions = computed(() =>
  (state.opmlCompareResult || []).filter((result) => result.tip === '不存在')
)

function exportMissingSubscriptions() {
  if (!missingSubscriptions.value.length) return

  const doc = document.implementation.createDocument(null, 'opml')
  doc.documentElement.setAttribute('version', '1.1')
  const head = doc.createElement('head')
  const title = doc.createElement('title')
  title.textContent = '不存在的订阅'
  head.appendChild(title)
  doc.documentElement.appendChild(head)

  const body = doc.createElement('body')
  for (const subscription of missingSubscriptions.value) {
    const outline = doc.createElement('outline')
    outline.setAttribute('type', 'rss')
    outline.setAttribute('text', subscription.title || '')
    outline.setAttribute('title', subscription.title || '')
    outline.setAttribute('xmlUrl', subscription.feedUrl)
    outline.setAttribute('htmlUrl', subscription.siteUrl || '')
    body.appendChild(outline)
  }
  doc.documentElement.appendChild(body)

  const xml = '<?xml version="1.0" encoding="UTF-8"?>\n' + new XMLSerializer().serializeToString(doc)
  const url = URL.createObjectURL(new Blob([xml], { type: 'application/xml;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = 'missing-subscriptions.opml'
  document.body.appendChild(link)
  link.click()
  link.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
</script>

<template>
  <Modal :open="state.settings == 'compare-opml'" @hide="state.settings = ''">
    <button
      class="btn btn-link outline-none float-right p-2 mr-n2 mt-n2"
      style="line-height: 1"
      @click="state.settings = ''"
    >
      <Icon name="x" />
    </button>
    <div>
      <p class="cursor-default"><b>Compare OPML</b></p>
      <button
        class="btn btn-primary"
        :disabled="!missingSubscriptions.length"
        @click="exportMissingSubscriptions"
      >
        导出不存在订阅
      </button>
      <div v-if="state.opmlCompareResult" class="mt-4">
        <!-- 使用表格展示对比结果信息 -->
        <table class="table table-bordered">
          <thead>
            <tr>
              <th>标题</th>
              <th>订阅地址</th>
              <th>网站地址</th>
              <th>是否存在</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="result in state.opmlCompareResult" :key="result.url">
              <td>{{ result.title }}</td>
              <td class="text-center"><a :href="result.feedUrl" target="_blank">打开</a></td>
              <td class="text-center"><a v-if="result.siteUrl" :href="result.siteUrl" target="_blank">打开</a></td>
              <td class="text-center">{{ result.tip }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </Modal>
</template>
