<!--
// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

<template>
  <window
    id="connect"
    flash-class="home-window-display"
    :display="display"
    @display="$emit('display', $event)"
    @close="$emit('close')"
  >
    <div id="connect-frame">
      <h1 class="window-title">Establish connection with</h1>

      <slot v-if="inputting"></slot>

      <connect-switch
        v-if="!inputting"
        :knowns-length="knowns.length"
        :tab="tab"
        @switch="switchTab"
      ></connect-switch>

      <connect-new
        v-if="tab === 'new' && !inputting"
        :connectors="connectors"
        @select="selectConnector"
      ></connect-new>

      <connect-known
        v-if="tab === 'known' && !inputting"
        :presets="presets"
        :restricted-to-presets="restrictedToPresets"
        :preset-editing="presetEditing"
        :presets-api="presetsApi"
        :connectors="connectors"
        :knowns="knowns"
        :launcher-builder="knownsLauncherBuilder"
        :knowns-export="knownsExport"
        :knowns-import="knownsImport"
        @select="selectKnown"
        @select-preset="selectPreset"
        @remove="removeKnown"
        @clear-session="clearSessionKnown"
      ></connect-known>

      <div id="connect-warning">
        <span id="connect-warning-icon" class="icon icon-warning1"></span>
        <div id="connect-warning-msg">
          <p>
            <strong>An insecured service may steal your secrets.</strong>
            Always exam the safety of the service before using it.
          </p>

          <p>
            Sshwifty is a free software, you can deploy it on your own trusted
            infrastructure.
            <a href="https://github.com/nirui/sshwifty" target="_blank"
              >Learn more</a
            >
          </p>
        </div>
      </div>

      <div v-if="busy" id="connect-busy-overlay"></div>
    </div>
  </window>
</template>

<script>
import "./connect.css";

import Window from "./window.vue";
import ConnectSwitch from "./connect_switch.vue";
import ConnectKnown from "./connect_known.vue";
import ConnectNew from "./connect_new.vue";

export default {
  components: {
    window: Window,
    "connect-switch": ConnectSwitch,
    "connect-known": ConnectKnown,
    "connect-new": ConnectNew,
  },
  props: {
    display: {
      type: Boolean,
      default: false,
    },
    inputting: {
      type: Boolean,
      default: false,
    },
    presets: {
      type: Array,
      default: () => [],
    },
    restrictedToPresets: {
      type: Boolean,
      default: () => false,
    },
    presetEditing: {
      type: Boolean,
      default: () => false,
    },
    presetsApi: {
      type: Object,
      default: () => null,
    },
    knowns: {
      type: Array,
      default: () => [],
    },
    knownsLauncherBuilder: {
      type: Function,
      default: () => [],
    },
    knownsExport: {
      type: Function,
      default: () => [],
    },
    knownsImport: {
      type: Function,
      default: () => [],
    },
    connectors: {
      type: Array,
      default: () => [],
    },
    busy: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      tab: this.defaultTab(),
      canSelect: true,
    };
  },
  watch: {
    display(newVal) {
      if (!newVal) {
        return;
      }

      // The window is never destroyed, so without this it would keep both
      // the tab the user last switched to and a choice made back when there
      // was nothing known yet
      if (this.defaultTab() === "new") {
        this.openNewTab();
        return;
      }

      this.tab = "known";
    },
  },
  methods: {
    // Reconnecting to a remote used before is the common case, so the known
    // remotes are what the window opens on whenever there are any.
    // A method rather than a computed: Vue 2 sets methods up before data,
    // computed properties only after it
    defaultTab() {
      return this.restrictedToPresets ||
        this.knowns.length > 0 ||
        this.presets.length > 0
        ? "known"
        : "new";
    },
    switchTab(to) {
      if (this.inputting) {
        return;
      }

      if (to === "new") {
        this.openNewTab();
        return;
      }

      this.tab = to;
    },
    // With only one protocol there is nothing to pick, so go straight to its
    // form. Only on opening: after the user cancels the form the one-entry
    // list stays, re-opening it here would loop
    openNewTab() {
      this.tab = "new";

      if (this.connectors.length === 1 && !this.inputting && !this.busy) {
        this.selectConnector(this.connectors[0]);
      }
    },
    selectConnector(connector) {
      if (this.inputting) {
        return;
      }

      this.$emit("connector-select", connector);
    },
    selectKnown(known) {
      if (this.inputting) {
        return;
      }

      this.$emit("known-select", known);
    },
    removeKnown(uid) {
      if (this.inputting) {
        return;
      }

      this.$emit("known-remove", uid);
    },
    selectPreset(preset) {
      if (this.inputting) {
        return;
      }

      this.$emit("preset-select", preset);
    },
    clearSessionKnown(uid) {
      if (this.inputting) {
        return;
      }

      this.$emit("known-clear-session", uid);
    },
  },
};
</script>
