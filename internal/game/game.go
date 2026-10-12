package game

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	etext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

type Game struct {
	mode        playState
	art         *artwork
	scene       *ebiten.Image
	regular     *etext.GoTextFaceSource
	bold        *etext.GoTextFaceSource
	sound       *soundBank
	initialized bool
	rng         *rand.Rand

	fruits     []fruit
	halves     []fruitHalf
	particles  []particle
	popups     []floatingText
	trail      []trailPoint
	cursor     Vec2
	prevCursor Vec2
	held       bool

	// Track one finger for a complete swipe; switching fingers starts a new stroke.
	activeTouch bool
	touchID     ebiten.TouchID
	touchIDs    []ebiten.TouchID

	score, best int
	lives       int
	remaining   float64
	elapsed     float64
	spawnDelay  float64
	chain       int
	lastSlice   float64
	shake       float64
	redFlash    float64
	clock       float64
	muted       bool
	loadedBest  bool
	totalSliced int
}

// NewGame only prepares pure-Go state. Android's JNI context is not yet
// guaranteed to exist during mobile package initialization, so image/audio
// resources are created lazily on the first Ebitengine Update tick.
func NewGame() (*Game, error) {
	return &Game{
		mode:      stateMenu,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		remaining: 60,
	}, nil
}

func (g *Game) initResources() error {
	if g.initialized {
		return nil
	}
	regular, err := etext.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return fmt.Errorf("regular font: %w", err)
	}
	bold, err := etext.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		return fmt.Errorf("bold font: %w", err)
	}
	g.regular = regular
	g.bold = bold
	g.art = newArtwork()
	g.scene = ebiten.NewImage(screenWidth, screenHeight)
	g.sound = newSoundBank()
	g.initialized = true
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) startRound() {
	g.mode = statePlaying
	g.fruits = g.fruits[:0]
	g.halves = g.halves[:0]
	g.particles = g.particles[:0]
	g.popups = g.popups[:0]
	g.trail = g.trail[:0]
	g.score = 0
	g.totalSliced = 0
	g.lives = 3
	g.elapsed = 0
	g.remaining = 60
	g.spawnDelay = 0.38
	g.chain = 0
	g.lastSlice = -100
	g.shake = 0
	g.redFlash = 0
	g.held = false
	g.sound.play(sndStart, g.muted)
}

func (g *Game) endRound() {
	if g.mode != statePlaying {
		return
	}
	g.mode = stateOver
	if g.score > g.best {
		g.best = g.score
		saveBestScore(g.best)
	}
	g.sound.play(sndEnd, g.muted)
}

func (g *Game) Update() error {
	if err := g.initResources(); err != nil {
		return err
	}
	const dt = 1.0 / framesPerSec
	g.clock += dt
	if !g.loadedBest {
		// On Android, the Activity sets the app-private save directory
		// before the first update is dispatched.
		g.best = loadBestScore()
		g.loadedBest = true
	}
	g.sound.cleanup()

	pointer, click, down := g.readPointer()
	g.cursor = pointer

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.muted = !g.muted
		if !g.muted {
			g.sound.play(sndClick, false)
		}
	}

	switch g.mode {
	case stateMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			(click && pointInRect(g.cursor, 470, 497, 340, 71)) {
			g.startRound()
		}
	case statePlaying:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
			(click && pointInRect(g.cursor, 1180, 15, 90, 90)) {
			g.mode = statePaused
			g.held = false
			g.trail = g.trail[:0]
			g.sound.play(sndClick, g.muted)
			break
		}
		if click && pointInRect(g.cursor, 18, 647, 170, 61) {
			g.muted = !g.muted
			if !g.muted {
				g.sound.play(sndClick, false)
			}
			g.held = false
			g.trail = g.trail[:0]
			break
		}
		g.updatePlaying(dt, down)
	case statePaused:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) || (click && pointInRect(g.cursor, 484, 413, 312, 69)) {
			g.mode = statePlaying
			g.held = false
			g.sound.play(sndClick, g.muted)
		} else if click && pointInRect(g.cursor, 523, 512, 234, 45) {
			g.mode = stateMenu
			g.sound.play(sndClick, g.muted)
		}
	case stateOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) || (click && pointInRect(g.cursor, 468, 502, 344, 69)) {
			g.startRound()
		} else if click && pointInRect(g.cursor, 529, 588, 222, 39) {
			g.mode = stateMenu
			g.sound.play(sndClick, g.muted)
		}
	}
	// The cursor's previous position is recorded even on non-playing frames,
	// so resuming the game cannot create a huge accidental slash.
	g.prevCursor = g.cursor
	return nil
}

