# jar-cart

A fast, cross-platform CLI tool written in Go to fetch, cache, and execute Java artifacts using an enterprise-grade dependency pipeline.

---

## 🛒 Overview

**jar-cart** is a modern, zero-configuration package manager and runner for the Java ecosystem.

By leveraging native filesystem **Hard Links**, a **Content Addressable Storage (CAS)** cache, and **Isolated Runtime Provisioning**, `jar-cart` eliminates the friction of traditional build systems while providing near-instant dependency reuse and project-specific Java runtimes.

---

## 🚀 Key Features

### ⚡ Performance & Efficiency

- **CAS Architecture:** Artifacts are downloaded once and stored globally.
- **Hard-Linking:** Instant dependency linking; no duplicated disk usage across projects.

### ☕ Isolated Runtimes (New!)

- **Automated JDK Provisioning:** Forget `JAVA_HOME` configuration. `jar-cart` automatically downloads and isolates specific JDK versions (17, 21, 25, etc.) for each project.

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

### Initialize a Project

```bash
jar-cart init my-app
cd my-app
```

### Sync & Execute

Download dependencies and provision the required Java runtime automatically:

```bash
jar-cart sync
jar-cart run-jar
```

---

## 📋 Commands

| Command          | Description                                                            |
| ---------------- | ---------------------------------------------------------------------- |
| `init`           | Creates an interactive or default project layout.                      |
| `add <pkg>`      | Adds an artifact to the manifest and resolves dependencies.            |
| `sync`           | Downloads dependencies and synchronizes the project runtime.           |
| `run <file>`     | Compiles source files and launches the JVM.                            |
| `run-jar`        | Runs the built JAR using the project's isolated JDK and native access. |
| `remove <pkg>`   | Removes a dependency and cleans associated links.                      |
| `convert <type>` | Converts manifest formats (e.g., `json` to `xml`).                     |
| `cache-clear`    | Clears cached artifacts and metadata.                                  |
| `watch <path>`   | Starts a reactive file-watcher for live reloads.                       |
| `build`          | Packages the project into a standalone, portable Fat JAR.              |

---

## 🏗 Architecture

### Isolated Java Runtimes

`jar-cart` manages Java runtimes inside `~/.jar-cart/jdks/`. When you run a project, the tool ensures the requested version is provisioned, keeping your system clean and your projects reproducible.

### Content Addressable Storage (CAS)

Every artifact is stored in:

```text
~/.jar-cart/cache
```

This enables efficient reuse across multiple projects, drastically reducing storage footprint.

### Hard-Linking

Instead of duplicating files, `jar-cart` creates hard links from the global cache into the project's `lib/` directory, providing near O(1) dependency resolution.

---

## 🎯 Philosophy

`jar-cart` is designed for developers who value:

- **Speed:** Instant builds and execution.
- **Reproducibility:** Projects run the same way on every machine.
- **Simplicity:** No global environment variables required.
- **Minimalism:** Low disk usage and zero unnecessary complexity.

---

_Built with ⚡ in Go._  
_Designed for developers who value performance and simplicity._ 🏎️💨✨
