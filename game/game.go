package game

import (
	"math/rand"
	"syscall/js"
	"time"

	"github.com/Baptiste-lg/GORL/audio"
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
	StateShop
	StateEvent
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
	streak        *KillStreak
	unlocks       *UnlockManager

	// Audio
	audioEngine *audio.Engine
	music       *audio.MusicEngine
	sfx         *audio.SFX

	// Level-up state
	levelChoices    []StatBoost
	levelUpSelected int

	// Event state
	eventState *EventState

	// Loadout selection
	selectedLoadout int
}

func New() *Game {
	ae := audio.NewEngine()
	g := &Game{
		state:       StateMenu,
		renderer:    render.NewRenderer(),
		keys:        make(map[string]bool),
		particles:   render.NewParticleSystem(),
		floor:       1,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		audioEngine: ae,
		music:       audio.NewMusicEngine(ae),
		sfx:         audio.NewSFX(ae),
		streak:      NewKillStreak(),
		unlocks:     NewUnlockManager(),
	}
	g.registerInput()
	return g
}

func (g *Game) startNewGame() {
	g.audioEngine.Resume()
	g.floor = 1
	g.killCount = 0
	g.player = nil
	g.particles = render.NewParticleSystem()
	g.groundItems = nil
	g.streak = NewKillStreak()
	g.generateFloor()

	// Apply selected loadout
	loadouts := g.unlocks.UnlockedLoadouts()
	if g.selectedLoadout >= 0 && g.selectedLoadout < len(loadouts) {
		lo := loadouts[g.selectedLoadout]
		g.player.Stats = lo.Stats
		g.player.Stats.HP = g.player.Stats.MaxHP()
		if lo.Weapon != nil {
			// Copy the weapon so we don't share pointers across runs
			w := *lo.Weapon
			g.player.Inventory.Weapon = &w
		}
	}

	g.state = StatePlaying
}

func (g *Game) generateFloor() {
	seed := g.rng.Int63()
	g.dungeonResult = dungeon.Generate(seed)
	g.fov = dungeon.NewFOV(dungeon.MapWidth, dungeon.MapHeight)
	g.world = NewWorld(g.dungeonResult, g.floor, g.rng)
	g.groundItems = g.world.GroundItems // treasure/secret room items

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

	// Check for new unlocks
	newUnlocks := g.unlocks.CheckUnlocks(g.floor)
	for _, name := range newUnlocks {
		g.particles.SpawnText(g.player.X, g.player.Y, "Unlocked: "+name+"!", "#FFD700")
	}

	// 50% chance of an event between floors
	if g.rng.Intn(2) == 0 {
		if evt := RollEvent(g.floor, g.rng); evt != nil {
			g.eventState = &EventState{Event: evt}
			g.state = StateEvent
			// Don't generate the new floor yet — wait for event completion
			return
		}
	}

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
		loadouts := g.unlocks.UnlockedLoadouts()
		switch key {
		case "Enter":
			g.startNewGame()
		case "ArrowLeft":
			if g.selectedLoadout > 0 {
				g.selectedLoadout--
			}
		case "ArrowRight":
			if g.selectedLoadout < len(loadouts)-1 {
				g.selectedLoadout++
			}
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
		case "m", "M":
			g.audioEngine.ToggleMute()
		case "e", "E", "Enter":
			g.interact()
		case "q", "Q":
			g.useActiveItem()
		}
	case StateInventory:
		g.handleInventoryKey(key)
	case StatePaused:
		g.handlePauseKey(key)
	case StateLevelUp:
		g.handleLevelUpKey(key)
	case StateShop:
		g.handleShopKey(key)
	case StateEvent:
		g.handleEventKey(key)
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
			g.player.RecalcSynergies()
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
			xpNeeded := g.player.Stats.XPToNextLevel()
			g.player.Stats.Level++
			g.player.Stats.XP -= xpNeeded
			if g.player.Stats.XP < 0 {
				g.player.Stats.XP = 0
			}
			g.player.Stats.HP = g.player.EffectiveStats().MaxHP()
			g.state = StatePlaying
		}
	}
}