func (g *Game) updatePlaying(dt float64, down bool) {
	g.elapsed += dt
	g.remaining -= dt
	if g.remaining <= 0 {
		g.remaining = 0
		g.endRound()
		return
	}
	if g.shake > 0 {
		g.shake *= 0.83
		if g.shake < 0.15 {
			g.shake = 0
		}
	}
	if g.redFlash > 0 {
		g.redFlash = math.Max(0, g.redFlash-dt*2.5)
	}

	// Continuous movement creates a sharp, short-lived blade trail.
	if down {
		if !g.held {
			g.trail = append(g.trail, trailPoint{Pos: g.cursor})
		} else {
			delta := g.cursor.Sub(g.prevCursor)
			if delta.Length() > 3 {
				g.trail = append(g.trail, trailPoint{Pos: g.cursor})
				g.slash(g.prevCursor, g.cursor)
			}
		}
	}
	g.held = down
	n := 0
	for _, p := range g.trail {
		p.Age += dt
		if p.Age < 0.24 {
			g.trail[n] = p
			n++
		}
	}
	g.trail = g.trail[:n]
	if len(g.trail) > 27 {
		g.trail = g.trail[len(g.trail)-27:]
	}

	g.spawnDelay -= dt
	if g.spawnDelay <= 0 {
		g.spawnWave()
		difficulty := clamp(g.elapsed/60, 0, 1)
		g.spawnDelay = mix(1.08, 0.53, difficulty) + g.rng.Float64()*0.38
	}

	n = 0
	for _, f := range g.fruits {
		f.Age += dt
		f.Pos = f.Pos.Add(f.Vel.Mul(dt))
		f.Vel.Y += 650 * dt
		f.Angle += f.Spin * dt
		if f.Pos.Y > screenHeight+f.Radius+30 && f.Vel.Y > 0 {
			if f.Kind != bomb {
				g.loseLife(f.Pos.X)
			}
			continue
		}
		g.fruits[n] = f
		n++
	}
	g.fruits = g.fruits[:n]

	n = 0
	for _, h := range g.halves {
		h.Age += dt
		h.Pos = h.Pos.Add(h.Vel.Mul(dt))
		h.Vel.Y += 790 * dt
		h.Angle += h.Spin * dt
		if h.Age < h.Lifetime && h.Pos.Y < screenHeight+150 {
			g.halves[n] = h
			n++
		}
	}
	g.halves = g.halves[:n]

	n = 0
	for _, p := range g.particles {
		p.Life -= dt
		p.Pos = p.Pos.Add(p.Vel.Mul(dt))
		p.Vel.Y += p.Gravity * dt
		p.Vel.X *= 0.995
		if p.Life > 0 {
			g.particles[n] = p
			n++
		}
	}
	g.particles = g.particles[:n]

	n = 0
	for _, p := range g.popups {
		p.Life -= dt
		p.Pos.Y -= 43 * dt
		if p.Life > 0 {
			g.popups[n] = p
			n++
		}
	}
	g.popups = g.popups[:n]
	if g.chain > 0 && g.elapsed-g.lastSlice > 0.66 {
		g.chain = 0
	}
	if g.lives <= 0 {
		g.endRound()
	}
}

func (g *Game) spawnWave() {
	number := 1
	if g.rng.Float64() < 0.68 {
		number++
	}
	if g.elapsed > 12 && g.rng.Float64() < 0.38 {
		number++
	}
	if g.elapsed > 35 && g.rng.Float64() < 0.25 {
		number++
	}
	positions := make([]float64, 0, number)
	for i := 0; i < number; i++ {
		x := 75 + g.rng.Float64()*float64(screenWidth-150)
		// Prevent new fruits from stacking directly on top of each other.
		for attempts := 0; attempts < 12; attempts++ {
			good := true
			for _, used := range positions {
				if math.Abs(x-used) < 107 {
					good = false
					break
				}
			}
			if good {
				break
			}
			x = 75 + g.rng.Float64()*float64(screenWidth-150)
		}
		positions = append(positions, x)
	}
	for i, x := range positions {
		kind := fruitKind(g.rng.Intn(int(plum) + 1))
		roll := g.rng.Float64()
		if roll < 0.045 {
			kind = golden
		}
		if g.elapsed > 5 && roll > 0.84 && i > 0 {
			kind = bomb
		}
		radius := 36 + g.rng.Float64()*11
		if kind == bomb {
			radius = 39
		}
		if kind == golden {
			radius = 35
		}
		// Vertical speed is measured in pixels/second, not pixels/tick.
		launchSpeed := 670 + g.rng.Float64()*160
		vx := (float64(screenWidth)/2-x)*0.18 + (g.rng.Float64()*2-1)*90
		g.fruits = append(g.fruits, fruit{
			Kind: kind, Pos: Vec2{x, screenHeight + radius + 12},
			Vel: Vec2{vx, -launchSpeed}, Radius: radius,
			Angle: g.rng.Float64()*0.4 - 0.2, Spin: (g.rng.Float64()*2 - 1) * 2.5,
		})
	}
}

