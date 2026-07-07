# Changelog

All notable changes to this project will be documented in this file.

## [0.4.0] - 2026-07-07

### 🚀 Added

#### Manifest-Driven Runtime Optimization

- Introduced `optimize` configuration object in `jar-cart.json` and `jar-cart.xml`.
- Users can now declaratively control `compression`, `strip_debug` (symbols), and `strip_native` (native symbols) settings directly from the manifest.
- Integrated `models` package to provide type-safe optimization configuration across the codebase.

### ⚡ Improved

#### Optimized Deployment Pipeline

- Enhanced `optimize` command to dynamically resolve optimization profiles based on the project manifest.
- Improved runtime packaging with safety checks (automatic cleaning, existence validation, and manifest-aware version resolution).
- Achieved ~87% reduction in runtime footprint (from ~290MB to ~12MB) for standard Java applications.
- Unified `JLinkOptimizer` to utilize shared configuration models, reducing hardcoded flags.

#### Performance Telemetry

- Implemented global command execution timing.
- Added automatic duration tracking (`Done duration=Xs`) to all CLI commands for improved developer feedback and performance monitoring.

#### Metadata & Consistency

- Synchronized internal versioning system (`v0.4.0`) across all metadata files and build resources.
- Updated `HelpTable()` documentation to reflect the new manifest-configured workflow for the `optimize` command.

---

## [0.3.2] - 2026-07-06

### 🐛 Fixed

#### XML Manifest JDK Resolution

- Fixed an issue where `GetJDKPaths` always assumed `jar-cart.json` as the manifest source.
- Projects using `jar-cart.xml` would incorrectly fall back to **JDK 25**, ignoring the configured JDK version.
- JDK resolution now correctly detects and reads whichever manifest (`jar-cart.json` or `jar-cart.xml`) is present.

### ✨ Improved

#### 📄 Manifest Detection

- Introduced a reusable `DetectManifestFile` utility.
- Ensures consistent manifest detection and resolution across all commands.
- Reduces duplicated manifest lookup logic.

#### ⚡ Incremental Builds

- Added **SHA-256 content-aware hashing** to the `RunProject` lifecycle.
- `jar-cart` now verifies project integrity before recompiling.
- Eliminates redundant build cycles when source files are unchanged.
- Avoids unnecessary deletion and recreation of the `bin/` directory.

#### 🔄 Unified Hashing Utility

- Consolidated file hashing functionality into the `utils` package.
- Removes duplicated hashing logic between the watcher and runner components.
- Improves maintainability and consistency.

#### 👀 Watcher Debouncing & Stability

- Improved the file watcher with:
  - Content verification before rebuilds.
  - **500 ms debounce** interval.
- Provides more reliable hot reloads.
- Better handles atomic file saves from modern editors and IDEs.
- Reduces unnecessary rebuilds triggered by duplicate file system events.

---

## [0.3.1] - 2026-07-03

### Improved

- **Update Notifications**: Automatic update checks now display a non-blocking notification when a newer version is available, allowing users to easily discover new releases while keeping command execution uninterrupted.
- **CLI Experience**: Improved update notification workflow to better align with modern package managers such as npm, pnpm, and bun by displaying update information using cached background checks.

---

## [0.3.0] - 2026-07-03

### Added

- **Application Argument Forwarding**: `run`, `run-jar`, `watch`, and custom `run` scripts now forward application arguments to Java programs using the familiar `--` separator, following npm-style CLI conventions.

### Improved

- **CLI Consistency**: Unified argument forwarding behavior across `run`, `run-jar`, `watch`, and custom scripts, providing a consistent developer experience.
- **Watch Mode**: `watch` now preserves forwarded application arguments across automatic recompilation and restart cycles.
- **Help Documentation**: Updated the built-in help output and README to document forwarded application arguments and npm-style usage patterns.

---

## [0.2.3] - 2026-07-03

### Added

- **Version Lifecycle Management**: Added support for switching to specific released versions using `jar-cart self-update <version>`.
- **Minimum Supported Version**: Enforced a minimum supported version (`v0.2.1`) to prevent downgrading to deprecated or unsupported releases.
- **Configurable Dependency Resolution**: Introduced `resolutionDepth` with `shallow` and `full` modes, allowing projects to install only direct dependencies or the complete transitive dependency graph.

