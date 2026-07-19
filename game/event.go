package game

import "math/rand"

// Event is a text-based encounter shown between floors or at special tiles.
type Event struct {
	Title   string
	Text    string
	Choices []EventChoice
}

// EventChoice is one option the player can pick.
type EventChoice struct {
	Label  string
	Action func(p *Player, rng *rand.Rand) EventResult
}

// EventResult is the outcome of choosing an event option.
type EventResult struct {
	Message string
	Color   string
}

// EventState tracks which event is active and what the player selected.
type EventState struct {
	Event    *Event
	Selected int
	Result   *EventResult // nil until a choice is made
}

// RollEvent picks a random event appropriate for the floor.
func RollEvent(floor int, rng *rand.Rand) *Event {
	pool := eventsForFloor(floor)
	if len(pool) == 0 {
		return nil
	}
	return pool[rng.Intn(len(pool))]
}

func eventsForFloor(floor int) []*Event {
	// All floors get the base pool
	pool := make([]*Event, len(baseEvents))
	copy(pool, baseEvents)

	// Deeper floors get riskier events
	if floor >= 5 {
		pool = append(pool, deepEvents...)
	}
	if floor >= 10 {
		pool = append(pool, abyssEvents...)
	}
	return pool
}

// --- Event Definitions ---

var baseEvents = []*Event{
	{
		Title: "Wandering Merchant",
		Text:  "A cloaked figure beckons from the shadows.\n\"I have wares... if you have coin.\"",
		Choices: []EventChoice{
			{Label: "Buy a potion (30g)", Action: eventBuyPotion},
			{Label: "Trade HP for gold", Action: eventTradeHPGold},
			{Label: "Walk away", Action: eventNothing},
		},
	},
	{
		Title: "Mysterious Fountain",
		Text:  "A cracked fountain still flows with shimmering liquid.\nIt smells faintly of iron.",
		Choices: []EventChoice{
			{Label: "Drink deeply", Action: eventFountainDrink},
			{Label: "Wash your weapon", Action: eventFountainWeapon},
			{Label: "Ignore it", Action: eventNothing},
		},
	},
	{
		Title: "Fallen Adventurer",
		Text:  "A skeleton slumps against the wall, clutching a satchel.\nSomething glints inside.",
		Choices: []EventChoice{
			{Label: "Loot the body", Action: eventLootBody},
			{Label: "Pay respects (+1 LCK)", Action: eventPayRespects},
			{Label: "Leave it", Action: eventNothing},
		},
	},
	{
		Title: "Strange Mushrooms",
		Text:  "Bioluminescent mushrooms carpet the floor.\nTheir spores tickle your nose.",
		Choices: []EventChoice{
			{Label: "Eat one (risky)", Action: eventEatMushroom},
			{Label: "Harvest for later (+1 potion)", Action: eventHarvestMushroom},
			{Label: "Pass through carefully", Action: eventNothing},
		},
	},
	{
		Title: "Trapped Chest",
		Text:  "A chest sits suspiciously in the open.\nYou notice a thin wire attached to the lid.",
		Choices: []EventChoice{
			{Label: "Disarm and open (DEX check)", Action: eventDisarmChest},
			{Label: "Force it open (take damage)", Action: eventForceChest},
			{Label: "Leave it alone", Action: eventNothing},
		},
	},
}

var deepEvents = []*Event{
	{
		Title: "Blood Pact",
		Text:  "A demonic circle pulses on the floor.\nPower radiates from its center.",
		Choices: []EventChoice{
			{Label: "Sacrifice 25% HP for +3 STR", Action: eventBloodSTR},
			{Label: "Sacrifice 25% HP for +3 DEX", Action: eventBloodDEX},
			{Label: "Back away slowly", Action: eventNothing},
		},
	},
	{
		Title: "Cursed Mirror",
		Text:  "Your reflection grins at you independently.\nIt mouths something...",
		Choices: []EventChoice{
			{Label: "Touch the mirror", Action: eventMirrorTouch},
			{Label: "Smash it", Action: eventMirrorSmash},
			{Label: "Look away", Action: eventNothing},
		},
	},
}

