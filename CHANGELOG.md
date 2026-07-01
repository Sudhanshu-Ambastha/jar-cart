# Changelog

All notable changes to this project will be documented in this file.

## [0.1.1] - 2026-07-01

### Added

- **Integrated Decompiler**: Added `decompile` command for on-demand provisioning and execution of industry-standard Java decompilers (Vineflower, CFR, Procyon).
- **Automated Provisioning**: Decompiler engines are now automatically downloaded, cached, and managed by `jar-cart` upon first use.
- **Manifest Conversion**: Added `convert` command to transform project manifests between supported formats (JSON, XML).
- **Documentation**: Updated `help` table and README to include new commands and engine support.

### Improved

- **Build Pipeline**: Enhanced build logic to automatically detect `Main-Class` and filter artifacts, facilitating a seamless Decompile-Modify-Rebuild loop.

---

## [0.1.0] - 2026-06-30

### Added

- **Core CLI**: Initial production release of `jar-cart`.
- **Package Management**: Native support for searching, scaffolding, and managing Java project dependencies.
- **Build System**: Automated cross-platform build pipeline for Windows, Linux, and macOS.
- **Security**:
  - Implemented SHA256 integrity verification for all binary releases.
  - Automated code-signing support for Windows binaries.
- **Professionalism**:
  - Embedded Windows metadata (Version, Copyright, Description).
  - Integrated self-update mechanism with hash validation.
- **User Experience**: Professional console UI with loading spinners and clean table outputs.

---
