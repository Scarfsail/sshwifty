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

import Vue from "vue";
import "./app.css";
import Auth from "./auth.vue";
import { Colors as ControlColors } from "./commands/color.js";
import { Commands } from "./commands/commands.js";
import { Controls } from "./commands/controls.js";
import { Presets } from "./commands/presets.js";
import * as ssh from "./commands/ssh.js";
import * as telnet from "./commands/telnet.js";
import "./common.css";
import * as sshctl from "./control/ssh.js";
import * as telnetctl from "./control/telnet.js";
import * as cipher from "./crypto.js";
import Home from "./home.vue";
import "./landing.css";
import Loading from "./loading.vue";
import { Socket } from "./socket.js";
import * as stream from "./stream/common.js";
import * as xhr from "./xhr.js";

const backendQueryRetryDelay = 2000;

const maxTimeDiff = 30000;

const updateIndicatorMaxDisplayTime = 3000;

const mainTemplate = `
<home
  v-if="page == 'app'"
  :host-path="hostPath"
  :query="query"
  :connection="socket"
  :controls="controls"
  :commands="commands"
  :server-message="serverMessage"
  :preset-data="presetData.presets"
  :restricted-to-presets="presetData.restricted"
  :preset-editing="presetData.editing"
  :presets-api="presetsApi"
  :bypass-clipboard-write-approval="bypassClipboardWriteApproval"
  :skip-preset-prompt-when-all-set="skipPresetPromptWhenAllSet"
  :view-port="viewPort"
  @navigate-to="changeURLHash"
  @tab-opened="tabOpened"
  @tab-closed="tabClosed"
  @tab-updated="tabUpdated"
></home>
<auth
  v-else-if="page == 'auth'"
  :error="authErr"
  @auth="submitAuth"
></auth>
<loading class="app-error-message" v-else :error="loadErr"></loading>
`.trim();

const socksInterface = "/sshwifty/socket";
const socksVerificationInterface = socksInterface + "/verify";
const socksKeyTimeTruncater = 100 * 1000;
const presetsInterface = "/sshwifty/presets";

function allCommands() {
  return [new telnet.Command(), new ssh.Command()];
}