var abyssEvents = []*Event{
	{
		Title: "The Void Whispers",
		Text:  "Reality thins here. You feel your stats shifting,\nas if the dungeon itself is rearranging you.",
		Choices: []EventChoice{
			{Label: "Embrace the chaos", Action: eventVoidChaos},
			{Label: "Resist (+2 VIT)", Action: eventVoidResist},
			{Label: "Flee", Action: eventNothing},
		},
	},
}

// --- Event Action Implementations ---

func eventNothing(p *Player, rng *rand.Rand) EventResult {
	return EventResult{Message: "You move on.", Color: "#888888"}
}

func eventBuyPotion(p *Player, rng *rand.Rand) EventResult {
	if p.Gold < 30 {
		return EventResult{Message: "Not enough gold!", Color: "#ff4444"}
	}
	if p.Inventory.IsFull() {
		return EventResult{Message: "Inventory full!", Color: "#ff4444"}
	}
	p.Gold -= 30
	p.Inventory.Add(&Item{
		Name: "Health Potion", Type: ItemPotion,
		Rarity: RarityCommon,
	})
	return EventResult{Message: "-30g, gained Health Potion", Color: "#44ff44"}
}

func eventTradeHPGold(p *Player, rng *rand.Rand) EventResult {
	cost := p.Stats.HP / 4
	if cost < 1 {
		cost = 1
	}
	p.Entity.TakeDamage(cost)
	gold := cost*3 + rng.Intn(10)
	p.Gold += gold
	return EventResult{Message: "-" + itoa(cost) + " HP, +" + itoa(gold) + "g", Color: "#FFD700"}
}

func eventFountainDrink(p *Player, rng *rand.Rand) EventResult {
	if rng.Intn(2) == 0 {
		healed := p.Heal(p.EffectiveStats().MaxHP() / 2)
		return EventResult{Message: "Refreshing! +" + itoa(healed) + " HP", Color: "#22cc22"}
	}
	p.Entity.TakeDamage(5)
	return EventResult{Message: "Poison! -5 HP", Color: "#ff4444"}
}

func eventFountainWeapon(p *Player, rng *rand.Rand) EventResult {
	p.Stats.STR += 1
	return EventResult{Message: "+1 STR (permanent)", Color: "#ffaa00"}
}

func eventLootBody(p *Player, rng *rand.Rand) EventResult {
	gold := 15 + rng.Intn(30)
	p.Gold += gold
	return EventResult{Message: "Found " + itoa(gold) + "g!", Color: "#FFD700"}
}

func eventPayRespects(p *Player, rng *rand.Rand) EventResult {
	p.Stats.LCK += 1
	return EventResult{Message: "+1 LCK. Rest well, stranger.", Color: "#aaaaff"}
}

func eventEatMushroom(p *Player, rng *rand.Rand) EventResult {
	roll := rng.Intn(3)
	switch roll {
	case 0:
		p.Stats.VIT += 2
		return EventResult{Message: "Invigorating! +2 VIT", Color: "#44ff44"}
	case 1:
		p.Entity.TakeDamage(8)
		return EventResult{Message: "Toxic! -8 HP", Color: "#ff4444"}
	default:
		p.Stats.DEX += 1
		p.Stats.STR -= 1
		if p.Stats.STR < 1 {
			p.Stats.STR = 1
		}
		return EventResult{Message: "Strange... +1 DEX, -1 STR", Color: "#cc44ff"}
	}
}

