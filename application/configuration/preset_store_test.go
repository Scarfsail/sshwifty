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
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nirui/sshwifty/application/log"
)

const testPresetStoreConfig = `{
  "SharedKey": "key",
  "Servers": [{"ListenPort": 8182}],
  "Presets": [
    {"Title": "Server", "Type": "SSH", "Host": "server:22",
     "Meta": {"Private Key": "file://%s"}}
  ],
  "OnlyAllowPresetRemotes": true
}`

// testPresetStoreReconfigure keeps SSH presets only, like the commands do
func testPresetStoreReconfigure(p []Preset) ([]Preset, error) {
	pp := make([]Preset, 0, len(p))
	for _, v := range p {
		if v.Type == "SSH" {
			pp = append(pp, v)
		}
	}
	return pp, nil
}

func testPresetStore(t *testing.T) (*PresetStore, string) {
	return testPresetStoreFrom(t, testPresetStoreConfig)
}

func testPresetStoreFrom(t *testing.T, config string) (*PresetStore, string) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key")
	if err := os.WriteFile(keyFile, []byte("SECRET"), 0600); err != nil {
		t.Fatal(err)
	}
	cfgFile := filepath.Join(dir, "sshwifty.conf.json")
	err := os.WriteFile(cfgFile, []byte(
		strings.Replace(config, "%s", keyFile, 1)), 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, cfg, err := loadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SourceFile != cfgFile {
		t.Fatalf("Expecting SourceFile %q, got %q", cfgFile, cfg.SourceFile)
	}
	if _, err := cfg.Presets.Reconfigure(testPresetStoreReconfigure); err != nil {
		t.Fatal(err)
	}
	return cfg.Presets, cfgFile
}

