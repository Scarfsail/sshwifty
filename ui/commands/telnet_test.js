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

import assert from "assert";
import * as presets from "./presets.js";
import * as telnet from "./telnet.js";

describe("Telnet", () => {
  const cmd = new telnet.Command();

  it("presetFromKnown builds a preset that matches the record", () => {
    const k = {
      type: "Telnet",
      title: "192.168.0.41",
      data: { host: "192.168.0.41", charset: "ibm866" },
    };
    const p = cmd.presetFromKnown(k);
    assert.deepStrictEqual(p, {
      Title: "192.168.0.41",
      Type: "Telnet",
      Host: "192.168.0.41",
      TabColor: "",
      Meta: { Encoding: "ibm866" },
    });
    const loaded = cmd.represet(
      new presets.Preset({
        title: p.Title,
        type: p.Type,
        host: p.Host,
        tab_color: p.TabColor,
        meta: p.Meta,
      }),
    );
    assert.strictEqual(cmd.matchesKnown(loaded, k), true);
  });

  it("presetFromKnown leaves out an empty charset", () => {
    const p = cmd.presetFromKnown({
      type: "Telnet",
      title: "192.168.0.41",
      data: { host: "192.168.0.41" },
    });
    assert.deepStrictEqual(p.Meta, {});
  });
});
