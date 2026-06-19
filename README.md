# selcall

A Go library and CLI tool for generating Selective Calling (Selcall) tone sequences as WAV audio or raw PCM.

Supports four standard modes: **ZVEI 1**, **ZVEI 2**, **CCIR**, and **EEA**.

## Library

### Install

```sh
go get selcall
```

### Usage

```go
import "selcall"

sc := selcall.NewZVEI1()

// Write a WAV file
f, _ := os.Create("out.wav")
defer f.Close()
sc.WriteWav("12345", f)

// Write raw PCM to any io.Writer
sc.Write("12345", w)
```

### Constructors

| Function | Mode | Default tone duration |
|---|---|---|
| `NewZVEI1()` | ZVEI 1 | 70 ms |
| `NewZVEI2()` | ZVEI 2 | 70 ms |
| `NewCCIR()` | CCIR | 100 ms |
| `NewEEA()` | EEA | 40 ms |

All constructors default to 8000 Hz sample rate and 80% volume.

### Configuration methods

```go
sc.SetDigitDuration(70 * time.Millisecond)
sc.SetInterDigitDuration(0)           // silence between digits
sc.SetFirstDigitDuration(200 * time.Millisecond) // optional longer first digit
sc.SetVolumePercent(80)               // 0–100
sc.SetSampleRate(44100)
```

### Tone frequencies (Hz)

| Digit | ZVEI 1 | ZVEI 2 | CCIR | EEA |
|---|---|---|---|---|
| 0 | 2400 | 2400 | 1981 | 1981 |
| 1 | 1060 | 1060 | 1124 | 1124 |
| 2 | 1160 | 1160 | 1197 | 1197 |
| 3 | 1270 | 1270 | 1275 | 1275 |
| 4 | 1400 | 1400 | 1358 | 1358 |
| 5 | 1530 | 1530 | 1446 | 1446 |
| 6 | 1670 | 1670 | 1540 | 1540 |
| 7 | 1830 | 1830 | 1640 | 1640 |
| 8 | 2000 | 2000 | 1747 | 1747 |
| 9 | 2200 | 2200 | 1860 | 1860 |
| A | 2800 | 886  | 2400 | 1055 |
| B | 810  | 810  | 930  | 930  |
| C | 970  | 740  | 2247 | 2247 |
| D | 886  | 680  | 991  | 991  |
| E | 2600 | 970  | 2110 | 2110 |

**E is the repeat tone.** Consecutive identical digits are automatically transmitted as `E` — for example `"1122"` is encoded as `1, E, 2, E` on the wire.

## CLI

### Build

```sh
go build -o selcall ./cmd
```

### Usage

```
selcall [-a MODE] VALUE OUTPUT
```

| Argument | Description |
|---|---|
| `-a MODE` | Mode: `ZVEI1`, `ZVEI2`, `CCIR`, `EEA` (default: `ZVEI1`) |
| `VALUE` | String of digits to transmit (`0`–`9`, `A`–`E`) |
| `OUTPUT` | Output `.wav` filename, or `-` for raw 16-bit mono PCM on stdout |

### Examples

```sh
# Write a ZVEI 1 WAV file
selcall 12345 out.wav

# Write a CCIR WAV file
selcall -a CCIR 12345 out.wav

# Pipe raw PCM to a sound device (8000 Hz, 16-bit signed, mono)
selcall 12345 - | aplay -r 8000 -f S16_LE -c 1

# Verify output with multimon-ng
selcall 12345 out.wav && multimon-ng -t wav -a ZVEI1 out.wav
```

## Testing

Tests use [multimon-ng](https://github.com/EliasOenal/multimon-ng) to decode generated WAV files and verify correctness. Tests are skipped automatically if `multimon-ng` is not found in `PATH`.

```sh
go test ./...
```
