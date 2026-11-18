package term

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/schors/velourbbs/internal/line"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

const (
	markStart = "\x1b[200~" // the paste start
	markEnd   = "\x1b[201~" // the paste end
)

// lineResult is the result of ReadLine.
type lineResult struct {
	text string
	ok   bool
	err  error
}

// readLine starts ReadLine in the field f and waits until it blocks.
func readLine(term *Terminal, f line.Field) <-chan lineResult {
	done := make(chan lineResult, 1)
	go func() {
		text, ok, err := term.ReadLine(f)
		done <- lineResult{text, ok, err}
	}()
	synctest.Wait()
	return done
}

// field is a field of M = 10 in the default style.
var field = line.Field{Row: 1, Col: 1, Max: 10}

func TestDecoderPasteMarks(t *testing.T) {
	var d Decoder
	now := time.Now()
	k1, r1 := d.Feed(now, []byte(markStart+"a\x1b[20"))
	k2, r2 := d.Feed(now, []byte("1~b"))
	got := append(k1, k2...)
	if want := []string{PasteStart, "a", PasteEnd, "b"}; !slices.Equal(got, want) || r1 != nil || r2 != nil {
		t.Errorf("keys %q, replies %v %v; want %q and no replies", got, r1, r2, want)
	}
}

// TestReadLineBytes covers the bytes of a draw, with ESC[?2004h at the start
// and ESC[?2004l at the end of the input (line-input, section 7).
func TestReadLineBytes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		before := len(f.output())
		red := screen.Style{Fg: screen.RGBColour(255, 0, 0), Attr: screen.Bold}
		done := readLine(term, line.Field{Row: 2, Col: 3, Max: 4, Style: red})
		// The keys of one read give one draw. XTERM gives 256 colours, so the RGB colour
		// becomes the index 196.
		f.write("a")
		f.write("bc\x1b[D\r")
		r := <-done
		if r.text != "abc" || !r.ok || r.err != nil {
			t.Fatalf("ReadLine() = %q, %v, %v; want abc", r.text, r.ok, r.err)
		}
		want := "\x1b[?2004h" +
			"\x1b[2;3H\x1b[0;1;38;5;196m     \x1b[0m\x1b[2;3H" +
			"\x1b[2;3H\x1b[0;1;38;5;196ma    \x1b[0m\x1b[2;4H" +
			"\x1b[2;3H\x1b[0;1;38;5;196mabc  \x1b[0m\x1b[2;5H" +
			"\x1b[?2004l"
		if got := f.output()[before:]; got != want {
			t.Errorf("output\n%q, want\n%q", got, want)
		}
	})
}

// TestReadLineCancel covers ESC: no line, and no draw after keys that change nothing.
func TestReadLineCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		done := readLine(term, field)
		before := len(f.output())
		f.write("\x1b[D\x1b")
		if r := <-done; r.text != "" || r.ok || r.err != nil {
			t.Errorf("ReadLine() = %q, %v, %v; want cancelled", r.text, r.ok, r.err)
		}
		if got := f.output()[before:]; got != "\x1b[?2004l" {
			t.Errorf("output %q, want only ESC[?2004l", got)
		}
	})
}

// TestReadLinePasteTwoReads covers a paste of two lines in two reads: the line gets
// the first line, and the next key after the input is a key after the paste end.
func TestReadLinePasteTwoReads(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		done := readLine(term, field)
		f.write(markStart + "one\rtw")
		if r := <-done; r.text != "one" || !r.ok || r.err != nil {
			t.Fatalf("ReadLine() = %q, %v, %v; want one", r.text, r.ok, r.err)
		}
		f.write("o\r" + markEnd + "k")
		if got := keys(t, term, 1); got[0] != "k" {
			t.Errorf("key after the input %q, want k", got[0])
		}
	})
}

// TestReadLineClear covers the queue clear at the end of the input: the next line input
// gets no key of the previous one and no key of its open paste (line-input, section 6).
func TestReadLineClear(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		for _, in := range [][]string{
			{"one\rtyped", "two\r"},
			{markStart + "one\rtw", "o\r" + markEnd + "two\r"},
		} {
			done := readLine(term, field)
			f.write(in[0])
			if r := <-done; r.text != "one" || r.err != nil {
				t.Fatalf("%q: first ReadLine() = %q, %v; want one", in, r.text, r.err)
			}
			done = readLine(term, field)
			f.write(in[1])
			if r := <-done; r.text != "two" || r.err != nil {
				t.Errorf("%q: second ReadLine() = %q, %v; want two", in, r.text, r.err)
			}
		}
	})
}

// TestReadLinePasteFull covers a paste of M = 254 characters and its ENTER in one read.
func TestReadLinePasteFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cl := puttyClient
		cl.naws = []byte{0, 0xff, 0xff, 0, 30} // 255x30: the subnegotiation doubles IAC
		term, f, _, _, err := probeLog(t, cl, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		done := readLine(term, line.Field{Row: 1, Col: 1, Max: 254})
		text := strings.Repeat("ы", 254)
		f.write(markStart + text + "\r" + markEnd)
		if r := <-done; r.text != text || !r.ok || r.err != nil {
			t.Errorf("ReadLine() = %q, %v, %v; want 254 characters", r.text, r.ok, r.err)
		}
	})
}

