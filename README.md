# jar-cart

A fast, cross-platform CLI tool written in Go to fetch, cache, and execute Java artifacts using an enterprise-grade dependency pipeline.

---

## 🛒 Overview

**jar-cart** is a modern, zero-configuration package manager and runner for the Java ecosystem.

By leveraging native filesystem **Hard Links** and a **Content Addressable Storage (CAS)** cache, `jar-cart` eliminates the friction of traditional build systems while providing near-instant dependency reuse and minimal disk overhead.

---

## 🚀 Key Features

### ⚡ Performance

- Native Hard Link (CAS) architecture
- Instant dependency linking across projects
- No duplicated disk usage

### 🔒 Security

- SHA256 integrity verification
- Atomic download protection
- Corruption-resistant synchronization

### 🧩 Zero Configuration

Works out of the box with support for:

- `jar-cart.json`
- `jar-cart.xml`

No complex setup required.

### 🏢 Enterprise Ready

- Custom repository mirrors
- Corporate proxy support
- Strict dependency isolation

### 👨‍💻 Developer Friendly

Simple and intuitive commands:

- `run`
- `add`
- `sync`
- `convert`

---

## 🛠 Quick Start

### Installation

_(Assuming the `jar-cart` binary is already available in your PATH.)_

### Initialize a Project

```bash
jar-cart init my-app
cd my-app
```

### Add Dependencies

```bash
jar-cart add org.slf4j:slf4j-api:2.0.7
```

### Sync & Execute

Download dependencies and hard-link them into the project's `lib/` directory:

```bash
jar-cart sync
```

Compile and run your application:

```bash
jar-cart run src/App.java
```

---

## 📋 Commands

| Command          | Description                                                                                                |
| ---------------- | ---------------------------------------------------------------------------------------------------------- |
| `init`           | Creates an interactive or default project layout.                                                          |
| `add <pkg>`      | Adds an artifact to the manifest, resolves the dependency tree, and automatically updates the lockfile.    |
| `sync`           | Downloads dependencies, validates integrity, and creates hard links.                                       |
| `run <file>`     | Compiles source files and launches the JVM.                                                                |
| `remove <pkg>`   | Removes a dependency and cleans associated links.                                                          |
| `convert <type>` | Converts and replaces the existing manifest (e.g., `json` to `xml`), maintaining a single source of truth. |
| `cache-clear`    | Clears cached artifacts and metadata.                                                                      |

---

## 🏗 Architecture

### Content Addressable Storage (CAS)

Every artifact is downloaded exactly once and stored in:

```text
~/.jar-cart/cache
```

This enables efficient reuse across multiple projects.

### Hard-Linking

Instead of duplicating files, `jar-cart` creates hard links from the global cache into the project's `lib/` directory, providing near O(1) dependency resolution regardless of project size.

### Manifest Strategy

- Clean dependency definitions
- Automatic deduplication
- Support for multiple manifest formats

---

## 🛡 Built for Reliability

### Atomic Operations

Downloads use temporary staging to prevent corruption caused by interrupted transfers.

### Integrity Checks

Every `sync` operation performs SHA256 verification to guard against dependency drift and corrupted artifacts.

### Cross-Platform

Native support for:

- Windows (hard-link optimized)
- Linux
- macOS
- Other POSIX-compliant systems

---

## 🎯 Philosophy

`jar-cart` is designed for developers who value:

- Speed
- Simplicity
- Reproducibility
- Minimal disk usage
- Zero unnecessary complexity

---

_Built with ⚡ in Go._

_Designed for developers who value performance and simplicity._ 🏎️💨✨
