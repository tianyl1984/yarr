<script setup>
import { ref } from 'vue'
import { opml, compareOPML } from '@/state/opml.js'
import {
  state,
  showSettings,
  fetchAllFeeds,
  refreshFeedStates,
  logout,
} from '@/state/store.js'
import Dropdown from '@/components/common/Dropdown.vue'
import Icon from '@/components/common/Icon.vue'

const menuDropdown = ref(null)

function showFeedStates() {
  menuDropdown.value.hide()
  refreshFeedStates()
  state.settings = 'feed-states'
}

function openCompare() {
  menuDropdown.value.hide()
  state.settings = 'compare-opml'
  if (opml.file) compareOPML()
}
</script>

<template>
  <Dropdown
    class="settings-dropdown"
    toggle-class="btn btn-link toolbar-item px-2"
    ref="menuDropdown"
    drop="right"
    title="Settings"
  >
    <template v-slot:button>
      <Icon name="more-horizontal" />
    </template>

    <button class="dropdown-item" @click="showSettings('create')">
      <Icon name="plus" class="mr-1" />
      New Feed
    </button>
    <div class="dropdown-divider"></div>
    <button class="dropdown-item" @click="fetchAllFeeds()">
      <Icon name="rotate-cw" class="mr-1" />
      Refresh Feeds
    </button>
    <button class="dropdown-item" @click="showFeedStates()">
      <Icon name="alert-circle" class="mr-1" />
      抓取结果
    </button>

    <div class="dropdown-divider"></div>
    <header class="dropdown-header" role="heading" aria-level="2">
      Subscriptions
    </header>
    <button class="dropdown-item" @click="openCompare()">
      <Icon name="compare" class="mr-1" />
      对比opml
    </button>
    <div class="dropdown-divider"></div>
    <button class="dropdown-item" @click="showSettings('shortcuts')">
      <Icon name="help-circle" class="mr-1" />
      Shortcuts
    </button>
    <div class="dropdown-divider" v-if="state.authenticated"></div>
    <button
      class="dropdown-item"
      v-if="state.authenticated"
      @click="logout()"
    >
      <Icon name="log-out" class="mr-1" />
      Log out
    </button>
  </Dropdown>
</template>
