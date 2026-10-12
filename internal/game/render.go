package game

import (
	"fmt"
	"image/color"
	"math"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	etext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}
	screen.DrawImage(g.art.Background, nil)
	g.drawAmbience(screen)
	if g.mode == stateMenu {
		g.drawMenu(screen)
		return
	}

	g.scene.Clear()
	g.drawWorld(g.scene)
	shakeX := math.Sin(g.clock*93) * g.shake * 0.54
	shakeY := math.Cos(g.clock*127) * g.shake * 0.37
	o := &ebiten.DrawImageOptions{}
	o.GeoM.Translate(shakeX, shakeY)
	screen.DrawImage(g.scene, o)
	g.drawHUD(screen)

	if g.redFlash > 0 {
		alpha := uint8(clamp(g.redFlash*90, 0, 90))
		vector.FillRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{255, 52, 78, alpha}, false)
	}
	switch g.mode {
	case statePaused:
		g.drawPause(screen)
	case stateOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) drawAmbience(dst *ebiten.Image) {
	// The soft-moving orbs make the scene feel alive even when nothing is cut.
	for i := 0; i < 11; i++ {
		x := float32(70 + i*113)
		y := float32(210+(i*127)%430) + float32(math.Sin(g.clock*(0.30+float64(i)*0.044)+float64(i))*15)
		c := color.RGBA{104, 243, 213, 21}
		if i%3 == 0 {
			c = color.RGBA{245, 108, 160, 22}
		}
		vector.StrokeCircle(dst, x, y, float32(24+i%5*15), 1.1, c, true)
	}
}

func drawRounded(dst *ebiten.Image, x, y, w, h, r float32, c color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	r = float32(clamp(float64(r), 0, math.Min(float64(w), float64(h))*0.5))
	if r < 0.5 {
		vector.FillRect(dst, x, y, w, h, c, false)
		return
	}
	// A single antialiased filled path avoids double-blended translucent
	// overlaps between corner circles and the center rectangles.
	const k float32 = 0.55228475
	p := &vector.Path{}
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.CubicTo(x+w-r+k*r, y, x+w, y+r-k*r, x+w, y+r)
	p.LineTo(x+w, y+h-r)
	p.CubicTo(x+w, y+h-r+k*r, x+w-r+k*r, y+h, x+w-r, y+h)
	p.LineTo(x+r, y+h)
	p.CubicTo(x+r-k*r, y+h, x, y+h-r+k*r, x, y+h-r)
	p.LineTo(x, y+r)
	p.CubicTo(x, y+r-k*r, x+r-k*r, y, x+r, y)
	p.Close()
	options := &vector.DrawPathOptions{AntiAlias: true}
	options.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, p, nil, options)
}

func drawPanel(dst *ebiten.Image, x, y, w, h, r float32, fill, border color.RGBA) {
	drawRounded(dst, x, y+7, w, h, r, color.RGBA{3, 9, 22, 80})
	drawRounded(dst, x, y, w, h, r, border)
	drawRounded(dst, x+1.7, y+1.7, w-3.4, h-3.4, r-1.1, fill)
}

func (g *Game) write(dst *ebiten.Image, s string, x, y float64, size int, c color.RGBA, bold bool, align etext.Align) {
	source := g.regular
	if bold {
		source = g.bold
	}
	face := &etext.GoTextFace{Source: source, Size: float64(size)}
	op := &etext.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.PrimaryAlign = align
	op.SecondaryAlign = etext.AlignCenter
	op.ColorScale.ScaleWithColor(c)
	etext.Draw(dst, s, face, op)
}

func (g *Game) shadowText(dst *ebiten.Image, s string, x, y float64, size int, c color.RGBA, align etext.Align) {
	g.write(dst, s, x+3, y+5, size, color.RGBA{2, 10, 27, 190}, true, align)
	g.write(dst, s, x, y, size, c, true, align)
}

