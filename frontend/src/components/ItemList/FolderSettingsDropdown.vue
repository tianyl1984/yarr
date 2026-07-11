<script setup>
import { current, renameFolder, deleteFolder } from '@/state/store.js'
import Dropdown from '@/components/common/Dropdown.vue'
import Icon from '@/components/common/Icon.vue'
</script>

<template>
  <Dropdown
    class="settings-dropdown"
    toggle-class="btn btn-link toolbar-item px-2 ml-2"
    title="Folder Settings"
    drop="right"
  >
    <template v-slot:button>
      <Icon name="more-horizontal" />
    </template>
    <header class="dropdown-header" role="heading" aria-level="2">
      {{ current.folder.title }}
    </header>
    <button class="dropdown-item" @click="renameFolder(current.folder)">
      <Icon name="edit" class="mr-1" />
      Rename
    </button>
    <div class="dropdown-divider"></div>
    <button class="dropdown-item text-danger" @click="deleteFolder(current.folder)">
      <Icon name="trash" class="mr-1" />
      Delete
    </button>
  </Dropdown>
</template>
