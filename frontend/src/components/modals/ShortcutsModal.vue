<script setup>
import { state } from '@/state/store.js'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/common/Icon.vue'
</script>

<template>
  <Modal :open="state.settings == 'shortcuts'" @hide="state.settings = ''">
    <button
      class="btn btn-link outline-none float-right p-2 mr-n2 mt-n2"
      style="line-height: 1"
      @click="state.settings = ''"
    >
      <Icon name="x" />
    </button>
    <div>
      <p class="cursor-default"><b>Keyboard Shortcuts</b></p>

      <table class="table table-borderless table-sm table-compact m-0">
        <tbody>
          <tr>
            <td><kbd>1</kbd> <kbd>2</kbd> <kbd>3</kbd></td>
            <td>show unread / starred / all feeds</td>
          </tr>
          <tr><td><kbd>/</kbd></td><td>focus the search bar</td></tr>

          <tr><td colspan="2">&nbsp;</td></tr>
          <tr><td><kbd>j</kbd> <kbd>k</kbd></td><td>next / prev article</td></tr>
          <tr><td><kbd>l</kbd> <kbd>h</kbd></td><td>next / prev feed</td></tr>
          <tr><td><kbd>q</kbd></td><td>close article</td></tr>

          <tr><td colspan="2">&nbsp;</td></tr>
          <tr><td><kbd>R</kbd></td><td>mark all read</td></tr>
          <tr><td><kbd>r</kbd></td><td>mark read / unread</td></tr>
          <tr><td><kbd>s</kbd></td><td>mark starred / unstarred</td></tr>
          <tr><td><kbd>o</kbd></td><td>open link</td></tr>
          <tr><td><kbd>i</kbd></td><td>read here</td></tr>
          <tr><td><kbd>f</kbd> <kbd>b</kbd></td><td>scroll content forward / backward</td></tr>
        </tbody>
      </table>
    </div>
  </Modal>
</template>
