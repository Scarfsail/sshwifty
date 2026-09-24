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

package controller

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nirui/sshwifty/application/command"
	"github.com/nirui/sshwifty/application/configuration"
	"github.com/nirui/sshwifty/application/log"
)

func testPresetsCtl(t *testing.T, editing bool) presets {
	cfgFile := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	err := os.WriteFile(cfgFile, []byte(`{
  "SharedKey": "key",
  "AllowPresetEditing": true,
  "Servers": [{"ListenPort": 8182}],
  "Presets": [{"Title": "Server", "Type": "SSH", "Host": "server:22"}]
}`), 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, cfg, err := configuration.CustomFile(cfgFile)(log.NewDitch())
	if err != nil {
		t.Fatal(err)
	}
	_, err = cfg.Presets.Reconfigure(
		func(p []configuration.Preset) ([]configuration.Preset, error) {
			return p, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	cfg.AllowPresetEditing = editing
	commonCfg := cfg.Common()
	return presets{verify: newSocketVerification(
		socket{commonCfg: commonCfg}, cfg.Servers[0], commonCfg,
		command.Commands{})}
}

func testPresetsRequest(
	p presets, method string, key string, body string,
) (*httptest.ResponseRecorder, error) {
	r := httptest.NewRequest(method, "/sshwifty/presets",
		strings.NewReader(body))
	if key == "" {
		key = base64.StdEncoding.EncodeToString(p.verify.authKey(r))
	}
	r.Header.Set("X-Key", key)
	rec := httptest.NewRecorder()
	w := newResponseWriter(rec)
	return rec, serveController(p, &w, r, log.NewDitch())
}

func testPresetsErrorCode(err error) int {
	if e, ok := err.(Error); ok {
		return e.Code()
	}
	return 0
}

func TestPresetsEditingDisabled(t *testing.T) {
	_, err := testPresetsRequest(testPresetsCtl(t, false), "GET", "", "")
	if testPresetsErrorCode(err) != http.StatusForbidden {
		t.Errorf("Expecting 403, got %v", err)
	}
}

func TestPresetsBadKey(t *testing.T) {
	_, err := testPresetsRequest(testPresetsCtl(t, true), "GET",
		base64.StdEncoding.EncodeToString([]byte("wrong")), "")
	if testPresetsErrorCode(err) != http.StatusForbidden {
		t.Errorf("Expecting 403, got %v", err)
	}
}

func TestPresetsGetAndPut(t *testing.T) {
	p := testPresetsCtl(t, true)
	rec, err := testPresetsRequest(p, "GET", "", "")
	if err != nil {
		t.Fatalf("Expecting no error, got %s", err)
	}
	loaded := presetsData{}
	if err := json.Unmarshal(rec.Body.Bytes(), &loaded); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Presets) != 1 || loaded.Presets[0].Title != "Server" {
		t.Fatalf("Unexpected presets %v", loaded.Presets)
	}
	if !strings.Contains(rec.Body.String(), `"Title":"Server"`) {
		t.Errorf("Expecting the file keys, got %s", rec.Body.String())
	}
	loaded.Presets[0].Title = "Renamed"
	body, _ := json.Marshal(loaded)
	rec, err = testPresetsRequest(p, "PUT", "", string(body))
	if err != nil || rec.Code != http.StatusOK {
		t.Fatalf("Expecting 200, got %d %v", rec.Code, err)
	}
	saved := presetsData{}
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Revision == loaded.Revision ||
		saved.Presets[0].Title != "Renamed" {
		t.Errorf("Unexpected PUT respond %s", rec.Body.String())
	}
	_, err = testPresetsRequest(p, "PUT", "", string(body))
	if testPresetsErrorCode(err) != http.StatusConflict {
		t.Errorf("Expecting 409 for a stale revision, got %v", err)
	}
}