func (g *Game) drawButton(dst *ebiten.Image, x, y, w, h float32, label string, hot bool) {
	if hot {
		drawRounded(dst, x-5, y-5, w+10, h+10, h/2+5, color.RGBA{86, 241, 212, 25})
	}
	drawRounded(dst, x, y+5, w, h, h/2, color.RGBA{3, 49, 54, 180})
	base := color.RGBA{102, 234, 205, 255}
	if hot {
		base = color.RGBA{156, 255, 225, 255}
	}
	drawRounded(dst, x, y, w, h, h/2, base)
	drawRounded(dst, x+4, y+3, w-8, h/2-3, h/3, color.RGBA{201, 255, 236, 30})
	g.write(dst, label, float64(x+w/2), float64(y+h/2), 23, color.RGBA{16, 54, 67, 255}, true, etext.AlignCenter)
	arrowX, arrowY := x+w-37, y+h/2
	arrowColor := color.RGBA{17, 73, 75, 255}
	vector.StrokeLine(dst, arrowX-5, arrowY-7, arrowX+2, arrowY, 3, arrowColor, true)
	vector.StrokeLine(dst, arrowX+2, arrowY, arrowX-5, arrowY+7, 3, arrowColor, true)
}

func (g *Game) drawMenu(dst *ebiten.Image) {
	t := g.clock
	// Rotating fruit props and their ambient glows frame the menu card.
	decorations := []struct {
		kind           fruitKind
		x, y, r, phase float64
	}{
		{watermelon, 194, 326, 95, 0.2}, {orangeFruit, 1033, 268, 79, 1.6},
		{plum, 151, 617, 57, 2.4}, {kiwi, 1096, 576, 72, 3.4},
		{lemon, 350, 610, 48, 2.9}, {golden, 944, 143, 39, 4.0},
	}
	for _, d := range decorations {
		y := d.y + math.Sin(t*1.12+d.phase)*12
		x := d.x + math.Cos(t*0.61+d.phase)*7
		g.drawGlow(dst, Vec2{x, y}, d.r*2.55, fruitPalettes[d.kind].Juice, 0.3)
		g.drawSprite(dst, g.art.Whole[d.kind], Vec2{x, y}, d.r, math.Sin(t*0.43+d.phase)*0.17, 1)
	}
	vector.StrokeLine(dst, 125, 444, 358, 482, 4, color.RGBA{96, 242, 219, 67}, true)
	vector.StrokeLine(dst, 955, 440, 1170, 392, 4, color.RGBA{255, 143, 164, 67}, true)

	drawPanel(dst, 348, 137, 584, 466, 31, color.RGBA{14, 28, 48, 230}, color.RGBA{90, 194, 190, 95})
	drawRounded(dst, 504, 169, 272, 31, 15, color.RGBA{91, 241, 213, 24})
	g.write(dst, "THE ULTIMATE SLICE ARCADE", 640, 184, 14, teal, true, etext.AlignCenter)
	g.shadowText(dst, "FRUIT", 640, 270, 95, white, etext.AlignCenter)
	g.shadowText(dst, "FURY", 640, 362, 101, lemonYellow, etext.AlignCenter)
	// A diagonal blade strike crossing the logotype.
	vector.StrokeLine(dst, 418, 370, 864, 249, 19, color.RGBA{70, 242, 223, 25}, true)
	vector.StrokeLine(dst, 415, 372, 862, 250, 7, color.RGBA{100, 249, 231, 176}, true)
	vector.StrokeLine(dst, 417, 370, 862, 250, 2, color.RGBA{246, 255, 253, 243}, true)
	g.write(dst, "SLICE FRUIT. DODGE BOMBS. GET THE HIGH SCORE.", 640, 426, 15, muted, true, etext.AlignCenter)
	drawRounded(dst, 512, 451, 256, 27, 13, color.RGBA{254, 255, 255, 11})
	g.write(dst, "60 SECONDS  /  3 LIVES", 640, 465, 14, lemonYellow, true, etext.AlignCenter)
	g.drawButton(dst, 470, 497, 340, 71, "START SLICING", pointInRect(g.cursor, 470, 497, 340, 71))
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		g.write(dst, "DRAG YOUR FINGER TO SLICE", 640, 633, 18, white, true, etext.AlignCenter)
		g.write(dst, "TAP PAUSE TO REST   /   TAP SOUND TO MUTE", 640, 667, 15, muted, false, etext.AlignCenter)
	} else {
		g.write(dst, "HOLD LEFT MOUSE BUTTON + SWIPE", 640, 633, 18, white, true, etext.AlignCenter)
		g.write(dst, "P TO PAUSE   •   M TO MUTE   •   ESC TO PAUSE", 640, 667, 15, muted, false, etext.AlignCenter)
	}
	g.write(dst, fmt.Sprintf("PERSONAL BEST  %d", g.best), 36, 42, 18, lemonYellow, true, etext.AlignStart)
	g.write(dst, "MADE WITH GO + EBITENGINE", 1241, 42, 14, muted, true, etext.AlignEnd)
}