func (g *Game) slash(a, b Vec2) {
	distance := b.Sub(a).Length()
	if distance < 5 {
		return
	}
	angle := math.Atan2(b.Y-a.Y, b.X-a.X)
	hitAny := false
	n := 0
	for _, f := range g.fruits {
		if !segmentHitsCircle(a, b, f.Pos, f.Radius*0.90) {
			g.fruits[n] = f
			n++
			continue
		}
		hitAny = true
		if f.Kind == bomb {
			g.explode(f.Pos)
		} else {
			g.sliceFruit(f, angle)
		}
	}
	g.fruits = g.fruits[:n]
	if !hitAny && g.rng.Float64() < 0.18 {
		g.sound.play(sndSwipe, g.muted)
	}
}

func (g *Game) sliceFruit(f fruit, angle float64) {
	g.totalSliced++
	if g.elapsed-g.lastSlice < 0.66 {
		g.chain++
	} else {
		g.chain = 1
	}
	g.lastSlice = g.elapsed
	points := 10
	if g.chain >= 3 {
		points += g.chain * 3
	}
	if f.Kind == golden {
		points += 25
		g.remaining = math.Min(75, g.remaining+3)
		g.addPopup("+3 SECONDS", f.Pos.Add(Vec2{0, -75}), lemonYellow, 22, 1.1)
	}
	g.score += points
	g.addPopup(fmt.Sprintf("+%d", points), f.Pos.Add(Vec2{0, -43}), white, 30, 0.72)
	if g.chain == 3 || g.chain == 5 || g.chain == 8 || g.chain == 12 {
		g.addPopup(fmt.Sprintf("%d HIT COMBO!", g.chain), f.Pos.Add(Vec2{0, -89}), teal, 22, 1.0)
		g.sound.play(sndCombo, g.muted)
	} else {
		g.sound.play(sndSlice, g.muted)
	}
	normal := Vec2{-math.Sin(angle), math.Cos(angle)}
	for side := 0; side < 2; side++ {
		direction := -1.0
		if side == 1 {
			direction = 1
		}
		separation := normal.Mul(direction * 42)
		h := fruitHalf{
			Kind: f.Kind, Side: side, Pos: f.Pos.Add(normal.Mul(direction * 5)),
			Vel:    f.Vel.Mul(0.42).Add(separation).Add(Vec2{(g.rng.Float64() - 0.5) * 45, -55 - g.rng.Float64()*65}),
			Radius: f.Radius, Angle: angle, Spin: direction * (2.1 + g.rng.Float64()*2.1), Lifetime: 1.7,
		}
		g.halves = append(g.halves, h)
	}
	for i := 0; i < 25; i++ {
		v := g.rng.Float64() * math.Pi * 2
		speed := 60 + g.rng.Float64()*340
		g.particles = append(g.particles, particle{
			Pos:    f.Pos.Add(Vec2{(g.rng.Float64() - 0.5) * 22, (g.rng.Float64() - 0.5) * 22}),
			Vel:    Vec2{math.Cos(v) * speed, math.Sin(v)*speed - 35},
			Radius: 2.0 + g.rng.Float64()*5, Life: 0.44 + g.rng.Float64()*0.5,
			MaxLife: 0.95, Gravity: 330, Color: fruitPalettes[f.Kind].Juice,
			Streak: i%4 == 0,
		})
	}
	g.shake = math.Max(g.shake, 2.8)
}

func (g *Game) explode(pos Vec2) {
	g.takeLife(pos.X, false)
	g.shake = 16
	g.redFlash = 0.52
	g.sound.play(sndBomb, g.muted)
	g.addPopup("BOMB! -1 LIFE", pos.Add(Vec2{0, -65}), coral, 27, 1.2)
	for i := 0; i < 55; i++ {
		a := g.rng.Float64() * 2 * math.Pi
		speed := 100 + g.rng.Float64()*440
		c := coral
		if i%3 == 0 {
			c = lemonYellow
		}
		if i%7 == 0 {
			c = white
		}
		g.particles = append(g.particles, particle{
			Pos: pos, Vel: Vec2{math.Cos(a) * speed, math.Sin(a) * speed},
			Life: 0.34 + g.rng.Float64()*0.48, MaxLife: 0.82,
			Radius: 2 + g.rng.Float64()*6, Gravity: 240, Color: c, Streak: i%2 == 0,
		})
	}
}

func (g *Game) loseLife(x float64) {
	g.takeLife(x, true)
}

func (g *Game) takeLife(x float64, missed bool) {
	if g.lives <= 0 {
		return
	}
	g.lives--
	g.shake = math.Max(g.shake, 8)
	g.redFlash = math.Max(g.redFlash, 0.22)
	if missed {
		g.addPopup("MISSED!", Vec2{clamp(x, 100, screenWidth-100), 600}, coral, 20, 0.8)
	}
	if g.lives <= 0 {
		g.endRound()
	}
}

func (g *Game) addPopup(s string, p Vec2, c color.RGBA, size int, lifetime float64) {
	g.popups = append(g.popups, floatingText{
		Text: s, Pos: p, Size: size, Color: c,
		Life: lifetime, MaxLife: lifetime,
	})
}
