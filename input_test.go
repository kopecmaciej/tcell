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

package tcell

import "testing"

func scanKeys(t *testing.T, bareBSIsCtrlH bool, b []byte) []*EventKey {
	t.Helper()
	ch := make(chan Event, 16)
	ip := NewInputProcessor(ch).(*inputProcessor)
	ip.bareBSIsCtrlH = bareBSIsCtrlH
	ip.ScanUTF8(b)

	var keys []*EventKey
	for {
		select {
		case ev := <-ch:
			if ke, ok := ev.(*EventKey); ok {
				keys = append(keys, ke)
			}
		default:
			return keys
		}
	}
}

func scanEvents(t *testing.T, b []byte) []Event {
	t.Helper()
	ch := make(chan Event, 256)
	ip := NewInputProcessor(ch).(*inputProcessor)
	ip.ScanUTF8(b)

	var evs []Event
	for {
		select {
		case ev := <-ch:
			evs = append(evs, ev)
		default:
			return evs
		}
	}
}

func TestBracketedPasteDeliversRawContent(t *testing.T) {
	content := "line1\n\tindented\nline3"
	input := append([]byte("\x1b[200~"), content...)
	input = append(input, "\x1b[201~"...)

	evs := scanEvents(t, input)

	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d: %#v", len(evs), evs)
	}
	start, ok := evs[0].(*EventPaste)
	if !ok || !start.Start() {
		t.Fatalf("event 0: want paste start, got %#v", evs[0])
	}
	end, ok := evs[1].(*EventPaste)
	if !ok || !end.End() {
		t.Fatalf("event 1: want paste end, got %#v", evs[1])
	}
	if string(end.Data()) != content {
		t.Errorf("paste data: got %q, want %q", end.Data(), content)
	}
}

func TestBracketedPasteSplitAcrossReads(t *testing.T) {
	ch := make(chan Event, 256)
	ip := NewInputProcessor(ch).(*inputProcessor)
	ip.ScanUTF8([]byte("\x1b[200~hello\nwor"))
	ip.ScanUTF8([]byte("ld\x1b["))
	ip.ScanUTF8([]byte("201~"))

	var data []byte
	var ends int
	for done := false; !done; {
		select {
		case ev := <-ch:
			if pe, ok := ev.(*EventPaste); ok && pe.End() {
				data = pe.Data()
				ends++
			}
		default:
			done = true
		}
	}
	if ends != 1 {
		t.Fatalf("expected 1 paste end event, got %d", ends)
	}
	if string(data) != "hello\nworld" {
		t.Errorf("paste data: got %q, want %q", data, "hello\nworld")
	}
}

func TestBareBackspaceKeptAsBackspace(t *testing.T) {
	keys := scanKeys(t, false, []byte{0x08, 0x7f})
	if len(keys) != 2 {
		t.Fatalf("expected 2 events, got %d", len(keys))
	}
	for i, ev := range keys {
		if ev.Name() != "Backspace" {
			t.Errorf("event %d: got %q, want Backspace", i, ev.Name())
		}
	}
}

func TestBareBackspaceSplitsCtrlHFromBackspace(t *testing.T) {
	keys := scanKeys(t, true, []byte{0x08, 0x7f})
	if len(keys) != 2 {
		t.Fatalf("expected 2 events, got %d", len(keys))
	}
	if keys[0].Name() != "Ctrl+H" || keys[0].Modifiers() != ModCtrl {
		t.Errorf("0x08: got %q mod=%d, want Ctrl+H ModCtrl", keys[0].Name(), keys[0].Modifiers())
	}
	if keys[1].Name() != "Backspace" || keys[1].Modifiers() != ModNone {
		t.Errorf("0x7f: got %q mod=%d, want Backspace ModNone", keys[1].Name(), keys[1].Modifiers())
	}
}