func (g *Game) drawWorld(dst *ebiten.Image) {
	for _, h := range g.halves {
		sprite := g.art.Top[h.Kind]
		if h.Side == 1 {
			sprite = g.art.Bottom[h.Kind]
		}
		opacity := clamp((h.Lifetime-h.Age)*3, 0, 1)
		g.drawSprite(dst, sprite, h.Pos, h.Radius, h.Angle, opacity)
	}
	for _, f := range g.fruits {
		if f.Kind == golden {
			pulse := 0.65 + 0.25*math.Sin(g.clock*7)
			g.drawGlow(dst, f.Pos, f.Radius*3.35, lemonYellow, pulse)
		}
		if f.Kind == bomb {
			g.drawGlow(dst, f.Pos, f.Radius*2.5, coral, 0.27)
		}
		g.drawSprite(dst, g.art.Whole[f.Kind], f.Pos, f.Radius, f.Angle, 1)
	}
	for _, p := range g.particles {
		life := clamp(p.Life/p.MaxLife, 0, 1)
		radius := float32(math.Max(0.5, p.Radius*life))
		c := fade(p.Color, life)
		vector.FillCircle(dst, float32(p.Pos.X), float32(p.Pos.Y), radius, c, true)
		if p.Streak {
			vector.StrokeLine(dst, float32(p.Pos.X), float32(p.Pos.Y),
				float32(p.Pos.X-p.Vel.X*0.025), float32(p.Pos.Y-p.Vel.Y*0.025),
				float32(math.Max(1, float64(radius)*0.65)), fade(c, 0.52), true)
		}
	}
	for _, p := range g.popups {
		a := clamp(p.Life/p.MaxLife, 0, 1)
		c := fade(p.Color, math.Min(1, a*3))
		g.write(dst, p.Text, p.Pos.X+1.5, p.Pos.Y+2, p.Size, color.RGBA{7, 16, 33, uint8(float64(c.A) * 0.8)}, true, etext.AlignCenter)
		g.write(dst, p.Text, p.Pos.X, p.Pos.Y, p.Size, c, true, etext.AlignCenter)
	}
	g.drawBlade(dst)
}

func (g *Game) drawSprite(dst *ebiten.Image, sprite *ebiten.Image, pos Vec2, radius, angle, opacity float64) {
	if sprite == nil {
		return
	}
	o := &ebiten.DrawImageOptions{}
	o.Filter = ebiten.FilterLinear
	scale := radius / spriteRadius
	o.GeoM.Translate(-spriteCenter, -spriteCenter)
	o.GeoM.Scale(scale, scale)
	o.GeoM.Rotate(angle)
	o.GeoM.Translate(pos.X, pos.Y)
	o.ColorScale.ScaleAlpha(float32(clamp(opacity, 0, 1)))
	dst.DrawImage(sprite, o)
}

func (g *Game) drawGlow(dst *ebiten.Image, pos Vec2, radius float64, c color.RGBA, opacity float64) {
	o := &ebiten.DrawImageOptions{}
	o.Filter = ebiten.FilterLinear
	o.GeoM.Translate(-128, -128)
	o.GeoM.Scale(radius/128, radius/128)
	o.GeoM.Translate(pos.X, pos.Y)
	o.ColorScale.ScaleWithColor(fade(c, opacity))
	dst.DrawImage(g.art.Glow, o)
}

