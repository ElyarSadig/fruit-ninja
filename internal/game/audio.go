package game

import (
	"encoding/binary"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Small sound effects are synthesized once at startup as headerless 16-bit
// stereo PCM. No third-party sound assets or internet connection are needed.
type soundID uint8

const (
	sndSwipe soundID = iota
	sndSlice
	sndCombo
	sndBomb
	sndClick
	sndStart
	sndEnd
)

type soundBank struct {
	ctx    *audio.Context
	clips  map[soundID][]byte
	voices []*audio.Player
}

func newSoundBank() *soundBank {
	s := &soundBank{
		ctx:   audio.NewContext(44100),
		clips: make(map[soundID][]byte),
	}
	s.clips[sndSwipe] = generateSound(0.16, 31, func(t, noise float64) float64 {
		e := math.Pow(math.Max(0, 1-t/0.16), 2)
		return 0.24*e*noise + 0.16*e*math.Sin(2*math.Pi*(1300*t-2500*t*t))
	})
	s.clips[sndSlice] = generateSound(0.20, 17, func(t, noise float64) float64 {
		e := math.Exp(-t * 23)
		chirp := math.Sin(2 * math.Pi * (510*t + 2500*t*t))
		bass := math.Sin(2 * math.Pi * (190*t - 190*t*t))
		return 0.31*e*noise + 0.45*e*chirp + 0.16*math.Exp(-t*16)*bass
	})
	s.clips[sndCombo] = generateSound(0.44, 89, func(t, noise float64) float64 {
		e := math.Min(1, t*120) * math.Exp(-t*7)
		freq := 660.0
		if t > 0.12 {
			freq = 840
		}
		if t > 0.24 {
			freq = 1120
		}
		return 0.49 * e * (math.Sin(2*math.Pi*freq*t) + 0.28*math.Sin(4*math.Pi*freq*t))
	})
	s.clips[sndBomb] = generateSound(0.68, 71, func(t, noise float64) float64 {
		e := math.Exp(-t * 6.5)
		rumble := math.Sin(2 * math.Pi * (145*t - 75*t*t))
		crack := noise * math.Exp(-t*18)
		return 0.64*e*rumble + 0.42*crack + 0.11*e*noise
	})
	s.clips[sndClick] = generateSound(0.085, 19, func(t, noise float64) float64 {
		return 0.32 * math.Exp(-t*44) * math.Sin(2*math.Pi*1100*t)
	})
	s.clips[sndStart] = generateSound(0.52, 41, func(t, noise float64) float64 {
		e := math.Exp(-t * 4.6)
		frequency := 440.0
		if t > 0.13 {
			frequency = 554
		}
		if t > 0.27 {
			frequency = 660
		}
		if t > 0.39 {
			frequency = 880
		}
		return 0.3 * e * (math.Sin(2*math.Pi*frequency*t) + 0.3*math.Sin(4*math.Pi*frequency*t))
	})
	s.clips[sndEnd] = generateSound(0.75, 41, func(t, noise float64) float64 {
		e := math.Exp(-t * 3.8)
		frequency := 440.0
		if t > 0.21 {
			frequency = 349
		}
		if t > 0.42 {
			frequency = 261
		}
		return 0.29 * e * math.Sin(2*math.Pi*frequency*t)
	})
	return s
}

func generateSound(seconds float64, seed int64, wave func(t, noise float64) float64) []byte {
	const sampleRate = 44100
	count := int(seconds * sampleRate)
	samples := make([]byte, count*4) // signed little-endian L + R
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < count; i++ {
		t := float64(i) / sampleRate
		sample := clamp(wave(t, rng.Float64()*2-1), -0.94, 0.94)
		v := int16(sample * 32767)
		binary.LittleEndian.PutUint16(samples[i*4:], uint16(v))
		binary.LittleEndian.PutUint16(samples[i*4+2:], uint16(v))
	}
	return samples
}

func (s *soundBank) play(id soundID, muted bool) {
	if s == nil || muted || !s.ctx.IsReady() {
		return
	}
	p := s.ctx.NewPlayerFromBytes(s.clips[id])
	if p == nil {
		return
	}
	p.SetVolume(0.47)
	p.Play()
	s.voices = append(s.voices, p)
}

func (s *soundBank) cleanup() {
	if s == nil {
		return
	}
	n := 0
	for _, p := range s.voices {
		if p.IsPlaying() {
			s.voices[n] = p
			n++
		} else {
			p.PauseAndStopReading()
		}
	}
	s.voices = s.voices[:n]
}
