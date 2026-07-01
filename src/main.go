package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui/components"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils"
	"github.com/charmbracelet/log"
)

const (
	ManifestJSON = "jar-cart.json"
	ManifestXML  = "jar-cart.xml"
)

func printHelp() {
    fmt.Println("🛒 jar-cart — Modern, zero-config package manager & runner for Java\n")
    t := components.HelpTable()
    fmt.Println(t.View())
}

func isPresentInLib(query string) bool {
    libDir := "lib"
    files, err := os.ReadDir(libDir)
    if err != nil {
        return false
    }
    for _, f := range files {
        if strings.Contains(f.Name(), query) {
            fmt.Printf("✅ %s is already present in lib/\n", f.Name())
            return true
        }
    }
    return false
}

func main() {
	logger := log.New(os.Stderr)
	logger.SetLevel(log.InfoLevel)

	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("jar-cart v0.1.0")
		return
	}

	utils.AutoCheckUpdate("v0.1.0")

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	manifestFile := ManifestJSON
	if _, err := os.Stat(ManifestXML); err == nil {
		manifestFile = ManifestXML
	}

	var filteredArgs []string
	var frozen, useXML, forceJSON bool

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--frozen":
			frozen = true
		case (arg == "-m" || arg == "--manifest") && i+1 < len(os.Args):
			manifestFile = os.Args[i+1]; i++
		case arg == "--xml":
			useXML = true
		case arg == "--json":
			forceJSON = true
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if useXML {
		manifestFile = ManifestXML
	} else if forceJSON {
		manifestFile = ManifestJSON
	}

	switch command {
	case "self-update":
        if err := utils.SelfUpdate("v0.1.0"); err != nil {
            logger.Error("Self-update failed", "error", err)
        }

	case "init":
		projectName := "."
		if len(filteredArgs) > 0 {
			projectName = filteredArgs[0]
		}
		manifestFormat, javaVersion := utils.InteractiveInit("25")
		manifestType := manifestFormat
		if useXML { manifestType = "xml" }

		targetDir, err := utils.HandleInit(projectName, manifestType)
		if err != nil {
			logger.Fatal("Failed to initialize", "error", err)
		}

		logger.Info("Scaffolding project", "path", targetDir, "format", manifestType)
		if err := utils.ExecuteScaffold(targetDir, projectName, "Vanilla", "no-build", "Java", javaVersion, manifestType); err != nil {
			logger.Error("Scaffold failed", "error", err)
		} else {
			_ = utils.GenerateLockFile(targetDir, []models.Dependency{})
			logger.Info("Project ready!")
		}

	case "cache-clear":
		logger.Info("Clearing cache...")
		if err := utils.CleanCache(); err != nil {
			logger.Error("Cache clear failed", "error", err)
		} else {
			logger.Info("Cache cleared successfully.")
		}

	case "search":
		if len(filteredArgs) < 1 {
			logger.Error("Search requires a query string.")
			os.Exit(1)
		}
		utils.SearchMavenCentral(filteredArgs[0])

	case "sync":
		logger.Info("Synchronizing dependencies", "manifest", manifestFile, "frozen", frozen)
		manifest, err := utils.LoadManifest(manifestFile)
		if err != nil {
			logger.Fatal("Manifest load failure", "error", err)
		}

		javaVersion := manifest.JavaVersion
		if javaVersion == "" { javaVersion = "25" }
		
		if err := utils.EnsureJavaVersion(javaVersion); err != nil {
			logger.Fatal("Java provisioning failed", "error", err)
		}

		lockEntries, err := utils.ResolveParallelDependencies(".", manifest.Dependencies, true)
		if err != nil {
			logger.Warn("Sync completed with some errors", "error", err)
		}

		if err := utils.CleanupLibDir(".", lockEntries); err != nil {
			logger.Error("Cleanup failed", "error", err)
		}
		logger.Info("Dependencies synced successfully.")

	case "add":
		if len(filteredArgs) < 1 {
			logger.Error("Add requires a package name.")
			os.Exit(1)
		}

		for _, query := range filteredArgs {
			var g, a, v string
			matches := utils.ScanLocalCache(query)

			if len(matches) > 0 {
				if len(matches) > 1 {
					options := make(map[string]string)
					for _, path := range matches {
						display := strings.ReplaceAll(path, string(os.PathSeparator), ":")
						options[display] = path
					}

					selectedPath, ok := utils.InteractiveSelection(options)
					if !ok { continue }

					dir := filepath.Dir(filepath.Clean(selectedPath))
					filename := filepath.Base(filepath.Clean(selectedPath))
					g = strings.ReplaceAll(dir, string(os.PathSeparator), ".")
					base := strings.TrimSuffix(filename, ".jar")
					lastDash := strings.LastIndex(base, "-")
					if lastDash != -1 {
						a = base[:lastDash]
						v = base[lastDash+1:]
					}
				} else {
					path := matches[0]
					dir := filepath.Dir(filepath.Clean(path))
					filename := filepath.Base(filepath.Clean(path))
					g = strings.ReplaceAll(dir, string(os.PathSeparator), ".")
					base := strings.TrimSuffix(filename, ".jar")
					lastDash := strings.LastIndex(base, "-")
					if lastDash != -1 {
						a = base[:lastDash]
						v = base[lastDash+1:]
					}
					logger.Info("Found in cache", "package", fmt.Sprintf("%s:%s:%s", g, a, v))
				}
			} else {
				if !utils.IsOnline() {
					logger.Error("Network is offline and package not in cache", "query", query)
					continue
				}

				suggestions := utils.GetSearchSuggestions(query)
				if len(suggestions) == 0 {
					logger.Warn("No results found", "query", query)
					continue
				}

				onlineOptions := make(map[string]string)
				for _, res := range suggestions {
					display := fmt.Sprintf("%s:%s:%s", res.G, res.A, res.LatestVersion)
					onlineOptions[display] = display
				}

				targetCoord, ok := utils.InteractiveSelection(onlineOptions)
				if !ok { continue }

				parts := strings.Split(targetCoord, ":")
				if len(parts) == 3 {
					g, a, v = parts[0], parts[1], parts[2]
				} else {
					continue
				}
			}

			target := fmt.Sprintf("%s:%s:%s", g, a, v)
			logger.Info("Processing dependency", "target", target)

			if err := utils.AddDependency(manifestFile, target, false, "lib"); err != nil {
				logger.Error("Sync failed", "target", target, "error", err)
			} else {
				logger.Info("Processed successfully", "target", target)
			}
		}

	case "remove", "rm":
		if len(filteredArgs) < 1 {
			logger.Error("Remove requires target coordinates (group:lib).")
			os.Exit(1)
		}
		if err := utils.RemoveDependency(manifestFile, filteredArgs[0], "lib"); err != nil {
			logger.Error("Removal failed", "error", err)
		} else {
			logger.Info("Dependency removed.")
		}

	case "run":
		if len(filteredArgs) < 1 {
			logger.Error("Run requires a target.")
			return
		}
		target := filteredArgs[0]
		
		manifest, err := utils.LoadManifest(manifestFile)
		if err == nil && manifest.Scripts != nil {
			if _, ok := manifest.Scripts[target]; ok {
				logger.Info("Executing script", "script", target)
				if err := utils.RunScript(target, manifest); err != nil {
					logger.Error("Script execution failed", "error", err)
				}
				return 
			}
		}
		
		logger.Info("Executing target", "target", target)
		utils.RunProject(target)

	case "watch":
			target := "src"
			if len(filteredArgs) > 0 {
				target = filteredArgs[0]
			}
			utils.WatchAndRun(target)

	case "build":
		if err := utils.RunBuild(); err != nil {
			logger.Error("Build failed", "error", err)
			os.Exit(1)
		}
		logger.Info("Build successful!")
	
	case "run-jar":
		if len(filteredArgs) < 1 {
			logger.Error("run-jar requires a jar path.")
			return
		}
		jarPath := filteredArgs[0]
		mainClass := ""
		if len(filteredArgs) > 1 {
			mainClass = filteredArgs[1]
		}
		if err := utils.RunJar(jarPath, mainClass); err != nil {
			logger.Error("Failed to run JAR", "error", err)
		}

	case "help", "-h", "--help":
		printHelp()

	default:
		logger.Warn("Unknown command", "command", command)
		printHelp()
	}
}