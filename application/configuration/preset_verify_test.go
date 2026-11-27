// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2025 Ni Rui <ranqus@gmail.com>
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

package configuration

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPinnedConfig = `{
  "SharedKey": "key",
  "AllowPresetEditing": true,
  "PinnedSSHFingerprints": {"server": "SHA256:NAo="},
  "Servers": [{"ListenPort": 8182}],
  "Presets": [
    {"Title": "Server", "Type": "SSH", "Host": "server:22",
     "Meta": {"Fingerprint": "%s"}}
  ]
}`

func testPinnedConfiguration(t *testing.T, fingerprint string) Configuration {
	cfgFile := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	err := os.WriteFile(cfgFile, []byte(
		strings.Replace(testPinnedConfig, "%s", fingerprint, 1)), 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, cfg, err := loadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Presets.Reconfigure(testPresetStoreReconfigure); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestVerifyRejectsPresetAgainstPinnedFingerprint(t *testing.T) {
	if err := testPinnedConfiguration(t, "SHA256:NAo=").Verify(); err != nil {
		t.Errorf("Expecting a matching Fingerprint to verify, got %s", err)
	}
	if err := testPinnedConfiguration(t, "SHA256:AAAA").Verify(); err == nil {
		t.Error("Expecting a mismatching Fingerprint to fail the verification")
	}
}

// A save must never write a preset that would stop the next start up
func TestPresetStoreReplaceRunsVerifier(t *testing.T) {
	cfg := testPinnedConfiguration(t, "SHA256:NAo=")
	if err := cfg.Verify(); err != nil {
		t.Fatal(err)
	}
	cfg.Presets.SetVerifier(cfg.PresetVerifier())
	raw, revision := cfg.Presets.Raw()
	bad := append(PresetInputs(nil), raw...)
	bad = append(bad, PresetInput{
		Title: "Same server, another key", Type: "SSH", Host: "server",
		Meta: map[string]String{"Fingerprint": "SHA256:AAAA"},
	})
	err := cfg.Presets.Replace(bad, revision)
	if !errors.Is(err, ErrPresetInvalid) {
		t.Fatalf("Expecting ErrPresetInvalid, got %v", err)
	}
	if got, _ := cfg.Presets.Raw(); len(got) != 1 {
		t.Errorf("Expecting the presets to be left unchanged, got %v", got)
	}
	bad[1].Host = "[::1"
	if err := cfg.Presets.Replace(bad, revision); !errors.Is(err, ErrPresetInvalid) {
		t.Errorf("Expecting an invalid Host to be refused, got %v", err)
	}
	good := append(PresetInputs(nil), raw...)
	good = append(good, PresetInput{
		Title: "Other", Type: "SSH", Host: "other:2222",
		Meta: map[string]String{"Fingerprint": "SHA256:AAAA"},
	})
	if err := cfg.Presets.Replace(good, revision); err != nil {
		t.Errorf("Expecting no error, got %s", err)
	}
}
