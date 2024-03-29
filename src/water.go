package main

import (
	"encoding/gob"
	"math/rand"

	rl "github.com/zaklaus/raylib-go/raylib"
	"github.com/zaklaus/raylib-go/raymath"

	"github.com/zaklaus/resolv/resolv"
	"github.com/zaklaus/rurik/src/core"
	"github.com/zaklaus/rurik/src/system"
)

const (
	waterTileSize  int32   = 4
	waterPushForce float32 = 10
	waterNoise     float32 = 0.65

	waterParticleMaxLifetime     float32 = 5
	waterParticleSpreadFactor    float32 = 3
	waterParticleSplashForce     float32 = 2
	waterParticleEnableCollision bool    = true
	waterParticleMinCount        int     = 6
	waterParticleMaxCount        int     = 10
)

var (
	waterCalm = rl.Color{R: 0, G: 120, B: 200, A: 255}
	waterWild = rl.Color{R: 80, G: 180, B: 240, A: 255}

	waterParticles []waterParticle
)

type water struct {
	gridWidth   int32
	gridHeight  int32
	gridSize    int32
	waveHeights [waterVertexCount]float32
	energy      []float32
}

type waterParticle struct {
	world     *core.World
	position  []rl.Vector2
	direction []rl.Vector2
	color     []rl.Color
	lifetime  float32
}

func (w *water) Serialize(enc *gob.Encoder)   {}
func (w *water) Deserialize(dec *gob.Decoder) {}

// NewWater water
func NewWater(o *core.Object) {
	o.IsCollidable = true
	o.CollisionType = core.CollisionTrigger
	o.Size = []int32{int32(o.Meta.Width), int32(o.Meta.Height)}
	o.DebugVisible = false
	o.IsOverlay = true

	waterGrid := &water{}
	waterGrid.gridWidth = (int32(o.Meta.Width) + int32(o.Meta.Width)%waterTileSize) / waterTileSize
	waterGrid.gridHeight = (int32(o.Meta.Height) + int32(o.Meta.Height)%waterTileSize) / waterTileSize
	waterGrid.gridSize = waterGrid.gridWidth * waterGrid.gridHeight

	for idx := 0; idx < int(waterGrid.gridSize); idx++ {
		waterGrid.energy = append(waterGrid.energy, 0)
	}

	o.UserData = waterGrid

	o.Update = func(o *core.Object, dt float32) {
		w := o.UserData.(*water)
		for _, v := range o.ContainedObjects {
			other := v.Object

			if rand.Int()%3 == 0 {
				xpos := int32(other.Position.X-o.Position.X) / waterTileSize
				ypos := int32(other.Position.Y-o.Position.Y) / waterTileSize

				if ypos < 1 {
					ypos = 1
				} else if ypos > w.gridHeight-1 {
					ypos = w.gridHeight - 2
				}

				if xpos < 1 {
					xpos = 1
				} else if xpos > w.gridWidth-1 {
					xpos = w.gridWidth - 2
				}

				idx := (ypos * w.gridWidth) + xpos
				w.energy[idx] = raymath.Vector2Length(other.Movement) * waterPushForce // * rand.Float32()
			}

			other.Movement.Y = core.ScalarLerp(other.Movement.Y, buoyancy*system.FrameTime, 0.30)
		}

		w.updateWater()
	}

	o.Draw = func(o *core.Object) {
		w := o.UserData.(*water)

		var y int32
		var x int32

		var waterEdge int32 = 1
		//var waveVertexMargin int32 = int32(float32(o.Meta.Width) / float32(waterVertexCount))

		for ; y < waterEdge; y++ {
			for x = 0; x < w.gridWidth; x++ {
				//col := w.getTileColor(x, y)
				w.drawWaterTile(o, x, y)
				//rl.DrawLineEx()
			}
			x = 0
		}

		for y = waterEdge; y < w.gridHeight; y++ {
			for x = 0; x < w.gridWidth; x++ {
				w.drawWaterTile(o, x, y)
			}

			x = 0
		}
	}

	o.GetAABB = core.GetSolidAABB

	o.HandleCollisionEnter = func(res *resolv.Collision, o, other *core.Object) {
		pushWaterParticle(o.GetWorld(), other.Position)
	}

}

func (w *water) getTileColor(x, y int32) rl.Color {
	idx := (y * w.gridWidth) + x
	en := w.energy[idx]
	if en > 1 {
		en = 1
	} else if en < -1 {
		en = 0
	}
	colV := raymath.Vector3Lerp(core.ColorToVec3(waterCalm), core.ColorToVec3(waterWild), en)
	col := core.Vec3ToColor(colV)
	col.A = 120
	return col
}

