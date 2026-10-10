<script setup>
import { computed, watch } from 'vue'
import { opml, compareOPML, addComparedFeed } from '@/state/opml.js'
import { state } from '@/state/store.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'

watch(() => state.settings, (value, previous) => {
  if (value === 'compare-opml' && previous === 'create') compareOPML()
})

async function selectFile(event) {
  const file = event.target.files[0]
  event.target.value = ''
  if (file) await compareOPML(file)
}

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
  head.appendChild(doc.createTextNode('\n    '))
  head.appendChild(title)
  head.appendChild(doc.createTextNode('\n  '))
  doc.documentElement.appendChild(doc.createTextNode('\n  '))
  doc.documentElement.appendChild(head)

  const body = doc.createElement('body')
  for (const subscription of missingSubscriptions.value) {
    const outline = doc.createElement('outline')
    outline.setAttribute('type', 'rss')
    outline.setAttribute('text', subscription.title || '')
    outline.setAttribute('title', subscription.title || '')
    outline.setAttribute('xmlUrl', subscription.feedUrl)
    outline.setAttribute('htmlUrl', subscription.siteUrl || '')
    body.appendChild(doc.createTextNode('\n    '))
    body.appendChild(outline)
  }
  body.appendChild(doc.createTextNode('\n  '))
  doc.documentElement.appendChild(doc.createTextNode('\n  '))
  doc.documentElement.appendChild(body)
  doc.documentElement.appendChild(doc.createTextNode('\n'))

  const xml = '<?xml version="1.0" encoding="UTF-8"?>\n' + new XMLSerializer().serializeToString(doc) + '\n'
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
  <Modal class="compare-opml-modal" :open="state.settings == 'compare-opml'" @hide="state.settings = ''">
    <button
      class="btn btn-link outline-none float-right p-2 mr-n2 mt-n2"
      style="line-height: 1"
      @click="state.settings = ''"
    >
      <Icon name="x" />
    </button>
    <div>
      <p class="cursor-default"><b>Compare OPML</b></p>
      <div class="mb-3">
        <label for="compare-opml-file">选择 OPML 文件（最大 5 MiB）</label>
        <input id="compare-opml-file" type="file" accept=".opml,.xml" class="d-block" :disabled="opml.loading" @change="selectFile" />
        <p v-if="opml.filename" class="mt-2 mb-2">当前文件：{{ opml.filename }}</p>
        <p v-if="opml.error" class="text-danger mt-2" role="alert">{{ opml.error }}</p>
      </div>
      <div class="d-flex align-items-center">
        <button class="btn btn-default mr-2" :disabled="opml.loading" @click="compareOPML()">
          {{ opml.loading ? '对比中…' : '刷新对比结果' }}
        </button>
      <button
        class="btn btn-primary"
        :disabled="!missingSubscriptions.length"
        @click="exportMissingSubscriptions"
      >
        导出不存在订阅
      </button>
      </div>
      <div v-if="state.opmlCompareResult" class="mt-4">
        <!-- 使用表格展示对比结果信息 -->
        <table class="table table-bordered">
          <thead>
            <tr>
              <th>标题</th>
              <th>订阅地址</th>
              <th>网站地址</th>
              <th>是否存在</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(result, index) in state.opmlCompareResult" :key="index">
              <td>{{ result.title }}</td>
              <td class="text-center"><a :href="result.feedUrl" target="_blank">打开</a></td>
              <td class="text-center"><a v-if="result.siteUrl" :href="result.siteUrl" target="_blank">打开</a></td>
              <td class="text-center">{{ result.tip }}</td>
              <td class="text-center">
                <button v-if="result.tip === '不存在'" class="btn btn-sm btn-primary" :disabled="opml.loading" @click="addComparedFeed(result)">添加</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.compare-opml-modal :deep(.modal-dialog) {
  max-width: 1100px;
  width: calc(100% - 2rem);
  margin-right: auto;
  margin-left: auto;
}
</style>
