<script setup>
import { state } from '@/state/store.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'
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
