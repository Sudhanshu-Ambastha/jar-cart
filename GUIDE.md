# 🛒 jar-cart v0.6.0: Complete User Guide

Welcome to the official guide for **jar-cart** (`jc`), a high-performance, zero-configuration package manager, build orchestrator, and runtime manager designed for modern Java development.

---

# 📑 Table of Contents

- [🛒 jar-cart v0.6.0: Complete User Guide](#-jar-cart-v060-complete-user-guide)
- [📑 Table of Contents](#-table-of-contents)
- [1. Introduction \& Core Philosophy](#1-introduction--core-philosophy)
  - [🚀 Zero-Configuration Development](#-zero-configuration-development)
  - [☕ Isolated Java Runtimes](#-isolated-java-runtimes)
  - [💾 Content-Addressable Storage (CAS)](#-content-addressable-storage-cas)
- [2. Installation \& Package Managers](#2-installation--package-managers)
  - [Official Installation Scripts](#official-installation-scripts)
    - [Windows (PowerShell)](#windows-powershell)
    - [Linux / macOS](#linux--macos)
  - [Package Manager Integration (Homebrew \& Chocolatey)](#package-manager-integration-homebrew--chocolatey)
    - [🍏 Homebrew (macOS \& Linux)](#-homebrew-macos--linux)
    - [🍫 Chocolatey (Windows)](#-chocolatey-windows)
  - [Verifying the `jc` CLI Alias](#verifying-the-jc-cli-alias)
- [3. Project Scaffolding \& Initialization Modes](#3-project-scaffolding--initialization-modes)
  - [🚀 The Interactive `init` Workflow](#-the-interactive-init-workflow)
  - [🏗 Scaffolding Architecture Modes](#-scaffolding-architecture-modes)
    - [📄 `flat`](#-flat)
    - [🏢 `backend`](#-backend)
  - [🎨 Custom Local Templates](#-custom-local-templates)
- [4. Dependency Management \& CAS Architecture](#4-dependency-management--cas-architecture)
  - [📦 Content-Addressable Storage (CAS)](#-content-addressable-storage-cas-1)
    - [⚡ Hard-Linking](#-hard-linking)
    - [💾 Disk Efficiency](#-disk-efficiency)
    - [🔒 Integrity Verification](#-integrity-verification)
  - [🌐 Google `deps.dev` Insights API (v3)](#-google-depsdev-insights-api-v3)
  - [➕ Adding \& Synchronizing Dependencies](#-adding--synchronizing-dependencies)
- [5. Multi-Module Workspace Orchestrator](#5-multi-module-workspace-orchestrator)
  - [🏗 The Workspace Manifest](#-the-workspace-manifest)
  - [🔄 Workspace-Wide Synchronization \& Auditing](#-workspace-wide-synchronization--auditing)
  - [🧩 Topological Multi-Module Builds](#-topological-multi-module-builds)
- [6. Execution, Scripting \& Hot Reloading](#6-execution-scripting--hot-reloading)
  - [▶️ Running Projects \& Scripts](#️-running-projects--scripts)
  - [⌨️ NPM-Style Argument Forwarding](#️-npm-style-argument-forwarding)
  - [🔥 Live Hot Reloading with `watch`](#-live-hot-reloading-with-watch)
- [7. Enterprise Optimization, JLink \& Reverse Engineering](#7-enterprise-optimization-jlink--reverse-engineering)
  - [⚡ Manifest-Driven Optimization (`optimize`)](#-manifest-driven-optimization-optimize)
  - [☕ Custom Lightweight Runtimes (`jlink`)](#-custom-lightweight-runtimes-jlink)
  - [🛠 Built-In Decompilation Toolkit](#-built-in-decompilation-toolkit)
    - [Supported Engines \& Commands](#supported-engines--commands)
- [8. Advanced JLink \& Module Resolution Mechanics](#8-advanced-jlink--module-resolution-mechanics)
  - [🧠 Understanding Module Graphs](#-understanding-module-graphs)
  - [📊 Verifying Optimized Builds](#-verifying-optimized-builds)
- [9. Troubleshooting \& FAQ](#9-troubleshooting--faq)
  - [❓ Why isn't my global `JAVA_HOME` being used?](#-why-isnt-my-global-java_home-being-used)
  - [❓ How do I clean orphaned cache entries or installed JDKs?](#-how-do-i-clean-orphaned-cache-entries-or-installed-jdks)
  - [❓ Does `jar-cart` support multi-module workspaces?](#-does-jar-cart-support-multi-module-workspaces)
- [🎉 You're Ready!](#-youre-ready)

---

# 1. Introduction & Core Philosophy

Traditional Java build tooling often introduces unnecessary complexity through heavyweight configuration, duplicated dependencies, and global environment conflicts.

**jar-cart** addresses these challenges with a modern, enterprise-grade, content-addressable dependency pipeline.

## 🚀 Zero-Configuration Development

Run Java source files directly without cumbersome XML/POM boilerplate while retaining the flexibility to scale into sophisticated multi-module workspaces.

## ☕ Isolated Java Runtimes

Each project specifies its required JDK version.

`jar-cart` automatically provisions and manages local runtimes, eliminating reliance on global `JAVA_HOME` configuration.

## 💾 Content-Addressable Storage (CAS)

Dependencies are downloaded once into a global cache and shared across projects using native filesystem hard links.

Benefits include:

- Near-zero duplicate disk usage
- Faster project synchronization
- Instant dependency reuse

---

# 2. Installation & Package Managers

## Official Installation Scripts

Install the latest version of **jar-cart** using the official installer for your platform.

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.ps1 -UseBasicParsing | iex
```

### Linux / macOS

```bash
curl -sSL https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.sh | bash
```

---

## Package Manager Integration (Homebrew & Chocolatey)

`jar-cart` supports native package manager installation for simplified updates and system-wide management.

### 🍏 Homebrew (macOS & Linux)

```bash
brew tap Sudhanshu-Ambastha/tap
brew trust Sudhanshu-Ambastha/tap
brew install jar-cart
```

### 🍫 Chocolatey (Windows)

```powershell
choco install jar-cart
```

---

## Verifying the `jc` CLI Alias

When installed via supported package managers, `jar-cart` automatically registers the lightweight `jc` alias.

Both commands invoke the same executable:

```bash
jar-cart --version
jc --version
```

---

# 3. Project Scaffolding & Initialization Modes

`jar-cart` includes a flexible project scaffolding system that can generate complete project layouts, manifests, and runtime configuration in seconds.

Whether you're creating a simple script, an enterprise backend, or a reusable SDK, `jar-cart` provides templates to get you started quickly.

## 🚀 The Interactive `init` Workflow

Create a new project interactively with:

```bash
jc init my-app
```

The interactive wizard guides you through:

1. Project Name
2. Project Directory
3. Target Java Version (e.g. `17`, `21`, `25`)
4. Manifest Format (`jar-cart.json` or `jar-cart.xml`)
5. Project Architecture Template

Once complete, the project structure, manifest, and runtime configuration are generated automatically.

## 🏗 Scaffolding Architecture Modes

### 📄 `flat`

A minimal project layout ideal for:

- Small utilities
- Command-line applications
- Learning Java
- Quick prototypes

```text
my-app/
├── jar-cart.json
└── Main.java
```

### 🏢 `backend`

A structured layout designed for enterprise services and web applications.

```text
my-app/
├── jar-cart.json
├── src/
│   ├── controller/
│   ├── service/
│   ├── repository/
│   ├── model/
│   │   └── User.java
│   ├── resources/
│   └── sql/
```

Suitable for:

- REST APIs
- Backend services
- Spring-style architectures
- Database-driven applications

## 🎨 Custom Local Templates

Register reusable project layouts in your local configuration to instantly scaffold company-standard or personal templates.

---

# 4. Dependency Management & CAS Architecture

`jar-cart` modernizes Java dependency management through a **Content-Addressable Storage (CAS)** engine.

## 📦 Content-Addressable Storage (CAS)

Artifacts are downloaded once into:

```text
~/.jar-cart/cache
```

### ⚡ Hard-Linking

Dependencies are linked into each project using native filesystem hard links.

Benefits include:

- Near-instant linking
- Zero duplicated files
- Minimal disk usage

### 💾 Disk Efficiency

Multiple projects reuse identical dependencies automatically.

### 🔒 Integrity Verification

SHA-256 verification ensures downloaded artifacts remain authentic.

## 🌐 Google `deps.dev` Insights API (v3)

`jar-cart` integrates directly with Google's **deps.dev** ecosystem for:

- Accurate transitive dependency resolution
- Security auditing
- Dependency graph analysis
- Package health verification

## ➕ Adding & Synchronizing Dependencies

```bash
jc add gson
jc sync
```

---

# 5. Multi-Module Workspace Orchestrator

Manage multiple Java projects from a single workspace using `jar-cart.workspace.json`.

## 🏗 The Workspace Manifest

```json
{
  "workspace": "enterprise-suite",
  "modules": {
    "core-sdk": {
      "path": "modules/core-sdk"
    },
    "web-api": {
      "path": "modules/web-api"
    }
  }
}
```

## 🔄 Workspace-Wide Synchronization & Auditing

Synchronize every module:

```bash
jc sync
```

Audit all dependencies:

```bash
jc audit
```

## 🧩 Topological Multi-Module Builds

Projects are automatically built in dependency order.

```bash
jc build
jc script test
```

---

# 6. Execution, Scripting & Hot Reloading

## ▶️ Running Projects & Scripts

```bash
jc run src/Main.java
jc run hello
```

## ⌨️ NPM-Style Argument Forwarding

```bash
jc run src -- --server-port=8080 --debug
```

## 🔥 Live Hot Reloading with `watch`

```bash
jc watch src -- --live-reload
```

---

# 7. Enterprise Optimization, JLink & Reverse Engineering

## ⚡ Manifest-Driven Optimization (`optimize`)

```json
"optimize": {
  "compression": 2,
  "strip_debug": true,
  "strip_native": true
}
```

Supported features:

- Compression
- Debug symbol stripping
- Native metadata stripping
- Production packaging

---

## ☕ Custom Lightweight Runtimes (`jlink`)

Generate optimized Java runtimes containing only the required platform modules.

```bash
jc optimize dist/app.jar dist/my-runtime
```

Benefits:

- Smaller runtime size
- Faster startup
- Reduced container footprint
- Lower memory usage

---

## 🛠 Built-In Decompilation Toolkit

Need to inspect or understand third-party libraries without source code?

`jar-cart` includes support for multiple Java decompilation engines and automatically provisions the required tooling.

### Supported Engines & Commands

- **Vineflower** _(default)_: The modern, highly accurate decompiler fork of Fernflower.

  ```bash
  jc decompile libs/legacy-library.jar --engine vineflower
  ```

- **CFR**: Excellent support for modern Java language features (lambdas, records, switch expressions).

  ```bash
  jc decompile libs/legacy-library.jar --engine cfr
  ```

- **Procyon**: Robust decompiler known for handling complex anonymous classes and modern constructs gracefully.
  ```bash
  jc decompile libs/legacy-library.jar --engine procyon
  ```

This is useful for:

- Library inspection
- Debugging
- Compatibility analysis
- Migration projects
- Reverse engineering

---

# 8. Advanced JLink & Module Resolution Mechanics

## 🧠 Understanding Module Graphs

`jar-cart` combines **jdeps** with recursive dependency tracing (`-R`) and your project's `lib/*` classpath to build accurate Java module graphs.

When external libraries such as **Gson**, **Jackson**, or other third-party dependencies are present, `jar-cart` automatically determines the required platform modules (for example `java.base`, `java.logging`, or `java.sql`) and supplies them to `jlink`.

This produces minimal runtime images that include everything required to execute the application while excluding unnecessary JDK modules.

---

## 📊 Verifying Optimized Builds

Inspect optimization metrics during runtime generation:

```bash
jc optimize dist/app.jar dist/my-runtime
```

The optimization report includes:

- Original application size
- Generated runtime size
- Required Java modules
- Compression statistics
- Overall size reduction

---

# 9. Troubleshooting & FAQ

## ❓ Why isn't my global `JAVA_HOME` being used?

`jar-cart` intentionally ignores the global `JAVA_HOME`.

Every project declares its required Java version in its manifest, allowing reproducible and isolated builds.

---

## ❓ How do I clean orphaned cache entries or installed JDKs?

```bash
jc cache list
jc cache remove java25
jc cache remove gson-2.14.0.jar
jc cache-clear
```

---

## ❓ Does `jar-cart` support multi-module workspaces?

Yes.

Commands such as:

- `jc sync`
- `jc audit`
- `jc build`
- `jc run`

automatically discover modules declared in `jar-cart.workspace.json` and execute operations using topological ordering.

---

# 🎉 You're Ready!

You now have everything needed to build, manage, optimize, and deploy Java applications with **jar-cart**.

Explore the repository for additional examples, advanced workflows, and community contributions.

Happy coding! ☕🚀
