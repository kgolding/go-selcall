package selcall

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	modeFlagZVEI1 = "ZVEI1"
	modeFlagZVEI2 = "ZVEI2"
	modeFlagCCIR  = "CCIR"
	modeFlagEEA   = "EEA"
)

func skipIfNoMultimonNg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("multimon-ng"); err != nil {
		t.Skip("multimon-ng not found in PATH")
	}
}

// writeTempWav writes value to a temp WAV file and returns its path.
// The file is automatically removed when the test ends.
func writeTempWav(t *testing.T, sc Selcall, value string) string {
	t.Helper()
	f, err := os.CreateTemp("", "selcall-*.wav")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	defer f.Close()
	if err := sc.WriteWav(value, f); err != nil {
		t.Fatalf("WriteWav(%q): %v", value, err)
	}
	return f.Name()
}

// decodeWav runs multimon-ng on a WAV file and returns the decoded digit string for modeFlag.
func decodeWav(t *testing.T, filename, modeFlag string) string {
	t.Helper()
	out, _ := exec.Command("multimon-ng", "-t", "wav", "-a", modeFlag, filename).Output()
	prefix := modeFlag + ": "
	for _, line := range strings.Split(string(out), "\n") {
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

// applyRepeat returns the sequence as it will actually be transmitted:
// consecutive identical digits are replaced with 'E'.
func applyRepeat(value string) string {
	runes := []rune(value)
	out := make([]rune, len(runes))
	var last rune
	for i, r := range runes {
		if r == last {
			out[i] = 'E'
		} else {
			out[i] = r
		}
		last = r
	}
	return string(out)
}

func testMode(t *testing.T, sc Selcall, modeFlag string, tests []string) {
	t.Helper()
	skipIfNoMultimonNg(t)

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			path := writeTempWav(t, sc, value)
			got := decodeWav(t, path, modeFlag)
			want := applyRepeat(value)
			if got != want {
				t.Errorf("%s: decoded %q, want %q", modeFlag, got, want)
			}
		})
	}
}

var standardTests = []string{
	"12345",
	"0011223",
	"0",
	"0123456789ABCDD",
}

// zvei2Tests omits extended tones B/C/D (810/740/680 Hz): multimon-ng ZVEI2 decoder
// does not detect them, even though ZVEI1 uses the same frequencies successfully.
var zvei2Tests = []string{
	"12345",
	"0011223",
	"0",
	"0123456789",
}

var eeaTests = []string{
	"12345",
	"00112", // truncated: EEA max 5 digits
	"0",
	"01234", // truncated: EEA max 5 digits
}

func TestZVEI1(t *testing.T) {
	sc := NewZVEI1()
	sc.SetDigitDuration(100 * time.Millisecond)
	testMode(t, sc, modeFlagZVEI1, standardTests)
}

func TestZVEI2(t *testing.T) {
	sc := NewZVEI2()
	sc.SetDigitDuration(100 * time.Millisecond)
	testMode(t, sc, modeFlagZVEI2, zvei2Tests)
}

func TestCCIR(t *testing.T) {
	testMode(t, NewCCIR(), modeFlagCCIR, standardTests)
}

func TestEEA(t *testing.T) {
	sc := NewEEA()
	sc.SetDigitDuration(100 * time.Millisecond) // EEA default 40ms may be too short to detect
	testMode(t, sc, modeFlagEEA, eeaTests)
}
