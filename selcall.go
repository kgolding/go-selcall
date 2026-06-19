package selcall

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
)

// Selcall generates selcall tone sequences as raw PCM or WAV data.
type Selcall interface {
	Write(value string, w io.Writer) error
	WriteWav(value string, w io.Writer) error
	SetDigitDuration(d time.Duration)
	SetInterDigitDuration(d time.Duration)
	SetFirstDigitDuration(d time.Duration)
	SetVolumePercent(v int)
	SetSampleRate(hz int)
}

type selcall struct {
	table         map[rune]float64
	sampleRate    int
	digitDur      time.Duration
	interDigitDur time.Duration
	firstDigitDur time.Duration
	maxLen        int // max value length; 0 = unlimited
	volumePercent int
}

func NewZVEI1() Selcall {
	return &selcall{
		table: map[rune]float64{
			'0': 2400, '1': 1060, '2': 1160, '3': 1270, '4': 1400,
			'5': 1530, '6': 1670, '7': 1830, '8': 2000, '9': 2200,
			'A': 2800, 'B': 810, 'C': 970, 'D': 886, 'E': 2600,
		},
		digitDur:      70 * time.Millisecond,
		interDigitDur: 0,
		sampleRate:    8000,
		volumePercent: 80,
	}
}

func NewCCIR() Selcall {
	return &selcall{
		table: map[rune]float64{
			'0': 1981, '1': 1124, '2': 1197, '3': 1275, '4': 1358,
			'5': 1446, '6': 1540, '7': 1640, '8': 1747, '9': 1860,
			'A': 2400, 'B': 930, 'C': 2247, 'D': 991, 'E': 2110,
		},
		digitDur:      100 * time.Millisecond,
		interDigitDur: 0,
		sampleRate:    8000,
		volumePercent: 80,
	}
}

func NewZVEI2() Selcall {
	return &selcall{
		table: map[rune]float64{
			'0': 2400, '1': 1060, '2': 1160, '3': 1270, '4': 1400,
			'5': 1530, '6': 1670, '7': 1830, '8': 2000, '9': 2200,
			'A': 886, 'B': 810, 'C': 740, 'D': 680, 'E': 970,
		},
		digitDur:      70 * time.Millisecond,
		interDigitDur: 0,
		sampleRate:    8000,
		volumePercent: 80,
	}
}

func NewEEA() Selcall {
	return &selcall{
		table: map[rune]float64{
			'0': 1981, '1': 1124, '2': 1197, '3': 1275, '4': 1358,
			'5': 1446, '6': 1540, '7': 1640, '8': 1747, '9': 1860,
			'A': 1055, 'B': 930, 'C': 2247, 'D': 991, 'E': 2110,
		},
		digitDur:      40 * time.Millisecond,
		interDigitDur: 0,
		maxLen:        5,
		sampleRate:    8000,
		volumePercent: 80,
	}
}

func (s *selcall) SetDigitDuration(d time.Duration)      { s.digitDur = d }
func (s *selcall) SetInterDigitDuration(d time.Duration) { s.interDigitDur = d }
func (s *selcall) SetFirstDigitDuration(d time.Duration) { s.firstDigitDur = d }
func (s *selcall) SetVolumePercent(v int)                { s.volumePercent = max(0, min(100, v)) }
func (s *selcall) SetSampleRate(hz int)                  { s.sampleRate = hz }

// Write generates raw 16-bit mono PCM samples for the selcall sequence and writes them to w.
// Consecutive identical digits are transmitted as the 'E' repeat tone.
func (s *selcall) Write(value string, w io.Writer) error {
	runes := []rune(value)
	if s.maxLen > 0 && len(runes) > s.maxLen {
		return fmt.Errorf("value length %d exceeds maximum of %d", len(runes), s.maxLen)
	}
	amplitude := (float64(s.volumePercent) / 100.0) * 0.9 * math.MaxInt16
	gap := make([]int16, int(s.interDigitDur.Seconds()*float64(s.sampleRate)))

	var lastRune rune
	for i, r := range runes {
		send := r
		if r == lastRune {
			send = 'E'
		}
		lastRune = r

		freq, ok := s.table[send]
		if !ok {
			return fmt.Errorf("character %q at position %d has no entry in table", send, i)
		}

		dur := s.digitDur
		if i == 0 && s.firstDigitDur > 0 {
			dur = s.firstDigitDur
		}

		if err := binary.Write(w, binary.LittleEndian, generateTone(freq, int(dur.Seconds()*float64(s.sampleRate)), amplitude, s.sampleRate)); err != nil {
			return err
		}
		if i < len(runes)-1 {
			if err := binary.Write(w, binary.LittleEndian, gap); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteWav writes a complete WAV file for the selcall sequence to w.
func (s *selcall) WriteWav(value string, w io.Writer) error {
	runes := []rune(value)
	if s.maxLen > 0 && len(runes) > s.maxLen {
		return fmt.Errorf("value length %d exceeds maximum of %d", len(runes), s.maxLen)
	}
	sr := float64(s.sampleRate)
	firstSamples := int(s.digitDur.Seconds() * sr)
	if s.firstDigitDur > 0 {
		firstSamples = int(s.firstDigitDur.Seconds() * sr)
	}
	toneSamples := int(s.digitDur.Seconds() * sr)
	gapSamples := int(s.interDigitDur.Seconds() * sr)

	numSamples := 0
	if len(runes) > 0 {
		numSamples = firstSamples + (len(runes)-1)*toneSamples + max(0, len(runes)-1)*gapSamples
	}

	if err := writeWavHeader(w, numSamples, s.sampleRate); err != nil {
		return err
	}
	return s.Write(value, w)
}

// writeWavHeader writes a RIFF/WAV header for 16-bit mono PCM with the given sample count.
func writeWavHeader(w io.Writer, numSamples, sampleRate int) error {
	const (
		numChannels   = 1
		bitsPerSample = 16
	)
	dataSize := numSamples * 2
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8

	for _, v := range []any{
		[]byte("RIFF"),
		uint32(36 + dataSize),
		[]byte("WAVE"),
		[]byte("fmt "),
		uint32(16),
		uint16(1), // PCM format
		uint16(numChannels),
		uint32(sampleRate),
		uint32(byteRate),
		uint16(blockAlign),
		uint16(bitsPerSample),
		[]byte("data"),
		uint32(dataSize),
	} {
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}

// generateTone returns n samples of a sine wave at freq Hz, ramped in/out to avoid clicks.
func generateTone(freq float64, n int, amplitude float64, sampleRate int) []int16 {
	out := make([]int16, n)
	rampSamples := min(sampleRate/200, n/2) // ~5ms ramp

	for i := range n {
		v := amplitude * math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate))
		if i < rampSamples {
			v *= float64(i) / float64(rampSamples)
		} else if i >= n-rampSamples {
			v *= float64(n-1-i) / float64(rampSamples)
		}
		out[i] = int16(v)
	}
	return out
}
