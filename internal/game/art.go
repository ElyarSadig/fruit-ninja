package game

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	spriteSize   = 160
	spriteCenter = 80.0
	spriteRadius = 61.0
)

type artwork struct {
	Background *ebiten.Image
	Glow       *ebiten.Image
	Whole      [fruitKindCount]*ebiten.Image
	Top        [fruitKindCount]*ebiten.Image
	Bottom     [fruitKindCount]*ebiten.Image
}

func newArtwork() *artwork {
	a := &artwork{}
	a.Background = generateBackground()
	a.Glow = generateGlow()
	for kind := fruitKind(0); kind < fruitKindCount; kind++ {
		a.Whole[kind] = generateFruitSprite(kind)
		if kind != bomb {
			a.Top[kind] = generateHalfSprite(kind, 0)
			a.Bottom[kind] = generateHalfSprite(kind, 1)
		}
	}
	return a
}

// paint applies a straight-RGB color onto the premultiplied-alpha image.RGBA.
// We keep the drawings generated at startup, never in the per-frame game loop.
func paint(dst *image.RGBA, x, y int, c color.RGBA, opacity float64) {
	if x < 0 || y < 0 || x >= dst.Bounds().Dx() || y >= dst.Bounds().Dy() {
		return
	}
	a := clamp(opacity*float64(c.A)/255, 0, 1)
	if a <= 0 {
		return
	}
	p := dst.PixOffset(x, y)
	inv := 1 - a
	dst.Pix[p] = uint8(clamp(float64(c.R)*a+float64(dst.Pix[p])*inv, 0, 255))
	dst.Pix[p+1] = uint8(clamp(float64(c.G)*a+float64(dst.Pix[p+1])*inv, 0, 255))
	dst.Pix[p+2] = uint8(clamp(float64(c.B)*a+float64(dst.Pix[p+2])*inv, 0, 255))
	dst.Pix[p+3] = uint8(clamp(255*a+float64(dst.Pix[p+3])*inv, 0, 255))
}

func shaded(c color.RGBA, amount float64) color.RGBA {
	return color.RGBA{
		R: uint8(clamp(float64(c.R)*amount, 0, 255)),
		G: uint8(clamp(float64(c.G)*amount, 0, 255)),
		B: uint8(clamp(float64(c.B)*amount, 0, 255)), A: c.A,
	}
}

