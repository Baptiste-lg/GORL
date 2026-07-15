package game

import (
	"math/rand"
	"syscall/js"
	"time"

	"github.com/Baptiste-lg/GORL/dungeon"
	"github.com/Baptiste-lg/GORL/render"
)

type State int

const (
	StateMenu State = iota
	StatePlaying
	StatePaused
	StateLevelUp
	StateDead
	StateInventory
)

const (
	fovRadius     = 8
	regenDelay    = 3.0
	regenInterval = 5.0
	lootDropRate  = 0.35
)

type Game struct {
	state    State
	renderer *render.Renderer
	lastTime float64
	keys     map[string]bool

	dungeonResult *dungeon.GenerateResult
	fov           *dungeon.FOV
	player        *Player
	world         *World
	particles     *render.ParticleSystem
	groundItems   []*Item
	floor         int
	rng           *rand.Rand
	killCount     int

	// Level-up state
	levelChoices    []StatBoost
	levelUpSelected int
}

func New() *Game {
	g := &Game{
		state:     StateMenu,
		renderer:  render.NewRenderer(),
		keys:      make(map[string]bool),
		particles: render.NewParticleSystem(),
		floor:     1,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.registerInput()
	return g
}

func (g *Game) startNewGame() {
	g.floor = 1
	g.killCount = 0
	g.player = nil
	g.particles = render.NewParticleSystem()
	g.groundItems = nil
	g.generateFloor()
	g.state = StatePlaying
}

func (g *Game) generateFloor() {
	seed := g.rng.Int63()
	g.dungeonResult = dungeon.Generate(seed)
	g.fov = dungeon.NewFOV(dungeon.MapWidth, dungeon.MapHeight)
	g.world = NewWorld(g.dungeonResult, g.floor, g.rng)
	g.groundItems = nil

	sx, sy := g.dungeonResult.SpawnX, g.dungeonResult.SpawnY
	if g.player == nil {
		g.player = NewPlayer(sx, sy)
	} else {
		g.player.X = sx
		g.player.Y = sy
	}

	g.renderer.CenterCamera(g.player.X, g.player.Y)
	g.fov.Compute(g.dungeonResult.Map, g.player.X, g.player.Y, fovRadius)
}

func (g *Game) nextFloor() {
	g.floor++
	g.generateFloor()
}

func (g *Game) registerInput() {
	doc := js.Global().Get("document")

	doc.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		g.keys[key] = true
		g.handleKeyDown(key)
		args[0].Call("preventDefault")
		return nil
	}))

	doc.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		g.keys[key] = false
		return nil
	}))
}

func (g *Game) handleKeyDown(key string) {
	switch g.state {
	case StateMenu:
		if key == "Enter" {
			g.startNewGame()
		}
	case StateDead:
		if key == "Enter" {
			g.state = StateMenu
		}
	case StatePlaying:
		switch key {
		case "i", "I":
			g.state = StateInventory
		case "Escape":
			g.state = StatePaused
		}
	case StateInventory:
		g.handleInventoryKey(key)
	case StatePaused:
		g.handlePauseKey(key)
	case StateLevelUp:
		g.handleLevelUpKey(key)
	}
}

func (g *Game) handlePauseKey(key string) {
	switch key {
	case "Escape", "r", "R":
		g.state = StatePlaying
	case "q", "Q":
		g.state = StateMenu
	}
}

func (g *Game) handleInventoryKey(key string) {
	inv := g.player.Inventory
	switch key {
	case "i", "I", "Escape":
		g.state = StatePlaying
	case "ArrowUp":
		if inv.Selected > 0 {
			inv.Selected--
		}
	case "ArrowDown":
		if inv.Selected < len(inv.Items)-1 {
			inv.Selected++
		}
	case "Enter":
		item := inv.SelectedItem()
		if item == nil {
			return
		}
		if item.Type == ItemWeapon || item.Type == ItemArmor {
			inv.Equip(inv.Selected)
		} else {
			g.player.UseItem(inv.Selected)
		}
	case "d", "D":
		item := inv.Remove(inv.Selected)
		if item != nil {
			item.X = g.player.X
			item.Y = g.player.Y
			g.groundItems = append(g.groundItems, item)
		}
	}
}

