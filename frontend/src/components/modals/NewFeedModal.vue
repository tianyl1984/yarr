<script setup>
import { ref, nextTick } from 'vue'
import {
  state,
  current,
  createFeed,
  createNewFeedFolder,
  resetFeedChoice,
} from '@/state/store.js'
import { focus as vFocus } from '@/directives/focus.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'

const newFeedFolder = ref(null)

function newFolder() {
  const result = createNewFeedFolder()
  if (!result) return
  result.then(function (folder) {
    nextTick(function () {
      if (folder && newFeedFolder.value) newFeedFolder.value.value = folder.id
    })
  })
}
</script>

<template>
  <Modal :open="state.settings == 'create'" @hide="state.settings = ''">
    <button
      class="btn btn-link outline-none float-right p-2 mr-n2 mt-n2"
      style="line-height: 1"
      @click="state.settings = ''"
    >
      <Icon name="x" />
    </button>
    <div>
      <p class="cursor-default"><b>New Feed</b></p>
      <form action="" @submit.prevent="createFeed($event)" class="mt-4">
        <label for="feed-url">URL</label>
        <input
          id="feed-url"
          name="url"
          type="url"
          class="form-control"
          required
          autocomplete="off"
          :readonly="state.feedNewChoice.length > 0"
          placeholder=""
          v-focus
        />
        <label for="feed-folder" class="mt-3 d-block">
          Folder
          <a href="#" class="float-right text-decoration-none" @click.prevent="newFolder()">new folder</a>
        </label>
        <select class="form-control" id="feed-folder" name="folder_id" ref="newFeedFolder">
          <option value="">---</option>
          <option
            :value="folder.id"
            v-for="folder in state.folders"
            :key="folder.id"
            :selected="folder.id === current.feed.folder_id || folder.id === current.folder.id"
          >
            {{ folder.title }}
          </option>
        </select>
        <label for="feed-use-proxy" class="mt-3 d-block">Proxy</label>
        <select class="form-control" id="feed-use-proxy" name="use_proxy">
          <option value="false" selected>不使用</option>
          <option value="true">使用</option>
        </select>
        <div class="mt-4" v-if="state.feedNewChoice.length">
          <p class="mb-2">
            Multiple feeds found. Choose one below:
            <a href="#" class="float-right text-decoration-none" @click.prevent="resetFeedChoice()">cancel</a>
          </p>
          <label class="selectgroup" v-for="choice in state.feedNewChoice" :key="choice.url">
            <input type="radio" name="feedToAdd" :value="choice.url" v-model="state.feedNewChoiceSelected" />
            <div class="selectgroup-label">
              <div class="text-truncate">{{ choice.title }}</div>
              <div class="text-truncate" :class="{ light: choice.title }">{{ choice.url }}</div>
            </div>
          </label>
        </div>
        <button class="btn btn-block btn-default mt-3" :class="{ loading: state.loading.newfeed }" type="submit">
          Add
        </button>
      </form>
    </div>
  </Modal>
</template>
