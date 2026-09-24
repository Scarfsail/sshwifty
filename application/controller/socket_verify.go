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
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/nirui/sshwifty/application/command"
	"github.com/nirui/sshwifty/application/configuration"
	"github.com/nirui/sshwifty/application/log"
)

type socketVerification struct {
	socket

	heartbeat string
	timeout   string
	protocols []string
}

type socketRemotePreset struct {
	Title    string            `json:"title"`
	Type     string            `json:"type"`
	Host     string            `json:"host"`
	TabColor string            `json:"tab_color"`
	Meta     map[string]string `json:"meta"`
}

type socketAccessConfiguration struct {
	Presets                      []socketRemotePreset `json:"presets"`
	ServerMessage                string               `json:"server_message"`
	BypassClipboardWriteApproval bool                 `json:"bypass_clipboard_write_approval"`
	SkipPresetPromptWhenAllSet   bool                 `json:"skip_preset_prompt_when_all_set"`
	EnabledProtocols             []string             `json:"enabled_protocols"`
	PresetEditing                bool                 `json:"preset_editing"`
}

func newSocketAccessConfiguration(
	remotes []configuration.Preset,
	serverMessage string,
	bypassClipboardWriteApproval bool,
	skipPresetPromptWhenAllSet bool,
	enabledProtocols []string,
	presetEditing bool,
) socketAccessConfiguration {
	presets := make([]socketRemotePreset, len(remotes))
	for i := range presets {
		presets[i] = socketRemotePreset{
			Title:    remotes[i].Title,
			Type:     remotes[i].Type,
			Host:     remotes[i].Host,
			TabColor: remotes[i].TabColor,
			Meta:     remotes[i].Meta,
		}
	}
	return socketAccessConfiguration{
		Presets:                      presets,
		ServerMessage:                parseServerMessage(html.EscapeString(serverMessage)),
		BypassClipboardWriteApproval: bypassClipboardWriteApproval,
		SkipPresetPromptWhenAllSet:   skipPresetPromptWhenAllSet,
		EnabledProtocols:             enabledProtocols,
		PresetEditing:                presetEditing,
	}
}

func buildAccessConfigRespondBody(accessCfg socketAccessConfiguration) []byte {
	mData, mErr := json.Marshal(accessCfg)
	if mErr != nil {
		panic(fmt.Errorf("unable to marshal remote data: %s", mErr))
	}
	return mData
}

func newSocketVerification(
	s socket,
	srvCfg configuration.Server,
	commCfg configuration.Common,
	cmds command.Commands,
) socketVerification {
	return socketVerification{
		socket: s,
		heartbeat: strconv.FormatFloat(
			srvCfg.HeartbeatTimeout.Seconds(), 'g', 2, 64),
		timeout: strconv.FormatFloat(
			srvCfg.ReadTimeout.Seconds(), 'g', 2, 64),
		protocols: cmds.Names(),
	}
}

func (s socketVerification) authKey(r *http.Request) []byte {
	timeMixer := strconv.FormatInt(time.Now().Unix()/100, 10)
	if len(s.commonCfg.SharedKey) > 0 {
		return hashCombineSocketKeys(
			timeMixer,
			s.commonCfg.SharedKey,
		)[:32]
	}
	return hashCombineSocketKeys(
		timeMixer,
		"DEFAULT VERIFY KEY",
	)[:32]
}

func (s socketVerification) setServerConfigRespond(
	hd *http.Header, w http.ResponseWriter) {
	hd.Add("X-Heartbeat", s.heartbeat)
	hd.Add("X-Timeout", s.timeout)
	if s.commonCfg.OnlyAllowPresetRemotes {
		hd.Add("X-OnlyAllowPresetRemotes", "yes")
	}
	hd.Add("Content-Type", "text/json; charset=utf-8")
	// Built on every request, as the presets can be edited at runtime
	w.Write(buildAccessConfigRespondBody(newSocketAccessConfiguration(
		s.commonCfg.Presets.Presets(),
		s.serverCfg.ServerMessage,
		s.commonCfg.BypassClipboardWriteApproval,
		s.commonCfg.SkipPresetPromptWhenAllSet,
		s.protocols,
		s.commonCfg.AllowPresetEditing,
	)))
}

func (s socketVerification) Get(
	w *ResponseWriter, r *http.Request, l log.Logger) error {
	hd := w.Header()
	hd.Add("Cache-Control", "no-store")
	hd.Add("Pragma", "no-store")
	if len(r.Header.Get("X-Key")) <= 0 {
		hd.Add("X-Key", base64.StdEncoding.EncodeToString(s.mixerKey(r)))
		if len(s.commonCfg.SharedKey) <= 0 {
			s.setServerConfigRespond(&hd, w)
			return nil
		}
		return ErrSocketInvalidAuthKey
	}
	if err := s.verifyKey(r); err != nil {
		return err
	}
	hd.Add("X-Key", base64.StdEncoding.EncodeToString(s.mixerKey(r)))
	s.setServerConfigRespond(&hd, w)
	return nil
}

// verifyKey checks the X-Key header of r against the shared key
func (s socketVerification) verifyKey(r *http.Request) error {
	key := r.Header.Get("X-Key")
	if len(key) <= 0 || len(key) > 64 {
		return ErrSocketInvalidAuthKey
	}
	// Delay the brute force attack. Use it with connection limits (via
	// iptables or nginx etc)
	time.Sleep(500 * time.Millisecond)
	decodedKey, decodedKeyErr := base64.StdEncoding.DecodeString(key)
	if decodedKeyErr != nil {
		return NewError(http.StatusBadRequest, decodedKeyErr.Error())
	}
	if !hmac.Equal(s.authKey(r), decodedKey) {
		return ErrSocketAuthFailed
	}
	return nil
}
