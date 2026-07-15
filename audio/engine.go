package audio

import "syscall/js"

// Engine wraps the Web Audio API via syscall/js.
type Engine struct {
	ctx        js.Value
	masterGain js.Value
	muted      bool
}

// NewEngine creates an AudioContext and master gain node.
func NewEngine() *Engine {
	ctx := js.Global().Get("AudioContext").New()
	master := ctx.Call("createGain")
	master.Get("gain").Set("value", 0.3)
	master.Call("connect", ctx.Get("destination"))

	return &Engine{
		ctx:        ctx,
		masterGain: master,
	}
}

// Resume unlocks the AudioContext (must be called from a user gesture).
func (e *Engine) Resume() {
	if e.ctx.Get("state").String() == "suspended" {
		e.ctx.Call("resume")
	}
}

// CurrentTime returns the audio context's current time in seconds.
func (e *Engine) CurrentTime() float64 {
	return e.ctx.Get("currentTime").Float()
}

// SetVolume sets the master volume (0.0 to 1.0).
func (e *Engine) SetVolume(v float64) {
	e.masterGain.Get("gain").Set("value", v)
}

// ToggleMute toggles audio on/off.
func (e *Engine) ToggleMute() {
	e.muted = !e.muted
	if e.muted {
		e.masterGain.Get("gain").Set("value", 0)
	} else {
		e.masterGain.Get("gain").Set("value", 0.3)
	}
}

// PlayTone plays an oscillator at the given frequency for duration seconds.
func (e *Engine) PlayTone(freq float64, startTime, duration float64, waveform string, volume float64) {
	osc := e.ctx.Call("createOscillator")
	gain := e.ctx.Call("createGain")

	osc.Get("type").Set("value", waveform) // ignored, set below
	osc.Set("type", waveform)
	osc.Get("frequency").Set("value", freq)

	gain.Get("gain").Call("setValueAtTime", volume, startTime)
	gain.Get("gain").Call("linearRampToValueAtTime", 0, startTime+duration)

	osc.Call("connect", gain)
	gain.Call("connect", e.masterGain)

	osc.Call("start", startTime)
	osc.Call("stop", startTime+duration+0.05)
}

// PlayNoise plays a short burst of noise (for percussion/SFX).
func (e *Engine) PlayNoise(startTime, duration, volume float64) {
	// Create a buffer of random samples
	sampleRate := e.ctx.Get("sampleRate").Float()
	length := int(sampleRate * duration)
	if length < 1 {
		length = 1
	}

	buffer := e.ctx.Call("createBuffer", 1, length, int(sampleRate))
	data := buffer.Call("getChannelData", 0)

	// Fill with random values
	for i := 0; i < length; i++ {
		data.SetIndex(i, js.ValueOf((js.Global().Get("Math").Call("random").Float()*2)-1))
	}

	src := e.ctx.Call("createBufferSource")
	src.Set("buffer", buffer)

	gain := e.ctx.Call("createGain")
	gain.Get("gain").Call("setValueAtTime", volume, startTime)
	gain.Get("gain").Call("linearRampToValueAtTime", 0, startTime+duration)

	src.Call("connect", gain)
	gain.Call("connect", e.masterGain)
	src.Call("start", startTime)
}

// Note frequency table (octave 0-7).
var noteFreqs = map[string]float64{
	"C": 16.35, "C#": 17.32, "D": 18.35, "D#": 19.45,
	"E": 20.60, "F": 21.83, "F#": 23.12, "G": 24.50,
	"G#": 25.96, "A": 27.50, "A#": 29.14, "B": 30.87,
}

// NoteFreq returns the frequency for a note name and octave (e.g., "A", 4 = 440Hz).
func NoteFreq(note string, octave int) float64 {
	base, ok := noteFreqs[note]
	if !ok {
		return 440.0
	}
	for i := 0; i < octave; i++ {
		base *= 2
	}
	return base
}
