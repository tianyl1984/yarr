<script setup>
import { ref } from 'vue'
import { api } from '@/api/api.js'
import {
  state,
  showSettings,
  fetchAllFeeds,
  refreshFeeds,
  refreshStats,
  logout,
} from '@/state/store.js'
import Dropdown from '@/components/common/Dropdown.vue'
import Icon from '@/components/common/Icon.vue'

const menuDropdown = ref(null)
const importForm = ref(null)
const compareForm = ref(null)

function importOPML(event) {
  const input = event.target
  menuDropdown.value.hide()
  api.upload_opml(importForm.value).then(function () {
    input.value = ''
    refreshFeeds()
    refreshStats()
  })
}

function compareOPML(event) {
  state.opmlCompareResult = []
  const input = event.target
  menuDropdown.value.hide()
  api.compare_opml(compareForm.value).then(function (result) {
    input.value = ''
    state.opmlCompareResult = result
    state.settings = 'compare-opml'
  })
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

    <div class="dropdown-divider"></div>
    <header class="dropdown-header" role="heading" aria-level="2">
      Subscriptions
    </header>
    <form
      id="opml-compare-form"
      ref="compareForm"
      enctype="multipart/form-data"
      tabindex="-1"
    >
      <input
        type="file"
        id="opml-compare"
        @change="compareOPML"
        name="opml"
        style="opacity: 0; width: 1px; height: 0; position: absolute; z-index: -1"
      />
      <label class="dropdown-item mb-0 cursor-pointer" for="opml-compare" @click.stop="">
        <Icon name="compare" class="mr-1" />
        对比opml
      </label>
    </form>
    <form
      id="opml-import-form"
      ref="importForm"
      enctype="multipart/form-data"
      tabindex="-1"
    >
      <input
        type="file"
        id="opml-import"
        @change="importOPML"
        name="opml"
        style="opacity: 0; width: 1px; height: 0; position: absolute; z-index: -1"
      />
      <label class="dropdown-item mb-0 cursor-pointer" for="opml-import" @click.stop="">
        <Icon name="download" class="mr-1" />
        Import
      </label>
    </form>
    <a class="dropdown-item" href="./api/opml/export">
      <Icon name="upload" class="mr-1" />
      Export
    </a>
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
