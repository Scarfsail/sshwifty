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
  <div id="home">
    <header id="home-header">
      <h1 id="home-hd-title">Sshwifty</h1>

      <a id="home-hd-delay" href="javascript:;" @click="showDelayWindow">
        <span
          id="home-hd-delay-icon"
          class="icon icon-point1"
          :class="socket.classStyle"
        ></span>
        <span v-if="socket.message.length > 0" id="home-hd-delay-value">{{
          socket.message
        }}</span>
      </a>

      <a
        id="home-hd-plus"
        class="icon icon-plus1"
        href="javascript:;"
        :class="{
          working: connector.inputting,
          intensify: connector.inputting && !windows.connect,
        }"
        @click="showConnectWindow"
      ></a>

      <tabs
        id="home-hd-tabs"
        :tab="tab.current"
        :tabs="tab.tabs"
        tabs-class="tab1"
        list-trigger-class="icon icon-more1"
        @current="switchTab"
        @retap="retapTab"
        @list="showTabsWindow"
        @close="closeTab"
      ></tabs>
    </header>

    <screens
      id="home-content"
      :screen="tab.current"
      :screens="tab.tabs"
      :view-port="viewPort"
      :bypass-clipboard-write-approval="bypassClipboardWriteApproval"
      @indicated="tabIndicate"
      @indicationDismissed="tabDismissIndicator"
      @updated="tabUpdated"
      @stopped="tabStopped"
    >
      <div id="home-content-wrap">
        <h1>Hi, this is Sshwifty</h1>

        <p>
          An Open Source Web SSH Client that enables you to connect to SSH
          servers without downloading any additional software.
        </p>

        <p>
          To get started, click the
          <span
            id="home-content-connect"
            class="icon icon-plus1"
            @click="showConnectWindow"
          ></span>
          icon near the top left corner.
        </p>

        <div v-if="serverMessage.length > 0">
          <hr />
          <p class="secondary" v-html="serverMessage"></p>
        </div>
      </div>
    </screens>

    <connect-widget
      :inputting="connector.inputting"
      :display="windows.connect"
      :connectors="connector.connectors"
      :presets="presets"
      :restricted-to-presets="restrictedToPresets"
      :preset-editing="presetEditing"
      :presets-api="presetsApi"
      :knowns="enabledKnowns"
      :knowns-launcher-builder="buildknownLauncher"
      :knowns-export="exportKnowns"
      :knowns-import="importKnowns"
      :busy="connector.busy"
      @display="windows.connect = $event"
      @close="connectWindowClosed"
      @connector-select="connectNew"
      @known-select="connectKnown"
      @known-remove="removeKnown"
      @preset-select="connectPreset"
      @known-clear-session="clearSessionKnown"
    >
      <connector
        :connector="connector.connector"
        @cancel="cancelConnection"
        @done="connectionSucceed"
      >
      </connector>
    </connect-widget>
    <status-widget
      :class="socket.windowClass"
      :display="windows.delay"
      :status="socket.status"
      @display="windows.delay = $event"
    ></status-widget>
    <tab-window
      :tab="tab.current"
      :tabs="tab.tabs"
      :display="windows.tabs"
      tabs-class="tab1 tab1-list"
      @display="windows.tabs = $event"
      @current="switchTab"
      @retap="retapTab"
      @close="closeTab"
    ></tab-window>
  </div>
</template>

<script>
import "./home.css";

import ConnectWidget from "./widgets/connect.vue";
import StatusWidget from "./widgets/status.vue";
import Connector from "./widgets/connector.vue";
import Tabs from "./widgets/tabs.vue";
import TabWindow from "./widgets/tab_window.vue";
import Screens from "./widgets/screens.vue";
import {
  Indicators as ScreenIndicators,
  Indicator as ScreenIndicator,
  Action as ScreenIndicatorAction,
} from "./widgets/screen_indicator.vue";

import * as home_socket from "./home_socketctl.js";
import * as home_history from "./home_historyctl.js";

import * as presets from "./commands/presets.js";

const BACKEND_CONNECT_ERROR =
  "Unable to connect to the Sshwifty backend server: ";
const BACKEND_REQUEST_ERROR = "Unable to perform request: ";

const INDICATOR_STOPPED = "STOPPED";
const INDICATOR_RECONNECT_FAILED = "RECONNECT_FAILED";

