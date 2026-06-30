# jar-cart

A fast, cross-platform CLI tool written in Go to fetch, cache, and execute Java artifacts using an enterprise-grade dependency pipeline.

---

## 🛒 Overview

**jar-cart** is a modern, zero-configuration package manager and runner for the Java ecosystem.

By leveraging native filesystem **Hard Links**, a **Content Addressable Storage (CAS)** cache, and **Isolated Runtime Provisioning**, `jar-cart` eliminates the friction of traditional build systems while providing near-instant dependency reuse and project-specific Java runtimes.

_Note: This is version 1.0 of the manager._

---

## 🚀 Key Features

### ⚡ Performance & Efficiency

- **CAS Architecture:** Artifacts are downloaded once and stored globally.
- **Hard-Linking:** Instant dependency linking; no duplicated disk usage across projects.

### ☕ Isolated Runtimes

- **Project-Level Version Locking:** Specify any JDK version in `jar-cart.json`. `jar-cart` handles the provisioning, isolation, and version-switching automatically, ensuring your project always runs on the exact runtime it expects.

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

| Command          | Description                                                                                                                                                                                                                                             |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `init`           | Creates an interactive or default project layout.                                                                                                                                                                                                       |
| `add <pkg>`      | Adds an artifact to the manifest and resolves dependencies.                                                                                                                                                                                             |
| `sync`           | Downloads dependencies and synchronizes the project runtime.                                                                                                                                                                                            |
| `run <path>`     | Compiles with the **project-locked JDK** and executes.                                                                                                                                                                                                  |
| `run-jar`        | Runs the built JAR using the project's isolated JDK.                                                                                                                                                                                                    |
| `remove <pkg>`   | Removes a dependency and cleans associated links.                                                                                                                                                                                                       |
| `convert <type>` | Converts manifest formats (e.g., `json` to `xml`).                                                                                                                                                                                                      |
| `cache-clear`    | Clears cached artifacts and metadata.                                                                                                                                                                                                                   |
| `watch <path>`   | Starts a reactive, content-verified file-watcher. Unlike standard watchers that trigger on every file save, `jar-cart` uses SHA256 hashing to ensure re-compilation only occurs when meaningful code changes are detected, preventing redundant builds. |
| `build`          | Packages the project into a standalone, portable Fat JAR.                                                                                                                                                                                               |

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

_Built with ⚡ in Go._

_Designed for developers who value performance and simplicity._ 🏎️💨✨