func (g *Game) drawBlade(dst *ebiten.Image) {
	if len(g.trail) > 1 {
		for i := 1; i < len(g.trail); i++ {
			a, b := g.trail[i-1], g.trail[i]
			freshness := 1 - clamp(math.Max(a.Age, b.Age)/0.24, 0, 1)
			if freshness <= 0 {
				continue
			}
			width := float32(5 + 10*freshness)
			x1, y1 := float32(a.Pos.X), float32(a.Pos.Y)
			x2, y2 := float32(b.Pos.X), float32(b.Pos.Y)
			vector.StrokeLine(dst, x1, y1, x2, y2, width+16, color.RGBA{50, 254, 230, uint8(46 * freshness)}, true)
			vector.StrokeLine(dst, x1, y1, x2, y2, width, color.RGBA{63, 236, 216, uint8(188 * freshness)}, true)
			vector.StrokeLine(dst, x1, y1, x2, y2, 2.8, color.RGBA{251, 255, 252, uint8(245 * freshness)}, true)
		}
	}
	if g.mode == statePlaying && g.held {
		x, y := float32(g.cursor.X), float32(g.cursor.Y)
		vector.StrokeCircle(dst, x, y, 13, 1.9, color.RGBA{221, 255, 243, 165}, true)
		vector.FillCircle(dst, x, y, 2.8, color.RGBA{246, 255, 246, 243}, true)
	}
}

func (g *Game) drawHUD(dst *ebiten.Image) {
	drawPanel(dst, 26, 23, 225, 88, 20, color.RGBA{16, 32, 54, 222}, color.RGBA{98, 225, 215, 51})
	g.write(dst, "SCORE", 46, 44, 14, teal, true, etext.AlignStart)
	g.write(dst, fmt.Sprintf("%06d", g.score), 44, 84, 37, white, true, etext.AlignStart)

	drawPanel(dst, 520, 23, 240, 88, 20, color.RGBA{16, 32, 54, 222}, color.RGBA{255, 227, 143, 69})
	g.write(dst, "TIME LEFT", 640, 44, 14, muted, true, etext.AlignCenter)
	c := white
	if g.remaining <= 10 {
		c = coral
	}
	g.write(dst, fmt.Sprintf("%02d", int(math.Ceil(g.remaining))), 640, 84, 39, c, true, etext.AlignCenter)
	drawRounded(dst, 555, 100, 170, 4, 2, color.RGBA{105, 136, 152, 100})
	drawRounded(dst, 555, 100, float32(170*clamp(g.remaining/60, 0, 1)), 4, 2, teal)

	drawPanel(dst, 964, 23, 208, 88, 20, color.RGBA{16, 32, 54, 222}, color.RGBA{255, 112, 143, 67})
	g.write(dst, "LIVES", 986, 44, 14, muted, true, etext.AlignStart)
	for i := 0; i < 3; i++ {
		active := i < g.lives
		g.drawHeart(dst, float32(997+i*51), 80, active)
	}
	pauseHot := pointInRect(g.cursor, 1189, 24, 58, 58)
	border := color.RGBA{103, 206, 206, 70}
	if pauseHot {
		border = color.RGBA{110, 244, 219, 160}
	}
	drawPanel(dst, 1189, 24, 58, 58, 17, color.RGBA{18, 38, 60, 239}, border)
	vector.FillRect(dst, 1207, 43, 6, 22, teal, false)
	vector.FillRect(dst, 1222, 43, 6, 22, teal, false)

	if g.chain >= 3 && g.elapsed-g.lastSlice < 0.66 {
		drawRounded(dst, 506, 136, 268, 34, 17, color.RGBA{86, 238, 210, 36})
		g.write(dst, fmt.Sprintf("%d × JUICY COMBO", g.chain), 640, 153, 17, teal, true, etext.AlignCenter)
	}
	if g.mode == statePlaying {
		g.write(dst, "SWIPE TO SLICE    /    AVOID THE BOMBS", 640, 692, 14, color.RGBA{161, 207, 223, 159}, true, etext.AlignCenter)
	}
	soundHot := pointInRect(g.cursor, 18, 647, 170, 61)
	soundColor := color.RGBA{25, 45, 66, 207}
	if soundHot {
		soundColor = color.RGBA{37, 71, 86, 225}
	}
	drawRounded(dst, 18, 657, 170, 46, 16, soundColor)
	g.write(dst, "SOUND  "+map[bool]string{true: "OFF", false: "ON"}[g.muted], 103, 680, 16, white, true, etext.AlignCenter)
}

