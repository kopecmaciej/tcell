// Copyright 2025 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package tcell

import "golang.org/x/sys/unix"

// eraseCharIsDEL reports whether the terminal's erase (Backspace) key sends
// DEL (0x7f) rather than BS (0x08). When it does, a bare 0x08 byte can only be
// Ctrl+H, so the two can be told apart even without an enhanced keyboard
// protocol (e.g. under tmux with extended-keys off).
func eraseCharIsDEL(fd int) bool {
	tio, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return false
	}
	return tio.Cc[unix.VERASE] == 0x7f
}
