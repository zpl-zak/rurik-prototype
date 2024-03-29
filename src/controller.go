package main

import (
	"github.com/zaklaus/rurik/src/core"
	"github.com/zaklaus/rurik/src/system"
)

type characterController struct {
	Object *core.Object
	physicsProps
}

func (c *characterController) move(factor float32) {
	if c.IsFalling && factor == 0 {
		return
	}

	speed := movementSpeed

	if c.IsFalling {
		speed *= movementFallSpeedFactor
	}

	c.Object.Movement.X = core.ScalarLerp(
		c.Object.Movement.X,
		factor*speed*system.FrameTime,
		movementSmoothingFactor,
	)
}

func (c *characterController) jump() {
	if c.IsGrounded {
		c.Object.Movement.Y = -jumpForce * system.FrameTime
	}

	if c.IsInWater {
		c.Object.Movement.Y = -upwardWaterForce * system.FrameTime
	}

	if c.IsOnLadder {
		c.Object.Movement.Y = -ladderClimbSpeed * system.FrameTime
	}
}

func (c *characterController) down() {
	if c.IsOnLadder {
		c.Object.Movement.Y = ladderClimbSpeed * system.FrameTime
	}

	if c.IsInWater {
		c.Object.Movement.Y = (upwardWaterForce + buoyancy/2) * system.FrameTime
	}
}

func (c *characterController) update() {
	// Handle free fall
	{
		down, _ := core.CheckForCollisionEx([]uint32{}, c.Object, 0, 4)
		c.IsGrounded = down.Colliding() && !c.IsOnLadder
		if !c.IsGrounded && !c.IsInWater && !c.IsOnLadder {
			g := gravity

			c.Object.Movement.Y += g * system.FrameTime

			if c.Object.Movement.Y > maxFallSpeed {
				c.Object.Movement.Y = maxFallSpeed
			}
		}
	}

	x := core.RoundFloat(c.Object.Movement.X)
	y := core.RoundFloat(c.Object.Movement.Y)

	// Handle collision
	{
		// Handle slope movement
		if res, _ := core.CheckForCollisionEx([]uint32{core.CollisionSlope}, c.Object, core.RoundFloatToInt32(x), core.RoundFloatToInt32(y)+4); res.Colliding() {
			y, c.Object.Movement.Y = calculateContactResponse(&c.physicsProps, res.ResolveY)
		}

		// Handle solid+trigger collisions
		if res, _ := core.CheckForCollisionEx([]uint32{core.CollisionSolid, core.CollisionTrigger}, c.Object, core.RoundFloatToInt32(x), 0); res.Colliding() && !res.Teleporting {
			x, c.Object.Movement.X = calculateContactResponse(&c.physicsProps, res.ResolveX)
		}

		if res, _ := core.CheckForCollisionEx([]uint32{core.CollisionSolid, core.CollisionTrigger}, c.Object, 0, core.RoundFloatToInt32(y)); res.Colliding() && !res.Teleporting {
			y, c.Object.Movement.Y = calculateContactResponse(&c.physicsProps, res.ResolveY)
		}

		// Apply motion
		c.Object.Position.X += x
		c.Object.Position.Y += y
	}

	c.IsFalling = c.Object.Movement.Y > 0 && !c.IsInWater && !c.IsOnLadder

	if c.IsOnLadder {
		c.Object.Movement.Y = 0
	}

	if c.IsGettingOnLadder && c.IsFalling {
		c.IsGettingOnLadder = false
		c.IsOnLadder = true
	}
}
