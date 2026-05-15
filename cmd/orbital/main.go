package main

import (
	"fmt"
	"os"

	"github.com/srmdn/orbital/internal/clean"
	"github.com/srmdn/orbital/internal/scan"
	"github.com/srmdn/orbital/internal/serve"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printBanner()
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "scan", "s":
		if len(os.Args) > 2 && (os.Args[2] == "--stacks" || os.Args[2] == "-s") {
			scan.ScanStacks(os.Getenv("HOME"))
		} else {
			scan.Run()
		}
	case "serve", "ui":
		if err := serve.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "size":
		scan.DiskSize()
	case "hogs":
		scan.TopHogs()
	case "git-trap":
		scan.CheckGitTrap()
	case "version", "--version", "-v":
		fmt.Printf("orbital v%s\n", version)
	case "help", "--help", "-h", "h":
		printHelp()
	case "clean":
		clean.Run()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println("  ██████╗ ██████╗ ██████╗ ██╗████████╗ █████╗ ██╗     ")
	fmt.Println("  ██╔═══██╗██╔══██╗██╔══██╗██║╚══██╔══╝██╔══██╗██║     ")
	fmt.Println("  ██║   ██║██████╔╝██████╔╝██║   ██║   ███████║██║     ")
	fmt.Println("  ██║   ██║██╔══██╗██╔══██╗██║   ██║   ██╔══██║██║     ")
	fmt.Println("  ╚██████╔╝██║  ██║██████╔╝██║   ██║   ██║  ██║███████╗")
	fmt.Println("   ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝")
	fmt.Println("")
	fmt.Println("  the friendly mac disk doctor")
	fmt.Println("")
}

func printHelp() {
	fmt.Println("commands:")
	fmt.Println("  scan, s       audit your mac — find reclaimable space")
	fmt.Println("                  --stacks, -s    show per-stack breakdown")
	fmt.Println("  serve, ui     open the dashboard in your browser")
	fmt.Println("  size          quick disk space check")
	fmt.Println("  hogs          space hogs in ~ grouped by tier")
	fmt.Println("  git-trap      check for accidental .git in home")
	fmt.Println("  clean         interactive cleanup")
	fmt.Println("  version       show version")
	fmt.Println("  help          show this help")
	fmt.Println("")
	fmt.Println("docs:  https://github.com/srmdn/orbital")
}
