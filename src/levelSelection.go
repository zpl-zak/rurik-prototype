package main

import (
	"fmt"
	"math"

	rl "github.com/zaklaus/raylib-go/raylib"

	"github.com/zaklaus/rurik/src/core"
	"github.com/zaklaus/rurik/src/system"
)

type level struct {
	title   string
	mapName string
}

var levelSelection struct {
	selectedChoice       int
	levels               []level
	waveTime             int32
	banner               string
	mouseDoublePressTime int32
}

func initLevels() {
	levelSelection.levels = []level{
		level{
			title:   "Intro scene",
			mapName: "intro",
		},
		level{
			title:   "------------",
			mapName: "",
		},
		level{
			title:   "Movement test",
			mapName: "movement",
		},
		level{
			title:   "Water particles",
			mapName: "water",
		},
		level{
			title:   "Exit game",
			mapName: "$exitGame",
		},
	}

	levelSelection.banner = "Debug level selection"
}

func (g *gameMode) drawLevelSelection() {
	levelSelection.waveTime = int32(math.Round(math.Sin(float64(rl.GetTime()) * 40)))

	width := system.ScreenWidth
	start := system.ScreenHeight / 2

	rl.DrawText(levelSelection.banner, 15, 30, 23, rl.RayWhite)

	// choices
	chsX := width / 2
	chsY := start + 40

	rl.DrawRectangle(chsX-120+levelSelection.waveTime, chsY-20, 240+levelSelection.waveTime, int32(len(levelSelection.levels))*15+40, rl.Fade(rl.Black, 0.25))

	if levelSelection.mouseDoublePressTime > 0 {
		levelSelection.mouseDoublePressTime -= int32(1000 * (system.FrameTime * float32(core.TimeScale)))
	} else if levelSelection.mouseDoublePressTime < 0 {
		levelSelection.mouseDoublePressTime = 0
	}

	var ySpacing int32 = 19

	if len(levelSelection.levels) > 0 {
		for idx, ch := range levelSelection.levels {
			ypos := chsY + int32(idx)*ySpacing - 2
			if idx == levelSelection.selectedChoice {
				rl.DrawRectangle(chsX-100, ypos, 200, ySpacing, rl.DarkPurple)
			}

			mapName := ""

			if ch.mapName != "" {
				mapName = fmt.Sprintf(" (%s)", ch.mapName)
			}

			core.DrawTextCentered(
				fmt.Sprintf("%s%s", ch.title, mapName),
				chsX,
				chsY+int32(idx)*ySpacing,
				16,
				rl.White,
			)

			if core.IsMouseInRectangle(chsX-100, ypos, 200, ySpacing) {
				if rl.IsMouseButtonDown(rl.MouseLeftButton) {
					rl.DrawRectangleLines(chsX-100, ypos, 200, ySpacing, rl.Pink)
				} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
					levelSelection.selectedChoice = idx

					if levelSelection.mouseDoublePressTime > 0 {
						g.playLevelSelection()
					} else {
						levelSelection.mouseDoublePressTime = MouseDoublePress
					}
				} else {
					rl.DrawRectangleLines(chsX-100, ypos, 200, ySpacing, rl.Purple)
				}
			}
		}
	}
}

func (g *gameMode) updateLevelSelection() {
	if system.IsKeyPressed("down") {
		levelSelection.selectedChoice++

		if levelSelection.selectedChoice >= len(levelSelection.levels) {
			levelSelection.selectedChoice = 0
		}
	}

	if system.IsKeyPressed("up") {
		levelSelection.selectedChoice--

		if levelSelection.selectedChoice < 0 {
			levelSelection.selectedChoice = len(levelSelection.levels) - 1
		}
	}

	if system.IsKeyPressed("use") {
		g.quests.quests = []quest{}
		g.playLevelSelection()

		//temp
		// g.quests.addQuest("TEST0", nil)
		// g.quests.addQuest("EXAMPLE", nil)
		// g.quests.addQuest("EVENTS", nil)
		// g.quests.callEvent("_TestIncrementCounter_", []int{120})
	}
}

func (g *gameMode) loadLevel(mapName string) {
	core.FlushMaps()
	core.LoadMap(mapName)
	core.InitMap()
}

func (g *gameMode) playLevelSelection() {
	mapName := levelSelection.levels[levelSelection.selectedChoice].mapName

	if mapName == "" {
		return
	}

	if mapName == "$exitGame" {
		core.CloseGame()
		return
	}

	g.loadLevel(mapName)
	g.playState = statePlay
}
