# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This Verdant Thane, a reimplementation of the classic Flash game **Azul Baronis**, a top-down 2D space combat game featuring large-scale battles with hundreds of AI-controlled ships. The player controls a single ship within massive battles, respawning into new ships when destroyed until victory or their faction is eliminated.

**Technology Stack**: Go + Ebiten (Ebitengine) game engine

**Current Status**: Milestone 6 complete! All core features implemented. Currently fixing remaining bugs and tweaking gameplay.

**Targets**:
- 60 FPS with 800 ships
- Test coverage of all core functionality
- Feels like the original to play

**Design goals**
- Clean, simple code
- Lots of tests. So many tests. Always be writing tests.

## Build and Development Commands

Once the Go project is initialized, standard Go commands will be used:

```bash
# Initialize the project (Milestone 0)
go mod init github.com/username/azul

# Build the game
go build -o verdant-thane .

# Run the game
go run .

# Run tests
go test ./...

# Run a specific test
go test -run TestName ./path/to/package

# Performance testing with large fleets
./verdant-thane -perf -ships 800 -factions 4 -fps -profile
```

**Performance testing flags:**
- `-perf`: Skip title screen, spawn large fleet immediately
- `-ships N`: Total number of ships to spawn (default: 800)
- `-factions N`: Number of factions (default: 4)
- `-fps`: Show FPS/TPS counter in-game
- `-profile`: Display detailed in-game performance profiling data every second
- `-profile-startup`: Profile startup time from main() to title screen (one-time report)
- `-profile-newgame`: Profile new game initialization from "Play Game" to game loaded (one-time report)

## Architecture and Game Systems

### Important considerations

* Do not pull in modules unnecessarily. Prefer modules distributed from strongly reputable sources when they must be pulled in.
* This game is targeting both web and local execution. Do not compromise this cross-platform ideal.

### Current Architecture (Interface-Based)

The game uses a clean interface-based architecture centered around the **EntityManager**:

**Core Interfaces** (defined in `entity/interfaces.go`):
- **Entity**: Base interface with Update(), Render(), GetID(), GetPosition(), IsAlive()
- **Ship**: Extends Entity with combat, identity, physics, control, and weapon methods
- **Projectile**: Extends Entity with owner tracking, damage, and collision detection
- **GameContext**: Provides entities access to game state without circular dependencies

**EntityManager** (`entity_manager.go`):
- Central orchestrator that manages all entities (ships, projectiles, explosions)
- Maintains entity maps by ID for O(1) lookups
- Builds spatial grid once per frame for efficient collision detection and AI targeting
- Implements discrete update passes for accurate profiling:
  1. Weapons Update: Capacitor charging for all weapons
  2. AI/Player Update: AI targeting/rotation/firing + player input handling
  3. Beam Weapons: Testudon beam weapon updates
  4. Movement: Applies velocity to position with world wrapping
  5. Missile Tracking: Updates homing missile trajectories
  6. Projectile Lifetime: Removes expired projectiles
  7. Collisions: Projectile-ship collision detection using spatial grid
  8. Explosions: Advances explosion animation frames

**Spatial Grid Optimization** (`spatial_grid.go`):
- 128×128 pixel cells for spatial partitioning
- Reduces collision detection from O(n²) to O(n×k)
- AI targeting uses expanding ring search for O(k) nearest-enemy lookup
- Built once per frame, reused for both collisions and targeting

**Performance characteristics (800 ships @ 60 FPS):**

Current implementation achieves excellent performance:
- **Startup time**: ~10ms (218x improvement after optimization - see note below)
- Total frame time: ~1-4ms (6-24% of 16.67ms budget, varies with projectile count)
- AI Update: 0.11-0.32ms (10-35x faster than old donburi implementation!)
- Collision detection: 0.08-2.29ms (varies with projectile count)
- Weapons Update: 0.15-0.42ms
- Movement: 0.01-0.05ms
- Rendering: 1.46-2.92ms total (stars, ships, projectiles, minimap)

