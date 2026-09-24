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
  <div id="connect-known-list" :class="{ reloaded: reloaded }">
    <preset-manager
      v-if="manager"
      :api="presetsApi"
      :types="connectors.map((c) => c.name())"
      :draft="manager.draft"
      @close="manager = null"
    ></preset-manager>
    <div v-else-if="sections.length <= 0" id="connect-known-list-empty">
      No known remote available
    </div>
    <div v-else>
      <div
        v-for="(section, sk) in sections"
        :key="section.title"
        class="known-section"
        :class="{ 'last-planel': sk > 0 }"
      >
        <h3>
          {{ section.title }}
          <a
            v-if="section.presets && presetEditing"
            class="manage"
            href="javascript:;"
            @click="manager = { draft: null }"
          >
            Manage
          </a>
        </h3>

        <ul class="hlst lstcl1">
          <li
            v-for="card in section.cards"
            :key="card.key"
            class="known-card"
            :class="{ disabled: card.preset && presetDisabled(card.preset) }"
          >
            <div class="labels">
              <span class="type" :style="'background-color: ' + card.color">
                {{ card.type }}
              </span>

              <template v-if="card.known">
                <a
                  class="opt link"
                  href="javascript:;"
                  @click="launcher(card.known, $event)"
                >
                  {{ copying[card.known.uid] || "Copy link" }}
                </a>

                <a
                  v-if="card.known.session"
                  class="opt clr"
                  href="javascript:;"
                  title="Clear session data"
                  @click="clearSession(card.known.uid)"
                >
                  Clear
                </a>
                <a
                  v-else-if="card.preset"
                  class="opt del"
                  href="javascript:;"
                  title="Forget every connection to this preset"
                  @click="forget(card.knowns)"
                >
                  Forget
                </a>
                <a
                  v-else
                  class="opt del"
                  href="javascript:;"
                  @click="remove(card.known.uid)"
                >
                  Remove
                </a>

                <a
                  v-if="!card.preset && presetEditing"
                  class="opt save"
                  href="javascript:;"
                  @click="saveAsPreset(card.known)"
                >
                  Save as preset
                </a>
              </template>
            </div>

            <div class="lst-wrap" @click="selectCard(card)">
              <h4
                :title="card.title"
                :class="{ highlight: card.known && card.known.session }"
              >
                {{ card.title }}
              </h4>
              <div class="info">
                <template v-if="card.known">
                  Last: {{ card.known.last.toLocaleString() }}
                </template>
              </div>
            </div>
          </li>
        </ul>

        <div
          v-if="section.presets && restrictedToPresets"
          id="connect-known-list-presets-alert"
        >
          The operator has restricted the outgoing connections. You can only
          connect to remotes from the pre-defined presets.
        </div>
      </div>
    </div>

    <div v-if="!manager" id="connect-known-list-import">
      Tip: You can
      <a href="javascript:;" @click="importHosts">import</a> and
      <a href="javascript:;" @click="exportHosts">export</a>
      known remotes from and to a file.
    </div>
  </div>
</template>

<script>
import "./connect_known.css";

import { groupKnowns } from "../commands/commands.js";
import PresetManager from "./preset_manager.vue";

