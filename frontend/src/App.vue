<script setup>
import { ref, onMounted } from 'vue'
import { api } from '@/api/api.js'
import {
  state,
  hydrate,
  registerWatchers,
  refreshStats,
  refreshFeeds,
  refreshItems,
} from '@/state/store.js'
import { registerKeybindings } from '@/keybindings.js'
import FeedList from '@/components/FeedList/FeedList.vue'
import ItemList from '@/components/ItemList/ItemList.vue'
import ItemPane from '@/components/ItemPane/ItemPane.vue'
import NewFeedModal from '@/components/modals/NewFeedModal.vue'
import ShortcutsModal from '@/components/modals/ShortcutsModal.vue'
import CompareOpmlModal from '@/components/modals/CompareOpmlModal.vue'

const ready = ref(false)

onMounted(async () => {
  const [s, status] = await Promise.all([api.settings.get(), api.status()])
  hydrate(s, status)

  registerWatchers()
  registerKeybindings()
  ready.value = true

  refreshStats()
    .then(refreshFeeds)
    .then(() => refreshItems(false))
  api.feeds.list_errors().then(function (errors) {
    state.feed_errors = errors
  })
})
</script>

<template>
  <div
    id="app"
    class="d-flex"
    :class="{
      'feed-selected': state.feedSelected !== null,
      'item-selected': state.itemSelected !== null,
    }"
    v-if="ready"
  >
    <FeedList />
    <ItemList />
    <ItemPane />

    <NewFeedModal />
    <ShortcutsModal />
    <CompareOpmlModal />
  </div>
</template>