func (g *Game) handleLevelUpKey(key string) {
	switch key {
	case "ArrowUp":
		if g.levelUpSelected > 0 {
			g.levelUpSelected--
		}
	case "ArrowDown":
		if g.levelUpSelected < len(g.levelChoices)-1 {
			g.levelUpSelected++
		}
	case "Enter":
		if g.levelUpSelected >= 0 && g.levelUpSelected < len(g.levelChoices) {
			choice := g.levelChoices[g.levelUpSelected]
			choice.Apply(&g.player.Stats)
			g.player.Stats.Level++
			g.player.Stats.XP -= g.player.Stats.XPToNextLevel()
			if g.player.Stats.XP < 0 {
				g.player.Stats.XP = 0
			}
			// Refill HP on level up
			g.player.Stats.HP = g.player.EffectiveStats().MaxHP()
			g.state = StatePlaying
		}
	}
}

func (g *Game) checkLevelUp() {
	if g.player.Stats.CanLevelUp() {
		// Pick 3 random boosts (or all 4 if few enough)
		perm := g.rng.Perm(len(allBoosts))
		count := 3
		if count > len(allBoosts) {
			count = len(allBoosts)
		}
		g.levelChoices = make([]StatBoost, count)
		for i := 0; i < count; i++ {
			g.levelChoices[i] = allBoosts[perm[i]]
		}
		g.levelUpSelected = 0
		g.state = StateLevelUp
	}
}

func (g *Game) processMovement() {
	if g.player == nil || !g.player.CanMove() {
		return
	}

	dx, dy := 0, 0
	if g.keys["w"] || g.keys["ArrowUp"] {
		dy = -1
	} else if g.keys["s"] || g.keys["ArrowDown"] {
		dy = 1
	}
	if g.keys["a"] || g.keys["ArrowLeft"] {
		dx = -1
	} else if g.keys["d"] || g.keys["ArrowRight"] {
		dx = 1
	}

	if dx == 0 && dy == 0 {
		return
	}

	nx, ny := g.player.X+dx, g.player.Y+dy
	dm := g.dungeonResult.Map
	if !dm.At(nx, ny).Passable() {
		return
	}

	// Attack enemy if one is in the way
	enemy := g.world.EnemyAt(nx, ny)
	if enemy != nil {
		g.playerAttack(enemy)
		g.player.ResetMoveCooldown()
		return
	}

	g.player.X = nx
	g.player.Y = ny
	g.player.ResetMoveCooldown()
	g.renderer.CenterCamera(g.player.X, g.player.Y)
	g.fov.Compute(dm, g.player.X, g.player.Y, fovRadius)

	// Pick up items
	g.pickupItems()

	// Check stairs
	if dm.At(g.player.X, g.player.Y) == dungeon.TileStairsDown {
		g.nextFloor()
	}
}

func (g *Game) pickupItems() {
	remaining := g.groundItems[:0]
	for _, item := range g.groundItems {
		if item.X == g.player.X && item.Y == g.player.Y {
			if g.player.Inventory.Add(item) {
				g.particles.SpawnText(g.player.X, g.player.Y, item.Name, item.Rarity.Color())
			} else {
				remaining = append(remaining, item)
			}
		} else {
			remaining = append(remaining, item)
		}
	}
	g.groundItems = remaining
}

func (g *Game) playerAttack(enemy *Enemy) {
	attacker := *g.player.Entity
	attacker.Stats = g.player.EffectiveStats()
	result := ResolveAttack(&attacker, enemy.Entity, g.rng)
	g.player.LastCombat = 0

	if result.IsDodge {
		g.particles.SpawnMiss(enemy.X, enemy.Y)
		return
	}
	if result.IsCrit {
		g.particles.SpawnCrit(enemy.X, enemy.Y, result.Damage)
	} else {
		g.particles.SpawnDamage(enemy.X, enemy.Y, result.Damage, "#ffffff")
	}

	if result.IsDeath {
		xp := XPForKill(enemy.BaseXP, g.floor)
		g.player.Stats.XP += xp
		g.killCount++
		g.particles.SpawnText(enemy.X, enemy.Y, "+"+itoa(xp)+"xp", "#FFD700")

		// Boss drops guaranteed rare+ loot
		if enemy.Type.IsBoss() {
			loot := GenerateLoot(g.rng, g.floor+5) // boosted floor for rarity
			loot.X = enemy.X
			loot.Y = enemy.Y
			g.groundItems = append(g.groundItems, loot)
		} else if g.rng.Float64() < lootDropRate {
			loot := GenerateLoot(g.rng, g.floor)
			loot.X = enemy.X
			loot.Y = enemy.Y
			g.groundItems = append(g.groundItems, loot)
		}
	}
}