func generateFruitSprite(kind fruitKind) *ebiten.Image {
	dst := image.NewRGBA(image.Rect(0, 0, spriteSize, spriteSize))
	pal := fruitPalettes[kind]
	for y := 0; y < spriteSize; y++ {
		for x := 0; x < spriteSize; x++ {
			dx := float64(x) - spriteCenter
			dy := float64(y) - spriteCenter - 3
			px, py := dx, dy
			if kind == lemon {
				px *= 1.03
				py *= 0.94
			}
			if kind == apple {
				px *= 0.97
				py *= 1.03
			}
			d := math.Hypot(px, py)
			if d > spriteRadius+1.5 {
				continue
			}
			aa := clamp((spriteRadius+0.5-d)/1.4, 0, 1)
			light := clamp(0.92-0.22*dx/60-0.20*dy/60, 0.58, 1.28)
			light += 0.018 * math.Sin(float64(x)*0.6+float64(y)*0.14)
			if kind == watermelon {
				stripe := math.Sin((dx + 8*math.Sin(dy*0.05)) * 0.18)
				light *= 0.89 + 0.12*math.Max(0, stripe)
			}
			if kind == orangeFruit || kind == lemon {
				light *= 0.98 + 0.024*math.Sin(float64(x)*1.4)*math.Sin(float64(y)*1.7)
			}
			if d > spriteRadius-4 {
				light *= 0.83
			}
			if kind == bomb {
				light = clamp(0.78-0.24*dx/60-0.22*dy/60, 0.42, 1.1)
				if d > spriteRadius-5 && d < spriteRadius-1 {
					light = 1.65
				}
			}
			paint(dst, x, y, shaded(pal.Skin, light), aa)

			// Glossy highlight, gently feathered into the skin.
			gl := math.Pow(math.Max(0, 1-math.Pow((dx+25)/21, 2)-math.Pow((dy+29)/12, 2)), 2)
			paint(dst, x, y, color.RGBA{255, 255, 245, 255}, aa*gl*0.43)
			gloss := math.Pow(math.Max(0, 1-math.Pow((dx+29)/8, 2)-math.Pow((dy+24)/5, 2)), 2)
			paint(dst, x, y, color.RGBA{255, 255, 255, 255}, aa*gloss*0.54)

			if kind == golden && d < 42 {
				// A radial six-point glint in the special time-bonus fruit.
				theta := math.Atan2(dy, dx)
				starR := 24.0 + 9.0*math.Cos(5*theta)
				if d < starR {
					a := clamp((starR-d)/2, 0, 1) * 0.88
					paint(dst, x, y, color.RGBA{255, 250, 190, 255}, a)
				}
			}
			if kind == bomb {
				ring := math.Abs(d - 48)
				if ring < 2.5 {
					paint(dst, x, y, color.RGBA{136, 159, 179, 255}, clamp((2.5-ring), 0, 1)*0.72)
				}
				// Striking orange-red crosshair / warning symbol.
				warningD := math.Hypot(dx, dy+1)
				if warningD < 26 {
					paint(dst, x, y, color.RGBA{247, 96, 106, 255}, clamp((27-warningD)/2, 0, 1))
					if warningD < 21 {
						paint(dst, x, y, color.RGBA{65, 45, 67, 255}, 1)
					}
					if math.Abs(dx) < 3.2 && dy > -13 && dy < 7 {
						paint(dst, x, y, color.RGBA{255, 201, 116, 255}, 1)
					}
					if math.Hypot(dx, dy-15) < 3.7 {
						paint(dst, x, y, color.RGBA{255, 201, 116, 255}, 1)
					}
				}
				for i := 0; i < 8; i++ {
					t := float64(i) * math.Pi / 4
					px, py := 46*math.Cos(t), 46*math.Sin(t)
					if math.Hypot(dx-px, dy-py) < 3.3 {
						paint(dst, x, y, color.RGBA{207, 219, 220, 255}, 0.9)
					}
				}
			}
		}
	}
	if kind == bomb {
		addStem(dst, true)
	} else {
		addStem(dst, false)
	}
	return ebiten.NewImageFromImage(dst)
}

func addStem(dst *image.RGBA, isBomb bool) {
	for y := 2; y < 45; y++ {
		for x := 30; x < 117; x++ {
			fx, fy := float64(x), float64(y)
			if isBomb {
				// Brass fuse ferrule and the glowing burning fuse.
				d := distPointToSegment(Vec2{fx, fy}, Vec2{79, 26}, Vec2{89, 8})
				if d < 5.5 && fy > 14 {
					paint(dst, x, y, color.RGBA{179, 128, 75, 255}, clamp(6.3-d, 0, 1))
				}
				if d < 2.9 {
					paint(dst, x, y, color.RGBA{254, 203, 104, 255}, clamp(3.2-d, 0, 1))
				}
				if math.Hypot(fx-89, fy-8) < 5.4 {
					paint(dst, x, y, color.RGBA{255, 138, 76, 255}, 0.95)
				}
				if math.Hypot(fx-89, fy-8) < 2.8 {
					paint(dst, x, y, color.RGBA{255, 250, 174, 255}, 1)
				}
				continue
			}
			// Fruit stem and a tilted green leaf. Both survive sprite rotation.
			d := distPointToSegment(Vec2{fx, fy}, Vec2{80, 34}, Vec2{83, 19})
			if d < 3 {
				paint(dst, x, y, color.RGBA{102, 75, 49, 255}, clamp(3.5-d, 0, 1))
			}
			// Rotated ellipse around (96,21).
			ex, ey := fx-96, fy-21
			u := ex*0.89 + ey*0.46
			v := -ex*0.46 + ey*0.89
			leafD := math.Sqrt((u*u)/(20*20) + (v*v)/(7*7))
			if leafD < 1.08 {
				c := color.RGBA{69, 203, 114, 255}
				if v > 0.3 {
					c = color.RGBA{35, 153, 104, 255}
				}
				paint(dst, x, y, c, clamp((1.06-leafD)*11, 0, 1))
			}
		}
	}
}

