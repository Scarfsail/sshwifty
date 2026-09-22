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
  <div
    v-if="indicator.message.length > 0"
    class="screen-error"
    :class="[ 'screen-error-level-' + indicator.level ]"
  >
    {{ indicator.message }}
    <ul
      v-if="indicator.actions && indicator.actions.length > 0"
      class="screen-error-options lst-nostyle"
    >
      <li
        v-for="(action, aIdx) in indicator.actions"
        :key="aIdx"
        class="screen-error-option"
        @click="click(action, indicator.uid)"
      >
        {{ actionTitle(action) }}
      </li>
    </ul>
  </div>
</template>

<script>
import "./screen_indicator.css";

export class Action {
  constructor(title, action) {
    this.title = title;
    this.action = action;
  }

  do(uid, nonCancel) {
    this.action(uid, nonCancel);
  }
}

export class Indicator {
  constructor(uid, message, level, actions) {
    this.uid = uid;
    this.message = message;
    this.level = level;
    this.actions = actions && actions.length > 0 ? actions : [];
  }

  cancelAllActions() {
    for (var i in this.actions) {
      this.actions[i].do(this.uid, false);
    }
  }
}

export class Indicators {
  constructor() {
    this.indicators = [];
  }

  dismiss(uid) {
    this.indicators = this.indicators.filter((item) => uid !== item.uid);
  }

  length() {
    return this.indicators.length;
  }

  levelMoreCritical(level1, level2) {
    if (!level1) {
      return level2;
    } else if (!level2) {
      return level1;
    }
    if (level1 === level2) {
      return level1;
    }
    if (level1 == "info") {
      return level2;
    }
    if (level1 == "warning" && level2 == "error") {
      return level2;
    }
    return level1;
  }

  level() {
    let level = "";
    for (var i in this.indicators) {
      level = this.levelMoreCritical(level, this.indicators[i].level);
    }
    return level;
  }

  append(indicator) {
    for (var i in this.indicators) {
      if (this.indicators[i].uid !== indicator.uid) {
        continue;
      }
      this.indicators[i].cancelAllActions();
    }
    this.dismiss(indicator.uid);
    this.indicators.push(indicator);
  }

  clear() {
    for (var i in this.indicators) {
      this.indicators[i].cancelAllActions();
    }
    this.indicators = [];
  }
}

export default {
  props: {
    indicator: {
      type: Indicator,
      default: () => {
        return new Indicator("", "", "", []);
      },
    },
  },
  methods: {
    actionTitle(actions) {
      return actions.title;
    },
    click(actions, uid) {
      return actions.do(uid, true);
    },
  },
}
</script>