function startApp(rootEl) {
  const pageTitle = document.title;

  let uiControlColors = new ControlColors();

  function getCurrentKeyMixer() {
    return Number(
      Math.trunc(new Date().getTime() / socksKeyTimeTruncater),
    ).toString();
  }

  async function buildSocketKey(privateKey) {
    return new Uint8Array(
      await cipher.hmac512(
        stream.buildBufferFromString(privateKey),
        stream.buildBufferFromString(getCurrentKeyMixer()),
      ),
    ).slice(0, 16);
  }

  new Vue({
    el: rootEl,
    components: {
      loading: Loading,
      auth: Auth,
      home: Home,
    },
    data() {
      return {
        hostPath:
          window.location.protocol +
          "//" +
          window.location.host +
          window.location.pathname,
        query:
          window.location.hash.length > 0 &&
          window.location.hash.indexOf("#") === 0
            ? window.location.hash.slice(1, window.location.hash.length)
            : "",
        page: "loading",
        key: "",
        serverMessage: "",
        presetData: {
          presets: new Presets([]),
          restricted: false,
          editing: false,
        },
        presetsApi: {
          load: this.loadPresets,
          save: this.savePresets,
        },
        bypassClipboardWriteApproval: false,
        skipPresetPromptWhenAllSet: false,
        authErr: "",
        loadErr: "",
        socket: null,
        controls: new Controls([
          new telnetctl.Telnet(uiControlColors),
          new sshctl.SSH(uiControlColors),
        ]),
        commands: new Commands(allCommands()),
        tabUpdateIndicator: null,
        viewPort: {
          dim: {
            width: 0,
            height: 0,
            renew(width, height) {
              this.width = width;
              this.height = height;
            },
          },
        },
        viewPortUpdaters: {
          width: 0,
          height: 0,
          dimResizer: null,
        },
      };
    },
    watch: {
      loadErr() {
        this.isErrored()
          ? document.body.classList.add("app-error")
          : document.body.classList.remove("app-error");
      },
      authErr() {
        this.isErrored()
          ? document.body.classList.add("app-error")
          : document.body.classList.remove("app-error");
      },
    },
    created() {
      // The passphrase of the current session, kept out of the reactive
      // data. Preset editing sends it along with every request
      this.passphrase = "";
    },
    mounted() {
      const self = this;

      self.tryInitialAuth();

      self.viewPortUpdaters.dimResizer = () => {
        self.viewPortUpdaters.height = window.innerHeight;
        self.viewPortUpdaters.width = window.innerWidth;

        self.$nextTick(() => {
          self.viewPort.dim.renew(
            self.viewPortUpdaters.width,
            self.viewPortUpdaters.height,
          );
        });
      };

      window.addEventListener("resize", self.viewPortUpdaters.dimResizer);
    },
    beforeDestroy() {
      window.removeEventListener("resize", self.viewPortUpdaters.dimResizer);
    },
    methods: {
      changeTitleInfo(newTitleInfo) {
        document.title = newTitleInfo + " " + pageTitle;
      },
      resetTitleInfo() {
        document.title = pageTitle;
      },
      changeURLHash(newHash) {
        window.location.hash = newHash;
      },
      isErrored() {
        return this.authErr.length > 0 || this.loadErr.length > 0;
      },
      async getSocketAuthKey(privateKey) {
        const enc = new TextEncoder(),
          rTime = Number(Math.trunc(new Date().getTime() / 100000));

        var finalKey = "";

        if (privateKey.length <= 0) {
          finalKey = "DEFAULT VERIFY KEY";
        } else {
          finalKey = privateKey;
        }

        return new Uint8Array(
          await cipher.hmac512(enc.encode(finalKey), enc.encode(rTime)),
        ).slice(0, 32);
      },
      buildBackendSocketURLs() {
        let r = {
          webSocket: "",
          keepAlive: "",
        };

        switch (location.protocol) {
          case "https:":
            r.webSocket = "wss://";
            break;

          default:
            r.webSocket = "ws://";
        }

        r.webSocket += location.host + socksInterface;
        r.keepAlive = location.protocol + "//" + location.host + socksInterface;

        return r;
      },
      buildSocket(key, dialTimeout, heartbeatInterval) {
        return new Socket(
          this.buildBackendSocketURLs(),
          key,
          dialTimeout * 1000,
          heartbeatInterval * 1000,
        );
      },
      executeHomeApp(authResult, key) {
        let authData = JSON.parse(authResult.data);
        this.serverMessage = authData.server_message
          ? authData.server_message
          : "";
        this.presetData = {
          presets: new Presets(authData.presets ? authData.presets : []),
          restricted: authResult.onlyAllowPresetRemotes,
          editing: authData.preset_editing === true,
        };
        this.bypassClipboardWriteApproval =
          authData.bypass_clipboard_write_approval === true;
        this.skipPresetPromptWhenAllSet =
          authData.skip_preset_prompt_when_all_set === true;
        const enabled = authData.enabled_protocols
          ? authData.enabled_protocols
          : [];
        this.commands = new Commands(
          allCommands().filter((c) => enabled.includes(c.name())),
        );
        this.socket = this.buildSocket(
          key,
          authResult.timeout,
          authResult.heartbeat,
        );
        this.page = "app";
      },
      async doAuth(privateKey) {
        let result = await this.requestAuth(privateKey);

        if (result.key) {
          this.key = result.key;
        }

        return result;
      },
      async authHeaders(privateKey) {
        const authKey = await this.getSocketAuthKey(privateKey);

        return { "X-Key": btoa(String.fromCharCode.apply(null, authKey)) };
      },
      async requestAuth(privateKey) {
        let h = await xhr.get(
          socksVerificationInterface,
          !privateKey || !this.key
            ? { "X-Key": "" }
            : await this.authHeaders(privateKey),
        );

        let serverDate = h.getResponseHeader("Date");

        return {
          result: h.status,
          key: h.getResponseHeader("X-Key"),
          timeout: h.getResponseHeader("X-Timeout"),
          heartbeat: h.getResponseHeader("X-Heartbeat"),
          date: serverDate ? new Date(serverDate) : null,
          data: h.responseText,
          onlyAllowPresetRemotes:
            h.getResponseHeader("X-OnlyAllowPresetRemotes") === "yes",
        };
      },
      async tryInitialAuth() {
        try {
          let result = await this.doAuth("");

          if (result.date) {
            let serverRespondTime = result.date,
              serverRespondTimestamp = serverRespondTime.getTime(),
              clientCurrent = new Date(),
              clientTimestamp = clientCurrent.getTime(),
              timeDiff = Math.abs(serverRespondTimestamp - clientTimestamp);

            if (timeDiff > maxTimeDiff) {
              this.loadErr =
                "The datetime difference between current client " +
                "and the Sshwifty server is beyond the operational tolerance." +
                "\r\n\r\n" +
                "The server time was " +
                serverRespondTime +
                ", and the client time was " +
                clientCurrent +
                ", resulted a " +
                timeDiff +
                "ms time difference, exceeding the " +
                "limitation of " +
                maxTimeDiff +
                "ms.\r\n\r\n" +
                "Try reload the page, see if the problem persists. And if " +
                "it did, please make sure both the server and the client are " +
                "having the correct time settings";

              return;
            }
          }

          let self = this;
          switch (result.result) {
            case 200:
              this.executeHomeApp(result, {
                async fetch() {
                  let result = await self.doAuth("");

                  if (result.result !== 200) {
                    throw new Error(
                      "Unable to fetch key from remote, unexpected " +
                        "error code: " +
                        result.result,
                    );
                  }

                  return await buildSocketKey(atob(result.key) + "+");
                },
              });
              break;

            case 403:
              this.page = "auth";
              break;

            case 0:
              setTimeout(() => {
                this.tryInitialAuth();
              }, backendQueryRetryDelay);
              break;

            default:
              alert("Unexpected backend query status: " + result.result);
          }
        } catch (e) {
          this.loadErr = "Unable to initialize client application: " + e;
        }
      },
      async submitAuth(passphrase) {
        this.authErr = "";

        try {
          let result = await this.doAuth(passphrase);

          let self = this;
          switch (result.result) {
            case 200:
              this.passphrase = passphrase;
              this.executeHomeApp(result, {
                async fetch() {
                  let result = await self.doAuth(passphrase);

                  if (result.result !== 200) {
                    throw new Error(
                      "Unable to fetch key from remote, unexpected " +
                        "error code: " +
                        result.result,
                    );
                  }

                  return await buildSocketKey(
                    atob(result.key) + "+" + passphrase,
                  );
                },
              });
              break;

            case 403:
              this.authErr = "Authentication has failed. Wrong passphrase?";
              break;

            default:
              this.authErr =
                "Unexpected backend query status: " + result.result;
          }
        } catch (e) {
          this.authErr = "Unable to authenticate: " + e;
        }
      },
      async loadPresets() {
        const h = await xhr.get(
          presetsInterface,
          await this.authHeaders(this.passphrase),
        );

        if (h.status !== 200) {
          throw new Error("Unable to load presets, error code: " + h.status);
        }

        return JSON.parse(h.responseText);
      },
      // savePresets replaces all presets. It resolves to the new presets and
      // revision on success, and to the reason (conflict or error) otherwise
      async savePresets(revision, presets) {
        const headers = await this.authHeaders(this.passphrase);

        headers["Content-Type"] = "application/json";

        const h = await xhr.put(
          presetsInterface,
          headers,
          JSON.stringify({ revision: revision, presets: presets }),
        );

        switch (h.status) {
          case 200:
            break;

          case 409:
            return { conflict: true };

          case 400:
            return { error: h.responseText };

          default:
            return { error: "Unexpected backend status: " + h.status };
        }

        // Reload through the verify interface, the one place presets are
        // parsed from. The save is done already, so a failed reload must not
        // report it as failed
        try {
          const result = await this.doAuth(this.passphrase);

          if (result.result === 200) {
            const authData = JSON.parse(result.data);

            this.presetData.presets = new Presets(
              authData.presets ? authData.presets : [],
            );
          }
        } catch (e) {
          // Shown after the next reload
        }

        return { data: JSON.parse(h.responseText) };
      },
      updateTabTitleInfo(tabs, updated) {
        if (tabs.length <= 0) {
          this.resetTitleInfo();

          return;
        }

        this.changeTitleInfo("(" + tabs.length + (updated ? "*" : "") + ")");
      },
      tabOpened(tabs) {
        this.tabUpdated(tabs);
      },
      tabClosed(tabs) {
        if (tabs.length > 0) {
          this.updateTabTitleInfo(tabs, this.tabUpdateIndicator !== null);

          return;
        }

        if (this.tabUpdateIndicator) {
          clearTimeout(this.tabUpdateIndicator);
          this.tabUpdateIndicator = null;
        }

        this.updateTabTitleInfo(tabs, false);
      },
      tabUpdated(tabs) {
        if (this.tabUpdateIndicator) {
          clearTimeout(this.tabUpdateIndicator);
          this.tabUpdateIndicator = null;
        }

        this.updateTabTitleInfo(tabs, true);

        this.tabUpdateIndicator = setTimeout(() => {
          this.tabUpdateIndicator = null;
          this.updateTabTitleInfo(tabs, false);
        }, updateIndicatorMaxDisplayTime);
      },
    },
  });
}

function initializeClient() {
  let landingRoot = document.getElementById("landing");

  if (!landingRoot) {
    return;
  }

  if (process.env.NODE_ENV === "development") {
    console.log("Currently in Development environment");
  }

  window.addEventListener("unhandledrejection", function (e) {
    console.error("Error:", e);
  });

  window.addEventListener("error", function (e) {
    console.error("Error:", e);
  });

  landingRoot.parentNode.removeChild(landingRoot);

  let normalRoot = document.createElement("div");
  normalRoot.setAttribute("id", "app");
  normalRoot.innerHTML = mainTemplate;

  document.body.insertBefore(normalRoot, document.body.firstChild);

  startApp(normalRoot);
}

window.addEventListener("load", initializeClient);
document.addEventListener("load", initializeClient);
document.addEventListener("DOMContentLoaded", initializeClient);
