package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui/components"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils"
	"github.com/charmbracelet/log"
)

const (
	ManifestJSON = "jar-cart.json"
	ManifestXML  = "jar-cart.xml"
	Version = "v0.6.0"
)

func printHelp() {
    fmt.Println("🛒 jar-cart — Modern, zero-config package manager & runner for Java\n")
    fmt.Println(components.HelpTable())
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
	needsUpdate, latestVersion := utils.AutoCheckUpdate(Version)
	
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		if arg == "--version" || arg == "-v" {
			fmt.Printf("jar-cart %s\n", Version)
			return
		}
	}

	defer func() {
		if needsUpdate {
			fmt.Println("\n" + components.UpdateNotification(Version, latestVersion))
		}
	}()

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			logger.Info("Done", "duration", elapsed.Round(time.Millisecond).String())
		}
	}()

	manifestFile := ManifestJSON
	if _, err := os.Stat(ManifestXML); err == nil && os.Getenv("JAR_CART_XML") == "1" {
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
		if len(filteredArgs) > 0 {
			targetVersion := filteredArgs[0]

			logger.Info("Switching version...", "target", targetVersion)

			if err := utils.DowngradeTo(targetVersion); err != nil {
				logger.Error("Version switch failed", "error", err)
				return
			}

			logger.Info("Version switched successfully", "version", targetVersion)
			return
		}

		if err := utils.SelfUpdate(Version); err != nil {
			logger.Error("Self-update failed", "error", err)
			return
		}

		logger.Info("Successfully updated to the latest release.")
		return

	case "init":
		projectName := "."
		if len(filteredArgs) > 0 {
			projectName = filteredArgs[0]
		}
		
		strategyName := "flat"
		for i := 0; i < len(filteredArgs); i++ {
			if filteredArgs[i] == "--strategy" && i+1 < len(filteredArgs) {
				strategyName = filteredArgs[i+1]
				break
			}
		}

		manifestFormat, javaVersion := utils.InteractiveInit("")
		if manifestFormat == "" {
			logger.Fatal("Initialization cancelled or failed.")
		}

		manifestType := manifestFormat
		if useXML { 
			manifestType = "xml" 
		}
		
		targetDir, err := utils.HandleInit(projectName, manifestType, javaVersion, strategyName)
		if err != nil {
			logger.Fatal("Failed to initialize", "error", err)
		}

		logger.Info("Checking Java runtime...", "version", javaVersion)
		if err := utils.EnsureJavaVersion(javaVersion); err != nil {
			logger.Fatal("Java provisioning failed", "error", err)
		}

		logger.Info("Scaffolding project", "path", targetDir, "format", manifestType, "strategy", strategyName)
		if err := utils.ExecuteScaffold(targetDir, projectName, "Vanilla", strategyName, "Java", javaVersion, manifestType); err != nil {
			logger.Error("Scaffold failed", "error", err)
		} else {
			manifestName := ManifestJSON
			if manifestType == "xml" {
				manifestName = ManifestXML
			}

			manifestPath := filepath.Join(targetDir, manifestName)

			manifest, err := utils.LoadManifest(manifestPath)
			if err != nil {
				logger.Warn("Failed to load generated manifest", "error", err)
			} else if err := utils.GenerateLockFile(targetDir, manifest); err != nil {
				logger.Warn("Failed to generate lockfile", "error", err)
			}

			logger.Info("Project ready!")
		}

	case "cache-clear":
		logger.Info("Clearing cache...")
		if err := utils.CleanCache(); err != nil {
			logger.Error("Cache clear failed", "error", err)
		} else {
			logger.Info("Cache cleared successfully.")
		}
	
	case "list-java","ls-java":
        utils.ListJDKs()

	case "cache":
        cacheArgs := os.Args[2:]
        if len(cacheArgs) < 1 {
            fmt.Println("Usage: jar-cart cache [list|ls|remove|rm] [targets...]")
            return
        }

        subCmd := cacheArgs[0]
        switch subCmd {
        case "list", "ls":
            utils.ListCache()
        case "remove", "rm":
            if len(cacheArgs) < 2 {
                log.Error("Specify targets (e.g., java17, gson-2.14.0.jar)")
                return
            }
            
            for _, target := range cacheArgs[1:] {
                if strings.HasPrefix(target, "java") {
                    if err := utils.RemoveJDK(strings.TrimPrefix(target, "java")); err != nil {
                        log.Error("Failed to remove JDK", "version", target, "error", err)
                    } else {
                        log.Info("JDK removed successfully", "version", target)
                    }
                } else {
                    if err := utils.RemoveCachedItem(target); err != nil {
                        log.Error("Failed to remove artifact", "target", target, "error", err)
                    } else {
                        log.Info("Artifact removed successfully", "target", target)
                    }
                }
            }
        default:
            fmt.Println("Unknown cache command. Use 'list' or 'remove'.")
        }

	case "search":
		if len(filteredArgs) < 1 {
			logger.Error("Search requires a query string.")
			os.Exit(1)
		}
		utils.SearchMavenCentral(filteredArgs[0])

	case "convert":
        if len(filteredArgs) < 1 {
            logger.Error("Convert requires a target format (json/xml).")
            return
        }
        targetFormat := strings.ToLower(filteredArgs[0])
        if err := utils.ConvertManifest(manifestFile, targetFormat); err != nil {
            logger.Error("Conversion failed", "error", err)
        } else {
            logger.Info("Conversion successful", "to", targetFormat)
        }
	
	case "template", "register-template":
		if len(filteredArgs) < 1 {
			logger.Error("Template command requires a subcommand or key (e.g., jar-cart template list, jar-cart template remove <key>, or jar-cart template <key> [path]).")
			return
		}

		subCommand := filteredArgs[0]
		switch subCommand {
		case "list", "ls":
			logger.Info("Listing available templates...")
			if err := utils.ListCustomTemplates(); err != nil {
				logger.Error("Failed to list templates", "error", err)
			}
			return

		case "remove", "rm":
			if len(filteredArgs) < 2 {
				logger.Error("Template removal requires a key (e.g., jar-cart template remove <key>).")
				return
			}
			templateKey := filteredArgs[1]
			logger.Info("Removing custom template...", "key", templateKey)
			if err := utils.RemoveCustomTemplate(templateKey); err != nil {
				logger.Error("Failed to remove custom template", "error", err)
			} else {
				logger.Info("Custom template removed successfully from local registry!", "key", templateKey)
			}
			return

		default:
			templateKey := subCommand
			projectPath := "."
			if len(filteredArgs) > 1 {
				projectPath = filteredArgs[1]
			}

			logger.Info("Registering custom template...", "key", templateKey, "path", projectPath)
			if err := utils.RegisterCustomTemplate(templateKey, projectPath); err != nil {
				logger.Error("Failed to register custom template", "error", err)
			} else {
				logger.Info("Custom template registered successfully in local registry!", "key", templateKey)
			}
			return
		}

	case "sync":
		if _, err := os.Stat("jar-cart.workspace.json"); err == nil {
			logger.Info("Synchronizing workspace modules...")
			ws := utils.LoadWorkspaceManifest()
			for modName, modConfig := range ws.Modules {
				logger.Info("Syncing module", "module", modName, "path", modConfig.Path)
				if err := utils.RunSync(modConfig.Path); err != nil {
					logger.Error("Module sync failed", "module", modName, "error", err)
				}
			}
			logger.Info("Workspace synchronized successfully.")
			return
		}

		if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
			logger.Fatal(
				"Project manifest not found",
				"expected", "jar-cart.json or jar-cart.xml",
				"hint", "Run 'jar-cart init'",
			)
		}
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

		shouldResolve := utils.IsFullResolution(manifest)
		lockEntries, err := utils.ResolveParallelDependencies(
			".",
			manifest.Dependencies,
			shouldResolve,
		)
		if err != nil {
			logger.Warn("Sync completed with some errors", "error", err)
		}

		if err := utils.CleanupLibDir(".", lockEntries); err != nil {
			logger.Error("Cleanup failed", "error", err)
		}
		logger.Info("Dependencies synced successfully.")

	case "audit":
		if _, err := os.Stat("jar-cart.workspace.json"); err == nil {
			logger.Info("Auditing workspace-wide dependencies for vulnerabilities...")
			ws := utils.LoadWorkspaceManifest()
			vulns, err := utils.CheckWorkspaceVulnerabilities(ws)
			if err != nil {
				logger.Error("Workspace audit request failed", "error", err)
				return
			}

			found := false
			for _, result := range vulns.Results {
				for _, v := range result.Vulns {
					found = true
					fmt.Printf("⚠️  [%s] %s\n   %s\n\n", v.ID, v.Summary, v.Details)
				}
			}

			if !found {
				fmt.Println("✅ No known vulnerabilities found across workspace modules.")
			}
			return
		}

		manifest, err := utils.LoadManifest(manifestFile)
		if err != nil {
			logger.Fatal("Failed to load manifest for audit", "error", err)
		}

		logger.Info("Auditing dependencies for vulnerabilities...")
		vulns, err := utils.CheckVulnerabilities(manifest.Dependencies)
		if err != nil {
			logger.Error("Audit request failed", "error", err)
			return
		}

		found := false
		for _, result := range vulns.Results {
			for _, v := range result.Vulns {
				found = true
				fmt.Printf("⚠️  [%s] %s\n   %s\n\n", v.ID, v.Summary, v.Details)
			}
		}

		if !found {
			fmt.Println("✅ No known vulnerabilities found.")
		}

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
		appArgs := utils.GetForwardedArgs()
		if _, err := os.Stat("jar-cart.workspace.json"); err == nil {
			ws := utils.LoadWorkspaceManifest()
			if len(filteredArgs) > 0 {
				targetMod := filteredArgs[0]
				if modConfig, exists := ws.Modules[targetMod]; exists {
					logger.Info("Executing workspace module", "module", targetMod)
					utils.RunModuleProject(modConfig.Path, "", appArgs, ws)
					return
				}
			}

			logger.Info("Executing workspace monorepo automatically...")
			utils.RunWorkspaceAutomatic(ws)
			return
		}

		if len(filteredArgs) < 1 {
			logger.Error("Run requires a target.")
			return
		}

		target := filteredArgs[0]
		manifest, err := utils.LoadManifest(manifestFile)
		if err == nil && manifest.Scripts != nil {
			if _, ok := manifest.Scripts[target]; ok {
				logger.Info(
					"Executing script",
					"script", target,
					"args", appArgs,
				)

				if err := utils.RunScript(target, appArgs, manifest); err != nil {
					logger.Error("Script execution failed", "error", err)
				}
				return
			}
		}

		logger.Info(
			"Executing target",
			"target", target,
			"args", appArgs,
		)

		utils.RunProject(target, appArgs)

	case "watch":
		target := "src"
		if len(filteredArgs) > 0 {
			target = filteredArgs[0]
		}

		appArgs := utils.GetForwardedArgs()

		if _, err := os.Stat("jar-cart.workspace.json"); err == nil {
			ws := utils.LoadWorkspaceManifest()
			if modConfig, exists := ws.Modules[target]; exists {
				logger.Info("Watching workspace module", "module", target, "path", modConfig.Path, "args", appArgs)
				utils.WatchAndRun(filepath.Join(modConfig.Path, "src"), appArgs)
				return
			}
			
			logger.Warn("In workspace mode, specify the module name to watch (e.g., jar-cart watch <module-name>) or run from a module directory.")
			return
		}

		logger.Info(
			"Watching target",
			"target", target,
			"args", appArgs,
		)

		utils.WatchAndRun(target, appArgs)

	case "build":
		if _, err := os.Stat("jar-cart.workspace.json"); err == nil {
			ws := utils.LoadWorkspaceManifest()
			if err := utils.RunWorkspaceBuild(ws); err != nil {
				logger.Error("Workspace build failed", "error", err)
				os.Exit(1)
			}
			logger.Info("Workspace build successful!")
			return
		}

		if err := utils.RunBuild(); err != nil {
			logger.Error("Build failed", "error", err)
			os.Exit(1)
		}
		logger.Info("Build successful!")

	case "optimize":
		if len(os.Args) < 4 {
			log.Fatal("Usage: jar-cart optimize <jar-path> <output-dir>")
		}
		jarPath := os.Args[2]
		outputDir := os.Args[3]
		
		err := utils.OptimizeJarForDeployment(jarPath, outputDir)
		if err != nil {
			log.Fatal("Optimization failed", "error", err)
		}
		log.Info("Runtime optimized successfully!")
	
	case "run-jar":
		if len(filteredArgs) < 1 {
			logger.Error("run-jar requires a JAR path.")
			return
		}

		jarPath := filteredArgs[0]

		mainClass := ""
		if len(filteredArgs) > 1 {
			mainClass = filteredArgs[1]
		}

		appArgs := utils.GetForwardedArgs()

		logger.Info(
			"Executing JAR",
			"jar", jarPath,
			"main", mainClass,
			"args", appArgs,
		)

		if err := utils.RunJar(jarPath, mainClass, appArgs); err != nil {
			logger.Error("Failed to run JAR", "error", err)
		}
	
	case "decompile":
        if len(filteredArgs) < 1 {
            logger.Error("Decompile requires a jar file path.")
            return
        }
        
        jarPath := filteredArgs[0]
        decompileCmd := flag.NewFlagSet("decompile", flag.ExitOnError)
        enginePtr := decompileCmd.String("engine", "vineflower", "Decompiler engine to use (vineflower, cfr, procyon)")
        decompileCmd.Parse(filteredArgs[1:])
        logger.Info("Starting decompilation", "jar", jarPath, "engine", *enginePtr)
        
        if err := utils.Decompile(jarPath, *enginePtr); err != nil {
            logger.Error("Decompilation failed", "error", err)
        } else {
            logger.Info("Decompilation successful")
        }

	case "help", "-h", "--help":
		printHelp()

	default:
		logger.Warn("Unknown command", "command", command)
		printHelp()
	}
}