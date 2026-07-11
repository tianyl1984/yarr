<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import {
  state,
  refs,
  feedsById,
  itemSelectedContent,
  contentImages,
  contentAudios,
  contentVideos,
  toggleItemStarred,
  toggleItemRead,
  toggleReadability,
  navigateToItem,
  incrFont,
  formatDate,
} from '@/state/store.js'
import Dropdown from '@/components/common/Dropdown.vue'
import Icon from '@/components/common/Icon.vue'

const content = ref(null)

onMounted(() => {
  refs.content = content.value
})
onUnmounted(() => {
  if (refs.content === content.value) refs.content = null
})
</script>

<template>
  <div id="col-item" class="vh-100 d-flex flex-column w-100" style="min-width: 0">
    <div class="toolbar px-2 d-flex align-items-center" v-if="state.itemSelectedDetails">
      <button
        class="toolbar-item"
        @click="toggleItemStarred(state.itemSelectedDetails)"
        title="Mark Starred"
      >
        <Icon name="star-full" v-if="state.itemSelectedDetails.status == 'starred'" />
        <Icon name="star" v-else-if="state.itemSelectedDetails.status != 'starred'" />
      </button>
      <button
        class="toolbar-item"
        title="Mark Unread"
        @click="toggleItemRead(state.itemSelectedDetails)"
      >
        <Icon name="circle-full" v-if="state.itemSelectedDetails.status == 'unread'" />
        <Icon name="circle" v-if="state.itemSelectedDetails.status != 'unread'" />
      </button>
      <Dropdown
        class="settings-dropdown"
        toggle-class="toolbar-item px-2"
        drop="center"
        title="Appearance"
      >
        <template v-slot:button>
          <Icon name="sliders" />
        </template>

        <button
          class="dropdown-item"
          :class="{ active: !state.theme.font }"
          @click.stop="state.theme.font = ''"
        >
          sans-serif
        </button>
        <button
          class="dropdown-item font-serif"
          :class="{ active: state.theme.font == 'serif' }"
          @click.stop="state.theme.font = 'serif'"
        >
          serif
        </button>
        <button
          class="dropdown-item font-monospace"
          :class="{ active: state.theme.font == 'monospace' }"
          @click.stop="state.theme.font = 'monospace'"
        >
          monospace
        </button>

        <div class="d-flex text-center">
          <button class="dropdown-item" style="font-size: 0.8rem" @click.stop="incrFont(-1)">A</button>
          <button class="dropdown-item" style="font-size: 1.2rem" @click.stop="incrFont(1)">A</button>
        </div>
      </Dropdown>
      <button
        class="toolbar-item"
        :class="{ active: state.itemSelectedReadability }"
        @click="toggleReadability()"
        title="Read Here"
      >
        <Icon name="book-open" :class="{ 'icon-loading': state.loading.readability }" />
      </button>
      <a
        class="toolbar-item"
        :href="state.itemSelectedDetails.link"
        rel="noopener noreferrer"
        target="_blank"
        referrerpolicy="no-referrer"
        title="Open Link"
      >
        <Icon name="external-link" />
      </a>
      <div class="flex-grow-1"></div>
      <button
        class="toolbar-item"
        @click="navigateToItem(-1)"
        title="Previous Article"
        :disabled="state.itemSelected == state.items[0].id"
      >
        <Icon name="chevron-left" />
      </button>
      <button
        class="toolbar-item"
        @click="navigateToItem(+1)"
        title="Next Article"
        :disabled="state.itemSelected == state.items[state.items.length - 1].id"
      >
        <Icon name="chevron-right" />
      </button>
      <button class="toolbar-item" @click="state.itemSelected = null" title="Close Article">
        <Icon name="x" />
      </button>
    </div>
    <div
      v-if="state.itemSelectedDetails"
      ref="content"
      class="content px-4 pt-3 pb-5 border-top overflow-auto scroll-touch"
      :class="{
        'font-serif': state.theme.font == 'serif',
        'font-monospace': state.theme.font == 'monospace',
      }"
      :style="{ 'font-size': state.theme.size + 'rem' }"
    >
      <div class="content-wrapper">
        <h1><b>{{ state.itemSelectedDetails.title || 'untitled' }}</b></h1>
        <div class="text-muted">
          <div>
            <span
              class="cursor-pointer"
              @click="state.feedSelected = 'feed:' + (feedsById[state.itemSelectedDetails.feed_id] || {}).id"
            >
              {{ (feedsById[state.itemSelectedDetails.feed_id] || {}).title }}
            </span>
          </div>
          <time>{{ formatDate(state.itemSelectedDetails.date) }}</time>
        </div>
        <hr />
        <div v-if="!state.itemSelectedReadability">
          <div v-if="contentImages.length">
            <figure v-for="(media, i) in contentImages" :key="i">
              <img :src="media.url" loading="lazy" />
              <figcaption v-if="media.description">{{ media.description }}</figcaption>
            </figure>
          </div>
          <audio class="w-100" controls v-for="(media, i) in contentAudios" :key="'a' + i" :src="media.url"></audio>
          <video class="w-100" controls v-for="(media, i) in contentVideos" :key="'v' + i" :src="media.url"></video>
        </div>
        <div v-html="itemSelectedContent"></div>
      </div>
    </div>
  </div>
</template>
