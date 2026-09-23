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

package command

import (
	"testing"

	"github.com/nirui/sshwifty/application/configuration"
	"github.com/nirui/sshwifty/application/log"
)

func testCommands() Commands {
	cmd := func(
		l log.Logger,
		h Hooks,
		w StreamResponder,
		cfg Configuration,
		b *BufferPool,
	) FSMMachine {
		return nil
	}
	reload := func(p configuration.Preset) (configuration.Preset, error) {
		return p, nil
	}
	return Commands{
		Register("Telnet", cmd, reload),
		Register("SSH", cmd, reload),
	}
}

func TestCommandsEnableEmpty(t *testing.T) {
	c := testCommands()
	enabled, err := c.Enable(nil)
	if err != nil {
		t.Errorf("Expecting no error, got %s", err)
		return
	}
	if enabled[0].name != "Telnet" || enabled[1].name != "SSH" {
		t.Errorf("Expecting commands to be unchanged, got %v",
			enabled.Names())
		return
	}
}

func TestCommandsEnableOne(t *testing.T) {
	enabled, err := testCommands().Enable([]string{"ssh"})
	if err != nil {
		t.Errorf("Expecting no error, got %s", err)
		return
	}
	if enabled[0].command != nil {
		t.Error("Expecting Telnet slot to be empty")
		return
	}
	if enabled[1].name != "SSH" || enabled[1].command == nil {
		t.Error("Expecting SSH to keep its ID")
		return
	}
	_, err = enabled.Run(0, log.NewDitch(), Hooks{}, StreamResponder{},
		Configuration{}, nil)
	if err != ErrCommandRunUndefinedCommand {
		t.Errorf("Expecting ErrCommandRunUndefinedCommand, got %v", err)
		return
	}
	presets, err := enabled.Reconfigure([]configuration.Preset{
		{Title: "t", Type: "Telnet"},
		{Title: "s", Type: "SSH"},
	})
	if err != nil {
		t.Errorf("Expecting no error, got %s", err)
		return
	}
	if len(presets) != 1 || presets[0].Type != "SSH" {
		t.Errorf("Expecting only the SSH preset to remain, got %v", presets)
		return
	}
}

func TestCommandsEnableUnknown(t *testing.T) {
	_, err := testCommands().Enable([]string{"SSH", "Bogus"})
	if err == nil {
		t.Error("Expecting an error for unknown protocol")
		return
	}
}

func TestCommandsNames(t *testing.T) {
	names := testCommands().Names()
	if len(names) != 2 || names[0] != "Telnet" || names[1] != "SSH" {
		t.Errorf("Expecting [Telnet SSH], got %v", names)
		return
	}
}