func (g *Game) enemyAttackPlayer(enemy *Enemy) {
	defender := *g.player.Entity
	defender.Stats = g.player.EffectiveStats()
	if g.player.HasEffect(ScrollShield) {
		defender.Stats.VIT += defender.Stats.VIT / 2
	}

	result := ResolveAttack(enemy.Entity, &defender, g.rng)
	if !result.IsDodge {
		g.player.Entity.TakeDamage(result.Damage)
	}
	g.player.LastCombat = 0

	if result.IsDodge {
		g.particles.SpawnMiss(g.player.X, g.player.Y)
		return
	}
	if result.IsCrit {
		g.particles.SpawnCrit(g.player.X, g.player.Y, result.Damage)
	} else {
		g.particles.SpawnDamage(g.player.X, g.player.Y, result.Damage, "#ff4444")
	}

	if !g.player.IsAlive {
		g.state = StateDead
	}
}

func (g *Game) processEnemyCombat() {
	for _, e := range g.world.Enemies {
		if !e.IsAlive {
			continue
		}
		dx := e.X - g.player.X
		dy := e.Y - g.player.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 && (dx+dy) == 1 && e.ActionTimer <= 0 {
			g.enemyAttackPlayer(e)
			e.ActionTimer = e.Stats.MoveCooldownMS() / 1000.0
		}
	}
}

func (g *Game) processRegen(dt float64) {
	if g.player.LastCombat < regenDelay {
		return
	}
	g.player.RegenTimer += dt
	if g.player.RegenTimer >= regenInterval {
		g.player.RegenTimer -= regenInterval
		healed := g.player.Heal(1)
		if healed > 0 {
			g.particles.SpawnText(g.player.X, g.player.Y, "+1", "#44ff44")
		}
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		g.player.Update(dt)
		g.processMovement()
		g.world.UpdateEnemies(dt, g.player.X, g.player.Y)
		g.processEnemyCombat()
		g.world.RemoveDead()
		g.processRegen(dt)
		g.particles.Update(dt)
		g.checkLevelUp()
	}
}

func (g *Game) Render() {
	g.renderer.Clear()

	switch g.state {
	case StateMenu:
		g.renderMenu()

	case StatePlaying, StateInventory, StateLevelUp, StatePaused:
		g.renderGameWorld()

		// HUD
		es := g.player.EffectiveStats()
		hudData := render.HUDData{
			HP:    g.player.Stats.HP,
			MaxHP: es.MaxHP(),
			Level: g.player.Stats.Level,
			XP:    g.player.Stats.XP,
			XPNext: g.player.Stats.XPToNextLevel(),
			Floor: g.floor,
		}
		for _, eff := range g.player.Effects {
			hudData.Effects = append(hudData.Effects, render.HUDEffect{
				Name:      eff.Name,
				Remaining: eff.Remaining,
			})
		}
		g.renderer.DrawHUD(hudData)

		// Overlays
		switch g.state {
		case StateInventory:
			g.renderer.DrawInventory(g.buildInventoryData())
		case StateLevelUp:
			choices := make([]render.LevelUpChoice, len(g.levelChoices))
			for i, c := range g.levelChoices {
				choices[i] = render.LevelUpChoice{Label: c.Label, Desc: c.Desc}
			}
			g.renderer.DrawLevelUp(g.player.Stats.Level+1, choices, g.levelUpSelected)
		case StatePaused:
			g.renderPause()
		}

	case StateDead:
		g.renderDeath()
	}
}

func (g *Game) renderMenu() {
	// ASCII art title
	g.renderer.DrawText(25, 8, " ####   ####  #####  #     ", "#00ff00")
	g.renderer.DrawText(25, 9, "#    # #    # #    # #     ", "#00ff00")
	g.renderer.DrawText(25, 10, "#      #    # #    # #     ", "#00ff00")
	g.renderer.DrawText(25, 11, "#  ### #    # #####  #     ", "#00cc00")
	g.renderer.DrawText(25, 12, "#    # #    # #   #  #     ", "#00cc00")
	g.renderer.DrawText(25, 13, " ####   ####  #    # ######", "#009900")
	g.renderer.DrawText(28, 16, "Go Roguelike", "#888888")
	g.renderer.DrawText(24, 22, "Press ENTER to start", "#aaaaaa")
	g.renderer.DrawText(24, 25, "WASD/Arrows: Move", "#555555")
	g.renderer.DrawText(24, 26, "I: Inventory  ESC: Pause", "#555555")
}

