package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui/components"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils"
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
	utils.AutoCheckUpdate("v0.0.5")

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
		case "init":
			projectName := "."
			if len(filteredArgs) > 0 {
				projectName = filteredArgs[0]
			}

			manifestFormat, javaVersion := utils.InteractiveInit("25") 
			
			manifestType := manifestFormat
			if useXML { 
				manifestType = "xml" 
			}

			targetDir, err := utils.HandleInit(projectName, manifestType)
			if err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(ui.TitleStyle.Render("⚡ Scaffolding " + targetDir + " (" + manifestType + " mode)..."))
			
			if err := utils.ExecuteScaffold(targetDir, projectName, "Vanilla", "no-build", "Java", javaVersion, manifestType); err != nil {
				fmt.Printf("❌ Scaffold failed: %v\n", err)
			} else {
				_ = utils.GenerateLockFile(targetDir, []models.Dependency{})
				fmt.Println(ui.SuccessStyle.Render("\n✨ Project ready! 🛒"))
			}
		case "cache-clear":
			fmt.Println("🧹 Clearing all cached blueprints and registry data...")
			if err := utils.CleanCache(); err != nil {
				fmt.Printf("❌ Failed to clear cache: %v\n", err)
			} else {
				fmt.Println("✅ Cache cleared. Next run will fetch fresh metadata from the web.")
			}

		case "search":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'search' requires a query string.")
				os.Exit(1)
			}
			utils.SearchMavenCentral(filteredArgs[0])

		case "sync":
			fmt.Printf("📦 Synchronizing via: %s (Frozen: %v)\n", manifestFile, frozen)
			
			manifest, err := utils.LoadManifest(manifestFile)
			if err != nil {
				fmt.Printf("❌ Manifest loading failure: %v\n", err)
				os.Exit(1)
			}

			javaVersion := manifest.JavaVersion
			if javaVersion == "" {
				javaVersion = "25"
			}
			fmt.Printf("🔍 Using Java version: %s\n", javaVersion)

			if err := utils.EnsureJavaVersion(javaVersion); err != nil {
				fmt.Printf("❌ Failed to ensure Java %s: %v\n", javaVersion, err)
				os.Exit(1)
			}

			lockEntries, err := utils.ResolveParallelDependencies(".", manifest.Dependencies)
			if err != nil {
				fmt.Printf("❌ Workspace synchronization loop failed: %v\n", err)
				os.Exit(1)
			}

			err = utils.CleanupLibDir(".", lockEntries)
			if err != nil {
				fmt.Printf("❌ Cleanup failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("✨ Dependencies synced and linked perfectly!")

		case "add":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'add' requires a package name.")
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
						fmt.Printf("🔍 Found in cache: %s:%s:%s\n", g, a, v)
					}
				} else {
					if !utils.IsOnline() {
						fmt.Printf("❌ '%s' not found in cache and network is offline.\n", query)
						continue
					}

					suggestions := utils.GetSearchSuggestions(query)
					if len(suggestions) == 0 {
						fmt.Printf("❌ No results found for '%s' in cache or online.\n", query)
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
				fmt.Printf("➕ Processing: %s\n", target)

				if err := utils.AddDependency(manifestFile, target, false, "lib"); err != nil {
					fmt.Printf("⚠️ Sync failed: %v\n", err)
				} else {
					fmt.Printf("✅ %s processed successfully!\n", target)
				}
			}
		case "remove", "rm":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'remove' requires target coordinates.")
				os.Exit(1)
			}
			if err := utils.RemoveDependency(manifestFile, filteredArgs[0], "lib"); err != nil {
				fmt.Printf("❌ Dependency removal failure: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✨ Dependency removed and lockfile updated.")

		case "convert":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: Target transformation variant required (json|xml).")
				os.Exit(1)
			}
			if err := utils.ConvertManifest(manifestFile, filteredArgs[0]); err != nil {
				fmt.Printf("❌ Context conversion breakdown: %v\n", err)
				os.Exit(1)
			}

		case "run":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'run' requires a target (file or script).")
				return
			}

			target := filteredArgs[0]
			manifest, err := utils.LoadManifest(manifestFile)
			
			if err == nil && manifest.Scripts != nil {
				if _, ok := manifest.Scripts[target]; ok {
					fmt.Printf("🔍 Detected script: %s\n", target)
					err := utils.RunScript(target, manifest)
					if err != nil { 
						fmt.Printf("❌ Script failed: %v\n", err) 
					}
					return 
				}
			}
			
			fmt.Printf("🔍 Detected Java file: %s\n", target)
			utils.RunProject(target)
		
		case "watch":
			if len(filteredArgs) < 1 {
				utils.WatchAndRun("src")
			} else {
				utils.WatchAndRun(filteredArgs[0])
			}
		
		case "build":
			if err := utils.RunBuild(); err != nil {
				fmt.Printf("❌ Build failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✨ Build successful!")
		
		case "run-jar":
			mainClass := ""
			if len(filteredArgs) > 0 {
				mainClass = filteredArgs[0]
			}
			
			if err := utils.RunJar("dist/app.jar", mainClass); err != nil {
				fmt.Printf("❌ Execution failed: %v\n", err)
				os.Exit(1)
			}

		case "help", "-h", "--help":
			printHelp()

		default:
			fmt.Printf("❌ Unknown command: '%s'\n", command)
			printHelp()
			os.Exit(1)
	}
}