func (w *water) drawWaterTile(o *core.Object, x, y int32) {
	col := w.getTileColor(x, y)
	rl.DrawRectangle(
		int32(o.Position.X)+int32(x*waterTileSize),
		int32(o.Position.Y)+int32(y*waterTileSize),
		waterTileSize,
		waterTileSize,
		col,
	)
}

func (w *water) updateWater() {
	var y int32
	var x int32

	for ; y < w.gridHeight; y++ {
		for ; x < w.gridWidth; x++ {

			idx := (y * w.gridWidth) + x

			ie := x - 1
			iw := x + 1
			is := y + 1
			in := y - 1

			var ve, vw, vs, vn float32

			if ie < 0 {
				ve = 0
			} else {
				ve = w.energy[(y*w.gridWidth)+ie]
			}

			if is >= w.gridHeight-1 {
				vs = 0
			} else {
				vs = w.energy[(is*w.gridWidth)+x]
			}

			if iw >= w.gridWidth-1 {
				vw = 0
			} else {
				vw = w.energy[(y*w.gridWidth)+iw]
			}

			if in < 0 {
				vn = 0
			} else {
				vn = w.energy[(in*w.gridWidth)+x]
			}

			m := &w.energy[idx]

			if rand.Int()%4 == 0 {
				*m = rand.Float32() * waterNoise
			}

			*m = core.ScalarLerp(
				*m,
				(ve+vw+vs+vn+*m)/5,
				1,
			)
		}

		x = 0
	}
}

func pushWaterParticle(world *core.World, origin rl.Vector2) {
	part := waterParticle{}

	numParts := waterParticleMinCount + rand.Int()%int(waterParticleMaxCount-waterParticleMinCount)

	for i := 0; i < numParts; i++ {
		pos := origin
		posx := float32(i) / float32(numParts-1)
		dir := rl.Vector2{
			X: (posx*2 - 1) * waterParticleSpreadFactor,
			Y: -waterParticleSplashForce,
		}
		col := core.Vec3ToColor(raymath.Vector3Lerp(core.ColorToVec3(waterCalm), core.ColorToVec3(waterWild), rand.Float32()))

		part.position = append(part.position, pos)
		part.direction = append(part.direction, dir)
		part.color = append(part.color, col)
	}

	part.lifetime = waterParticleMaxLifetime
	part.world = world

	waterParticles = append(waterParticles, part)
}

func updateWaterParticles() {
	var newWaterParticles []waterParticle

	for _, v := range waterParticles {
		for idx := 0; idx < len(v.position); idx++ {
			p := &v.position[idx]
			d := &v.direction[idx]

			dy := float32(core.RoundFloatToInt32(d.Y))
			dx := float32(core.RoundFloatToInt32(d.X))

			if waterParticleEnableCollision && rand.Int()%3 == 0 {
				rect := rl.RectangleInt32{
					X:      core.RoundFloatToInt32(p.X),
					Y:      core.RoundFloatToInt32(p.Y),
					Width:  1,
					Height: 1,
				}

				if res, _ := core.CheckForCollisionRectangle(v.world, rect, []uint32{core.CollisionSolid, core.CollisionSlope}, int32(dx), 0); res.Colliding() && !res.Teleporting {
					dx = float32(res.ResolveX)
					d.X = float32(-res.ResolveX) / 2
				}

				if res, _ := core.CheckForCollisionRectangle(v.world, rect, []uint32{core.CollisionSolid, core.CollisionSlope}, 0, int32(dy)+4); res.Colliding() && !res.Teleporting {
					dy = float32(res.ResolveY)
					d.Y = float32(-res.ResolveY) / 4
				}
			}

			p.X += dx
			p.Y += dy

			d.Y += gravity * system.FrameTime
		}

		v.lifetime -= system.FrameTime

		if v.lifetime > 0 {
			newWaterParticles = append(newWaterParticles, v)
		}
	}

	waterParticles = newWaterParticles
}

func drawWaterParticles() {
	for _, v := range waterParticles {
		for idx := 0; idx < len(v.position); idx++ {
			p := v.position[idx]
			c := v.color[idx]
			c.A = 120
			rl.DrawRectangle(
				int32(p.X)-2,
				int32(p.Y)-2,
				2,
				2,
				c,
			)
		}
	}
}