func generateHalfSprite(kind fruitKind, side int) *ebiten.Image {
	dst := image.NewRGBA(image.Rect(0, 0, spriteSize, spriteSize))
	pal := fruitPalettes[kind]
	for y := 0; y < spriteSize; y++ {
		for x := 0; x < spriteSize; x++ {
			dx, dy := float64(x)-spriteCenter, float64(y)-spriteCenter
			if side == 0 && dy > 0.5 {
				continue
			}
			if side == 1 && dy < -0.5 {
				continue
			}
			d := math.Hypot(dx, dy)
			if d > spriteRadius+1 {
				continue
			}
			outerAlpha := clamp(spriteRadius+0.5-d, 0, 1)
			straightAlpha := clamp(2.0-math.Abs(dy), 0, 1)
			bodyAlpha := outerAlpha
			light := clamp(1.06-0.22*dx/70-0.12*dy/70, 0.75, 1.2)
			flesh := shaded(pal.Flesh, light)
			// A darker rim, pale rind and lush fresh-cut flesh.
			if d > spriteRadius-4 {
				flesh = shaded(pal.Skin, 0.86)
			}
			if d > spriteRadius-10 && d <= spriteRadius-4 {
				flesh = pal.Rind
			}
			if d < 18 && kind == kiwi {
				flesh = color.RGBA{239, 244, 183, 255}
			}
			if kind == watermelon && d > spriteRadius-17 && d <= spriteRadius-10 {
				flesh = color.RGBA{242, 252, 211, 255}
			}
			if kind == orangeFruit || kind == lemon {
				t := math.Atan2(dy, dx)
				if d > 10 && math.Abs(math.Sin(t*6)) < 0.045 {
					flesh = shaded(pal.Rind, 1.04)
				}
			}
			if kind == plum && d > 30 && d < 40 && math.Sin(dx*0.2+dy*0.22) > 0.72 {
				flesh = shaded(pal.Flesh, 1.09)
			}
			paint(dst, x, y, flesh, bodyAlpha)
			if kind == watermelon {
				for _, p := range []Vec2{{-29, -15}, {-9, -29}, {15, -27}, {30, -10}, {-28, 16}, {-5, 32}, {17, 23}} {
					if math.Hypot(dx-p.X, dy-p.Y) < 3.0 {
						paint(dst, x, y, pal.Detail, bodyAlpha)
					}
				}
			} else if kind == kiwi {
				for j := 0; j < 15; j++ {
					t := 2 * math.Pi * float64(j) / 15
					p := Vec2{27 * math.Cos(t), 27 * math.Sin(t)}
					dd := math.Hypot(dx-p.X, dy-p.Y)
					if dd < 2.5 {
						paint(dst, x, y, pal.Detail, bodyAlpha)
					}
				}
			} else if kind == apple {
				for _, p := range []Vec2{{-8, -9}, {8, -9}, {-8, 10}, {8, 10}} {
					if math.Hypot(dx-p.X, dy-p.Y) < 3 {
						paint(dst, x, y, pal.Detail, bodyAlpha*0.88)
					}
				}
			} else if kind == golden {
				theta := math.Atan2(dy, dx)
				starR := 22 + 7*math.Cos(5*theta)
				if d < starR {
					paint(dst, x, y, color.RGBA{255, 255, 229, 255}, 0.7*bodyAlpha)
				}
			}
			// Highlight the fresh, flat, juicy cut edge.
			if straightAlpha > 0 && d < spriteRadius-7 {
				paint(dst, x, y, pal.Rind, straightAlpha*0.4*bodyAlpha)
			}
		}
	}
	return ebiten.NewImageFromImage(dst)
}

