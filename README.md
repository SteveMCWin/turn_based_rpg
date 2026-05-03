# Turn-Based RPG

This is a turn based RPG made for the Nordeus Full Stack Challenge. Below are instructions on how to run the project, things I implemented that are and are not in the specification.

---

## How to Run

### With Docker (recommended)

```bash
docker build -t tbrpg .
docker run -p 5000:5000 tbrpg

```

Or run straight from my docker repo, no building required (though this way you cannot tweak the config).
```bash
docker run -p 5000:5000 stevemcwin/tbrpg
```

Then open `http://localhost:5000` in your browser.

### Without Docker

Requires Go 1.26.2 and `gcc` (CGO is needed for the SQLite driver).

```bash
go build -o tbrpg
./tbrpg
```

Then open `http://localhost:5000` in your browser.

---

## Note for Reviewers

The game is a roguelike; losing a run means starting over from the beginning. If you'd like an easier time getting through the game to see all the content, open `config/hero.json` and increase the base stats for whichever class you pick. The server needs to be restarted for config changes to take effect (no recompile needed, but if you are running it with docker, you need to build a new image, as docker bakes all the files into the image).

---

## Specification and Implementation

Some things worth noting:
- Since the spec mentions working with game designers, I focused on making the game easily configurable: item drop rates, hero stat scaling, number of floors per realm, etc. are all in JSON files, while game saves are stored in an SQLite database.
- Moves, monster templates, and similar data are cached (loaded once on startup and kept in memory), as they are read very frequently.

Short comments are left throughout the backend source code explaining the intent and quirks of each function.

### Recommended Reading Order

**`models`** : `stats` → `effect` → `move` → `item` → `entity` → `character` → `encounter` → `shop` → `floor`

The foundation of the game. Defines the core structures (`Stats`, `Entity`, etc.) and the functions that operate on them.

**`game`** : `config` → `game` → `battle` → `shop`

Where the actual game logic lives.

**`database`** : `database` → `save`

Handles database connections and all interactions with SQLite.

**`main.go`**

Initializes the database and config, then starts the server.

**`handlers`** : `handlers`

Connects the frontend to the backend. The frontend is statically served; changes to HTML/CSS take effect on server restart; JS changes require a browser refresh (Ctrl+Shift+R in Firefox) due to caching. All game data (heroes, monsters, fight results, etc.) is exchanged as JSON.

---

The following is a full comparison of the challenge specification against what was built.

### Core Game Mechanics

| Requirement | Notes |
|---|---|
| Turn-based combat | Player/monster turn alternation via `WaitingForMonster` flag |
| Hero fights 5 monsters in sequence (gauntlet) | Configurable amount of floors and rooms per floor, monster appearance chance, boss in last room |
| Hero learns random move from defeated monster | One random move from the monster's moveset is learned after each victory |
| Player can replay/grind fights | 20% chance monster permanently levels up on rematch |
| Equip learned moves before next fight | Max 4 equipped moves; must always keep at least one zero-cost move to avoid soft-locking |

### Stats System

| Requirement | Notes |
|---|---|
| Four stats: Health, Attack, Defense, Magic | Mana added as a 5th stat to balance magical attacks |
| Hero stats increase on level up | Automatic scaling based on config + point allocation per level up |
| Monsters have fixed stats | Defined in `config/monsters.json`, scaled to floor level on room entry |
| Difficulty progression by level | Monsters leveled to `floor.Idx + 1`; boss always at highest floor level |

### Move System & Damage Calculations

| Requirement | Notes |
|---|---|
| Physical moves scale off Attack, reduced by Defense | `max(1, power - defense)` |
| Magic moves scale off Magic, bypass Defense | No defense reduction applied, but require mana |
| Move base value + stat scaling formula | `power = scalingStat * scalingFactor * baseValue / 100` |
| Move levels increase damage | 10% bonus per level (configurable in `config/game.json`) |
| Healing moves | Same formula as damage, capped at max HP |

### Moves

The spec defined the 4 Knight starting moves and the movesets for 5 monsters (Witch, Giant Spider, Dragon, Goblin Warrior, Goblin Mage). All of these are implemented as specified.

