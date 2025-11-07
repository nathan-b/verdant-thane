package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommand
	subcommand := os.Args[1]

	// Create a new flag set for the subcommand
	var fs *flag.FlagSet

	switch subcommand {
	case "build-sounds":
		fs = flag.NewFlagSet("build-sounds", flag.ExitOnError)
		outputFile := fs.String("output", "laser.wav", "Output WAV file path")
		duration := fs.Float64("duration", 0.2, "Sound duration in seconds")
		fs.Parse(os.Args[2:])
		if err := makeLaserSound(*outputFile, *duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "convert-grayscale":
		fs = flag.NewFlagSet("convert-grayscale", flag.ExitOnError)
		fs.Parse(os.Args[2:])
		if fs.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: tools convert-grayscale <input.png> <output.png>\n")
			os.Exit(1)
		}
		if err := convertGrayscale(fs.Arg(0), fs.Arg(1)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "extract-palette":
		fs = flag.NewFlagSet("extract-palette", flag.ExitOnError)
		outputFile := fs.String("output", "palette_strip.png", "Output palette strip file")
		fs.Parse(os.Args[2:])
		if fs.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Usage: tools extract-palette [--output palette.png] <input.png>\n")
			os.Exit(1)
		}
		if err := extractPalette(fs.Arg(0), *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand '%s'\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Verdant Thane Development Tools")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./tools <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build-sounds         Generate laser sound effect WAV file")
	fmt.Println("  convert-grayscale    Convert a PNG image to grayscale")
	fmt.Println("  extract-palette      Extract color palette from an image")
	fmt.Println("  help                 Show this help message")
	fmt.Println()
	fmt.Println("Run 'go run ./tools <command> -h' for command-specific help")
}
