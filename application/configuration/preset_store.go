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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Errors
var (
	ErrPresetRevisionConflict = errors.New(
		"presets have been changed since they were loaded")

	ErrPresetInvalid = errors.New("invalid preset")
)

// PresetsReconfigurer lets the commands alter and filter the presets
type PresetsReconfigurer func(p []Preset) ([]Preset, error)

// PresetStore holds the presets while the application runs. Presets can be
// replaced at runtime, the replacement is written back to the configuration
// file
type PresetStore struct {
	lock        sync.RWMutex
	raw         PresetInputs
	presets     []Preset
	revision    string
	reconfigure PresetsReconfigurer
	sourceFile  string
}

func presetRevision(raw PresetInputs) (string, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func newPresetStore(raw PresetInputs, sourceFile string) (*PresetStore, error) {
	presets, err := raw.concretize()
	if err != nil {
		return nil, err
	}
	revision, err := presetRevision(raw)
	if err != nil {
		return nil, err
	}
	return &PresetStore{
		raw:        raw,
		presets:    presets,
		revision:   revision,
		sourceFile: sourceFile,
	}, nil
}

// Reconfigure runs the presets through r, and keeps r to validate later
// replacements. It returns how many presets r has dropped
func (s *PresetStore) Reconfigure(r PresetsReconfigurer) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	presets, err := r(s.presets)
	if err != nil {
		return 0, err
	}
	ignored := len(s.presets) - len(presets)
	s.presets = presets
	s.reconfigure = r
	return ignored, nil
}

// Presets returns the current presets
func (s *PresetStore) Presets() []Preset {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return append([]Preset(nil), s.presets...)
}

// Raw returns the presets as they were written, and their revision
func (s *PresetStore) Raw() (PresetInputs, string) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return append(PresetInputs(nil), s.raw...), s.revision
}

// Allowed returns whether or not given host is the Host of a preset
func (s *PresetStore) Allowed(host string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, p := range s.presets {
		if len(p.Host) > 0 && p.Host == host {
			return true
		}
	}
	return false
}

// Replace replaces all presets with raw and writes them to the configuration
// file. revision must be the one returned by Raw, or
// ErrPresetRevisionConflict is returned
func (s *PresetStore) Replace(raw PresetInputs, revision string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if revision != s.revision {
		return ErrPresetRevisionConflict
	}
	existing, references := s.known()
	presets := make([]Preset, 0, len(raw))
	for i, p := range raw {
		pp, err := s.check(p, references)
		if err != nil && existing[presetKey(p)] {
			// Kept as it was, and ignored just like on start up
			continue
		} else if err != nil {
			return fmt.Errorf("%w %d (titled %q): %s",
				ErrPresetInvalid, i+1, p.Title, err)
		}
		presets = append(presets, pp)
	}
	newRevision, err := presetRevision(raw)
	if err != nil {
		return err
	}
	if err := s.write(raw); err != nil {
		return err
	}
	s.raw = raw
	s.presets = presets
	s.revision = newRevision
	return nil
}

// presetKey returns a key that identifies p by its full content
func presetKey(p PresetInput) string {
	b, _ := json.Marshal(p)
	return string(b)
}

// isPresetReference returns whether v makes the server read a file or an
// environment variable when it is parsed
func isPresetReference(v String) bool {
	scheme, _, found := strings.Cut(string(v), "://")
	return found && (strings.EqualFold(scheme, "file") ||
		strings.EqualFold(scheme, "environment"))
}

// known returns the current presets by presetKey, and the file:// and
// environment:// Meta values they contain
func (s *PresetStore) known() (map[string]bool, map[String]bool) {
	existing := make(map[string]bool, len(s.raw))
	references := map[String]bool{}
	for _, p := range s.raw {
		existing[presetKey(p)] = true
		for _, v := range p.Meta {
			if isPresetReference(v) {
				references[v] = true
			}
		}
	}
	return existing, references
}

// check concretizes p and runs it through the commands. Unlike on start up, a
// preset the commands would drop is an error. Meta values that read a file or
// an environment variable are only accepted when they are in references, so a
// save can keep the ones written in the configuration file but never add one
func (s *PresetStore) check(
	p PresetInput, references map[String]bool) (Preset, error) {
	for k, v := range p.Meta {
		if isPresetReference(v) && !references[v] {
			return Preset{}, fmt.Errorf(
				"Meta %q: file:// and environment:// values can only be "+
					"added in the configuration file", k)
		}
	}
	pp, err := p.concretize()
	if err != nil {
		return Preset{}, err
	}
	ps, err := s.reconfigure([]Preset{pp})
	if err != nil {
		return Preset{}, err
	}
	if len(ps) != 1 {
		return Preset{}, fmt.Errorf(
			"Type %q is unknown or not enabled", p.Type)
	}
	return ps[0], nil
}

// write replaces the "Presets" key of the configuration file with raw, and
// keeps every other key
func (s *PresetStore) write(raw PresetInputs) error {
	info, err := os.Stat(s.sourceFile)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(s.sourceFile)
	if err != nil {
		return err
	}
	cfg := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	// Keys are matched case-insensitively when the file is loaded
	for k := range cfg {
		if strings.EqualFold(k, "Presets") {
			delete(cfg, k)
		}
	}
	if cfg["Presets"], err = json.Marshal(raw); err != nil {
		return err
	}
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	// Write to the file a symlink points to rather than replacing the link
	target, err := filepath.EvalSymlinks(s.sourceFile)
	if err != nil {
		return err
	}
	if s.writeAtomic(target, buf.Bytes(), info.Mode().Perm()) == nil {
		return nil
	}
	// The directory may not be writable, or the file may be a bind mount
	// that cannot be renamed over (Docker). Write it in place instead
	return os.WriteFile(target, buf.Bytes(), info.Mode().Perm())
}

// writeAtomic writes data to a temporary file next to target, then renames it
// over target
func (s *PresetStore) writeAtomic(
	target string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(
		filepath.Dir(target), "."+filepath.Base(target)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(data)
	if err == nil {
		err = tmp.Chmod(perm)
	}
	if cErr := tmp.Close(); err == nil {
		err = cErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}