Each move can also carry a mana cost (I forgot the spec also mentioned they could cost hp :^/ ). To avoid soft-locking, each character needs at least one 0-cost move. Moves also have an intent tag (`IntentDamage`, `IntentHeal`, `IntentBuff`, `IntentDebuff`) used by the AI to make context-aware decisions.

When a hero learns a move they already know, its level increases rather than adding a duplicate. Each level adds 10% to the move's damage (configurable). The learned move's starting level matches the defeated monster's level, so the reward for beating a higher level monster is reflected here as well.

The game includes 15 monsters, 3 bosses and 35+ moves. Each has their own set of 4 moves defined in `config/monsters.json`.

### Status Effects

| Requirement | Notes |
|---|---|
| Stat modifiers (buffs/debuffs) | `StatModifier` type, ticked each turn |
| Damage over time | `DamageOverTime` type, applied at the end of each turn while active |
| Effect duration & activation delay | `Duration` and `ActivationDelay` / `TurnsToActivate` fields |
| Target (self/opponent) | `EffectTarget` field on each effect |

### Leveling & XP

| Requirement | Notes |
|---|---|
| Battle awards XP | 35 XP × monster level (configurable) |
| XP thresholds trigger level up | Configurable threshold array in `config/game.json` |
| Level up increases stats | Scaling applied automatically; player also allocates points manually or randomly |

### API Endpoints

- GET endpoint to fetch run state/config
- Monster move decision endpoint

---

### Bonus Features (Game Designer's Backlog)

All 15 bonus features from the spec's backlog were implemented:

| Feature | Notes |
|---|---|
| Move descriptions | All moves have descriptions in `config/moves.json` |
| Attribute choices on level up | Player allocates points manually or randomly (amount or manual and rand points configurable, rand gives more) |
| Status effects (Bleed, Poison, Damage Reduction, Damage Increase) | DoT applied at the end of turn so player has a chance to avoid taking damage from it |
| Resource costs (HP/Mana) | Mana cost validation, per-turn regen, restore percentage of manx mana between fights, configurable |
| Save & Exit (mid-run save) | Full SQLite serialisation of game state in `database/save.go` |
| Battle log | Damage, heals, buffs, debuffs, level-ups, and loot all logged |
| Battle animations | Animations handled client-side: smooth bars, sprite animations indicating taking damage, receiving buffs/debuffs |
| Smarter AI | HP-aware move weighting, kill-on-sight logic, buff/debuff tracking |
| Items system | Equip/use/drop system with Armor, Weapon, Trinket, Consumable types, enemies have a chance of having an item equipped |
| Shop system | 4 random equippables + all consumables; buy/sell with gold (bigger cost when buying than selling, configurable) |
| More enemies & moves | 15 monsters + 3 bosses |
| Non-linear map (branching paths) | Non-crossing path algorithm, chance of branching and chance of monster appearance in each room, configurable |
| Environmental effects | 12 environments with per-character effect maps applied on room entry |
| Endless mode | New realm generated on boss defeat; infinite procedural progression, each realm ends with a bossfight |
| Hero classes | 3 classes (Knight, Wizard, Rogue), each with unique stats, moves, starting items and gold |

---

### Extra Features

These were not in the specification but I throught they would be fun to have:

| Feature | Notes |
|---|---|
| Monster level-up on re-grind | 20% chance the monster permanently gains a level (configurable), intended as a way to stop the player from excessive farming |
| Move level stacking | Learning the same move multiple times increases its level |
| Item drop system | Per-item configurable drop rates, items drawn from each monster's loot pool |
| Event system | Non-combat, mistery room encounters that apply permanent stat changes to the hero, cannot be revisited |
| Effect stat scaling | Effects can scale off the caster's stats rather than using a fixed delta |
| Configurable balance knobs | XP thresholds, move level bonus %, gold drops, and more all live in `config/game.json` |
| Boss monsters | Spawn at last floor of each realm, drop twice the gold, have a guaranteed item equipped and drop and cannot be re-fought in endless mode |

---

## Credits

| | |
|---|---|
| **Color palette** | [SLSO8](https://lospec.com/palette-list/slso8) by Luis Miguel Maldonado |
| **Art** | Made by me!! |

