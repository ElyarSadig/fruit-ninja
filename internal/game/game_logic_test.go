package game

import (
	"math/rand"
	"testing"
)

func TestFruitSliceScoresAndSplits(t *testing.T) {
	g := &Game{rng: rand.New(rand.NewSource(1)), remaining: 60, lastSlice: -100, mode: statePlaying}
	first := fruit{Kind: watermelon, Pos: Vec2{100, 200}, Vel: Vec2{60, -200}, Radius: 42}
	g.sliceFruit(first, 0)
	if g.score != 10 || g.totalSliced != 1 || len(g.halves) != 2 || len(g.particles) != 25 {
		t.Fatalf("one slice: score=%d sliced=%d halves=%d particles=%d", g.score, g.totalSliced, len(g.halves), len(g.particles))
	}
	g.elapsed = 0.1
	g.sliceFruit(first, 0)
	g.elapsed = 0.2
	g.sliceFruit(first, 0)
	// The third consecutive hit earns an extra 3*3 points.
	if g.score != 39 || g.chain != 3 {
		t.Fatalf("combo score=%d chain=%d, want 39 and 3", g.score, g.chain)
	}
}

func TestGoldenFruitAddsTime(t *testing.T) {
	g := &Game{rng: rand.New(rand.NewSource(3)), remaining: 12, lastSlice: -100, mode: statePlaying}
	g.sliceFruit(fruit{Kind: golden, Pos: Vec2{300, 400}, Radius: 37}, 0.5)
	if g.score != 35 || g.remaining != 15 {
		t.Fatalf("gold score %d time %.2f", g.score, g.remaining)
	}
}

func TestFastSwipeHitsMultipleFruits(t *testing.T) {
	g := &Game{rng: rand.New(rand.NewSource(8)), remaining: 60, lastSlice: -100, mode: statePlaying}
	g.fruits = []fruit{
		{Kind: apple, Pos: Vec2{200, 350}, Radius: 39},
		{Kind: lemon, Pos: Vec2{500, 350}, Radius: 41},
		{Kind: kiwi, Pos: Vec2{800, 450}, Radius: 40},
	}
	g.slash(Vec2{100, 350}, Vec2{620, 350})
	if g.score != 20 || len(g.fruits) != 1 || len(g.halves) != 4 {
		t.Fatalf("slash score=%d remaining=%d halves=%d", g.score, len(g.fruits), len(g.halves))
	}
}

func TestSpawnWaveBounds(t *testing.T) {
	g := &Game{rng: rand.New(rand.NewSource(22)), elapsed: 38}
	for i := 0; i < 20; i++ {
		g.spawnWave()
	}
	if len(g.fruits) == 0 {
		t.Fatal("no spawned fruits")
	}
	for _, f := range g.fruits {
		if f.Pos.X < 75 || f.Pos.X > screenWidth-75 || f.Vel.Y >= 0 || f.Radius < 35 || f.Radius > 48 {
			t.Fatalf("invalid spawned fruit %+v", f)
		}
	}
}