export default {
  components: {
    "connect-widget": ConnectWidget,
    "status-widget": StatusWidget,
    connector: Connector,
    tabs: Tabs,
    "tab-window": TabWindow,
    screens: Screens,
  },
  props: {
    hostPath: {
      type: String,
      default: "",
    },
    query: {
      type: String,
      default: "",
    },
    connection: {
      type: Object,
      default: () => null,
    },
    controls: {
      type: Object,
      default: () => null,
    },
    commands: {
      type: Object,
      default: () => null,
    },
    serverMessage: {
      type: String,
      default: "",
    },
    presetData: {
      type: Object,
      default: () => new presets.Presets([]),
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
    bypassClipboardWriteApproval: {
      type: Boolean,
      default: () => false,
    },
    skipPresetPromptWhenAllSet: {
      type: Boolean,
      default: () => false,
    },
    viewPort: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    let history = home_history.build(this);
    return {
      ticker: null,
      windows: {
        delay: false,
        connect: false,
        tabs: false,
      },
      socket: home_socket.build(this),
      connector: {
        historyRec: history,
        connector: null,
        connectors: this.commands.all(),
        inputting: false,
        acquired: false,
        busy: false,
        reconnectTabID: null,
        knowns: history.all(),
      },
      presets: this.commands.mergePresets(this.presetData),
      tab: {
        current: -1,
        lastID: 0,
        tabs: [],
      },
    };
  },
  computed: {
    // connectorBusy reports whether the single shared connector slot is
    // taken, either by a live wizard or by a stream acquisition that has not
    // produced one yet. A reconnect can only start when it is free
    connectorBusy() {
      return this.connector.inputting || this.connector.acquired;
    },
    // enabledKnowns hides history records of protocols the server has not
    // enabled. The records are kept, and reappear once re-enabled
    enabledKnowns() {
      return this.connector.knowns.filter(
        (k) => this.getConnectorByType(k.type) !== null,
      );
    },
  },
  watch: {
    // Presets are replaced after they have been edited
    presetData(newVal) {
      this.presets = this.commands.mergePresets(newVal);
    },
  },
  mounted() {
    this.ticker = setInterval(() => {
      this.tick();
    }, 1000);
    if (this.query.length > 1 && this.query.indexOf("+") === 0) {
      this.connectLaunch(this.query.slice(1, this.query.length), (success) => {
        if (!success) {
          return;
        }
        this.$emit("navigate-to", "");
      });
    } else {
      // Landing on an empty page, the next thing to do is always to pick a
      // remote, so open the window as though the + button had been clicked
      this.showConnectWindow();
    }
    window.addEventListener("beforeunload", this.onBrowserClose);
  },
  beforeDestroy() {
    window.removeEventListener("beforeunload", this.onBrowserClose);

    if (this.ticker === null) {
      clearInterval(this.ticker);
      this.ticker = null;
    }
  },
  methods: {
    onBrowserClose(e) {
      if (this.tab.current < 0) {
        return undefined;
      }
      const msg = "Some tabs are still open, are you sure you want to exit?";
      (e || window.event).returnValue = msg;
      return msg;
    },
    tick() {
      let now = new Date();
      this.socket.update(now, this);
    },
    closeAllWindow(e) {
      for (let i in this.windows) {
        this.windows[i] = false;
      }
    },
    showDelayWindow() {
      this.closeAllWindow();
      this.windows.delay = true;
    },
    showConnectWindow() {
      this.closeAllWindow();
      this.windows.connect = true;
    },
    connectWindowClosed() {
      if (this.connector.reconnectTabID === null) {
        return;
      }

      // The window has been dismissed while a reconnect was still pending or
      // running. Give it up, keeping the dead tab as it is so it can be
      // retried. Dropping the claim makes both the wizard's own late
      // cancellation callback and a still in-flight stream acquisition
      // no-ops, since neither owns the reconnect any more
      this.reconnectGiveUp(this.connector.reconnectTabID);
    },
    showTabsWindow() {
      this.closeAllWindow();
      this.windows.tabs = true;
    },
    async getStreamThenRun(run, end) {
      let errStr = null;
      try {
        let conn = await this.connection.get(this.socket);
        try {
          run(conn);
        } catch (e) {
          errStr = BACKEND_REQUEST_ERROR + e;
          process.env.NODE_ENV === "development" && console.trace(e);
        }
      } catch (e) {
        errStr = BACKEND_CONNECT_ERROR + e;
        process.env.NODE_ENV === "development" && console.trace(e);
      }
      end();
      if (errStr !== null) {
        alert(errStr);
      }
    },
    // runConnect hands a backend stream to callback. When no stream can be
    // delivered - the connector is already taken, or the backend cannot be
    // reached - failed is called instead, so the caller can undo whatever it
    // set up in anticipation
    runConnect(callback, failed = () => {}) {
      if (this.connector.acquired) {
        failed();

        return;
      }
      this.connector.acquired = true;
      this.connector.busy = true;
      let delivered = false;
      this.getStreamThenRun(
        (stream) => {
          this.connector.busy = false;
          callback(stream);

          // Only after callback has returned, so that a throw out of it
          // (getStreamThenRun catches those) still counts as undelivered
          // and lets the caller undo its own setup
          delivered = true;
        },
        () => {
          this.connector.busy = false;
          this.connector.acquired = false;

          if (!delivered) {
            failed();
          }
        },
      );
    },
    connectNew(connector) {
      const self = this;
      self.runConnect((stream) => {
        self.connector.connector = {
          id: connector.id(),
          name: connector.name(),
          description: connector.description(),
          wizard: connector.wizard(
            stream,
            self.controls,
            self.connector.historyRec,
            presets.emptyPreset(),
            null,
            false,
            () => {},
          ),
        };
        self.connector.inputting = true;
      });
    },
    connectPreset(preset) {
      const self = this;
      self.runConnect((stream) => {
        self.connector.connector = {
          id: preset.command.id(),
          name: preset.command.name(),
          description: preset.command.description(),
          wizard: preset.command.wizard(
            stream,
            self.controls,
            self.connector.historyRec,
            preset.preset,
            null,
            [],
            () => {},
            self.skipPresetPromptWhenAllSet,
          ),
        };
        self.connector.inputting = true;
      });
    },
    getConnectorByType(type) {
      let connector = null;
      for (let c in this.connector.connectors) {
        if (this.connector.connectors[c].name() !== type) {
          continue;
        }
        connector = this.connector.connectors[c];
      }
      return connector;
    },
    connectKnown(known) {
      const self = this;
      self.runConnect((stream) => {
        let connector = self.getConnectorByType(known.type);
        if (!connector) {
          alert("Unknown connector: " + known.type);
          self.connector.inputting = false;
          return;
        }
        self.connector.connector = {
          id: connector.id(),
          name: connector.name(),
          description: connector.description(),
          wizard: connector.execute(
            stream,
            self.controls,
            self.connector.historyRec,
            known.data,
            known.session,
            known.keptSessions,
            () => {
              self.connector.knowns = self.connector.historyRec.all();
            },
          ),
        };
        self.connector.inputting = true;
      });
    },
    parseConnectLauncher(ll) {
      let llSeparatorIdx = ll.indexOf(":");
      // Type must contain at least one charater
      if (llSeparatorIdx <= 0) {
        throw new Error("Invalid Launcher string");
      }
      return {
        type: ll.slice(0, llSeparatorIdx),
        query: ll.slice(llSeparatorIdx + 1, ll.length),
      };
    },
    connectLaunch(launcher, done) {
      this.showConnectWindow();
      this.runConnect((stream) => {
        let ll = this.parseConnectLauncher(launcher),
          connector = this.getConnectorByType(ll.type);
        if (!connector) {
          alert("Unknown connector: " + ll.type);
          this.connector.inputting = false;
          return;
        }
        const self = this;
        this.connector.connector = {
          id: connector.id(),
          name: connector.name(),
          description: connector.description(),
          wizard: connector.launch(
            stream,
            this.controls,
            this.connector.historyRec,
            ll.query,
            (n) => {
              self.connector.knowns = self.connector.historyRec.all();

              done(n.data().success);
            },
          ),
        };
        this.connector.inputting = true;
      });
    },
    buildknownLauncher(known) {
      let connector = this.getConnectorByType(known.type);
      if (!connector) {
        return;
      }
      return this.hostPath + "#+" + connector.launcher(known.data);
    },
    exportKnowns() {
      return this.connector.historyRec.export();
    },
    importKnowns(d) {
      this.connector.historyRec.import(d);
      this.connector.knowns = this.connector.historyRec.all();
    },
    removeKnown(uid) {
      this.connector.historyRec.del(uid);
      this.connector.knowns = this.connector.historyRec.all();
    },
    clearSessionKnown(uid) {
      this.connector.historyRec.clearSession(uid);
      this.connector.knowns = this.connector.historyRec.all();
    },
    tabIndexByID(id) {
      for (let i = 0; i < this.tab.tabs.length; i++) {
        if (this.tab.tabs[i].id !== id) {
          continue;
        }

        return i;
      }

      return -1;
    },
    reconnectTab(index) {
      if (this.connectorBusy) {
        return;
      }

      const self = this,
        tab = this.tab.tabs[index];

      if (!tab.reconnect) {
        return;
      }

      const recIdx = this.connector.historyRec.indexOf(tab.reconnect.uname);

      if (recIdx < 0) {
        this.tabIndicate(
          index,
          new ScreenIndicator(
            INDICATOR_RECONNECT_FAILED,
            "Unable to reconnect: the record of this remote is no longer " +
              "available",
            "error",
            [],
          ),
        );

        return;
      }

      const known = this.connector.historyRec.all()[recIdx],
        reconnectTabID = tab.id;

      this.showConnectWindow();

      // Claim the tab now rather than once the stream arrives: getting the
      // stream can await a re-dial of the backend socket, and closing the
      // window during that wait has to be able to call the reconnect off
      this.connector.reconnectTabID = reconnectTabID;

      this.runConnect(
        (stream) => {
          if (self.connector.reconnectTabID !== reconnectTabID) {
            return;
          }

          let connector = self.getConnectorByType(known.type);

          if (!connector) {
            alert("Unknown connector: " + known.type);

            self.reconnectGiveUp(reconnectTabID);

            return;
          }

          self.connector.connector = {
            id: connector.id(),
            name: connector.name(),
            description: connector.description(),
            wizard: connector.execute(
              stream,
              self.controls,
              self.connector.historyRec,
              known.data,
              known.session,
              known.keptSessions,
              (n) => {
                self.connector.knowns = self.connector.historyRec.all();

                if (self.connector.reconnectTabID !== reconnectTabID) {
                  return;
                }

                if (n.data().success) {
                  return;
                }

                self.reconnectFailed(reconnectTabID, n.data());
              },
            ),
          };

          self.connector.inputting = true;
        },
        () => {
          // No stream was delivered, so no wizard will ever run. Drop the
          // claim, otherwise the next successful connect of any kind would
          // replace this tab instead of opening its own
          self.reconnectGiveUp(reconnectTabID);
        },
      );
    },
    reconnectGiveUp(reconnectTabID) {
      if (this.connector.reconnectTabID !== reconnectTabID) {
        return;
      }

      this.connector.reconnectTabID = null;
      this.connector.inputting = false;

      // Stop the wizard right here instead of leaving it to the <connector>
      // unmount, which only happens on the next render tick: a wizard that
      // succeeded in between would arrive at connectionSucceed with nothing
      // left to say it was given up, and land as a brand new tab
      if (this.connector.connector && this.connector.connector.wizard) {
        this.connector.connector.wizard.close();
      }
    },
    reconnectFailed(reconnectTabID, data) {
      const index = this.tabIndexByID(reconnectTabID);

      if (index >= 0) {
        this.tabIndicate(
          index,
          new ScreenIndicator(
            INDICATOR_RECONNECT_FAILED,
            data.errorTitle + ": " + data.errorMessage,
            "error",
            [],
          ),
        );
      }

      this.connector.reconnectTabID = null;
      this.connector.inputting = false;
      this.windows.connect = false;
    },
    cancelConnection() {
      this.connector.reconnectTabID = null;
      this.connector.inputting = false;
      this.connector.acquired = false;
    },
    connectionSucceed(data) {
      const replaceIndex =
        this.connector.reconnectTabID !== null
          ? this.tabIndexByID(this.connector.reconnectTabID)
          : -1;

      this.connector.reconnectTabID = null;
      this.connector.inputting = false;
      this.connector.acquired = false;
      this.windows.connect = false;

      this.addToTab(data, replaceIndex);

      this.$emit("tab-opened", this.tab.tabs);
    },
    async addToTab(data, replaceIndex = -1) {
      const newTab = {
        id: this.tab.lastID++,
        name: data.name,
        info: data.info,
        control: data.control,
        ui: data.ui,
        reconnect: data.reconnect,
        toolbar: false,
        indicators: new ScreenIndicators(),
        updated: false,
        status: {
          closing: false,
        },
      };

      if (replaceIndex >= 0) {
        const replaced = this.tab.tabs[replaceIndex],
          replacedID = replaced.id;

        // The error indicator is raised by any failure of the screen's read
        // loop, not only by a remote termination, so the replaced session
        // can still be running. Shut that one down the way closeTab does,
        // to not orphan it.
        //
        // A session that did finish must not be closed again: its stream has
        // already been returned to the pool and the new session has most
        // likely been given the very same stream ID, so the close signal -
        // which the old control still sends, stamped with that ID - would
        // terminate the connection just established
        if (!replaced.control.closed) {
          try {
            replaced.control.disabled();

            await replaced.control.close();
          } catch (e) {
            process.env.NODE_ENV === "development" && console.trace(e);
          }
        }

        // Shutting the old session down is awaited, and by then the user may
        // have closed its tab, shifting everything after it. Resolve the
        // target again instead of trusting the index taken before the wait,
        // and fall back to appending when it is gone, rather than pushing
        // the new session into whichever tab now sits at that index
        const replaceAt = this.tabIndexByID(replacedID);

        if (replaceAt >= 0) {
          this.tab.tabs.splice(replaceAt, 1, newTab);

          await this.switchTab(replaceAt);

          return;
        }
      }

      await this.switchTab(this.tab.tabs.push(newTab) - 1);
    },
    removeFromTab(index) {
      let isLast = index === this.tab.tabs.length - 1;
      this.tab.tabs.splice(index, 1);
      this.tab.current = isLast ? this.tab.tabs.length - 1 : index;
    },
    async switchTab(to) {
      if (this.tab.current >= 0) {
        await this.tab.tabs[this.tab.current].control.disabled();
      }
      this.tab.current = to;
      this.tab.tabs[this.tab.current].updated = false;
      await this.tab.tabs[this.tab.current].control.enabled();
    },
    async retapTab(tab) {
      this.tab.tabs[tab].toolbar = !this.tab.tabs[tab].toolbar;
      await this.tab.tabs[tab].control.retap(this.tab.tabs[tab].toolbar);
    },
    async closeTab(index) {
      if (this.tab.tabs[index].status.closing) {
        return;
      }

      // Closing the tab a reconnect was claiming means the user no longer
      // wants that session. Give the reconnect up, otherwise it would come
      // back as a brand new tab once it succeeded
      if (this.connector.reconnectTabID === this.tab.tabs[index].id) {
        this.reconnectGiveUp(this.tab.tabs[index].id);
        this.windows.connect = false;
      }

      this.tab.tabs[index].status.closing = true;
      try {
        this.tab.tabs[index].control.disabled();
        await this.tab.tabs[index].control.close();
      } catch (e) {
        alert("Cannot close tab due to error: " + e);
        process.env.NODE_ENV === "development" && console.trace(e);
      }
      this.removeFromTab(index);
      this.$emit("tab-closed", this.tab.tabs);
    },
    clearIndicator(index) {
      this.tab.tabs[index].indicators.clear();
    },
    tabIndicate(index, indicator) {
      this.tab.tabs[index].indicators.append(indicator);
    },
    tabDismissIndicator(index, uid) {
      this.tab.tabs[index].indicators.dismiss(uid);
    },
    tabUpdated(index) {
      this.$emit("tab-updated", this.tab.tabs);
      this.tab.tabs[index].updated = index !== this.tab.current;
    },
    tabStopped(index, reason) {
      this.clearIndicator(index);
      this.tabIndicate(
        index,
        new ScreenIndicator(
          INDICATOR_STOPPED,
          "" + reason,
          "error",
          this.buildReconnectActions(this.tab.tabs[index]),
        ),
      );
    },
    // buildReconnectActions offers a reconnect for a session that carries a
    // history reference to dial again. A session without one - a command
    // that keeps no history - is given no action at all
    buildReconnectActions(tab) {
      if (!tab.reconnect) {
        return [];
      }

      const self = this,
        tabID = tab.id;

      return [
        new ScreenIndicatorAction("Reconnect", (uid, nonCancel) => {
          // The indicator is gone rather than clicked: it has been replaced
          // or cleared, and the reconnect it offered is no longer on offer
          if (!nonCancel) {
            return;
          }

          // The tab may have moved while the indicator was displayed, so
          // resolve it again instead of trusting the index it was built with
          const index = self.tabIndexByID(tabID);

          if (index < 0) {
            return;
          }

          self.reconnectTab(index);
        }),
      ];
    },
  },
};
</script>
