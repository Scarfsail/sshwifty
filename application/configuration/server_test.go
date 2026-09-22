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
	"testing"
	"time"
)

func TestServerNormalizePingInterval(t *testing.T) {
	for _, c := range []struct {
		name         string
		readTimeout  time.Duration
		pingInterval time.Duration
		expected     time.Duration
	}{
		{
			name:         "Unspecified PingInterval is derived from ReadTimeout",
			readTimeout:  120 * time.Second,
			pingInterval: 0,
			expected:     60 * time.Second,
		},
		{
			name:         "Too large PingInterval is clamped down",
			readTimeout:  100 * time.Second,
			pingInterval: 90 * time.Second,
			expected:     70 * time.Second,
		},
		{
			name:         "Too small PingInterval is clamped up",
			readTimeout:  10 * time.Second,
			pingInterval: 500 * time.Millisecond,
			expected:     serverMinValidSecond,
		},
		{
			name: "When ReadTimeout is too short for the lower limit, " +
				"the upper limit wins",
			readTimeout:  1 * time.Second,
			pingInterval: 0,
			expected:     700 * time.Millisecond,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			ss := Server{
				ListenInterface: "127.0.0.1",
				ReadTimeout:     c.readTimeout,
				PingInterval:    c.pingInterval,
			}.normalize()
			if ss.PingInterval != c.expected {
				t.Errorf(
					"Expecting the PingInterval to be %s, got %s instead",
					c.expected,
					ss.PingInterval,
				)
				return
			}
			// The Ping must land inside every read deadline window, otherwise
			// an idle connection will still be dropped
			if ss.PingInterval >= ss.ReadTimeout {
				t.Errorf(
					"Expecting the PingInterval %s to be smaller than the "+
						"ReadTimeout %s",
					ss.PingInterval,
					ss.ReadTimeout,
				)
				return
			}
		})
	}
}
