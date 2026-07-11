<script setup>
import {
  state,
  current,
  moveFeed,
  moveFeedToNewFolder,
  renameFeed,
  updateFeedLink,
  deleteFeed,
} from '@/state/store.js'
import Dropdown from '@/components/common/Dropdown.vue'
import Icon from '@/components/common/Icon.vue'
</script>

<template>
  <Dropdown
    class="settings-dropdown"
    toggle-class="btn btn-link toolbar-item px-2 ml-2"
    drop="right"
    title="Feed Settings"
  >
    <template v-slot:button>
      <Icon name="more-horizontal" />
    </template>
    <header class="dropdown-header" role="heading" aria-level="2">
      {{ current.feed.title }}
    </header>
    <a
      class="dropdown-item"
      :href="current.feed.link"
      rel="noopener noreferrer"
      target="_blank"
      referrerpolicy="no-referrer"
      v-if="current.feed.link"
    >
      <Icon name="globe" class="mr-1" />
      Website
    </a>
    <a
      class="dropdown-item"
      :href="current.feed.feed_link"
      rel="noopener noreferrer"
      target="_blank"
      referrerpolicy="no-referrer"
      v-if="current.feed.feed_link"
    >
      <Icon name="rss" class="mr-1" />
      Feed Link
    </a>
    <div class="dropdown-divider" v-if="current.feed.link || current.feed.feed_link"></div>
    <button class="dropdown-item" @click="renameFeed(current.feed)">
      <Icon name="edit" class="mr-1" />
      Rename
    </button>
    <button
      class="dropdown-item"
      @click="updateFeedLink(current.feed)"
      v-if="current.feed.feed_link"
    >
      <Icon name="edit" class="mr-1" />
      Change Link
    </button>
    <div class="dropdown-divider"></div>
    <header class="dropdown-header" role="heading" aria-level="2">Move to...</header>
    <template v-for="folder in state.folders" :key="folder.id">
      <button
        class="dropdown-item"
        v-if="folder.id != current.feed.folder_id"
        @click="moveFeed(current.feed, folder)"
      >
        <Icon name="folder" class="mr-1" />
        {{ folder.title }}
      </button>
    </template>
    <button
      class="dropdown-item text-muted"
      @click="moveFeed(current.feed, null)"
      v-if="current.feed.folder_id"
    >
      <Icon name="folder-minus" class="mr-1" />
      ──
    </button>
    <button class="dropdown-item text-muted" @click="moveFeedToNewFolder(current.feed)">
      <Icon name="folder-plus" class="mr-1" />
      new folder
    </button>
    <div class="dropdown-divider"></div>
    <button class="dropdown-item text-danger" @click.prevent="deleteFeed(current.feed)">
      <Icon name="trash" class="mr-1" />
      Delete
    </button>
  </Dropdown>
</template>