// TestReadLineQueueFull covers a full queue in a paste: the host drops the rest
// of the paste, also its later part (telnet-term, section 8.4).
func TestReadLineQueueFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		done := readLine(term, line.Field{Row: 1, Col: 1, Max: 5})
		// The queue takes 256 keys: abc and 253 x. The ENTER is lost.
		f.write(markStart + "abc" + strings.Repeat("x", 300) + "\rdef")
		// Without the drop, the ENTER of this part ends the input with abcxx.
		f.write("ghi\r" + markEnd + "\x7f\r")
		if r := <-done; r.text != "abcx" || !r.ok || r.err != nil {
			t.Errorf("ReadLine() = %q, %v, %v; want abcx", r.text, r.ok, r.err)
		}
	})
}

// TestReadLineNextPaste covers a paste start after a full queue: the new paste is not dropped.
func TestReadLineNextPaste(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		done := readLine(term, field)
		// The queue takes 256 TAB, the line input ignores them, and the next read
		// gives a new paste.
		f.write(markStart + strings.Repeat("\t", 300))
		synctest.Wait()
		f.write("x" + markStart + "ab\r")
		if r := <-done; r.text != "ab" || !r.ok || r.err != nil {
			t.Errorf("ReadLine() = %q, %v, %v; want ab", r.text, r.ok, r.err)
		}
	})
}

// TestKeyPaste covers the keys of a paste outside the line input: the host drops them.
func TestKeyPaste(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		f.write("a" + markStart + "b\rc" + markEnd + "d" + markStart + "e")
		f.write("f" + markEnd + "g")
		if got := keys(t, term, 3); !slices.Equal(got, []string{"a", "d", "g"}) {
			t.Errorf("keys %q, want a d g", got)
		}
	})
}

// TestPasteIdle covers the idle timeout: the paste marks do not reset it,
// the dropped keys of a paste do (telnet-term, section 8.3).
func TestPasteIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		to := testTimeouts
		to.Idle = 300 * time.Millisecond
		term, f := connect(t, to)
		begin := time.Now()
		go func() {
			time.Sleep(100 * time.Millisecond)
			f.write(markStart + "a")
			time.Sleep(100 * time.Millisecond)
			f.write(markEnd + markStart + markEnd)
		}()
		if k, err := term.Key(); !errors.Is(err, ErrIdle) {
			t.Fatalf("Key() = %q, %v; want ErrIdle", k, err)
		}
		if d := time.Since(begin); d != 100*time.Millisecond+to.Idle {
			t.Errorf("ErrIdle after %v, want %v", d, 100*time.Millisecond+to.Idle)
		}
	})
}

// TestReadLineModeOff covers ESC[?2004l at each end of the session in the line input
// (telnet-term, section 8.4).
func TestReadLineModeOff(t *testing.T) {
	tests := []struct {
		name string
		end  func(*Terminal, *fake)
		err  error
	}{
		{"idle", func(*Terminal, *fake) {}, ErrIdle},
		{"IP", func(_ *Terminal, f *fake) { f.write("a\xff\xf4") }, telnet.ErrHangup},
		{"BRK", func(_ *Terminal, f *fake) { f.write("\xff\xf3") }, telnet.ErrHangup},
		{"protocol error", func(_ *Terminal, f *fake) {
			f.write("\xff\xfa\x1f" + strings.Repeat("0", telnet.MaxSub+2))
		}, telnet.ErrSubTooLong},
		{"stop", func(term *Terminal, _ *fake) { term.Stop() }, ErrStop},
	}
	for _, tt := range tests {
		synctest.Test(t, func(t *testing.T) {
			to := testTimeouts
			to.Idle = time.Second
			term, f := connect(t, to)
			done := readLine(term, field)
			tt.end(term, f)
			if r := <-done; !errors.Is(r.err, tt.err) {
				t.Errorf("%s: ReadLine() error %v, want %v", tt.name, r.err, tt.err)
			}
			if out := f.output(); !strings.HasSuffix(out, "\x1b[?2004l") {
				t.Errorf("%s: output %q does not end with ESC[?2004l", tt.name, out)
			}
		})
	}
}

// TestReadLineFit covers a field that does not fit the screen: a panic before any output.
func TestReadLineFit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		before := len(f.output())
		func() {
			defer func() {
				if recover() == nil {
					t.Error("ReadLine with a field of 100 cells on 100 columns: no panic")
				}
			}()
			_, _, _ = term.ReadLine(line.Field{Row: 1, Col: 1, Max: 100})
		}()
		if got := f.output()[before:]; got != "" {
			t.Errorf("output %q, want none", got)
		}
	})
}