func (g *Game) drawHeart(dst *ebiten.Image, x, y float32, active bool) {
	colorHeart := color.RGBA{255, 106, 137, 255}
	if !active {
		colorHeart = color.RGBA{66, 83, 108, 235}
	}
	// A rounded heart formed from two lobes and a filled vector path.
	path := &vector.Path{}
	path.MoveTo(x, y+12)
	path.CubicTo(x-14, y+1, x-17, y-10, x-9, y-11)
	path.CubicTo(x-3, y-15, x, y-8, x, y-6)
	path.CubicTo(x+5, y-15, x+16, y-15, x+17, y-6)
	path.CubicTo(x+19, y+1, x+8, y+10, x, y+12)
	path.Close()
	opts := &vector.DrawPathOptions{AntiAlias: true}
	opts.ColorScale.ScaleWithColor(colorHeart)
	vector.FillPath(dst, path, nil, opts)
	if active {
		vector.FillCircle(dst, x-5, y-7, 2.2, color.RGBA{255, 220, 225, 155}, true)
	}
}

func (g *Game) modalBackground(dst *ebiten.Image) {
	vector.FillRect(dst, 0, 0, screenWidth, screenHeight, color.RGBA{6, 12, 29, 185}, false)
}

func (g *Game) drawPause(dst *ebiten.Image) {
	g.modalBackground(dst)
	drawPanel(dst, 366, 146, 548, 439, 31, color.RGBA{18, 34, 53, 246}, color.RGBA{109, 220, 213, 117})
	drawRounded(dst, 535, 181, 210, 33, 16, color.RGBA{97, 236, 216, 24})
	g.write(dst, "TAKE A BREATHER", 640, 198, 15, teal, true, etext.AlignCenter)
	g.shadowText(dst, "PAUSED", 640, 291, 67, white, etext.AlignCenter)
	g.write(dst, "YOUR FRUIT CAN WAIT.", 640, 351, 20, muted, false, etext.AlignCenter)
	g.write(dst, fmt.Sprintf("SCORE   %d", g.score), 640, 386, 24, lemonYellow, true, etext.AlignCenter)
	g.drawButton(dst, 484, 413, 312, 69, "RESUME GAME", pointInRect(g.cursor, 484, 413, 312, 69))
	g.write(dst, "BACK TO MENU", 640, 533, 19, muted, true, etext.AlignCenter)
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		g.write(dst, "TAP RESUME TO KEEP SLICING", 640, 615, 15, muted, false, etext.AlignCenter)
	} else {
		g.write(dst, "PRESS P OR ESC TO RESUME", 640, 615, 15, muted, false, etext.AlignCenter)
	}
}

func (g *Game) drawGameOver(dst *ebiten.Image) {
	g.modalBackground(dst)
	drawPanel(dst, 361, 109, 558, 547, 32, color.RGBA{18, 32, 53, 249}, color.RGBA{243, 157, 141, 99})
	drawRounded(dst, 506, 142, 268, 34, 17, color.RGBA{255, 176, 130, 31})
	title := "ROUND COMPLETE"
	if g.lives == 0 {
		title = "OUT OF LIVES"
	}
	g.write(dst, title, 640, 159, 15, lemonYellow, true, etext.AlignCenter)
	g.shadowText(dst, "NICE SLICING!", 640, 221, 52, white, etext.AlignCenter)
	g.write(dst, "YOUR FINAL SCORE", 640, 284, 17, muted, true, etext.AlignCenter)
	g.shadowText(dst, fmt.Sprintf("%d", g.score), 640, 355, 75, teal, etext.AlignCenter)
	vector.StrokeLine(dst, 464, 401, 816, 401, 1, color.RGBA{129, 210, 204, 70}, true)
	g.write(dst, fmt.Sprintf("BEST  %d", g.best), 547, 433, 20, lemonYellow, true, etext.AlignCenter)
	g.write(dst, fmt.Sprintf("FRUITS  %d", g.totalSliced), 737, 433, 20, white, true, etext.AlignCenter)
	g.drawButton(dst, 468, 502, 344, 69, "PLAY AGAIN", pointInRect(g.cursor, 468, 502, 344, 69))
	g.write(dst, "MAIN MENU", 640, 608, 18, muted, true, etext.AlignCenter)
}