export default {
  components: {
    "preset-manager": PresetManager,
  },
  props: {
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
    connectors: {
      type: Array,
      default: () => [],
    },
    knowns: {
      type: Array,
      default: () => [],
    },
    launcherBuilder: {
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
  },
  data() {
    return {
      copying: {},
      reloaded: false,
      busy: false,
      manager: null,
    };
  },
  computed: {
    sections() {
      const g = groupKnowns(this.presets, this.knowns);
      return [
        {
          title: "Presets",
          presets: true,
          cards: g.presets.map((p, i) => ({
            key: "preset-" + i,
            type: p.preset.command.name(),
            color: p.preset.command.color(),
            title: p.preset.preset.title(),
            preset: p.preset,
            known: p.knowns.length > 0 ? p.knowns[0] : null,
            knowns: p.knowns,
          })),
        },
        {
          title: "Other recent",
          presets: false,
          cards: g.others.map((k) => ({
            key: "known-" + k.uid,
            type: k.type,
            color: k.color,
            title: k.title,
            preset: null,
            known: k,
            knowns: [k],
          })),
        },
        // Keep the Presets section while editing, it holds the Manage link
      ].filter((s) => s.cards.length > 0 || (s.presets && this.presetEditing));
    },
  },
  watch: {
    knowns(newVal, oldVal) {
      // Only play reload animation when we're adding data into the records,
      // not reducing
      if (newVal.length <= oldVal.length) {
        return;
      }

      const self = this;

      self.reloaded = true;
      setTimeout(() => {
        self.reloaded = false;
      }, 500);
    },
  },
  methods: {
    selectCard(card) {
      if (card.preset) {
        this.selectPreset(card.preset);
      } else {
        this.select(card.known);
      }
    },
    select(known) {
      if (this.busy) {
        return;
      }

      this.$emit("select", known);
    },
    presetDisabled(preset) {
      if (!this.restrictedToPresets || preset.preset.host().length > 0) {
        return false;
      }

      return true;
    },
    selectPreset(preset) {
      if (this.busy || this.presetDisabled(preset)) {
        return;
      }

      this.$emit("select-preset", preset);
    },
    async launcher(known, ev) {
      if (this.copying[known.uid] || this.busy) {
        return;
      }

      ev.preventDefault();

      this.busy = true;
      this.$set(this.copying, known.uid, "Copying");

      let lnk = this.launcherBuilder(known);

      try {
        await navigator.clipboard.writeText(lnk);

        this.copying[known.uid] = "Copied!";
      } catch (e) {
        this.copying[known.uid] = "Failed";
        ev.target.setAttribute("href", lnk);
      }

      setTimeout(() => {
        this.$delete(this.copying, known.uid);
      }, 2000);

      this.busy = false;
    },
    saveAsPreset(known) {
      if (this.busy) {
        return;
      }

      const command = this.connectors.find((c) => c.name() === known.type);

      this.manager = { draft: command.presetFromKnown(known) };
    },
    remove(uid) {
      if (this.busy) {
        return;
      }

      this.$emit("remove", uid);
    },
    forget(knowns) {
      if (this.busy) {
        return;
      }

      for (let i = 0; i < knowns.length; i++) {
        this.$emit("remove", knowns[i].uid);
      }
    },
    clearSession(uid) {
      if (this.busy) {
        return;
      }

      this.$emit("clear-session", uid);
    },
    exportHosts() {
      let el = null;

      try {
        const dataStr = JSON.stringify(this.knownsExport());

        el = document.createElement("a");
        el.setAttribute(
          "href",
          "data:text/plain;charset=utf-8," + btoa(dataStr),
        );
        el.setAttribute("target", "_blank");
        el.setAttribute("download", "sshwifty.known-remotes.txt");
        el.setAttribute(
          "style",
          "overflow: hidden; opacity: 0; width: 1px; height: 1px; top: -1px;" +
            "left: -1px; position: absolute;",
        );

        document.body.appendChild(el);

        el.click();
      } catch (e) {
        alert("Unable to export known remotes: " + e);
      }

      if (el === null) {
        return;
      }

      document.body.removeChild(el);
    },
    importHosts() {
      const self = this;

      let el = null;

      try {
        el = document.createElement("input");
        el.setAttribute("type", "file");
        el.setAttribute(
          "style",
          "overflow: hidden; opacity: 0; width: 1px; height: 1px; top: -1px;" +
            "left: -1px; position: absolute;",
        );
        el.addEventListener("change", (ev) => {
          const t = ev.target;

          if (t.files.length <= 0) {
            return;
          }

          t.disabled = "disabled";

          let r = new FileReader();

          r.onload = () => {
            try {
              self.knownsImport(JSON.parse(atob(r.result)));
            } catch (e) {
              alert("Unable to import known remotes due to error: " + e);
            }
          };

          r.readAsText(t.files[0], "utf-8");
        });

        document.body.appendChild(el);

        el.click();
      } catch (e) {
        alert("Unable to load known remotes data due to error: " + e);
      }

      if (el === null) {
        return;
      }

      document.body.removeChild(el);
    },
  },
};
</script>
