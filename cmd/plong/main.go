package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/srmdn/plong/internal/clean"
	"github.com/srmdn/plong/internal/scan"
	"github.com/srmdn/plong/internal/serve"
)

var version = "0.2.6"

func main() {
	if len(os.Args) < 2 {
		printBanner()
		printHelp()
		os.Exit(0)
	}

	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h" || os.Args[2] == "help") {
		printBanner()
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "scan", "s":
		if len(os.Args) > 2 && os.Args[2] == "--stacks" {
			scan.ScanStacks(os.Getenv("HOME"))
		} else {
			mode := scan.ModeFast
			if slices.Contains(os.Args, "--deep") {
				mode = scan.ModeDeep
			}
			if slices.Contains(os.Args, "--fast") {
				mode = scan.ModeFast
			}
			scan.Run(mode)
		}
	case "serve", "ui":
		if err := serve.Run(version); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "size":
		scan.DiskSize()
	case "hogs":
		scan.TopHogs()
	case "git-trap":
		scan.CheckGitTrap()
	case "history":
		scan.PrintHistory(os.Getenv("HOME"))
	case "version", "--version", "-v":
		fmt.Printf("plong v%s\n", version)
	case "help", "--help", "-h", "h":
		printBanner()
		printHelp()
	case "clean":
		dryRun := slices.Contains(os.Args, "--dry-run") || slices.Contains(os.Args, "-n")
		clean.Run(dryRun)
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println()
	fmt.Println("  ██████╗ ██╗      ██████╗ ███╗   ██╗ ██████╗ ")
	fmt.Println("  ██╔══██╗██║     ██╔═══██╗████╗  ██║██╔════╝ ")
	fmt.Println("  ██████╔╝██║     ██║   ██║██╔██╗ ██║██║  ███╗")
	fmt.Println("  ██╔═══╝ ██║     ██║   ██║██║╚██╗██║██║   ██║")
	fmt.Println("  ██║     ███████╗╚██████╔╝██║ ╚████║╚██████╔╝")
	fmt.Println("  ╚═╝     ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ")
	fmt.Println("")
	fmt.Println("  relief sound, finally clear")
	fmt.Println("")
}

func printHelp() {
	fmt.Println("commands:")
	fmt.Println("  scan, s       audit your mac — find reclaimable space")
	fmt.Println("                  --stacks       show per-stack breakdown")
	fmt.Println("                  --fast         quick first-pass scan (default)")
	fmt.Println("                  --deep         full scan with discovery and old-file walk")
	fmt.Println("  serve, ui     open the dashboard in your browser")
	fmt.Println("  size          quick disk space check")
	fmt.Println("  hogs          space hogs in ~ grouped by tier")
	fmt.Println("  git-trap      check for accidental .git in home")
	fmt.Println("  history       view past scan snapshots with deltas")
	fmt.Println("  clean         interactive cleanup")
	fmt.Println("                  --dry-run, -n   preview without deleting")
	fmt.Println("  version       show version")
	fmt.Println("  help          show this help")
	fmt.Println("")
	fmt.Println("docs:  https://github.com/srmdn/plong")
}
