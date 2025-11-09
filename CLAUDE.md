# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a reimplementation of the classic Flash game **Azul Baronis** (to be named "Verdant Thane"), a top-down 2D space combat game featuring large-scale battles with hundreds of AI-controlled ships. The player controls a single ship within massive battles, respawning into new ships when destroyed until their faction is eliminated.

**Technology Stack**: Go + Ebiten (Ebitengine) game engine + donburi (ECS framework)

**Current Status**: Milestone 5 is complete. The game features large-scale multi-faction battles (2-4 factions, 800+ ships) with three ship types (fighter, destroyer, testudon), advanced AI targeting, full game loop with respawning, game over screen with high score saving, instructions screen, high scores display, victory/interstitial screens, and battle progression. Performance target of 60 FPS with 800 ships has been achieved.

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
- `-profile`: Display detailed performance profiling data every second

## Architecture and Game Systems

### Important considerations

* Do not pull in modules unnecessarily. Prefer modules distributed from strongly reputable sources when they must be pulled in.
* This game is targeting both web and local execution. Do not compromise this cross-platform ideal.

### ECS Architecture (donburi)

The game uses an Entity Component System (ECS) architecture via the donburi library:
- **Entities**: Ships, projectiles, explosions are all entities with unique IDs
- **Components**: Data-only structures (Position, Velocity, Health, Weapon, etc.) stored in components
- **Systems**: Pure functions that operate on entities with specific component combinations
- **Tags**: Used for entity classification (PlayerControlled, AIControlled, IsShip, IsProjectile, etc.)

Key systems include:
- `UpdatePlayerInput`: Handles WASD controls and rotation
- `UpdateAIMovement`: AI decision-making for movement with nearest-enemy targeting (~1 second retarget intervals)
- `UpdateAIFiring`: AI weapon firing logic with probabilistic targeting
- `UpdateWeapons`: Capacitor charging for all weapons
- `UpdateMovement`: Applies velocity to position with world wrapping
- `UpdateCollisions`: **Optimized** projectile-ship collision detection using spatial grid partitioning (see COLLISION.md)
- `UpdateProjectileLifetime`: Removes expired projectiles
- `UpdateExplosions`: Advances explosion animation frames
- Render systems: `RenderShips`, `RenderProjectiles`, `RenderExplosions`, `RenderMinimap`

**Performance characteristics (800 ships @ 60 FPS):**
- Total frame time: ~1.5ms (8.8% of 16.67ms budget)
- Collision detection: 0.46ms (spatial grid optimization)
- AI movement: 0.03ms (all ships updated every frame)
- Rendering: 0.74ms total (stars, ships, projectiles, minimap)

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
  - **Fighter**: Standard ship with forward-facing laser (30° arc), 8 HP
  - **Destroyer**: Fighter + rear-facing homing missile launcher (180° arc)
  - **Testudon ("Turtle")**: Slow, durable, 360° laser beam (AI-only in original)
  - **Mothership**: Large objective structure with turrets (enemy-only)
- Ships have inertial movement (acceleration/deceleration)
- Rotation-based steering (not instant)

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
- Space: Fire weapon

**Team/Faction System**
- Four team colors: Green (player faction 0), Blue (faction 1), Red (faction 2), Yellow (faction 3)
- Multi-faction battles support 2-4 factions with configurable ship counts per faction
- Fleet spawning: Each faction spawns in circular formation around designated spawn point
- Player always controls green team (faction 0) ships
- Palette swapping: Ships use grayscale base sprite with runtime color transformation via Ebiten ColorM
- Sprite caching: Faction-colored sprites are cached to avoid repeated palette transformations
- Permadeath with immediate respawn into nearby friendly ship (not yet implemented)

**AI System**
- Advanced targeting: AI ships use `SelectNearestEnemy()` to find closest enemy across world-wrapping boundaries
- Retargeting: AI re-evaluates targets every 60 ticks (~1 second) or when current target becomes invalid
- Pursuit behavior: AI rotates toward target and flies at 80-100% max speed (randomized for variety)
- Patrol behavior: When no target available, maintains heading at 50% max speed
- Firing logic: 50% accurate, 25% random within cone, 25% no fire (when target in 30° firing arc)
- All 800 AI ships updated every frame for responsive battlefield behavior
- Large-scale coordination between AI ships is a key feature

**UI Elements**
The game uses the ebitenui package to implement UI widgets. Avoid manually implementing widgets that could be more easily implemented via ebitenui (but do consider manually implementing buttons if the ebitenui container model would be cumbersome for the specific use case).

- Title screen: "Verdant Thane" title with manual buttons (Play Game, Settings, Instructions, High Scores)
- Game state management: Title screen → in-game transitions
- Minimap: Bottom-right corner, shows all ships as colored faction blips, player marked with light green plus sign
- HUD: Displays Shield (top-right), Score (top-left), Kills (top-right below shield)
- FPS/TPS counter: Optional display via `-fps` flag
- Visual focus: Black background with deterministic procedurally-generated white star field

**Physics Considerations**
- Inertial movement (velocity-based, not position-based)
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
7. **Future**: Audio (music and sound effects), pause functionality, afterburner, settings screen, additional polish

### Development notes

When instructed to work on a milestone and accomplishing multiple sub-features / development steps in a row, create a separate git commit for each step to maintain good source control discipline. The commit history should read like a list of features, not just "implemented a bunch of stuff". However, do not add the un-committed files in the repository (such as doc/ or CLAUDE.md) as part of the commit. Those are not committed for a reason. Just add any new files you specifically created for the work you have done.

Do not neglect to run `go fmt` when done editing a file.

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
- **doc/COLLISION.md**: Spatial grid collision detection design, performance analysis, and ECS integration considerations