### Improved

- **Unified Update Pipeline**: Consolidated download, checksum verification, and OS-aware binary replacement into a single safe update workflow with rollback protection.
- **Semantic Versioning**: Adopted `golang.org/x/mod/semver` for reliable version validation, comparison, and lifecycle management.
- **Automatic Update Check**: Package manager now checks for updates in the background and caches the latest release information for future notifications.
- **Dependency Resolution Consistency**: Standardized dependency resolution logic across `add`, `sync`, and `lockfile` generation using a single unified rule (`IsFullResolution`), ensuring consistent behavior for `shallow` and `full` modes throughout the system.
- **Project Validation**: Commands that require a project manifest now fail immediately when neither `jar-cart.json` nor `jar-cart.xml` is present, preventing unintended dependency resolution and misleading follow-up errors.
- **Windows Update Stability**: Improved executable replacement with stronger error handling, rollback behavior, and automatic cleanup of temporary update artifacts after successful self-updates.

---

## [0.2.2] - 2026-07-02

### Fixed

- **UI Clipping:** Fixed viewport clipping issues in the help and search tables.
- **Visual Polish:** Updated UI borders to use rounded corners and a vibrant neon-blue color scheme for improved readability.

---

## [0.2.1] - 2026-07-02

### Fixed

- **CLI Visibility:** Removed the `-H=windowsgui` linker flag to ensure proper stdout/stderr attachment to the console on Windows.
- **Flag Parsing:** Added support for `--v` as a valid alias for the version command.

---

## [0.2.0] - 2026-07-02

### Added

- **Java Runtime Discovery**: Added `ls-java` command to inventory locally managed JDKs.
- **Granular Cache Management**: Standardized `cache list` and `cache remove` with fuzzy-matching support for intelligent disk cleanup.
- **Environment Parity**: `init` command now enforces interactive Java version pinning.

### Improved

- **Smart Artifact Routing**: Unified routing engine for JDK management and JAR dependencies.
- **Self-Update Engine**: Completely rewritten self-update logic. Now supports **cross-platform ZIP/TAR extraction**, ensuring binary integrity and compatibility across Windows, Linux, and macOS.
- **Robust CLI**: Added audit logging for bulk operations and hardened CLI responsiveness.
- **Build Pipeline**: Optimized build scripts with explicit architecture targeting (Windows/POSIX) and automated SHA256 integrity verification for all release assets.

### Fixed

- **Windows Binary Compatibility**: Resolved executable linking issues by correctly generating and embedding Windows manifest resources (`resource.syso`).
- **Update Failure Recovery**: Added automatic binary swapping with fallback queueing to handle OS file locks during self-updates.

---

## [0.1.1] - 2026-07-01

### Added

- **Integrated Decompiler**: Added `decompile` command for on-demand provisioning and execution of industry-standard Java decompilers (Vineflower, CFR, Procyon).
- **Automated Provisioning**: Decompiler engines are now automatically downloaded, cached, and managed by `jar-cart` upon first use.
- **Manifest Conversion**: Added `convert` command to transform project manifests between supported formats (JSON, XML).
- **Documentation**: Updated `help` table and README to include new commands and engine support.

### Improved

- **Build Pipeline**: Enhanced build logic to automatically detect `Main-Class` and filter artifacts, facilitating a seamless Decompile-Modify-Rebuild loop.
- **Security & Infrastructure**: Upgraded self-update mechanism to use per-asset SHA256 sidecar verification, ensuring robust and tamper-proof binary updates.

---

## [0.1.0] - 2026-06-30

### Added

- **Core CLI**: Initial production release of `jar-cart`.
- **Package Management**: Native support for searching, scaffolding, and managing Java project dependencies.
- **Build System**: Automated cross-platform build pipeline for Windows, Linux, and macOS.

### Security

- Implemented SHA256 integrity verification for all binary releases.
- Automated code-signing support for Windows binaries.

### Professionalism

- Embedded Windows metadata (Version, Copyright, Description).
- Integrated self-update mechanism with hash validation.

### User Experience

- Professional console UI with loading spinners and clean table outputs.

---
