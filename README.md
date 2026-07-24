# GORL — Go Roguelike

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![WebAssembly](https://img.shields.io/badge/WebAssembly-654FF0?logo=webassembly&logoColor=white)](https://webassembly.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI/CD](https://github.com/Baptiste-lg/GORL/actions/workflows/deploy.yml/badge.svg)](https://github.com/Baptiste-lg/GORL/actions/workflows/deploy.yml)

A browser-based roguelike dungeon crawler written entirely in **Go**, compiled to **WebAssembly** and served via Nginx.

## [Play Online](https://baptiste-lg.github.io/GORL/)

Or run locally with Docker:

```bash
docker build -t gorl . && docker run -p 8080:8080 gorl
```

## Controls

| Key | Action |
|-----|--------|
| WASD / Arrows | Move & attack |
| E / Enter | Interact (shop, shrine) |
| Q | Use active item |
| I | Inventory |
| M | Toggle music |
| ESC | Pause |

## Features

### Core Gameplay
- Turn-based movement and combat on a grid
- BSP procedural dungeon generation with rooms, corridors, and doors
- Symmetric recursive shadowcasting field of vision
- 15+ floors of escalating difficulty with 4 visual themes (Stone, Crypt, Inferno, Abyss)

### Combat & Stats
- 4 core stats: STR, DEX, VIT, LCK with formula-driven mechanics
- Dodge, critical hits, and kill streak XP multipliers
- Level-up system with 3 random stat boost choices per level

### Items & Affixes
- 4 item types: weapons, armor, potions, scrolls
- 4 rarity tiers: Common, Uncommon, Rare, Epic
- **16 affixes** on weapons and armor (Vampiric, Burning, Freezing, Berserker, Thorns, Fortified, Stealth, etc.)
- Affix count scales with rarity

### Active Items
- 8 active abilities charged by kills: Dash, Fireball, Freeze, Heal Burst, Shield Wall, Poison Cloud, Blink, War Cry
- Dropped by bosses, one slot — press Q to activate

### Synergies (24)
- Specific affix + affix or affix + active combos trigger bonus effects
- Weapon combos: Toxic Flame, Shatter, Desperate Gambit, Whirlwind, Bloodthirst
- Cross-equipment combos: Flame Lord, Blood Mirror, Ice Dancer, Juggernaut, Shadow Step, and more
- Active combos: Inferno, Absolute Zero, Crimson Tide, Rage Unleashed, Flash Step, Vanish, Phantom, and more
- Synergies displayed in inventory screen

### Bosses
- 3 boss types: Minotaur (floor 5), Lich (floor 10), Dragon (floor 15)
- Phase system (100-50%, 50-25%, below 25% enraged)
- Unique abilities: charge, stomp, summon skeletons, life drain, fire breath, roar
- Telegraphed danger tiles shown before attacks land

### Special Rooms
- **Treasure rooms** with guaranteed loot and fewer enemies
- **Challenge rooms** with double enemies and rare reward on clear
- **Secret rooms** hidden behind cracked walls with high-value loot
- **Shops** selling randomized equipment for gold
- Color-coded minimap markers for discovered rooms

### Economy & Events
- Gold drops from all kills, spent in shops
- 8 branching text events between floors (Wandering Merchant, Mysterious Fountain, Blood Pact, Cursed Mirror, etc.)
- Risk/reward choices with stat trades, gold gambles, and inventory rewards

### World
- 5 enemy types + 3 bosses with AI states (Idle, Patrol, Chase, Flee)
- Traps (spike, poison, teleport), hazards (lava, poison gas), shrines (blood, fortune, healing)
- 4 starting loadouts unlocked by floor progression
- Persistent unlock state via localStorage

### Audio
- Procedural music engine with explore, combat, and boss tracks
- Full SFX suite (hits, crits, footsteps, pickups, level-up fanfare, death)

### Technical
- Pure Go → WASM, no JavaScript framework
- HTML5 Canvas rendering (80×40 character grid, 3×3 tile sprites)
- Multi-stage Docker build with Nginx, gzip, caching headers
- Panic recovery in game loop to prevent WASM crashes
- ~0 external dependencies

## Architecture

```
game/       Core logic: player, combat, items, affixes, synergies, events, bosses
dungeon/    BSP generation, FOV, tile map
render/     Canvas rendering, HUD, particles, sprites, UI overlays
audio/      Web Audio API engine, procedural music, SFX
web/        Static frontend (HTML, CSS, nginx config)
```

## License

MIT
