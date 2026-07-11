<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import {
  state,
  current,
  feedsById,
  refs,
  loadMoreItems,
  markItemsRead,
  resizeItemList,
  formatDate,
} from '@/state/store.js'
import { scroll as vScroll } from '@/directives/scroll.js'
import Drag from '@/components/common/Drag.vue'
import Icon from '@/components/common/Icon.vue'
import RelativeTime from '@/components/common/RelativeTime.vue'
import FeedSettingsDropdown from './FeedSettingsDropdown.vue'
import FolderSettingsDropdown from './FolderSettingsDropdown.vue'

const itemlist = ref(null)

onMounted(() => {
  refs.itemlist = itemlist.value
})
onUnmounted(() => {
  if (refs.itemlist === itemlist.value) refs.itemlist = null
})
</script>

<template>
  <div
    id="col-item-list"
    class="vh-100 position-relative d-flex flex-column border-right flex-shrink-0"
    :style="{ width: state.itemListWidth + 'px' }"
  >
    <Drag :width="state.itemListWidth" @resize="resizeItemList" />
    <div class="px-2 toolbar d-flex align-items-center">
      <button
        class="toolbar-item mr-2 d-block d-md-none"
        @click="state.feedSelected = null"
        title="Show Feeds"
      >
        <Icon name="chevron-left" />
      </button>
      <div class="input-icon flex-grow-1">
        <Icon name="search" />
        <!-- id used by keybindings -->
        <input
          id="searchbar"
          type=""
          class="d-block toolbar-search"
          v-model="state.itemSearch"
          @keydown.enter="$event.target.blur()"
        />
      </div>
      <button
        class="toolbar-item ml-2"
        @click="markItemsRead()"
        v-if="state.filterSelected == 'unread'"
        title="Mark All Read"
      >
        <Icon name="check" />
      </button>

      <button class="btn btn-link toolbar-item px-2 ml-2" v-if="!current.type" disabled>
        <Icon name="more-horizontal" />
      </button>
      <FeedSettingsDropdown v-if="current.type == 'feed'" />
      <FolderSettingsDropdown v-if="current.type == 'folder'" />
    </div>
    <div
      id="item-list-scroll"
      class="p-2 overflow-auto scroll-touch border-top flex-grow-1"
      v-scroll="loadMoreItems"
      ref="itemlist"
    >
      <label v-for="item in state.items" :key="item.id" class="selectgroup">
        <input type="radio" name="item" :value="item.id" v-model="state.itemSelected" />
        <div class="selectgroup-label d-flex flex-column">
          <div
            style="line-height: 1; opacity: 0.7; margin-bottom: 0.1rem"
            class="d-flex align-items-center"
          >
            <transition name="indicator">
              <Icon name="circle-full" class="icon-small mr-1" v-if="item.status == 'unread'" />
              <Icon name="star-full" class="icon-small mr-1" v-else-if="item.status == 'starred'" />
            </transition>
            <small class="flex-fill text-truncate mr-1">
              {{ (feedsById[item.feed_id] || {}).title }}
            </small>
            <small class="flex-shrink-0">
              <RelativeTime :title="formatDate(item.date)" :val="item.date" />
            </small>
          </div>
          <div>{{ item.title || 'untitled' }}</div>
        </div>
      </label>
      <button class="btn btn-link btn-block loading my-3" v-if="state.itemsHasMore"></button>
    </div>
    <div
      class="px-3 py-2 border-top text-danger text-break"
      v-if="state.feed_errors[current.feed.id]"
    >
      {{ state.feed_errors[current.feed.id] }}
    </div>
  </div>
</template>
