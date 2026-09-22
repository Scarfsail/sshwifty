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
  <main :class="{ active: screens.length > 0 }">
    <slot v-if="screens.length <= 0"></slot>

    <div
      v-for="(screenInfo, idx) in screens"
      :key="screenInfo.id"
      :class="{ 'screen-inactive': screen !== idx }"
      class="screen"
    >
      <h1 style="display: none">Main Interface</h1>

      <div
        v-if="screenInfo.indicators.length() > 0"
        class="screen-errors"
      >
        <ScreenIndicator
          v-for="(indicator, idx) in screenInfo.indicators.indicators"
          :key="indicator.uid"
          :indicator="indicator"
        />
      </div>

      <div class="screen-screen">
        <component
          :is="getComponent(screenInfo.ui)"
          :active="screen === idx"
          :control="screenInfo.control"
          :change="screenInfo.indicators"
          :toolbar="screenInfo.toolbar"
          :view-port="viewPort"
          :style="'background-color: ' + screenInfo.control.color()"
          class="screen-content"
          @indicated="indicated(idx, $event)"
          @indicationDismissed="indicationDismissed(idx, $event)"
          @updated="updated(idx)"
          @stopped="stopped(idx, $event)"
        ></component>
      </div>
    </div>
  </main>
</template>

<script>
import ScreenIndicator from "./screen_indicator.vue";
import ConsoleScreen from "./screen_console.vue";

import "./screens.css";

export default {
  components: {
    ScreenIndicator,
    ConsoleScreen,
  },
  props: {
    screen: {
      type: Number,
      default: 0,
    },
    screens: {
      type: Array,
      default: () => [],
    },
    viewPort: {
      type: Object,
      default: () => {},
    },
  },
  methods: {
    getComponent(ui) {
      switch (ui) {
        case "Console":
          return "ConsoleScreen";
        default:
          throw new Error("Unknown UI: " + ui);
      }
    },
    indicated(index, indicator) {
      this.$emit("indicated", index, indicator);
    },
    indicationDismissed(index, indicatorUID) {
      this.$emit("indicationDismissed", index, indicatorUID);
    },
    updated(index) {
      this.$emit("updated", index);
    },
    stopped(index, reason) {
      this.$emit("stopped", index, reason);
    },
  },
};
</script>