The spatial grid optimization provides 40-120x speedup over naive linear search for AI targeting.

**Note on startup optimization**: Initial implementation had 2+ second startup delay due to synchronously decoding entire MP3 files during playback. Optimized by using `stream.Length()` to get decoded size from metadata instead of `io.ReadAll()`, allowing streaming playback without pre-decoding. This reduced startup from 2075ms to 10ms.

### Core Game Loop (Ebiten)

Ebiten requires implementing two main methods:
- `Update()`: Game logic at 60 TPS (ticks per second)
- `Draw()`: Rendering logic

### Key Game Components

**World/Game Board**
- Top-down 2D space sector with wrapping edges (toroidal topology)
- Entities leaving one edge appear on the opposite edge
- Coordinate system needs to handle wrapping for collision detection and rendering

**Ship System**
- Four ship types with distinct capabilities:
  - **Fighter**: Standard ship with forward-facing laser (30° arc), 8 HP, has afterburner
  - **Destroyer**: Fighter + rear-facing homing missile launcher (180° arc), has afterburner
  - **Testudon ("Turtle")**: Slow, durable, 360° laser beam, no afterburner (AI-only in original)
  - **Mothership**: Large objective structure with turrets (enemy-only)
- Ships have inertial movement (acceleration/deceleration)
- Rotation-based steering (not instant)
- Afterburner: Fighters and destroyers can boost speed for short bursts, consumes charge that regenerates over time
- Ship sprites are all oriented upward (along the Y-axis) and must be rotated correctly in order to face the direction of acceleration

**Weapon System**
- **Main Gun**: Projectile-based (not instant), fires within weapon arc toward cursor
  - If cursor outside arc, fires along nearest arc edge
  - 750ms cooldown between shots
  - Projectile speed ~2x max ship speed
  - 1 damage per hit
  - Sometimes referred to as a laser even though it doesn't look or act like a laser
- **Homing Missiles** (Destroyer): Lock-on system with range check
- **Beam Weapon** (Testudon): Instant-fire, short-range, constant damage over time
- Player weapon targeting uses mouse cursor position relative to ship orientation

**Control Scheme**
- W: Accelerate
- S: Decelerate
- A: Rotate left (CCW)
- D: Rotate right (CW)
- Shift: Afterburner (fighters and destroyers only, consumes charge)
- Space: Fire weapon
- P: Pause/Resume
- M: Toggle music mute
- N: Toggle sound effects mute

**Team/Faction System**
- Four team colors: Green (player faction 0), Blue (faction 1), Red (faction 2), Yellow (faction 3)
- Multi-faction battles support 2-4 factions with configurable ship counts per faction
- Fleet spawning: Each faction spawns in circular formation around designated spawn point
- Player always controls green team (faction 0) ships
- Palette swapping: Ships use grayscale base sprite with runtime color transformation via Ebiten ColorM
- Sprite caching: Faction-colored sprites are cached to avoid repeated palette transformations
- Player respawning: When player ship dies, respawns into a random friendly ship (Milestone 5)

**AI System**
- Advanced targeting: AI ships use spatial grid expanding ring search via `GameContext.FindNearestEnemy()` to find closest enemy across world-wrapping boundaries
- Retargeting: AI re-evaluates targets every 60 ticks (~1 second) or when current target becomes invalid
- Pursuit behavior: AI rotates toward target and flies at 80-100% max speed (randomized for variety)
- Patrol behavior: When no target available, maintains heading at 50% max speed
- Firing logic: Probabilistic targeting with 50% accurate shots, 25% random within cone, 25% no fire (when target in 30° firing arc)
- All 800 AI ships updated every frame for responsive battlefield behavior
- Spatial grid optimization enables efficient nearest-enemy searches (O(k) instead of O(n))
- Large-scale coordination between AI ships emerges from individual behaviors

