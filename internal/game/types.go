package game

import "image/color"

type fruitKind int

const (
	watermelon fruitKind = iota
	orangeFruit
	apple
	kiwi
	lemon
	plum
	golden
	bomb
	fruitKindCount
)

type palette struct {
	Name   string
	Skin   color.RGBA
	Flesh  color.RGBA
	Rind   color.RGBA
	Juice  color.RGBA
	Detail color.RGBA
}

var fruitPalettes = [fruitKindCount]palette{
	{"MELON", color.RGBA{39, 179, 108, 255}, color.RGBA{251, 76, 102, 255}, color.RGBA{174, 245, 155, 255}, color.RGBA{255, 77, 127, 255}, color.RGBA{58, 43, 55, 255}},
	{"ORANGE", color.RGBA{255, 158, 56, 255}, color.RGBA{255, 188, 96, 255}, color.RGBA{255, 228, 152, 255}, color.RGBA{255, 184, 71, 255}, color.RGBA{255, 248, 216, 255}},
	{"APPLE", color.RGBA{247, 76, 104, 255}, color.RGBA{255, 235, 194, 255}, color.RGBA{255, 246, 213, 255}, color.RGBA{255, 88, 125, 255}, color.RGBA{106, 64, 64, 255}},
	{"KIWI", color.RGBA{172, 119, 70, 255}, color.RGBA{144, 223, 78, 255}, color.RGBA{235, 236, 143, 255}, color.RGBA{162, 236, 100, 255}, color.RGBA{35, 62, 44, 255}},
	{"LEMON", color.RGBA{255, 213, 67, 255}, color.RGBA{255, 240, 153, 255}, color.RGBA{255, 250, 204, 255}, color.RGBA{254, 235, 85, 255}, color.RGBA{255, 255, 230, 255}},
	{"PLUM", color.RGBA{160, 92, 226, 255}, color.RGBA{239, 139, 210, 255}, color.RGBA{246, 187, 223, 255}, color.RGBA{209, 95, 235, 255}, color.RGBA{120, 54, 137, 255}},
	{"GOLD", color.RGBA{255, 206, 70, 255}, color.RGBA{255, 236, 126, 255}, color.RGBA{255, 247, 175, 255}, color.RGBA{255, 222, 92, 255}, color.RGBA{255, 255, 229, 255}},
	{"BOMB", color.RGBA{39, 51, 76, 255}, color.RGBA{87, 104, 128, 255}, color.RGBA{127, 143, 159, 255}, color.RGBA{255, 93, 111, 255}, color.RGBA{255, 213, 110, 255}},
}

var (
	white       = color.RGBA{245, 253, 255, 255}
	muted       = color.RGBA{145, 170, 194, 255}
	teal        = color.RGBA{97, 240, 217, 255}
	lemonYellow = color.RGBA{255, 224, 111, 255}
	coral       = color.RGBA{255, 112, 143, 255}
	navy        = color.RGBA{14, 24, 43, 255}
)

type playState uint8

const (
	stateMenu playState = iota
	statePlaying
	statePaused
	stateOver
)

type fruit struct {
	Kind                fruitKind
	Pos, Vel            Vec2
	Radius, Angle, Spin float64
	Age                 float64
}

type fruitHalf struct {
	Kind                fruitKind
	Side                int
	Pos, Vel            Vec2
	Radius, Angle, Spin float64
	Age, Lifetime       float64
}

type particle struct {
	Pos, Vel      Vec2
	Radius        float64
	Life, MaxLife float64
	Gravity       float64
	Color         color.RGBA
	Streak        bool
}

type floatingText struct {
	Text          string
	Pos           Vec2
	Color         color.RGBA
	Life, MaxLife float64
	Size          int
}

type trailPoint struct {
	Pos Vec2
	Age float64
}
