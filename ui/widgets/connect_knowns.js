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

// hasKnowns reports whether there is anything to show on the Known remotes
// view. Restricted deployments always count as having something, since their
// backend only allows connections built from presets, making the New remote
// view useless there even when no preset is available
export function hasKnowns(knowns, presets, restrictedToPresets) {
  return restrictedToPresets || knowns.length > 0 || presets.length > 0;
}
