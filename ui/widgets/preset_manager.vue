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
  <div id="preset-manager">
    <a
      id="preset-manager-back"
      href="javascript:;"
      :class="{ disabled: busy }"
      @click="close"
    >
      Back
    </a>

    <h3>
      {{ form ? (form.index < 0 ? "Add preset" : "Edit preset") : "Presets" }}
    </h3>

    <div v-if="error" id="preset-manager-error">{{ error }}</div>

    <div v-if="loading" id="preset-manager-loading">Loading ...</div>

    <form
      v-else-if="form"
      class="form1"
      action="javascript:;"
      method="POST"
      @submit="submit"
    >
      <fieldset :disabled="busy">
        <div class="field">
          Title
          <input v-model="form.title" type="text" autocomplete="off" />
        </div>

        <div class="field">
          Type
          <select v-model="form.type" autocomplete="off">
            <option v-for="t in typeOptions" :key="t" :value="t">
              {{ t }}
            </option>
          </select>
        </div>

        <div class="field">
          Host
          <input
            v-model="form.host"
            type="text"
            autocomplete="off"
            placeholder="example.com:22"
          />
        </div>

        <div class="field">
          Tab color
          <input
            v-model="form.tabColor"
            type="text"
            autocomplete="off"
            placeholder="112233"
          />
        </div>

        <div v-for="(m, mk) in form.meta" :key="mk" class="field meta">
          <input
            v-model="m.key"
            type="text"
            autocomplete="off"
            placeholder="Field"
            list="preset-manager-meta-keys"
          />
          <a class="opt" href="javascript:;" @click="form.meta.splice(mk, 1)">
            Remove
          </a>
          <a
            v-if="isSecret(m) && !m.revealed"
            class="opt"
            href="javascript:;"
            @click="m.revealed = true"
          >
            Reveal
          </a>
          <div v-if="isSecret(m) && !m.revealed" class="textinfo">
            <div class="info">&bull;&bull;&bull;&bull;&bull;&bull;</div>
          </div>
          <textarea v-else v-model="m.value" autocomplete="off"></textarea>
        </div>

        <datalist id="preset-manager-meta-keys">
          <option v-for="k in suggestedKeys" :key="k" :value="k"></option>
        </datalist>

        <div class="field">
          <a href="javascript:;" @click="addMeta">Add field</a>
        </div>

        <div class="field">
          <button type="submit">Save</button>
          <button type="button" class="secondary" @click="form = null">
            Cancel
          </button>
        </div>
      </fieldset>
    </form>

    <div v-else>
      <ul class="lst1 lst-nostyle">
        <li v-for="(p, pk) in list" :key="pk">
          <div class="lst-wrap">
            <span class="opts">
              <a class="opt" href="javascript:;" @click="edit(pk)">Edit</a>
              <a class="opt del" href="javascript:;" @click="remove(pk)">
                Delete
              </a>
            </span>
            <h4>{{ p.Title }}</h4>
            {{ p.Type }} {{ p.Host }}
          </div>
        </li>
      </ul>

      <a id="preset-manager-add" href="javascript:;" @click="add">
        Add preset
      </a>
    </div>
  </div>
</template>

<script>
import "./preset_manager.css";

const metaKeys = {
  SSH: [
    "User",
    "Authentication",
    "Encoding",
    "Fingerprint",
    "Private Key",
    "Password",
  ],
  Telnet: ["Encoding"],
};

const secretKeys = ["Private Key", "Password"];

export default {
  props: {
    api: {
      type: Object,
      default: () => null,
    },
    types: {
      type: Array,
      default: () => [],
    },
    // A preset to open the add form with, in the configuration file form
    draft: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      loading: true,
      busy: false,
      error: "",
      revision: "",
      list: [],
      form: null,
    };
  },
  computed: {
    typeOptions() {
      // Keep the Type of a preset whose protocol is not enabled, so the
      // form does not silently change it
      return this.form && this.types.indexOf(this.form.type) < 0
        ? this.types.concat([this.form.type])
        : this.types;
    },
    suggestedKeys() {
      return this.form ? metaKeys[this.form.type] || [] : [];
    },
  },
  async mounted() {
    await this.load();

    if (this.draft) {
      this.open(-1, this.draft);
    }
  },
  methods: {
    async load() {
      this.loading = true;

      try {
        const d = await this.api.load();

        this.revision = d.revision;
        this.list = d.presets || [];
      } catch (e) {
        this.error = "" + e;
      }

      this.loading = false;
    },
    close() {
      if (this.busy) {
        return;
      }

      this.$emit("close");
    },
    isSecret(m) {
      return secretKeys.indexOf(m.key) >= 0;
    },
    open(index, p) {
      const meta = p.Meta || {};

      this.error = "";
      this.form = {
        index: index,
        // Finds the preset again if the list is reloaded before it is saved
        original: index < 0 ? "" : JSON.stringify(p),
        title: p.Title || "",
        type: p.Type || this.types[0] || "",
        host: p.Host || "",
        tabColor: p.TabColor || "",
        meta: Object.keys(meta).map((k) => ({
          key: k,
          value: meta[k],
          revealed: false,
        })),
      };
    },
    add() {
      this.open(-1, {});
    },
    edit(index) {
      this.open(index, this.list[index]);
    },
    addMeta() {
      this.form.meta.push({ key: "", value: "", revealed: true });
    },
    async remove(index) {
      if (
        this.busy ||
        !confirm('Delete preset "' + this.list[index].Title + '"?')
      ) {
        return;
      }

      const list = this.list.slice();

      list.splice(index, 1);

      await this.save(list);
    },
    async submit() {
      const meta = {};

      for (let i = 0; i < this.form.meta.length; i++) {
        const key = this.form.meta[i].key.trim();

        if (key.length > 0) {
          meta[key] = this.form.meta[i].value;
        }
      }

      const p = {
          Title: this.form.title,
          Type: this.form.type,
          Host: this.form.host,
          TabColor: this.form.tabColor,
          Meta: meta,
        },
        list = this.list.slice();

      if (this.form.index < 0 || this.form.index >= list.length) {
        list.push(p);
      } else {
        list[this.form.index] = p;
      }

      if (!(await this.save(list))) {
        return;
      }

      this.form = null;

      // Opened to save one remote as a preset, which is done now
      if (this.draft) {
        this.$emit("close");
      }
    },
    async save(list) {
      this.busy = true;
      this.error = "";

      try {
        const r = await this.api.save(this.revision, list);

        if (r.conflict) {
          this.error =
            "The presets have been changed elsewhere and were reloaded. " +
            "Please review your change and save again.";

          await this.load();

          // The edited preset may have moved or been deleted. If it is gone,
          // saving adds it back
          if (this.form && this.form.index >= 0) {
            this.form.index = this.list.findIndex(
              (p) => JSON.stringify(p) === this.form.original,
            );
          }

          return false;
        }

        if (r.error) {
          this.error = r.error;

          return false;
        }

        this.revision = r.data.revision;
        this.list = r.data.presets || [];

        return true;
      } catch (e) {
        this.error = "Unable to save presets: " + e;

        return false;
      } finally {
        this.busy = false;
      }
    },
  },
};
</script>