func eventHarvestMushroom(p *Player, rng *rand.Rand) EventResult {
	if p.Inventory.IsFull() {
		return EventResult{Message: "Inventory full!", Color: "#ff4444"}
	}
	p.Inventory.Add(&Item{
		Name: "Mushroom Potion", Type: ItemPotion,
		Rarity: RarityUncommon,
	})
	return EventResult{Message: "Gained Mushroom Potion", Color: "#44ff44"}
}

func eventDisarmChest(p *Player, rng *rand.Rand) EventResult {
	if rng.Intn(10) < p.EffectiveStats().DEX {
		gold := 30 + rng.Intn(40)
		p.Gold += gold
		return EventResult{Message: "Disarmed! Found " + itoa(gold) + "g", Color: "#FFD700"}
	}
	p.Entity.TakeDamage(10)
	return EventResult{Message: "Failed! Trap triggered, -10 HP", Color: "#ff4444"}
}

func eventForceChest(p *Player, rng *rand.Rand) EventResult {
	p.Entity.TakeDamage(5)
	gold := 20 + rng.Intn(25)
	p.Gold += gold
	return EventResult{Message: "-5 HP, found " + itoa(gold) + "g", Color: "#FFD700"}
}

func eventBloodSTR(p *Player, rng *rand.Rand) EventResult {
	cost := p.Stats.HP / 4
	if cost < 1 {
		cost = 1
	}
	p.Entity.TakeDamage(cost)
	p.Stats.STR += 3
	return EventResult{Message: "-" + itoa(cost) + " HP, +3 STR!", Color: "#cc2222"}
}

func eventBloodDEX(p *Player, rng *rand.Rand) EventResult {
	cost := p.Stats.HP / 4
	if cost < 1 {
		cost = 1
	}
	p.Entity.TakeDamage(cost)
	p.Stats.DEX += 3
	return EventResult{Message: "-" + itoa(cost) + " HP, +3 DEX!", Color: "#cc2222"}
}

func eventMirrorTouch(p *Player, rng *rand.Rand) EventResult {
	// Swap two random stats
	stats := []*int{&p.Stats.STR, &p.Stats.DEX, &p.Stats.VIT, &p.Stats.LCK}
	i, j := rng.Intn(4), rng.Intn(4)
	for i == j {
		j = rng.Intn(4)
	}
	*stats[i], *stats[j] = *stats[j], *stats[i]
	names := []string{"STR", "DEX", "VIT", "LCK"}
	return EventResult{
		Message: names[i] + " and " + names[j] + " swapped!",
		Color:   "#cc44ff",
	}
}

func eventMirrorSmash(p *Player, rng *rand.Rand) EventResult {
	if rng.Intn(2) == 0 {
		p.Stats.LCK += 2
		return EventResult{Message: "Shards of fortune! +2 LCK", Color: "#FFD700"}
	}
	p.Stats.LCK -= 1
	if p.Stats.LCK < 1 {
		p.Stats.LCK = 1
	}
	return EventResult{Message: "7 years bad luck! -1 LCK", Color: "#ff4444"}
}

func eventVoidChaos(p *Player, rng *rand.Rand) EventResult {
	// Randomize all stats within +/-2 of current
	p.Stats.STR += rng.Intn(5) - 2
	p.Stats.DEX += rng.Intn(5) - 2
	p.Stats.VIT += rng.Intn(5) - 2
	p.Stats.LCK += rng.Intn(5) - 2
	if p.Stats.STR < 1 {
		p.Stats.STR = 1
	}
	if p.Stats.DEX < 1 {
		p.Stats.DEX = 1
	}
	if p.Stats.VIT < 1 {
		p.Stats.VIT = 1
	}
	if p.Stats.LCK < 1 {
		p.Stats.LCK = 1
	}
	return EventResult{Message: "Your stats shift wildly!", Color: "#cc44ff"}
}

func eventVoidResist(p *Player, rng *rand.Rand) EventResult {
	p.Stats.VIT += 2
	return EventResult{Message: "You hold firm. +2 VIT", Color: "#44aaff"}
}