**UI Elements**
The game uses the ebitenui package to implement UI widgets. Avoid manually implementing widgets that could be more easily implemented via ebitenui (but do consider manually implementing buttons if the ebitenui container model would be cumbersome for the specific use case).

- Title screen: "Verdant Thane" title with manual buttons (Play Game, Settings, Instructions, High Scores)
- Settings screen: Audio volume controls (sound/music), mute toggles, chat enable/disable
- Game state management: Title screen → in-game transitions, victory/defeat screens, interstitial screens
- Minimap: Bottom-right corner, shows all ships as colored faction blips, player marked with light green plus sign
- Afterburner bar: Vertical bar left of minimap showing charge level (fighters and destroyers only)
- Chat window: Left of minimap, displays faction-colored messages from AI ships (optional, can be disabled in settings)
- HUD: Displays Shield (top-right), Score (top-left), Kills (top-right below shield)
- Pause overlay: Shows "PAUSED" message and control hints when game is paused (P key)
- FPS/TPS counter: Optional display via `-fps` flag
- Visual focus: Black background with deterministic procedurally-generated white star field

**Audio System**
- Background music: Menu music on title screen, 5 battle tracks that play randomly during gameplay
- Sound effects: Laser fire, projectile impact, ship explosion
- Audio manager handles music/sound playback with volume controls and mute toggles
- Music uses streaming MP3 playback with infinite looping
- Settings persist across sessions (volume levels, mute states)

**Physics Considerations**
- Physics are mostly arcade-style
- Collision detection across wrapping boundaries
- Projectile physics (constant velocity)

## Implementation Milestones

See PLAN.md for detailed milestone breakdowns. Key progression:

1. **Milestone 0**: Project setup, build environment, dependency management ✓
2. **Milestone 1**: Player ship, movement, shooting mechanics, game board ✓
3. **Milestone 2**: Single AI opponent, minimap, ship destruction ✓
4. **Milestone 3**: Multi-faction battles, advanced AI, UI system, performance optimization ✓
5. **Milestone 4**: Destroyer and testudon ship types, balanced fleet composition, spectate mode ✓
6. **Milestone 5**: Player respawning, game over screen, high scores, instructions, interstitial screen, battle progression ✓
7. **Milestone 6**: Audio (music and sound effects), pause functionality, afterburner, settings screen, chat window, additional polish ✓

### Development notes

When instructed to work on a milestone and accomplishing multiple sub-features / development steps in a row, create a separate git commit for each step to maintain good source control discipline. The commit history should read like a list of features, not just "implemented a bunch of stuff". However, do not add the un-committed files in the repository (such as doc/ or CLAUDE.md) as part of the commit. Those are not committed for a reason. Just add any new files you specifically created for the work you have done.

Do not neglect to run `go fmt` when done editing a file.

Write unit tests for any new functionality. When adding functionality, keep testability in mind.

While working, if you observe some code that needs to be cleaned up, an inefficient algorithm that could be improved, or any sort of technical debt, notate it in WORK.md. Also sync your todo list into WORK.md. In fact, use WORK.md as an additional scratch pad for anything to remember or that would be good to communicate with collaborators.

## Important Design Notes

- The fighter's main weapon is sometimes called a "laser" weapon but behaves as a projectile weapon (visible, travels over time)
- Game board wrapping is crucial for spatial calculations
- Mouse cursor position drives weapon aiming within arc constraints
- Scale is important: battles can involve hundreds of entities (target ~800 ships)
- The aesthetic is deliberately simple: pixel art, limited palette, focus on combat

## Reference documents

- **PLAN.md**: Implementation plan with detailed milestone breakdowns
- **WORK.md**: Working document, to be kept up to date regularly
- **doc/DETAILS.md**: Description of Azul Baronis, the game which serves as inspiration for Verdant Thane
- **doc/COLLISION.md**: Spatial grid collision detection design and performance analysis
- **doc/REFACTORING.md**: Documentation of the major refactoring from donburi ECS to interface-based architecture