func (g *Game) checkLevelUp() {
	if g.player.Stats.CanLevelUp() {
		g.sfx.LevelUp()
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

func (g *Game) checkChallengeCleared() {
	if g.world.ChallengeCleared || g.world.ChallengeRoom < 0 {
		return
	}
	if g.world.EnemiesInRoom(g.world.ChallengeRoom) == 0 {
		g.world.ChallengeCleared = true
		r := g.world.Dungeon.Rooms[g.world.ChallengeRoom]
		cx, cy := r.CenterX(), r.CenterY()
		// Spawn rare+ reward
		loot := GenerateLoot(g.rng, g.floor+5)
		loot.X = cx
		loot.Y = cy
		g.groundItems = append(g.groundItems, loot)
		g.particles.SpawnText(cx, cy, "CHALLENGE CLEARED!", "#FFD700")
		g.sfx.LevelUp()
	}
}

func (g *Game) interact() {
	// Check shop first
	if s := g.world.Shop; s != nil && s.X == g.player.X && s.Y == g.player.Y {
		s.Selected = 0
		g.state = StateShop
		return
	}

	// Then shrine
	shrine := g.world.ShrineAt(g.player.X, g.player.Y)
	if shrine == nil || shrine.Used {
		return
	}
	result := UseShrine(shrine, g.player, g.rng)
	g.particles.SpawnText(g.player.X, g.player.Y, result.Message, result.Color)
	g.sfx.Pickup()

	if !g.player.IsAlive {
		g.sfx.Death()
		g.renderer.Shake(10.0, 0.5)
		g.state = StateDead
	}
}

func (g *Game) useActiveItem() {
	a := g.player.Active
	if a == nil {
		g.particles.SpawnText(g.player.X, g.player.Y, "No active item", "#555555")
		return
	}
	if !a.Use() {
		g.particles.SpawnText(g.player.X, g.player.Y, "Not charged!", "#ff4444")
		return
	}

	g.sfx.Pickup()
	px, py := g.player.X, g.player.Y

	syn := g.player.Synergies

	switch a.ID {
	case ActiveHealBurst:
		healAmt := g.player.EffectiveStats().MaxHP() * 40 / 100
		// Synergy: Crimson Tide (HealBurst+Vampiric: +5 HP per visible enemy)
		if HasSynergy(syn, SynCrimsonTide) {
			for _, e := range g.world.Enemies {
				if e.IsAlive && g.fov.IsVisible(e.X, e.Y) {
					healAmt += 5
				}
			}
		}
		healed := g.player.Heal(healAmt)
		g.particles.SpawnText(px, py, "+"+itoa(healed)+" HP", "#44ff44")
		// Synergy: Fountain of Life (HealBurst+Regeneration: doubled regen 10s)
		if HasSynergy(syn, SynFountainOfLife) {
			g.player.AddEffect(StatusEffect{
				Name: "Fountain", Remaining: 10.0, Kind: ScrollKind(102),
			})
			g.particles.SpawnText(px, py, "FOUNTAIN OF LIFE!", "#44ffaa")
		}

	case ActiveShieldWall:
		vitBonus := 10
		// Synergy: Unbreakable (ShieldWall+Fortified: +20 VIT)
		if HasSynergy(syn, SynUnbreakable) {
			vitBonus = 20
		}
		g.player.Stats.VIT += vitBonus
		g.player.AddEffect(StatusEffect{
			Name: "Shield Wall", Remaining: 8.0, Kind: ScrollKind(104),
			Magnitude: vitBonus,
		})
		g.particles.SpawnText(px, py, "SHIELD WALL! +"+itoa(vitBonus)+" VIT", "#4488ff")

	case ActiveWarCry:
		strBonus := 5
		// Synergy: Rage Unleashed (WarCry+Berserker: +10 STR)
		if HasSynergy(syn, SynRageUnleashed) {
			strBonus = 10
		}
		g.player.Stats.STR += strBonus
		g.particles.SpawnText(px, py, "WAR CRY! +"+itoa(strBonus)+" STR", "#ff8800")
		g.player.AddEffect(StatusEffect{
			Name: "War Cry", Remaining: 6.0, Kind: ScrollKind(101),
			Magnitude: strBonus,
		})

	case ActiveFreeze:
		freezeDur := 2.0
		// Synergy: Absolute Zero (Freeze+Freezing: 4s freeze)
		if HasSynergy(syn, SynAbsoluteZero) {
			freezeDur = 4.0
		}
		count := 0
		for _, e := range g.world.Enemies {
			if e.IsAlive && g.fov.IsVisible(e.X, e.Y) {
				e.ActionTimer += freezeDur
				count++
			}
		}
		g.particles.SpawnText(px, py, "FREEZE! ("+itoa(count)+")", "#44aaff")

	case ActiveFireball:
		dx, dy := 0, 0
		if g.keys["w"] || g.keys["ArrowUp"] {
			dy = -1
		} else if g.keys["s"] || g.keys["ArrowDown"] {
			dy = 1
		} else if g.keys["a"] || g.keys["ArrowLeft"] {
			dx = -1
		} else if g.keys["d"] || g.keys["ArrowRight"] {
			dx = 1
		}
		if dx == 0 && dy == 0 {
			dy = -1
		}
		fbRange := 3
		fbDmg := 15
		// Synergy: Inferno (Fireball+Burning: 5 tiles, +10 damage)
		if HasSynergy(syn, SynInferno) {
			fbRange = 5
			fbDmg = 25
		}
		for i := 1; i <= fbRange; i++ {
			tx, ty := px+dx*i, py+dy*i
			if !g.dungeonResult.Map.At(tx, ty).Passable() {
				break
			}
			if e := g.world.EnemyAt(tx, ty); e != nil {
				e.Entity.TakeDamage(fbDmg)
				g.particles.SpawnDamage(tx, ty, fbDmg, "#ff4400")
				if !e.IsAlive {
					g.killCount++
				}
			}
			g.particles.SpawnText(tx, ty, "*", "#ff4400")
		}
		g.renderer.Shake(4.0, 0.2)

	case ActivePoisonCloud:
		// Synergy: Plague Bearer (PoisonCloud+Venomous: double radius)
		radius := 1
		if HasSynergy(syn, SynPlagueBearer) {
			radius = 2
		}
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				tx, ty := px+dx, py+dy
				if e := g.world.EnemyAt(tx, ty); e != nil {
					e.Entity.TakeDamage(3)
					e.ActionTimer += 1.0
					g.particles.SpawnText(tx, ty, "POISON!", "#44aa44")
				}
			}
		}
		g.particles.SpawnText(px, py, "POISON CLOUD!", "#44aa44")

	case ActiveDash:
		dx, dy := 0, 0
		if g.keys["w"] || g.keys["ArrowUp"] {
			dy = -1
		} else if g.keys["s"] || g.keys["ArrowDown"] {
			dy = 1
		} else if g.keys["a"] || g.keys["ArrowLeft"] {
			dx = -1
		} else if g.keys["d"] || g.keys["ArrowRight"] {
			dx = 1
		}
		if dx == 0 && dy == 0 {
			dy = -1
		}
		dashRange := 3
		// Synergy: Flash Step (Dash+Swift: 5 tiles)
		if HasSynergy(syn, SynFlashStep) {
			dashRange = 5
		}
		for i := 1; i <= dashRange; i++ {
			tx, ty := px+dx*i, py+dy*i
			if !g.dungeonResult.Map.At(tx, ty).Passable() || g.world.EnemyAt(tx, ty) != nil {
				break
			}
			g.player.X = tx
			g.player.Y = ty
		}
		g.renderer.CenterCamera(g.player.X, g.player.Y)
		g.fov.Compute(g.dungeonResult.Map, g.player.X, g.player.Y, fovRadius)
		g.particles.SpawnText(g.player.X, g.player.Y, "DASH!", "#44aaff")
		// Synergy: Phantom (Dash+Evasion: 100% dodge 1s)
		if HasSynergy(syn, SynPhantom) {
			g.player.AddEffect(StatusEffect{
				Name: "Phantom", Remaining: 1.0, Kind: ScrollKind(103),
			})
		}
		// Synergy: Flash Step cooldown reset
		if HasSynergy(syn, SynFlashStep) {
			g.player.MoveCooldown = 0
		}

	case ActiveBlink:
		dm := g.dungeonResult.Map
		for attempts := 0; attempts < 100; attempts++ {
			tx := g.rng.Intn(dungeon.MapWidth)
			ty := g.rng.Intn(dungeon.MapHeight)
			if dm.At(tx, ty).Passable() && g.fov.IsExplored(tx, ty) && g.world.EnemyAt(tx, ty) == nil {
				g.player.X = tx
				g.player.Y = ty
				g.renderer.CenterCamera(tx, ty)
				g.fov.Compute(dm, tx, ty, fovRadius)
				break
			}
		}
		g.particles.SpawnText(g.player.X, g.player.Y, "BLINK!", "#cc44ff")
		// Synergy: Vanish (Blink+Stealth: enemies lose aggro)
		if HasSynergy(syn, SynVanish) {
			for _, e := range g.world.Enemies {
				if e.AI == AIChase {
					e.AI = AIPatrol
				}
			}
			g.particles.SpawnText(g.player.X, g.player.Y, "VANISH!", "#aaaaff")
		}
	}
}

