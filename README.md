# jar-cart

A fast, cross-platform CLI tool written in Go to fetch, cache, and execute Java artifacts using an enterprise-grade dependency pipeline.

---

## 🛒 Overview

**jar-cart** is a modern, zero-configuration package manager and runner for the Java ecosystem.

By leveraging native filesystem **Hard Links**, a **Content Addressable Storage (CAS)** cache, and **Isolated Runtime Provisioning**, `jar-cart` eliminates the friction of traditional build systems while providing near-instant dependency reuse and project-specific Java runtimes.

_Note: This is version ![Version](https://img.shields.io/github/v/release/Sudhanshu-Ambastha/jar-cart?label=Version&color=blue) of the manager._

---

## 🚀 Key Features

### ⚡ Performance & Efficiency

- **CAS Architecture:** Artifacts are downloaded once and stored globally.
- **Hard-Linking:** Instant dependency linking; no duplicated disk usage across projects.

### ☕ Isolated Runtimes

- **Project-Level Version Locking:** Specify any JDK version in `jar-cart.json`. `jar-cart` handles the provisioning, isolation, and version-switching automatically, ensuring your project always runs on the exact runtime it expects.

### 🛠 Reverse Engineering & Patching

- **Automated Decompilation:** Seamlessly extract source code from binary JARs. `jar-cart` automatically provisions, caches, and manages decompiler engines (Vineflower, CFR, Procyon) so you don't have to.

- **Full-Circle Rebuilds:** Decompiled code is "compilation-ready." Effortlessly patch, modify, and rebuild binary-only projects back into valid, executable JARs.

### 🧠 Intelligent CLI Experience

- **Auto-Main Class Detection:** Forget manual manifest configuration. `jar-cart` intelligently scans compiled binaries to locate the `main` method, ensuring your JARs are executable the moment they are built.

- **Smart Path Resolution:** `run-jar` understands your project context—simply provide the JAR name or directory, and jar-cart resolves the path and dependencies automatically.

### 📜 Custom Script Runner

- **NPM-style Automation:** Define project-specific command aliases (e.g., `test`, `build`, `hello`) in your `jar-cart.json`. `jar-cart` handles the lifecycle, including `pre-` and `post-` hook execution automatically.

### 🔒 Security & Reliability

- **SHA256 Verification:** Integrity checks for every artifact.
- **Atomic Operations:** Prevents corruption from interrupted downloads.

### 🧩 Zero Configuration

Works out of the box with `jar-cart.json` or `jar-cart.xml`.

---

## 🛠 Quick Start

### Installation

Use the official install scripts to get started instantly:

**Windows (PowerShell):**
```powershell
iwr https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.ps1 -UseBasicParsing | iex
```

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.sh | bash
```

> **Note:** By default, these scripts fetch the latest version. If you need to install a specific version, you can override this:

**Windows (PowerShell):**
```powershell
iwr https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.ps1 -OutFile install.ps1
```
```powershell
.\install.ps1 -Version v0.1.1
```

**Linux / macOS:**
```bash
VERSION=v0.1.1 curl -sSL https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.sh | bash
```

### Initialize a Project

```bash
jar-cart init my-app
cd my-app
```

---

## 📜 Custom Scripts

`jar-cart` allows you to define project-specific command aliases in your `jar-cart.json` file.

### Defining Scripts

```json
{
  "project": "my-app",
  "java_version": "25",
  "resolution_depth": "full",
  "scripts": { ... },
  "dependencies": []
}
```

### Executing Scripts

Run your scripts using the `run` command. `jar-cart` automatically triggers `pre-` and `post-` hooks if defined:

```bash
jar-cart run hello
jar-cart run test
```

---

## 📋 Commands

| Command | Description |
| :--- | :--- |
| `init` | Creates an interactive project layout with JDK locking. |
| `ls-java` | Inventory all managed JDK runtimes. |
| `cache list/ls` | Displays inventory and storage usage of cached artifacts. |
| `cache remove/rm` | Fuzzy-match removal for JAR artifacts and JDKs. |
| `cache-clear` | Wipes all cached artifacts and registry data. |
| `search <query>` | Searches Maven Central API for packages. |
| `sync` | Synchronizes dependencies and provisions local runtimes. |
| `add <pkg>` | Adds an artifact dependency to your manifest. |
| `remove <pkg>` | Removes dependency and cleans local links. |
| `convert <type>` | Translates manifest formats (e.g., `json` to `xml`). |
| `run <path>` | Compiles and executes with the project-locked JDK. |
| `run-jar <jar>` | Runs built JAR using the project's isolated JDK. |
| `decompile <jar>` | Extracts source code via --engine (vineflower/ cfr/ procyon). |
| `watch <path>` | Starts a reactive, hash-verified file-watcher for incremental builds. |
| `build` | Packages the project into a standalone, portable Fat JAR. |
| `help` | Displays this documentation. |

---
## 🏗 Architecture

### Isolated Java Runtimes

`jar-cart` bypasses the system's global `PATH` and `JAVA_HOME`. By locking your project to a specific JDK version in your `jar-cart.json`, the tool manages a local, immutable toolchain for each project, guaranteeing identical behavior across all environments.

### Content Addressable Storage (CAS)

Every artifact is stored in `~/.jar-cart/cache`, enabling efficient reuse across multiple projects.

### Hard-Linking

`jar-cart` creates hard links from the global cache into the project's `lib/` directory, providing near O(1) dependency resolution without duplicating disk usage.

---

## 🎯 Philosophy

`jar-cart` is designed for developers who value:

- **Speed:** Instant builds and execution.
- **Reproducibility:** Projects run the same way on every machine.
- **Simplicity:** No global environment variables required.
- **Minimalism:** Low disk usage and zero unnecessary complexity.

---

## License

jar-cart is licensed under the ![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg).

_Built with ⚡ in Go._

_Designed for developers who value performance and simplicity._ 🏎️💨✨
