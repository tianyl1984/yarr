<script setup>
import { ref, computed } from 'vue'
import { state, refreshFeedStates, formatDate } from '@/state/store.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'

const onlyErrors = ref(false)
const reloading = ref(false)

// One row per feed, whether or not it has ever been refreshed. Failures first,
// then never-refreshed feeds, then the rest alphabetically — the whole list is
// rendered at once (no paging), the table body just scrolls.
const rows = computed(function () {
  const rows = state.feeds.map(function (feed) {
    const st = state.feed_states[feed.id] || {}
    return {
      id: feed.id,
      title: feed.title,
      feed_link: feed.feed_link,
      last_refreshed: st.last_refreshed || null,
      last_success: st.last_success || null,
      item_count: st.item_count,
      error: st.error || null,
    }
  })
  const rank = function (row) {
    if (row.error) return 0
    if (!row.last_refreshed) return 1
    return 2
  }
  return rows.sort(function (a, b) {
    return rank(a) - rank(b) || (a.title || '').localeCompare(b.title || '')
  })
})

const visibleRows = computed(function () {
  return onlyErrors.value ? rows.value.filter((r) => r.error) : rows.value
})

const errorCount = computed(function () {
  return rows.value.filter((r) => r.error).length
})

function reload() {
  reloading.value = true
  refreshFeedStates().finally(function () {
    reloading.value = false
  })
}
</script>

<template>
  <Modal
    class="feed-states-modal"
    :open="state.settings == 'feed-states'"
    @hide="state.settings = ''"
  >
    <button
      class="btn btn-link outline-none float-right p-2 mr-n2 mt-n2"
      style="line-height: 1"
      @click="state.settings = ''"
    >
      <Icon name="x" />
    </button>
    <div>
      <p class="cursor-default"><b>抓取结果</b></p>

      <div class="d-flex align-items-center mb-2">
        <span class="text-muted mr-3">
          共 {{ rows.length }} 个订阅，<span :class="{ 'text-danger': errorCount }"
            >{{ errorCount }} 个失败</span
          >
        </span>
        <div class="form-check mb-0 mr-3">
          <input
            class="form-check-input"
            type="checkbox"
            id="feed-states-only-errors"
            v-model="onlyErrors"
          />
          <label class="form-check-label cursor-pointer" for="feed-states-only-errors">
            只看失败
          </label>
        </div>
        <button class="btn btn-sm btn-outline-secondary" :disabled="reloading" @click="reload">
          <Icon name="rotate-cw" class="mr-1" />
          刷新状态
        </button>
      </div>

      <div class="feed-states-scroll scroll-touch">
        <table class="table table-sm table-bordered mb-0">
          <thead>
            <tr>
              <th>订阅</th>
              <th class="text-nowrap">状态</th>
              <th class="text-nowrap">最近抓取</th>
              <th class="text-nowrap">最近成功</th>
              <th class="text-nowrap" title="最近一次成功抓取解析到的条目数（含已存在的条目）">
                条目数
              </th>
              <th>错误信息</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in visibleRows" :key="row.id">
              <td>
                <a :href="row.feed_link" target="_blank" :title="row.feed_link">
                  {{ row.title }}
                </a>
              </td>
              <td class="text-center text-nowrap">
                <span v-if="row.error" class="badge badge-danger">失败</span>
                <span v-else-if="!row.last_refreshed" class="badge badge-secondary">未抓取</span>
                <span v-else class="badge badge-success">成功</span>
              </td>
              <td class="text-nowrap">
                {{ row.last_refreshed ? formatDate(row.last_refreshed) : '-' }}
              </td>
              <td class="text-nowrap">
                {{ row.last_success ? formatDate(row.last_success) : '-' }}
              </td>
              <td class="text-center">
                {{ row.item_count === null || row.item_count === undefined ? '-' : row.item_count }}
              </td>
              <td class="text-danger text-break">{{ row.error || '' }}</td>
            </tr>
            <tr v-if="!visibleRows.length">
              <td colspan="6" class="text-center text-muted">没有符合条件的订阅</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </Modal>
</template>

<style>
.feed-states-modal .modal-dialog {
  max-width: 64rem;
}

.feed-states-modal .feed-states-scroll {
  max-height: 70vh;
  overflow-y: auto;
}

.feed-states-modal .feed-states-scroll thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #fff;
}

.feed-states-modal .table {
  color: unset;
}
</style>
