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

- **Content Addressable Storage (CAS):** Artifacts are downloaded once and shared across all projects.
- **Hard-Linking:** Dependencies are linked instantly into projects without duplicating disk usage.

### ☕ Isolated Runtimes

- **Project-Level Version Locking:** Pin a specific JDK version in `jar-cart.json`. `jar-cart` automatically provisions, isolates, and manages the required runtime, ensuring every project executes with the exact Java version it was developed against.

### 🛠 Reverse Engineering & Patching

- **Integrated Decompilation:** Automatically provisions and manages Vineflower, CFR, and Procyon so JARs can be decompiled without any manual setup.

- **Decompiler-to-Build Workflow:** Decompiled projects can be modified, rebuilt, and packaged back into executable JARs using the normal `jar-cart` toolchain.

### 🧠 Intelligent CLI Experience

- **Automatic Entry Point Detection:** `jar-cart` automatically locates the application's `main()` class, eliminating manual manifest configuration for most projects.

- **Flexible Target Resolution:** `run` and `run-jar` accept source files, directories, class names, or JAR names and automatically resolve the correct execution target.

- **NPM-style Argument Forwarding:** Everything after `--` is forwarded directly to the Java application, allowing runtime arguments without conflicting with `jar-cart`'s own CLI options.

  ```sh
  jar-cart run src -- --server
  jar-cart run App -- --port 8080 --debug
  jar-cart run-jar app.jar -- --profile production
  jar-cart watch src -- --server
  ```

### 📜 Custom Script Runner

- **Lifecycle Scripts:** Define reusable commands (such as `build`, `test`, or `lint`) in `jar-cart.json`. `jar-cart` automatically executes matching `pre-` and `post-` lifecycle hooks, similar to npm.

### 🔒 Security & Reliability

- **SHA256 Verification:** Downloaded artifacts and release binaries are verified before use.
- **Atomic Operations:** Prevents partially downloaded artifacts or interrupted updates from corrupting the local installation.

### 🚀 Developer Experience

- **Hot Reload Development:** `watch` automatically recompiles and restarts applications whenever Java source files change.
- **Persistent Runtime Arguments:** Forwarded application arguments are preserved across every automatic restart.
- **Consistent CLI:** `run`, `run-jar`, and `watch` all follow the same npm-style argument forwarding syntax.

### 🧩 Zero Configuration

Works out of the box with `jar-cart.json` or `jar-cart.xml`.

---

## 🛠 Quick Start

### Installation

Use the official install scripts to get started instantly:

**Windows (PowerShell):**

```sh
iwr https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.ps1 -UseBasicParsing | iex
```

**Linux / macOS:**

```bash
curl -sSL https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.sh | bash
```

> **Note:** By default, these scripts fetch the latest version. If you need to install a specific version, you can override this:

**Windows (PowerShell):**

```sh
iwr https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.ps1 -OutFile install.ps1
```

```sh
.\install.ps1 -Version v0.2.1
```

**Linux / macOS:**

```sh
VERSION=v0.2.1 curl -sSL https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/scripts/install.sh | bash
```

### Initialize a Project

```sh
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

| Command                      | Description                                                               |
| :--------------------------- | :------------------------------------------------------------------------ |
| `init`                       | Creates an interactive project layout with JDK locking.                   |
| `ls-java`                    | Lists all managed JDK runtimes.                                           |
| `cache list/ls`              | Displays cached artifacts and storage usage.                              |
| `cache remove/rm`            | Removes cached JARs and JDKs using fuzzy matching.                        |
| `cache-clear`                | Clears all cached artifacts and registry data.                            |
| `search <query>`             | Searches Maven Central for packages.                                      |
| `sync`                       | Synchronizes dependencies and provisions project runtimes.                |
| `add <pkg>`                  | Adds a dependency to the project manifest.                                |
| `remove <pkg>`               | Removes a dependency from the project manifest.                           |
| `convert <type>`             | Converts manifests between supported formats (`json`/`xml`).              |
| `run <path> [-- args...]`    | Compiles and runs a project, forwarding application arguments.            |
| `run-jar <jar> [-- args...]` | Runs a JAR, forwarding application arguments.                             |
| `decompile <jar>`            | Decompiles JARs using Vineflower, CFR, or Procyon.                        |
| `watch <path> [-- args...]`  | Watches, recompiles, and restarts while preserving application arguments. |
| `build`                      | Packages the project into a portable Fat JAR.                             |
| `help`                       | Displays this documentation.                                              |

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