func testPresetStoreReadFile(t *testing.T, cfgFile string) map[string]json.RawMessage {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestPresetStoreReplace(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	raw, revision := s.Raw()
	if len(raw) != 1 || s.Presets()[0].Meta["Private Key"] != "SECRET" {
		t.Fatalf("Unexpected presets %v", raw)
	}
	raw = append(raw, PresetInput{Title: "New", Type: "SSH", Host: "new:22"})
	if err := s.Replace(raw, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	_, cfg, err := loadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.OnlyAllowPresetRemotes || cfg.SharedKey != "key" ||
		len(cfg.Servers) != 1 || cfg.Servers[0].ListenPort != 8182 {
		t.Errorf("Expecting other settings to be kept, got %v", cfg)
	}
	reloaded, _ := cfg.Presets.Raw()
	if len(reloaded) != 2 || reloaded[1].Title != "New" {
		t.Errorf("Expecting the new preset in the file, got %v", reloaded)
	}
	if !strings.HasPrefix(string(reloaded[0].Meta["Private Key"]), "file://") {
		t.Errorf("Expecting the file:// reference to be kept, got %q",
			reloaded[0].Meta["Private Key"])
	}
	if _, newRevision := s.Raw(); newRevision == revision {
		t.Error("Expecting the revision to change")
	}
}

func TestPresetStoreReplaceConflict(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	before, _ := os.ReadFile(cfgFile)
	err := s.Replace(PresetInputs{}, "stale")
	if !errors.Is(err, ErrPresetRevisionConflict) {
		t.Fatalf("Expecting ErrPresetRevisionConflict, got %v", err)
	}
	if after, _ := os.ReadFile(cfgFile); string(after) != string(before) {
		t.Error("Expecting the file to be left unchanged")
	}
}

func TestPresetStoreReplaceUnknownType(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	before, _ := os.ReadFile(cfgFile)
	raw, revision := s.Raw()
	raw = append(raw, PresetInput{Title: "Bogus", Type: "Bogus"})
	err := s.Replace(raw, revision)
	if !errors.Is(err, ErrPresetInvalid) {
		t.Fatalf("Expecting ErrPresetInvalid, got %v", err)
	}
	if after, _ := os.ReadFile(cfgFile); string(after) != string(before) {
		t.Error("Expecting the file to be left unchanged")
	}
	if kept, _ := s.Raw(); len(kept) != 1 {
		t.Errorf("Expecting the presets to be left unchanged, got %v", kept)
	}
}

func TestPresetStoreReplaceKeepsPresetsKeyOnly(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	before := testPresetStoreReadFile(t, cfgFile)
	_, revision := s.Raw()
	if err := s.Replace(PresetInputs{}, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	after := testPresetStoreReadFile(t, cfgFile)
	if len(after) != len(before) || string(after["Presets"]) != "[]" {
		t.Errorf("Expecting only Presets to be emptied, got %v", after)
	}
}

func TestPresetStoreAllowed(t *testing.T) {
	s, _ := testPresetStore(t)
	if !s.Allowed("server:22") || s.Allowed("new:22") {
		t.Fatal("Unexpected allowed hosts before Replace")
	}
	_, revision := s.Raw()
	err := s.Replace(PresetInputs{{Title: "New", Type: "SSH", Host: "new:22"}},
		revision)
	if err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	if s.Allowed("server:22") || !s.Allowed("new:22") {
		t.Error("Expecting allowed hosts to follow Replace")
	}
}

func TestPresetStoreReplaceNewReference(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	raw, revision := s.Raw()
	// Editing the preset keeps its existing reference
	raw[0].Title = "Renamed"
	if err := s.Replace(raw, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	before, _ := os.ReadFile(cfgFile)
	for _, v := range []String{"file:///etc/passwd", "Environment://HOME"} {
		raw, revision = s.Raw()
		raw = append(raw, PresetInput{
			Title: "Sneaky", Type: "SSH", Host: "sneaky:22",
			Meta: Meta{"Password": v},
		})
		if err := s.Replace(raw, revision); !errors.Is(err, ErrPresetInvalid) {
			t.Errorf("Expecting ErrPresetInvalid for %q, got %v", v, err)
		}
	}
	if after, _ := os.ReadFile(cfgFile); string(after) != string(before) {
		t.Error("Expecting the file to be left unchanged")
	}
}

func TestPresetStoreReplaceKeepsIgnoredPreset(t *testing.T) {
	s, cfgFile := testPresetStoreFrom(t, `{
  "SharedKey": "key",
  "Servers": [{"ListenPort": 8182}],
  "Presets": [
    {"Title": "Router", "Type": "Telnet", "Host": "router:23"},
    {"Title": "Server", "Type": "SSH", "Host": "server:22"}
  ]
}`)
	raw, revision := s.Raw()
	if len(raw) != 2 || len(s.Presets()) != 1 {
		t.Fatalf("Expecting the Telnet preset to be ignored, got %v", raw)
	}
	raw[1].Title = "Renamed"
	if err := s.Replace(raw, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	_, cfg, err := loadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if kept, _ := cfg.Presets.Raw(); len(kept) != 2 || kept[0].Title != "Router" {
		t.Errorf("Expecting the ignored preset in the file, got %v", kept)
	}
	if s.Allowed("router:23") || !s.Allowed("server:22") {
		t.Error("Expecting the ignored preset to stay ignored")
	}
	// A changed preset of an unknown Type is still refused
	raw, revision = s.Raw()
	raw[0].Host = "router2:23"
	if err := s.Replace(raw, revision); !errors.Is(err, ErrPresetInvalid) {
		t.Errorf("Expecting ErrPresetInvalid, got %v", err)
	}
}

func TestPresetStoreReplaceSymlink(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	link := filepath.Join(t.TempDir(), "link.conf.json")
	if err := os.Symlink(cfgFile, link); err != nil {
		t.Skip("Unable to create symlink:", err)
	}
	s.sourceFile = link
	_, revision := s.Raw()
	if err := s.Replace(PresetInputs{}, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Expecting the symlink to be kept, got %v, %v", fi, err)
	}
	if after := testPresetStoreReadFile(t, cfgFile); string(after["Presets"]) != "[]" {
		t.Errorf("Expecting the link target to be written, got %s", after["Presets"])
	}
}

func TestPresetStoreReplaceFileChanged(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	data, _ := os.ReadFile(cfgFile)
	edited := strings.Replace(string(data), `"Title": "Server"`,
		`"Title": "Edited by hand"`, 1)
	if err := os.WriteFile(cfgFile, []byte(edited), 0600); err != nil {
		t.Fatal(err)
	}
	_, revision := s.Raw()
	err := s.Replace(PresetInputs{}, revision)
	if !errors.Is(err, ErrPresetFileChanged) {
		t.Fatalf("Expecting ErrPresetFileChanged, got %v", err)
	}
	if after, _ := os.ReadFile(cfgFile); string(after) != edited {
		t.Error("Expecting the hand edit to be kept")
	}
}

func TestPresetStoreReplaceReadOnlyDir(t *testing.T) {
	s, cfgFile := testPresetStore(t)
	dir := filepath.Dir(cfgFile)
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0700)
	if f, err := os.CreateTemp(dir, "probe"); err == nil {
		f.Close()
		os.Remove(f.Name())
		t.Skip("Directory is still writable (running as root?)")
	}
	_, revision := s.Raw()
	if err := s.Replace(PresetInputs{}, revision); err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	if after := testPresetStoreReadFile(t, cfgFile); string(after["Presets"]) != "[]" {
		t.Errorf("Expecting the file to be written in place, got %s",
			after["Presets"])
	}
}

func TestDirectLoaderWithoutPresets(t *testing.T) {
	_, cfg, err := Direct(Configuration{})(log.NewDitch())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Presets == nil || len(cfg.Presets.Presets()) != 0 {
		t.Errorf("Expecting an empty preset store, got %v", cfg.Presets)
	}
}