func (g *Game) renderGameWorld() {
	g.renderer.DrawDungeon(g.dungeonResult.Map, g.fov)

	// Ground items
	for _, item := range g.groundItems {
		if g.fov.IsVisible(item.X, item.Y) {
			vx := item.X - g.renderer.CamX
			vy := item.Y - g.renderer.CamY
			if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
				col := vx*render.TileCells + 1
				row := vy*render.TileCells + 1
				g.renderer.DrawChar(col, row, item.Glyph(), item.Rarity.Color())
			}
		}
	}

	// Enemies
	for _, e := range g.world.Enemies {
		if !e.IsAlive || !g.fov.IsVisible(e.X, e.Y) {
			continue
		}
		sprite := render.Sprites[e.Sprite]
		if sprite != nil {
			g.renderer.DrawSprite(sprite, 0, e.X, e.Y, sprite.Color)
		}
	}

	// Player
	sprite := render.Sprites[g.player.Sprite]
	if sprite != nil {
		g.renderer.DrawSprite(sprite, 0, g.player.X, g.player.Y, sprite.Color)
	}

	g.particles.Draw(g.renderer)
}

func (g *Game) renderPause() {
	boxW, boxH := 30, 10
	ox, oy := (render.GridCols-boxW)/2, (render.GridRows-boxH)/2
	g.renderer.DrawBox(ox, oy, boxW, boxH, "#888888", "#111111")
	g.renderer.DrawText(ox+10, oy+1, "PAUSED", "#ffffff")
	g.renderer.DrawText(ox+3, oy+4, "ESC/R: Resume", "#aaaaaa")
	g.renderer.DrawText(ox+3, oy+6, "Q: Quit to Menu", "#aaaaaa")
}

func (g *Game) renderDeath() {
	g.renderer.DrawText(24, 10, "#   # ####  #   #", "#ff0000")
	g.renderer.DrawText(24, 11, " # #  #   # #   #", "#ff0000")
	g.renderer.DrawText(24, 12, "  #   #   # #   #", "#cc0000")
	g.renderer.DrawText(24, 13, "  #   #   # #   #", "#cc0000")
	g.renderer.DrawText(24, 14, "  #   ####   ### ", "#990000")
	g.renderer.DrawText(22, 16, "D  I  E  D", "#ff0000")
	g.renderer.DrawText(20, 20, "Floor: "+itoa(g.floor), "#aaaaaa")
	g.renderer.DrawText(20, 21, "Level: "+itoa(g.player.Stats.Level), "#aaaaaa")
	g.renderer.DrawText(20, 22, "Kills: "+itoa(g.killCount), "#aaaaaa")
	g.renderer.DrawText(18, 26, "Press ENTER to restart", "#aaaaaa")
}

func (g *Game) buildInventoryData() render.InventoryData {
	inv := g.player.Inventory
	data := render.InventoryData{
		Selected: inv.Selected,
	}
	if inv.Weapon != nil {
		data.WeaponName = inv.Weapon.Name
		data.WeaponColor = inv.Weapon.Rarity.Color()
	}
	if inv.Armor != nil {
		data.ArmorName = inv.Armor.Name
		data.ArmorColor = inv.Armor.Rarity.Color()
	}
	for _, item := range inv.Items {
		detail := ""
		if item.BonusSTR > 0 {
			detail += "+" + itoa(item.BonusSTR) + " STR "
		}
		if item.BonusDEX > 0 {
			detail += "+" + itoa(item.BonusDEX) + " DEX "
		}
		if item.BonusVIT > 0 {
			detail += "+" + itoa(item.BonusVIT) + " VIT "
		}
		if item.BonusLCK > 0 {
			detail += "+" + itoa(item.BonusLCK) + " LCK "
		}
		data.Items = append(data.Items, render.InventoryItem{
			Name:   item.Name,
			Glyph:  item.Glyph(),
			Color:  item.Rarity.Color(),
			Detail: detail,
		})
	}
	return data
}

// Run starts the game loop using requestAnimationFrame.
func (g *Game) Run() {
	var frame js.Func
	frame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := args[0].Float()
		if g.lastTime == 0 {
			g.lastTime = now
		}
		dt := (now - g.lastTime) / 1000.0
		g.lastTime = now

		g.Update(dt)
		g.Render()

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
