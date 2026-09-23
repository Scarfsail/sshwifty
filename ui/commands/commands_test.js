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
import * as commands from "./commands.js";
import * as presets from "./presets.js";
import * as ssh from "./ssh.js";
import * as telnet from "./telnet.js";

describe("Commands", () => {
  const cmds = new commands.Commands([new ssh.Command(), new telnet.Command()]);

  const preset = (type, title, host, meta) => ({
    title: title,
    type: type,
    host: host,
    tab_color: "",
    meta: meta,
  });

  // Like History.all(): records are appended as they're used, so each one is
  // newer than the one before, unless last is given
  let lastUsed = 0;
  const known = (uid, type, host, user, last) => ({
    uid: uid,
    type: type,
    last: new Date(last === undefined ? ++lastUsed : last),
    data: { host: host, user: user },
  });

  const uids = (ks) => ks.map((k) => k.uid);

  it("groupKnowns attaches matched records to presets, newest first", () => {
    const ps = cmds.mergePresets(
      new presets.Presets([
        preset("SSH", "nas", "192.168.0.41", { User: "root" }),
        preset("Telnet", "router", "192.168.0.1", {}),
      ]),
    );
    const knowns = [
      known("a", "SSH", "192.168.0.41", "root"),
      known("b", "Telnet", "192.168.0.1:23", ""),
      known("c", "SSH", "192.168.0.41:22", "root"),
      known("d", "SSH", "192.168.0.41", "admin"),
      known("e", "SSH", "192.168.0.99", "root"),
    ];
    const g = commands.groupKnowns(ps, knowns);

    assert.deepStrictEqual(
      g.presets.map((p) => p.preset.preset.title()),
      ["nas", "router"],
    );
    assert.deepStrictEqual(uids(g.presets[0].knowns), ["c", "a"]);
    assert.deepStrictEqual(uids(g.presets[1].knowns), ["b"]);
    assert.deepStrictEqual(uids(g.others), ["e", "d"]);
  });

  it("groupKnowns lets a host-only preset absorb every user", () => {
    const ps = cmds.mergePresets(
      new presets.Presets([preset("SSH", "nas", "192.168.0.41", {})]),
    );
    const knowns = [
      known("a", "SSH", "192.168.0.41", "root"),
      known("b", "SSH", "192.168.0.41", "admin"),
    ];
    const g = commands.groupKnowns(ps, knowns);

    assert.deepStrictEqual(uids(g.presets[0].knowns), ["b", "a"]);
    assert.deepStrictEqual(g.others, []);
  });

  it("groupKnowns attaches a record to every preset it matches", () => {
    const ps = cmds.mergePresets(
      new presets.Presets([
        preset("SSH", "nas", "192.168.0.41", {}),
        preset("SSH", "nas-root", "192.168.0.41", { User: "root" }),
      ]),
    );
    const g = commands.groupKnowns(ps, [
      known("a", "SSH", "192.168.0.41", "root"),
    ]);

    assert.deepStrictEqual(uids(g.presets[0].knowns), ["a"]);
    assert.deepStrictEqual(uids(g.presets[1].knowns), ["a"]);
    assert.deepStrictEqual(g.others, []);
  });

  it("groupKnowns keeps presets without records and records without presets", () => {
    const ps = cmds.mergePresets(
      new presets.Presets([preset("SSH", "empty", "", {})]),
    );
    const g = commands.groupKnowns(ps, [
      known("a", "SSH", "192.168.0.41", "root"),
    ]);

    assert.deepStrictEqual(uids(g.presets[0].knowns), []);
    assert.deepStrictEqual(uids(g.others), ["a"]);
  });

  it("groupKnowns orders by last use, not by array order", () => {
    const ps = cmds.mergePresets(
      new presets.Presets([preset("SSH", "nas", "192.168.0.41", {})]),
    );
    // Imported records are appended to the end, whatever their age
    const g = commands.groupKnowns(ps, [
      known("a", "SSH", "192.168.0.41", "root", 2000),
      known("b", "SSH", "192.168.0.99", "root", 3000),
      known("c", "SSH", "192.168.0.41", "admin", 1000),
      known("d", "SSH", "192.168.0.98", "root", 500),
    ]);

    assert.deepStrictEqual(uids(g.presets[0].knowns), ["a", "c"]);
    assert.deepStrictEqual(uids(g.others), ["b", "d"]);
  });
});
