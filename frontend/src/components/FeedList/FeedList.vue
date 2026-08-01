<script setup>
import {
  state,
  filteredStats,
  foldersWithFeeds,
  toggleFolderExpanded,
  resizeFeedList,
  mustHideFolder,
  mustHideFeed,
} from '@/state/store.js'
import Drag from '@/components/common/Drag.vue'
import Icon from '@/components/common/Icon.vue'
import SettingsDropdown from './SettingsDropdown.vue'
</script>

<template>
  <div
    id="col-feed-list"
    class="vh-100 position-relative d-flex flex-column border-right flex-shrink-0"
    :style="{ width: state.feedListWidth + 'px' }"
  >
    <Drag :width="state.feedListWidth" @resize="resizeFeedList" />
    <div class="p-2 toolbar d-flex align-items-center">
      <Icon name="anchor" class="mx-2" />
      <div class="flex-grow-1"></div>
      <button
        class="toolbar-item ml-1"
        :class="{ active: state.filterSelected == 'unread' }"
        :aria-pressed="state.filterSelected == 'unread'"
        title="Unread"
        @click="state.filterSelected = 'unread'"
      >
        <Icon name="circle-full" />
      </button>
      <button
        class="toolbar-item mx-1"
        :class="{ active: state.filterSelected == '' }"
        :aria-pressed="state.filterSelected == ''"
        title="All"
        @click="state.filterSelected = ''"
      >
        <Icon name="assorted" />
      </button>
      <div class="flex-grow-1"></div>
      <SettingsDropdown />
    </div>
    <div
      id="feed-list-scroll"
      class="p-2 overflow-auto scroll-touch border-top flex-grow-1"
    >
      <label class="selectgroup">
        <input type="radio" name="feed" value="" v-model="state.feedSelected" />
        <div class="selectgroup-label d-flex align-items-center w-100">
          <Icon name="layers" class="mr-2" />
          <span class="flex-fill text-left text-truncate" v-if="state.filterSelected == 'unread'">All Unread</span>
          <span class="flex-fill text-left text-truncate" v-if="state.filterSelected == ''">All Feeds</span>
          <span class="counter text-right">{{ filteredStats.total }}</span>
        </div>
      </label>
      <div v-for="folder in foldersWithFeeds" :key="folder.id">
        <label
          class="selectgroup mt-1"
          :class="{ 'd-none': mustHideFolder(folder) }"
          v-if="folder.id"
        >
          <input
            type="radio"
            name="feed"
            :value="'folder:' + folder.id"
            v-model="state.feedSelected"
          />
          <div class="selectgroup-label d-flex align-items-center w-100">
            <Icon
              name="chevron-right"
              class="mr-2"
              :class="{ expanded: folder.is_expanded }"
              @click.prevent="toggleFolderExpanded(folder)"
            />
            <span class="flex-fill text-left text-truncate">{{ folder.title }}</span>
            <span class="counter text-right">{{ filteredStats.folders[folder.id] || '' }}</span>
          </div>
        </label>
        <div
          v-show="!folder.id || folder.is_expanded"
          class="mt-1"
          :class="{ 'pl-3': folder.id }"
        >
          <label
            class="selectgroup"
            :class="{ 'd-none': mustHideFeed(feed) }"
            v-for="feed in folder.feeds"
            :key="feed.id"
          >
            <input type="radio" name="feed" :value="'feed:' + feed.id" v-model="state.feedSelected" />
            <div class="selectgroup-label d-flex align-items-center w-100">
              <Icon name="rss" class="mr-2" v-if="!feed.has_icon" />
              <span class="icon mr-2" v-else><img :src="'./api/feeds/' + feed.id + '/icon'" alt="" loading="lazy" /></span>
              <span class="flex-fill text-left text-truncate">{{ feed.title }}</span>
              <span class="counter text-right">{{ filteredStats.feeds[feed.id] || '' }}</span>
              <Icon
                name="alert-circle"
                class="flex-shrink-0 mx-2"
                :title="state.feed_errors[feed.id]"
                v-if="!state.filterSelected && state.feed_errors[feed.id]"
              />
            </div>
          </label>
        </div>
      </div>
    </div>
  </div>
</template>
