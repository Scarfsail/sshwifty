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
	"errors"
	"net/http"

	"github.com/nirui/sshwifty/application/configuration"
	"github.com/nirui/sshwifty/application/log"
)

// Errors
var (
	ErrPresetsEditingDisabled = NewError(
		http.StatusForbidden, "Preset editing is not enabled")

	ErrPresetsRevisionConflict = NewError(
		http.StatusConflict, "Presets have been changed elsewhere")
)

// presets reads and replaces the presets for the web UI
type presets struct {
	baseController

	verify socketVerification
}

type presetsData struct {
	Revision string                     `json:"revision"`
	Presets  configuration.PresetInputs `json:"presets"`
}

func (p presets) auth(w *ResponseWriter, r *http.Request) error {
	hd := w.Header()
	hd.Add("Cache-Control", "no-store")
	hd.Add("Pragma", "no-store")
	if !p.verify.commonCfg.AllowPresetEditing {
		return ErrPresetsEditingDisabled
	}
	return p.verify.verifyKey(r)
}

func (p presets) respond(w *ResponseWriter) error {
	raw, revision := p.verify.commonCfg.Presets.Raw()
	data, err := json.Marshal(presetsData{Revision: revision, Presets: raw})
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	return err
}

func (p presets) Get(w *ResponseWriter, r *http.Request, l log.Logger) error {
	if err := p.auth(w, r); err != nil {
		return err
	}
	return p.respond(w)
}

// badRequest responds with the reason in plain text, which the error page
// would hide, so the UI can show it
func (p presets) badRequest(w *ResponseWriter, reason error) error {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte(reason.Error()))
	return err
}

func (p presets) Put(w *ResponseWriter, r *http.Request, l log.Logger) error {
	if err := p.auth(w, r); err != nil {
		return err
	}
	data := presetsData{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return p.badRequest(w, err)
	}
	err := p.verify.commonCfg.Presets.Replace(data.Presets, data.Revision)
	switch {
	case errors.Is(err, configuration.ErrPresetRevisionConflict):
		return ErrPresetsRevisionConflict
	case errors.Is(err, configuration.ErrPresetInvalid),
		errors.Is(err, configuration.ErrPresetFileChanged):
		return p.badRequest(w, err)
	case err != nil:
		return err
	}
	l.Info("Presets have been replaced")
	return p.respond(w)
}
