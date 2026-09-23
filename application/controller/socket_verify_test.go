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
	"encoding/json"
	"strings"
	"testing"

	"github.com/nirui/sshwifty/application/configuration"
)

func TestNewSocketAccessConfigurationBypassClipboardWriteApprovalDefault(t *testing.T) {
	cfg := newSocketAccessConfiguration(
		[]configuration.Preset{},
		"",
		false,
		false,
		nil,
	)
	if cfg.BypassClipboardWriteApproval {
		t.Error("Expecting BypassClipboardWriteApproval to be false, " +
			"got true instead")
		return
	}
	body := buildAccessConfigRespondBody(cfg)
	if !strings.Contains(string(body), `"bypass_clipboard_write_approval":false`) {
		t.Errorf("Expecting respond body to contain "+
			"\"bypass_clipboard_write_approval\":false, got %s instead",
			body)
		return
	}
	var decoded socketAccessConfiguration
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Errorf("Unable to unmarshal respond body: %s", err)
		return
	}
	if decoded.BypassClipboardWriteApproval {
		t.Error("Expecting decoded BypassClipboardWriteApproval to be " +
			"false, got true instead")
		return
	}
}

func TestNewSocketAccessConfigurationBypassClipboardWriteApprovalEnabled(t *testing.T) {
	cfg := newSocketAccessConfiguration(
		[]configuration.Preset{},
		"",
		true,
		false,
		nil,
	)
	if !cfg.BypassClipboardWriteApproval {
		t.Error("Expecting BypassClipboardWriteApproval to be true, " +
			"got false instead")
		return
	}
	body := buildAccessConfigRespondBody(cfg)
	if !strings.Contains(string(body), `"bypass_clipboard_write_approval":true`) {
		t.Errorf("Expecting respond body to contain "+
			"\"bypass_clipboard_write_approval\":true, got %s instead",
			body)
		return
	}
	var decoded socketAccessConfiguration
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Errorf("Unable to unmarshal respond body: %s", err)
		return
	}
	if !decoded.BypassClipboardWriteApproval {
		t.Error("Expecting decoded BypassClipboardWriteApproval to be " +
			"true, got false instead")
		return
	}
}

func TestNewSocketAccessConfigurationSkipPresetPromptWhenAllSetDefault(t *testing.T) {
	cfg := newSocketAccessConfiguration(
		[]configuration.Preset{},
		"",
		false,
		false,
		nil,
	)
	if cfg.SkipPresetPromptWhenAllSet {
		t.Error("Expecting SkipPresetPromptWhenAllSet to be false, " +
			"got true instead")
		return
	}
	body := buildAccessConfigRespondBody(cfg)
	if !strings.Contains(string(body), `"skip_preset_prompt_when_all_set":false`) {
		t.Errorf("Expecting respond body to contain "+
			"\"skip_preset_prompt_when_all_set\":false, got %s instead",
			body)
		return
	}
	var decoded socketAccessConfiguration
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Errorf("Unable to unmarshal respond body: %s", err)
		return
	}
	if decoded.SkipPresetPromptWhenAllSet {
		t.Error("Expecting decoded SkipPresetPromptWhenAllSet to be " +
			"false, got true instead")
		return
	}
}

func TestNewSocketAccessConfigurationSkipPresetPromptWhenAllSetEnabled(t *testing.T) {
	cfg := newSocketAccessConfiguration(
		[]configuration.Preset{},
		"",
		false,
		true,
		nil,
	)
	if !cfg.SkipPresetPromptWhenAllSet {
		t.Error("Expecting SkipPresetPromptWhenAllSet to be true, " +
			"got false instead")
		return
	}
	body := buildAccessConfigRespondBody(cfg)
	if !strings.Contains(string(body), `"skip_preset_prompt_when_all_set":true`) {
		t.Errorf("Expecting respond body to contain "+
			"\"skip_preset_prompt_when_all_set\":true, got %s instead",
			body)
		return
	}
	var decoded socketAccessConfiguration
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Errorf("Unable to unmarshal respond body: %s", err)
		return
	}
	if !decoded.SkipPresetPromptWhenAllSet {
		t.Error("Expecting decoded SkipPresetPromptWhenAllSet to be " +
			"true, got false instead")
		return
	}
}

func TestNewSocketAccessConfigurationEnabledProtocols(t *testing.T) {
	cfg := newSocketAccessConfiguration(
		[]configuration.Preset{},
		"",
		false,
		false,
		[]string{"SSH"},
	)
	body := buildAccessConfigRespondBody(cfg)
	if !strings.Contains(string(body), `"enabled_protocols":["SSH"]`) {
		t.Errorf("Expecting respond body to contain "+
			"\"enabled_protocols\":[\"SSH\"], got %s instead",
			body)
		return
	}
}
