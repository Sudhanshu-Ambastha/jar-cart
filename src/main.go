package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils"
)

const (
	ManifestJSON = "jar-cart.json"
	ManifestXML  = "jar-cart.xml"
)

func printHelp() {
	fmt.Println("🛒 jar-cart — Modern, zero-config package manager & runner for Java")
	fmt.Println("\nUsage:")
	fmt.Println("  jar-cart <command> [arguments] [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  init             Constructs an interactive or default blueprint layout")
	fmt.Println("  cache-clear      Clears all cached blueprints and registry data")
	fmt.Println("  search <query>   Searches Maven Central API for packages")
	fmt.Println("  sync             Synchronizes dependencies")
	fmt.Println("  add <pkg>        Appends an artifact dependency to your manifest")
	fmt.Println("  remove <pkg>     Strips an artifact marker and cleans up the local JAR")
	fmt.Println("  convert <type>   Translates configuration contexts (json|xml)")
	fmt.Println("  run <path>       Compiles and runs a target Java source file")
	fmt.Println("  run-jar <class>  Runs the built JAR with all dependencies/native access")
	fmt.Println("  watch <path>     Starts a reactive file-watcher for live reloads")
	fmt.Println("  build            Packages the project into a portable Fat JAR")
	fmt.Println("  help             Displays this documentation")
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
	var flagFramework, flagLang string

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--frozen":
			frozen = true
		case (arg == "-m" || arg == "--manifest") && i+1 < len(os.Args):
			manifestFile = os.Args[i+1]; i++
		case arg == "--framework" && i+1 < len(os.Args):
			flagFramework = os.Args[i+1]; i++
		case arg == "--lang" && i+1 < len(os.Args):
			flagLang = os.Args[i+1]; i++
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

			targetDir, err := utils.HandleInit(projectName)
			if err != nil {
				fmt.Printf("❌ Failed to initialize project: %v\n", err)
				os.Exit(1)
			}

			s := "no-build"
			f := flagFramework
			if f == "" {
				f = "Vanilla Java Application" 
			}
			l := flagLang
			if l == "" {
				l = "Java" 
			}

			fmt.Printf("\n⚡ Structuring scaffolding for \033[34m%s\033[0m (no-build mode)...\n", targetDir)
			if err := utils.ExecuteScaffold(targetDir, filepath.Base(targetDir), f, s, l, "25"); err != nil {
				fmt.Printf("❌ Scaffold failed: %v\n", err)
			} else {
				fmt.Println("\n\033[32m✨ Project ready! Happy coding! 🛒🏎️💨\033[0m")
				utils.LaunchWorkspace(targetDir)
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
			dependencies, err := utils.ParseManifest(manifestFile)
			if err != nil {
				fmt.Printf("❌ Manifest parsing failure: %v\n", err)
				os.Exit(1)
			}
			_ = utils.EnsureJavaVersion("25")
			lockEntries, err := utils.ResolveParallelDependencies(".", dependencies)
			if err != nil {
				fmt.Printf("❌ Workspace synchronization loop failed: %v\n", err)
				os.Exit(1)
			}
			lock := utils.LockFile{
				Version: 1, GeneratedAt: time.Now().Format(time.RFC3339), Dependencies: lockEntries,
			}
			_ = utils.WriteLockFile(".", &lock)
			fmt.Println("✨ Dependencies synced and linked perfectly!")

		case "add":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'add' requires at least one target coordinate.")
				os.Exit(1)
			}
			for _, coord := range filteredArgs {
				fmt.Printf("➕ Adding: %s\n", coord)
				if err := utils.AddDependency(manifestFile, coord, false, "lib"); err != nil {
					fmt.Printf("❌ Failed to add %s: %v\n", coord, err)
					os.Exit(1) 
				}
			}
			fmt.Println("✨ All dependencies updated and locked successfully!")

		case "remove", "rm":
			if len(filteredArgs) < 1 {
				fmt.Println("❌ Error: 'remove' requires target coordinates.")
				os.Exit(1)
			}
			if err := utils.RemoveDependency(manifestFile, filteredArgs[0], "lib"); err != nil {
				fmt.Printf("❌ Dependency removal failure: %v\n", err)
				os.Exit(1)
			}

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
				fmt.Println("❌ Error: 'run' requires a target file.")
				os.Exit(1)
			}
			if _, err := os.Stat(manifestFile); err == nil {
				deps, _ := utils.ParseManifest(manifestFile)
				_, _ = utils.ResolveParallelDependencies(".", deps)
			}
			utils.RunProject(filteredArgs[0])
		
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