func distPointToSegment(p, a, b Vec2) float64 {
	d := b.Sub(a)
	len2 := d.X*d.X + d.Y*d.Y
	if len2 < 0.000001 {
		return p.Sub(a).Length()
	}
	t := clamp(((p.X-a.X)*d.X+(p.Y-a.Y)*d.Y)/len2, 0, 1)
	return p.Sub(a.Add(d.Mul(t))).Length()
}

func generateGlow() *ebiten.Image {
	const size = 256
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			d := math.Hypot(float64(x)-127.5, float64(y)-127.5) / 125
			a := math.Pow(math.Max(0, 1-d), 2) * 0.64
			paint(dst, x, y, color.RGBA{255, 255, 255, 255}, a)
		}
	}
	return ebiten.NewImageFromImage(dst)
}

func generateBackground() *ebiten.Image {
	dst := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	for y := 0; y < screenHeight; y++ {
		yf := float64(y) / screenHeight
		for x := 0; x < screenWidth; x++ {
			xf := float64(x) / screenWidth
			cyanGlow := math.Exp(-math.Pow((float64(x)-270)/530, 2) - math.Pow((float64(y)-400)/460, 2))
			roseGlow := math.Exp(-math.Pow((float64(x)-1060)/480, 2) - math.Pow((float64(y)-410)/425, 2))
			limeGlow := math.Exp(-math.Pow((float64(x)-640)/620, 2) - math.Pow((float64(y)-800)/330, 2))
			vignette := 1 - clamp(math.Pow((xf-0.5)*1.33, 2)+math.Pow((yf-0.5)*1.14, 2), 0, 0.5)*0.48
			grain := 2 * math.Sin(float64(x)*0.21+float64(y)*0.37) * math.Sin(float64(x)*0.12-float64(y)*0.27)
			r := (11 + 8*yf + 4*cyanGlow + 24*roseGlow + 5*limeGlow + grain) * vignette
			g := (20 + 7*yf + 23*cyanGlow + 4*roseGlow + 17*limeGlow + grain) * vignette
			b := (37 + 13*yf + 29*cyanGlow + 20*roseGlow + 12*limeGlow + grain) * vignette
			idx := dst.PixOffset(x, y)
			dst.Pix[idx] = uint8(clamp(r, 0, 255))
			dst.Pix[idx+1] = uint8(clamp(g, 0, 255))
			dst.Pix[idx+2] = uint8(clamp(b, 0, 255))
			dst.Pix[idx+3] = 255
		}
	}
	bg := ebiten.NewImageFromImage(dst)
	// Architectural rings, horizon lines, and tiny stars are part of the
	// cached background, so none of this work repeats every frame.
	for i := 0; i < 6; i++ {
		r := float32(215 + i*80)
		vector.StrokeCircle(bg, 640, 478, r, 1.0, color.RGBA{100, 188, 200, uint8(18 - i*2)}, true)
	}
	for y := 420; y <= screenHeight; y += 47 {
		vector.StrokeLine(bg, 0, float32(y), screenWidth, float32(y), 1, color.RGBA{126, 213, 232, 13}, true)
	}
	for x := 0; x <= screenWidth; x += 62 {
		vector.StrokeLine(bg, float32(x), 420, float32(x), screenHeight, 1, color.RGBA{126, 213, 232, 9}, true)
	}
	// Horizon accent and subtle decorative stars.
	vector.StrokeLine(bg, 0, 615, screenWidth, 615, 2, color.RGBA{114, 224, 204, 28}, true)
	for i := 0; i < 65; i++ {
		x := float32((i*371 + 83) % screenWidth)
		y := float32((i*197 + 29) % 580)
		r := float32(1 + (i % 3))
		alpha := uint8(40 + (i%5)*15)
		vector.FillCircle(bg, x, y, r, color.RGBA{190, 232, 240, alpha}, true)
		if i%11 == 0 {
			vector.StrokeLine(bg, x-6, y, x+6, y, 1, color.RGBA{146, 246, 235, 42}, true)
			vector.StrokeLine(bg, x, y-6, x, y+6, 1, color.RGBA{146, 246, 235, 42}, true)
		}
	}
	return bg
}
