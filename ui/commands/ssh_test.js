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
import * as ssh from "./ssh.js";

describe("SSH", () => {
  const rec = (type, host, fingerprint, last) => ({
    type: type,
    last: new Date(last),
    data: { host: host, fingerprint: fingerprint },
  });

  it("findHistoryFingerprint picks the newest record", () => {
    const records = [
      rec("SSH", "192.168.0.41", "old", 1000),
      rec("SSH", "192.168.0.41", "new", 3000),
      rec("SSH", "192.168.0.41", "mid", 2000),
    ];
    assert.strictEqual(
      ssh.findHistoryFingerprint(records, "192.168.0.41"),
      "new",
    );
  });

  it("findHistoryFingerprint ignores non-SSH records", () => {
    const records = [
      rec("SSH", "192.168.0.41", "ssh", 1000),
      rec("Telnet", "192.168.0.41", "telnet", 2000),
    ];
    assert.strictEqual(
      ssh.findHistoryFingerprint(records, "192.168.0.41"),
      "ssh",
    );
  });

  it("findHistoryFingerprint matches host with and without default port", () => {
    assert.strictEqual(
      ssh.findHistoryFingerprint(
        [rec("SSH", "192.168.0.41:22", "fp", 1000)],
        "192.168.0.41",
      ),
      "fp",
    );
    assert.strictEqual(
      ssh.findHistoryFingerprint(
        [rec("SSH", "192.168.0.41", "fp", 1000)],
        "192.168.0.41:22",
      ),
      "fp",
    );
  });

  it("findHistoryFingerprint does not match a different port", () => {
    assert.strictEqual(
      ssh.findHistoryFingerprint(
        [rec("SSH", "192.168.0.41:2222", "fp", 1000)],
        "192.168.0.41",
      ),
      "",
    );
  });

  it("findHistoryFingerprint skips records without a fingerprint", () => {
    const records = [
      rec("SSH", "192.168.0.41", "fp", 1000),
      rec("SSH", "192.168.0.41", "", 2000),
      { type: "SSH", last: new Date(3000), data: { host: "192.168.0.41" } },
    ];
    assert.strictEqual(
      ssh.findHistoryFingerprint(records, "192.168.0.41"),
      "fp",
    );
  });

  it("findHistoryFingerprint returns empty string when nothing matches", () => {
    assert.strictEqual(ssh.findHistoryFingerprint([], "192.168.0.41"), "");
    assert.strictEqual(
      ssh.findHistoryFingerprint(
        [rec("SSH", "192.168.0.42", "fp", 1000)],
        "192.168.0.41",
      ),
      "",
    );
    assert.strictEqual(
      ssh.findHistoryFingerprint(
        [rec("SSH", "192.168.0.41", "fp", 1000)],
        "not a host",
      ),
      "",
    );
  });
  const preset = (meta) =>
    new presets.Preset({
      title: "Test",
      type: "SSH",
      host: "",
      tab_color: "",
      meta: meta,
    });

  it("presetConnectConfig returns config when all fields are set", () => {
    assert.deepStrictEqual(
      ssh.presetConnectConfig(
        preset({
          Host: "192.168.0.41",
          User: "root",
          Authentication: "Password",
        }),
      ),
      {
        user: "root",
        authentication: "Password",
        host: "192.168.0.41",
        charset: "utf-8",
      },
    );
  });

  it("presetConnectConfig returns null when a field is missing", () => {
    assert.strictEqual(
      ssh.presetConnectConfig(preset({ Host: "192.168.0.41", User: "root" })),
      null,
    );
  });

  it("presetConnectConfig returns null when a field is invalid", () => {
    assert.strictEqual(
      ssh.presetConnectConfig(
        preset({
          Host: "192.168.0.41",
          User: "root",
          Authentication: "Password",
          Encoding: "utf8",
        }),
      ),
      null,
    );
  });

  const known = (type, host, user) => ({
    type: type,
    data: { host: host, user: user },
  });
  const cmd = new ssh.Command();

  it("matchesKnown matches host with and without default port", () => {
    const p = preset({ Host: "192.168.0.41", User: "root" });
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41", "root")),
      true,
    );
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41:22", "root")),
      true,
    );
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41:2222", "root")),
      false,
    );
  });

  it("matchesKnown matches an explicit port", () => {
    const p = preset({ Host: "192.168.0.41:2222" });
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41:2222", "root")),
      true,
    );
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41", "root")),
      false,
    );
  });

  it("matchesKnown requires the same user when User is set", () => {
    const p = preset({ Host: "192.168.0.41", User: "root" });
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41", "admin")),
      false,
    );
  });

  it("matchesKnown matches any user when User is unset", () => {
    const p = preset({ Host: "192.168.0.41" });
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41", "root")),
      true,
    );
    assert.strictEqual(
      cmd.matchesKnown(p, known("SSH", "192.168.0.41", "admin")),
      true,
    );
  });

  it("matchesKnown does not match another command type", () => {
    const p = preset({ Host: "192.168.0.41" });
    assert.strictEqual(
      cmd.matchesKnown(p, known("Telnet", "192.168.0.41", "")),
      false,
    );
  });

  it("matchesKnown does not match an invalid host", () => {
    assert.strictEqual(
      cmd.matchesKnown(
        preset({ Host: "not a host" }),
        known("SSH", "not a host", "root"),
      ),
      false,
    );
    assert.strictEqual(
      cmd.matchesKnown(preset({}), known("SSH", "192.168.0.41", "root")),
      false,
    );
  });

  it("presetFromKnown builds a preset that matches the record", () => {
    const k = {
      type: "SSH",
      title: "root@192.168.0.41:2222",
      data: {
        host: "192.168.0.41:2222",
        user: "root",
        authentication: "Password",
        charset: "utf-8",
        fingerprint: "SHA256:fp",
      },
    };
    const p = cmd.presetFromKnown(k);
    assert.deepStrictEqual(p, {
      Title: "root@192.168.0.41:2222",
      Type: "SSH",
      Host: "192.168.0.41:2222",
      TabColor: "",
      Meta: {
        User: "root",
        Authentication: "Password",
        Encoding: "utf-8",
        Fingerprint: "SHA256:fp",
      },
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

  it("presetFromKnown leaves out empty fields", () => {
    const p = cmd.presetFromKnown({
      type: "SSH",
      title: "root@192.168.0.41",
      data: { host: "192.168.0.41", user: "root" },
    });
    assert.deepStrictEqual(p.Meta, { User: "root" });
  });
});