func (g *Game) handleShopKey(key string) {
	shop := g.world.Shop
	if shop == nil {
		g.state = StatePlaying
		return
	}

	switch key {
	case "Escape", "e", "E":
		g.state = StatePlaying
	case "ArrowUp":
		if shop.Selected > 0 {
			shop.Selected--
		}
	case "ArrowDown":
		if shop.Selected < len(shop.Items)-1 {
			shop.Selected++
		}
	case "Enter":
		if shop.Buy(shop.Selected, g.player) {
			g.sfx.Pickup()
			g.particles.SpawnText(g.player.X, g.player.Y, "Purchased!", "#44ff44")
		}
	}
}

func (g *Game) handleEventKey(key string) {
	es := g.eventState
	if es == nil {
		g.state = StatePlaying
		return
	}

	// If result is showing, Enter continues to next floor
	if es.Result != nil {
		if key == "Enter" {
			g.eventState = nil
			if !g.player.IsAlive {
				g.state = StateDead
				return
			}
			g.generateFloor()
			g.state = StatePlaying
		}
		return
	}

	switch key {
	case "ArrowUp":
		if es.Selected > 0 {
			es.Selected--
		}
	case "ArrowDown":
		if es.Selected < len(es.Event.Choices)-1 {
			es.Selected++
		}
	case "Enter":
		if es.Selected >= 0 && es.Selected < len(es.Event.Choices) {
			choice := es.Event.Choices[es.Selected]
			result := choice.Action(g.player, g.rng)
			es.Result = &result
		}
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

	// Destructible walls: attack cracked wall to break it
	if dm.At(nx, ny) == dungeon.TileCrackedWall {
		dm.Set(nx, ny, dungeon.TileFloor)
		g.sfx.Hit()
		g.renderer.Shake(2.0, 0.1)
		g.particles.SpawnText(nx, ny, "CRACK!", "#887755")
		g.player.ResetMoveCooldown()
		g.fov.Compute(dm, g.player.X, g.player.Y, fovRadius)
		return
	}

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
	g.sfx.Footstep()

	// Pick up items
	g.pickupItems()

	// Check traps
	g.checkTraps()

	// Check stairs
	if dm.At(g.player.X, g.player.Y) == dungeon.TileStairsDown {
		g.sfx.Stairs()
		g.nextFloor()
	}
}

func (g *Game) checkTraps() {
	trap := g.world.TrapAt(g.player.X, g.player.Y)
	if trap == nil {
		return
	}

	trap.Triggered = true
	trap.Revealed = true

	switch trap.Type {
	case TrapSpike:
		dmg := 5 + g.floor*2
		g.player.Entity.TakeDamage(dmg)
		g.particles.SpawnDamage(g.player.X, g.player.Y, dmg, "#aa4444")
		g.particles.SpawnText(g.player.X, g.player.Y, "SPIKE TRAP!", "#aa4444")
		g.sfx.Hit()
		g.renderer.Shake(4.0, 0.2)
	case TrapPoison:
		// Apply a poison status effect
		g.player.AddEffect(StatusEffect{
			Name:      "Poisoned",
			Remaining: 5.0,
			Kind:      ScrollKind(100), // special kind for poison
		})
		g.particles.SpawnText(g.player.X, g.player.Y, "POISON TRAP!", "#44aa44")
		g.sfx.Hit()
	case TrapTeleport:
		// Teleport to a random passable tile
		dm := g.dungeonResult.Map
		for attempts := 0; attempts < 100; attempts++ {
			tx := g.rng.Intn(dungeon.MapWidth)
			ty := g.rng.Intn(dungeon.MapHeight)
			if dm.At(tx, ty).Passable() && g.world.EnemyAt(tx, ty) == nil {
				g.player.X = tx
				g.player.Y = ty
				g.renderer.CenterCamera(g.player.X, g.player.Y)
				g.fov.Compute(dm, g.player.X, g.player.Y, fovRadius)
				break
			}
		}
		g.particles.SpawnText(g.player.X, g.player.Y, "TELEPORT!", "#4444ee")
		g.sfx.Stairs()
	}

	if !g.player.IsAlive {
		g.sfx.Death()
		g.renderer.Shake(10.0, 0.5)
		g.state = StateDead
	}
}

func (g *Game) processHazards(dt float64) {
	hazards := g.world.HazardsAt(g.player.X, g.player.Y)
	fireproof := g.player.ArmorHasAffix(AffixFireproof)
	for _, h := range hazards {
		// Fireproof armor negates fire/lava
		if fireproof && h.Type == HazardLava {
			continue
		}
		h.DmgAccum += h.DPS * dt
		if h.DmgAccum >= 1.0 {
			dmg := int(h.DmgAccum)
			h.DmgAccum -= float64(dmg)
			g.player.Entity.TakeDamage(dmg)

			color := "#ff4400"
			if h.Type == HazardPoisonGas {
				color = "#44cc44"
			}
			g.particles.SpawnDamage(g.player.X, g.player.Y, dmg, color)
			g.player.LastCombat = 0
		}
	}

	if !g.player.IsAlive {
		g.sfx.Death()
		g.renderer.Shake(10.0, 0.5)
		g.state = StateDead
	}
}

func (g *Game) pickupItems() {
	remaining := g.groundItems[:0]
	for _, item := range g.groundItems {
		if item.X == g.player.X && item.Y == g.player.Y {
			if g.player.Inventory.Add(item) {
				g.particles.SpawnText(g.player.X, g.player.Y, item.Name, item.Rarity.Color())
				g.sfx.Pickup()
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

	syn := g.player.Synergies

	// Weapon affix: Berserker (+50% damage below 30% HP)
	berserkerActive := false
	if g.player.WeaponHasAffix(AffixBerserker) {
		maxHP := attacker.Stats.MaxHP()
		if maxHP > 0 && float64(g.player.Stats.HP)/float64(maxHP) < 0.30 {
			attacker.Stats.STR += attacker.Stats.STR / 2
			berserkerActive = true
		}
	}

	// Weapon affix: Lucky (+5% crit — applied via stats)
	if g.player.WeaponHasAffix(AffixLucky) {
		attacker.Stats.LCK += 3
	}

	// Synergy: Desperate Gambit (Lucky+Berserker: +15% crit below 30%)
	if berserkerActive && HasSynergy(syn, SynDesperateGambit) {
		attacker.Stats.LCK += 8 // +16% extra crit
	}

	// Synergy: Juggernaut (Berserker+Fortified: extra STR at low HP)
	if berserkerActive && HasSynergy(syn, SynJuggernaut) {
		attacker.Stats.STR += attacker.Stats.STR / 4
	}

	result := ResolveAttack(&attacker, enemy.Entity, g.rng)
	g.player.LastCombat = 0

	if result.IsDodge {
		g.particles.SpawnMiss(enemy.X, enemy.Y)
		g.sfx.Miss()
		return
	}

	// Weapon affix: Burning (+3 fire damage)
	if g.player.WeaponHasAffix(AffixBurning) {
		fireDmg := 3
		// Synergy: Toxic Flame (Burning+Venomous: +2 fire)
		if HasSynergy(syn, SynToxicFlame) {
			fireDmg += 2
		}
		// Synergy: Flame Lord (Burning+Fireproof: +5 fire)
		if HasSynergy(syn, SynFlameLord) {
			fireDmg += 5
		}
		bonus := enemy.Entity.TakeDamage(fireDmg)
		result.Damage += bonus
		g.particles.SpawnText(enemy.X, enemy.Y, "+"+itoa(fireDmg)+" fire", "#ff4400")
	}

	// Weapon affix: Executioner (double damage to enemies below 25% HP)
	if g.player.WeaponHasAffix(AffixExecutioner) && !result.IsDeath {
		maxHP := enemy.Stats.MaxHP()
		if maxHP > 0 && float64(enemy.Stats.HP)/float64(maxHP) < 0.25 {
			bonus := enemy.Entity.TakeDamage(result.Damage) // deal same damage again
			result.Damage += bonus
			g.particles.SpawnText(enemy.X, enemy.Y, "EXECUTE!", "#cc0000")
		}
	}

	// Weapon affix: Freezing (15% chance to stun)
	if g.player.WeaponHasAffix(AffixFreezing) && g.rng.Intn(100) < 15 {
		freezeDur := 2.0
		// Synergy: Shatter (Freezing+Executioner: longer freeze)
		if HasSynergy(syn, SynShatter) {
			freezeDur = 3.0
			// Extra damage on frozen crit
			if result.IsCrit {
				bonus := enemy.Entity.TakeDamage(result.Damage) // 3x total
				result.Damage += bonus
				g.particles.SpawnText(enemy.X, enemy.Y, "SHATTER!", "#88ccff")
			}
		}
		enemy.ActionTimer += freezeDur
		g.particles.SpawnText(enemy.X, enemy.Y, "FROZEN!", "#44aaff")
	}

	// Weapon affix: Venomous (20% chance to poison)
	if g.player.WeaponHasAffix(AffixVenomous) && g.rng.Intn(100) < 20 {
		poisonDmg := 2
		// Synergy: Toxic Flame (Burning+Venomous: +2 poison)
		if HasSynergy(syn, SynToxicFlame) {
			poisonDmg += 2
		}
		enemy.Entity.TakeDamage(poisonDmg)
		g.particles.SpawnText(enemy.X, enemy.Y, "POISON!", "#44aa44")
	}

	result.IsDeath = !enemy.IsAlive

	if result.IsCrit {
		g.particles.SpawnCrit(enemy.X, enemy.Y, result.Damage)
		g.sfx.CritHit()
		// Synergy: Whirlwind (Swift+Lucky: crit halves next cooldown)
		if HasSynergy(syn, SynWhirlwind) {
			g.player.MoveCooldown /= 2
		}
	} else {
		g.particles.SpawnDamage(enemy.X, enemy.Y, result.Damage, "#ffffff")
		g.sfx.Hit()
	}

	if berserkerActive {
		g.particles.SpawnText(g.player.X, g.player.Y, "BERSERK!", "#ff4400")
	}

	if result.IsDeath {
		// Kill streak XP multiplier
		mult := g.streak.RegisterKill()
		xp := int(float64(XPForKill(enemy.BaseXP, g.floor)) * mult)
		g.player.Stats.XP += xp
		g.killCount++

		xpText := "+" + itoa(xp) + "xp"
		if mult > 1.0 {
			xpText += " x" + itoa(int(mult*100)) + "%"
		}
		g.particles.SpawnText(enemy.X, enemy.Y, xpText, "#FFD700")

		if g.streak.Active() {
			g.particles.SpawnText(g.player.X, g.player.Y, g.streak.Label(), "#ff8800")
		}

		// Weapon affix: Vampiric (heal 1 HP on kill)
		if g.player.WeaponHasAffix(AffixVampiric) {
			healAmt := 1
			// Synergy: Bloodthirst (Vampiric+Berserker: heal 3 at low HP)
			if berserkerActive && HasSynergy(syn, SynBloodthirst) {
				healAmt = 3
			}
			healed := g.player.Heal(healAmt)
			if healed > 0 {
				g.particles.SpawnText(g.player.X, g.player.Y, "+"+itoa(healed)+" HP", "#ff4488")
			}
		}

		// Charge active item on kill
		if g.player.Active != nil && !g.player.Active.Ready() {
			g.player.Active.AddCharge()
		}

		// Gold drop (always)
		goldAmt := goldForKill(enemy.BaseXP, g.floor, g.rng)
		g.player.Gold += goldAmt
		g.particles.SpawnText(enemy.X, enemy.Y-1, "+"+itoa(goldAmt)+"g", "#ccaa00")

		// Boss drops guaranteed rare+ loot + active item
		if enemy.Type.IsBoss() {
			loot := GenerateLoot(g.rng, g.floor+5)
			loot.X = enemy.X
			loot.Y = enemy.Y
			g.groundItems = append(g.groundItems, loot)

			// Boss always drops an active item
			newActive := NewActiveItem(RandomActiveID(g.rng))
			g.player.Active = newActive
			g.player.RecalcSynergies()
			g.particles.SpawnText(g.player.X, g.player.Y, "Active: "+newActive.Name+"!", "#cc44ff")
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

	dsyn := g.player.Synergies

	// Armor affix: Evasion (+5% dodge via stats)
	if g.player.ArmorHasAffix(AffixEvasion) {
		dexBonus := 3
		// Synergy: Ice Dancer (Freezing+Evasion: +10% dodge total)
		if HasSynergy(dsyn, SynIceDancer) {
			dexBonus = 7
		}
		defender.Stats.DEX += dexBonus
	}

	result := ResolveAttack(enemy.Entity, &defender, g.rng)

	// Synergy: Phantom (Dash+Evasion: 100% dodge while effect active)
	if !result.IsDodge && g.player.HasEffect(ScrollKind(103)) {
		result.IsDodge = true
	}

	if !result.IsDodge {
		dmg := result.Damage

		// Armor affix: Bulwark (reduce all damage by 1)
		if g.player.ArmorHasAffix(AffixBulwark) && dmg > 1 {
			reduction := 1
			// Synergy: Reaper's Guard (Executioner+Bulwark: reduce by 2)
			if HasSynergy(dsyn, SynReaperGuard) {
				reduction = 2
			}
			dmg -= reduction
			if dmg < 1 {
				dmg = 1
			}
		}

		// Armor affix: Absorbing (10% chance to heal instead)
		absorbChance := 10
		// Synergy: Fortune's Favor (Lucky+Absorbing: +15% absorb)
		if HasSynergy(dsyn, SynFortuneFavor) {
			absorbChance = 25
		}
		if g.player.ArmorHasAffix(AffixAbsorbing) && g.rng.Intn(100) < absorbChance {
			g.player.Heal(dmg)
			g.particles.SpawnText(g.player.X, g.player.Y, "ABSORB!", "#44ffaa")
			dmg = 0
		}

		if dmg > 0 {
			g.player.Entity.TakeDamage(dmg)
		}
	}
	g.player.LastCombat = 0

	// Armor affix: Thorns (reflect 2 damage)
	if !result.IsDodge && g.player.ArmorHasAffix(AffixThorns) {
		thornAmt := 2
		// Synergy: Iron Maiden (Thorns+ShieldWall: thorns x3 while shielded)
		if HasSynergy(dsyn, SynIronMaiden) && (g.player.HasEffect(ScrollShield) || g.player.HasEffect(ScrollKind(104))) {
			thornAmt *= 3
		}
		thornDmg := enemy.Entity.TakeDamage(thornAmt)
		if thornDmg > 0 {
			g.particles.SpawnDamage(enemy.X, enemy.Y, thornDmg, "#aa4444")
			// Synergy: Blood Mirror (Vampiric+Thorns: thorn damage heals)
			if HasSynergy(dsyn, SynBloodMirror) {
				g.player.Heal(thornDmg)
			}
		}
	}

	if result.IsDodge {
		g.particles.SpawnMiss(g.player.X, g.player.Y)
		g.sfx.Miss()
		return
	}
	if result.IsCrit {
		g.particles.SpawnCrit(g.player.X, g.player.Y, result.Damage)
		g.sfx.CritHit()
		g.renderer.Shake(6.0, 0.3)
	} else {
		g.particles.SpawnDamage(g.player.X, g.player.Y, result.Damage, "#ff4444")
		g.sfx.Hit()
		g.renderer.Shake(3.0, 0.15)
	}

	if !g.player.IsAlive {
		g.sfx.Death()
		g.renderer.Shake(10.0, 0.5)
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
		regenAmt := 1
		// Armor affix: Regeneration (+1 HP)
		if g.player.ArmorHasAffix(AffixRegeneration) {
			regenAmt += 1
			// Synergy: Toxic Resilience (Venomous+Regeneration: +2 more)
			if HasSynergy(g.player.Synergies, SynToxicResilience) {
				regenAmt += 2
			}
		}
		// Synergy: Fountain of Life (doubled regen while effect active)
		if g.player.HasEffect(ScrollKind(102)) {
			regenAmt *= 2
		}
		healed := g.player.Heal(regenAmt)
		if healed > 0 {
			g.particles.SpawnText(g.player.X, g.player.Y, "+"+itoa(healed), "#44ff44")
		}
	}
}

func (g *Game) updateMusicState() {
	combat := false
	boss := false
	for _, e := range g.world.Enemies {
		if !e.IsAlive {
			continue
		}
		if e.AI == AIChase {
			combat = true
			if e.Type.IsBoss() {
				boss = true
			}
		}
	}

	if boss {
		g.music.SetState(audio.MusicBoss)
	} else if combat {
		g.music.SetState(audio.MusicCombat)
	} else {
		g.music.SetState(audio.MusicExplore)
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		g.player.Update(dt)
		g.processMovement()
		detectRange := 6
		if g.player.ArmorHasAffix(AffixStealth) {
			detectRange = 4
			// Synergy: Shadow Step (Swift+Stealth: range 3)
			if HasSynergy(g.player.Synergies, SynShadowStep) {
				detectRange = 3
			}
		}
		g.world.UpdateEnemies(dt, g.player.X, g.player.Y, detectRange)
		g.processEnemyCombat()
		g.world.RemoveDead()
		g.checkChallengeCleared()
		g.processHazards(dt)
		g.processRegen(dt)
		g.streak.Update(dt)
		g.particles.Update(dt)
		g.renderer.UpdateCamera(dt)
		g.updateMusicState()
		g.music.Update(g.audioEngine.CurrentTime())
		g.checkLevelUp()
	}
}

func (g *Game) Render() {
	g.renderer.Clear()

	switch g.state {
	case StateMenu:
		g.renderMenu()

	case StatePlaying, StateInventory, StateLevelUp, StatePaused, StateShop:
		g.renderGameWorld()

		// HUD
		es := g.player.EffectiveStats()
		hudData := render.HUDData{
			HP:     g.player.Stats.HP,
			MaxHP:  es.MaxHP(),
			Level:  g.player.Stats.Level,
			XP:     g.player.Stats.XP,
			XPNext: g.player.Stats.XPToNextLevel(),
			Floor:  g.floor,
			Gold:   g.player.Gold,
			Streak: g.streak.Label(),
		}
		if a := g.player.Active; a != nil {
			hudData.ActiveName = a.Name
			hudData.ActiveReady = a.Ready()
			if a.MaxCharges > 0 {
				hudData.ActivePct = a.Charges * 100 / a.MaxCharges
			}
		}
		for _, eff := range g.player.Effects {
			hudData.Effects = append(hudData.Effects, render.HUDEffect{
				Name:      eff.Name,
				Remaining: eff.Remaining,
			})
		}
		g.renderer.DrawHUD(hudData)

		// Interaction hints
		if s := g.world.Shop; s != nil && s.X == g.player.X && s.Y == g.player.Y && s.HasItems() {
			g.renderer.DrawText(1, render.GridRows-3, "[E] Shop", "#FFD700")
		} else if shrine := g.world.ShrineAt(g.player.X, g.player.Y); shrine != nil && !shrine.Used {
			g.renderer.DrawText(1, render.GridRows-3, "[E] "+shrine.Name(), shrine.Color())
		}

		// Mini-map
		g.renderer.DrawMiniMap(render.MiniMapData{
			MapW:    dungeon.MapWidth,
			MapH:    dungeon.MapHeight,
			PlayerX: g.player.X,
			PlayerY: g.player.Y,
			IsExplored: func(x, y int) bool {
				return g.fov.IsExplored(x, y)
			},
			IsWall: func(x, y int) bool {
				t := g.dungeonResult.Map.At(x, y)
				return t == dungeon.TileWall || t == dungeon.TileCrackedWall
			},
			Markers: g.buildMiniMapMarkers(),
		})

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
		case StateShop:
			g.renderer.DrawShop(g.buildShopData())
		}

	case StateDead:
		g.renderDeath()

	case StateEvent:
		g.renderEvent()
	}
}

func (g *Game) renderMenu() {
	g.renderer.DrawText(25, 8, " ####   ####  #####  #     ", "#00ff00")
	g.renderer.DrawText(25, 9, "#    # #    # #    # #     ", "#00ff00")
	g.renderer.DrawText(25, 10, "#      #    # #    # #     ", "#00ff00")
	g.renderer.DrawText(25, 11, "#  ### #    # #####  #     ", "#00cc00")
	g.renderer.DrawText(25, 12, "#    # #    # #   #  #     ", "#00cc00")
	g.renderer.DrawText(25, 13, " ####   ####  #    # ######", "#009900")
	g.renderer.DrawText(28, 16, "Go Roguelike", "#888888")

	// Loadout selection
	loadouts := g.unlocks.UnlockedLoadouts()
	if len(loadouts) > 1 {
		g.renderer.DrawText(24, 19, "< Loadout: ", "#555555")
		if g.selectedLoadout >= 0 && g.selectedLoadout < len(loadouts) {
			lo := loadouts[g.selectedLoadout]
			g.renderer.DrawText(35, 19, lo.Name, "#FFD700")
			g.renderer.DrawText(24, 20, "  "+lo.Desc, "#888888")
		}
		g.renderer.DrawText(55, 19, " >", "#555555")
	}

	g.renderer.DrawText(24, 22, "Press ENTER to start", "#aaaaaa")
	g.renderer.DrawText(24, 25, "WASD/Arrows: Move  E: Interact", "#555555")
	g.renderer.DrawText(24, 26, "I: Inventory  ESC: Pause", "#555555")
	g.renderer.DrawText(24, 27, "M: Toggle Music", "#555555")
}

func (g *Game) renderGameWorld() {
	theme := dungeon.ThemeForFloor(g.floor)
	g.renderer.DrawDungeonThemed(g.dungeonResult.Map, g.fov, theme)

	// Revealed traps
	for _, trap := range g.world.Traps {
		if trap.Revealed && g.fov.IsVisible(trap.X, trap.Y) {
			vx := trap.X - g.renderer.CamX
			vy := trap.Y - g.renderer.CamY
			if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
				col := vx*render.TileCells + 1
				row := vy*render.TileCells + 1
				g.renderer.DrawChar(col, row, trap.Glyph(), trap.Color())
			}
		}
	}

	// Shop NPC
	if s := g.world.Shop; s != nil && g.fov.IsVisible(s.X, s.Y) {
		vx := s.X - g.renderer.CamX
		vy := s.Y - g.renderer.CamY
		if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
			col := vx*render.TileCells + 1
			row := vy*render.TileCells + 1
			g.renderer.DrawChar(col, row, "S", "#FFD700")
		}
	}

	// Challenge room marker (! until cleared)
	if ci := g.world.ChallengeRoom; ci >= 0 && !g.world.ChallengeCleared {
		r := g.world.Dungeon.Rooms[ci]
		cx, cy := r.CenterX(), r.CenterY()
		if g.fov.IsVisible(cx, cy) {
			vx := cx - g.renderer.CamX
			vy := cy - g.renderer.CamY
			if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
				col := vx*render.TileCells + 1
				row := vy*render.TileCells + 1
				g.renderer.DrawChar(col, row, "!", "#ff4444")
			}
		}
	}

	// Shrines
	for _, shrine := range g.world.Shrines {
		if g.fov.IsVisible(shrine.X, shrine.Y) {
			vx := shrine.X - g.renderer.CamX
			vy := shrine.Y - g.renderer.CamY
			if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
				col := vx*render.TileCells + 1
				row := vy*render.TileCells + 1
				g.renderer.DrawChar(col, row, shrine.Glyph(), shrine.Color())
			}
		}
	}

	// Hazards
	for _, haz := range g.world.Hazards {
		if g.fov.IsVisible(haz.X, haz.Y) {
			vx := haz.X - g.renderer.CamX
			vy := haz.Y - g.renderer.CamY
			if vx >= 0 && vx < render.ViewTilesX && vy >= 0 && vy < render.ViewTilesY {
				col := vx*render.TileCells + 1
				row := vy*render.TileCells + 1
				g.renderer.DrawChar(col, row, haz.Glyph(), haz.Color())
			}
		}
	}

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

	// Player with HP color fade
	sprite := render.Sprites[g.player.Sprite]
	if sprite != nil {
		hpRatio := 1.0
		maxHP := g.player.EffectiveStats().MaxHP()
		if maxHP > 0 {
			hpRatio = float64(g.player.Stats.HP) / float64(maxHP)
		}
		color := render.HPColor(sprite.Color, hpRatio)
		g.renderer.DrawSprite(sprite, 0, g.player.X, g.player.Y, color)
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
	g.renderer.DrawText(20, 23, "Gold:  "+itoa(g.player.Gold), "#FFD700")
	if g.streak.Count > 2 {
		g.renderer.DrawText(20, 24, "Best Streak: "+itoa(g.streak.Count), "#ff8800")
	}
	g.renderer.DrawText(18, 26, "Press ENTER to restart", "#aaaaaa")
}

func (g *Game) renderEvent() {
	es := g.eventState
	if es == nil {
		return
	}
	data := render.EventData{
		Title:    es.Event.Title,
		Text:     es.Event.Text,
		Selected: es.Selected,
	}
	for _, ch := range es.Event.Choices {
		data.Choices = append(data.Choices, render.EventChoiceData{Label: ch.Label})
	}
	if es.Result != nil {
		data.Result = es.Result.Message
		data.ResColor = es.Result.Color
	}
	g.renderer.DrawEvent(data)
}

func itemDetail(item *Item) string {
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
	if len(item.Affixes) > 0 {
		for _, id := range item.Affixes {
			if def, ok := AffixDefs[id]; ok {
				detail += "[" + def.Name + "] "
			}
		}
	}
	return detail
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
		data.Items = append(data.Items, render.InventoryItem{
			Name:   item.Name,
			Glyph:  item.Glyph(),
			Color:  item.Rarity.Color(),
			Detail: itemDetail(item),
		})
	}

	// Active synergies
	for _, sid := range g.player.Synergies {
		if s := SynergyByID(sid); s != nil {
			data.Synergies = append(data.Synergies, s.Name+" - "+s.Desc)
		}
	}
	return data
}

func (g *Game) buildShopData() render.ShopData {
	shop := g.world.Shop
	data := render.ShopData{
		Gold:     g.player.Gold,
		Selected: shop.Selected,
	}
	for i, item := range shop.Items {
		si := render.ShopItem{
			Price: shop.Prices[i],
			Sold:  item == nil,
		}
		if item != nil {
			si.Name = item.Name
			si.Glyph = item.Glyph()
			si.Color = item.Rarity.Color()
		}
		data.Items = append(data.Items, si)
	}
	return data
}

func (g *Game) buildMiniMapMarkers() []render.MiniMapMarker {
	var markers []render.MiniMapMarker
	for idx, role := range g.world.RoomRoles {
		if idx >= len(g.world.Dungeon.Rooms) {
			continue
		}
		r := g.world.Dungeon.Rooms[idx]
		var color string
		switch role {
		case RoleShop:
			color = "#44ff44"
		case RoleTreasure:
			color = "#FFD700"
		case RoleChallenge:
			if g.world.ChallengeCleared {
				color = "#444444"
			} else {
				color = "#ff4444"
			}
		default:
			continue
		}
		markers = append(markers, render.MiniMapMarker{
			X: r.CenterX(), Y: r.CenterY(), Color: color,
		})
	}
	return markers
}

// Run starts the game loop using requestAnimationFrame.
func (g *Game) Run() {
	var frame js.Func
	frame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if r := recover(); r != nil {
				msg := "unknown panic"
				if err, ok := r.(error); ok {
					msg = err.Error()
				} else if s, ok := r.(string); ok {
					msg = s
				}
				js.Global().Get("console").Call("error", "GORL panic recovered: "+msg)
				g.state = StateMenu
				g.player = nil
				js.Global().Call("requestAnimationFrame", frame)
			}
		}()

		now := args[0].Float()
		if g.lastTime == 0 {
			g.lastTime = now
		}
		dt := (now - g.lastTime) / 1000.0
		g.lastTime = now

		// Cap delta time to prevent spiral of death after tab switch
		if dt > 0.1 {
			dt = 0.1
		}

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
