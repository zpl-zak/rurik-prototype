package main

import (
	"github.com/zaklaus/resolv/resolv"
	"github.com/zaklaus/rurik/src/core"
)

type ladder struct{}

// NewLadder ladder
func NewLadder(o *core.Object) {
	o.IsCollidable = true
	o.CollisionType = core.CollisionTrigger
	o.Size = []int32{int32(o.Meta.Width), int32(o.Meta.Height)}
	o.DebugVisible = true

	o.GetAABB = core.GetSolidAABB

	o.HandleCollisionEnter = func(res *resolv.Collision, o, other *core.Object) {
		switch v := other.UserData.(type) {
		case *player:
			v.ctrl.IsGettingOnLadder = true
			v.ctrl.IsOnLadder = false
		}
	}

	o.HandleCollisionLeave = func(res *resolv.Collision, o, other *core.Object) {
		switch v := other.UserData.(type) {
		case *player:
			v.ctrl.IsGettingOnLadder = false
			v.ctrl.IsOnLadder = false
		}
	}
}
