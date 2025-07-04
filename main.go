package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"

	"simple/cmd"
	"simple/config"
)

var CLI struct {
	Config string `help:"Config file path" type:"path" default:"${config_file}"`

	// Commands
	TUI       cmd.TUICmd       `cmd:"" help:"Launch the terminal UI (default)" default:"1"`
	Configure cmd.ConfigureCmd `cmd:"" help:"Create default configuration file"`

	// API Commands
	Threads cmd.ThreadsCmd `cmd:"" help:"Manage threads"`
	Report  cmd.ReportCmd  `cmd:"" help:"Generate a report of threads"`
}

func main() {
	// Check if this is the configure command before doing anything else
	if len(os.Args) > 1 && os.Args[1] == "configure" {
		// Handle configure command directly without config validation
		force := false
		for _, arg := range os.Args[2:] {
			if arg == "-f" || arg == "--force" {
				force = true
				break
			}
			if arg == "-h" || arg == "--help" {
				fmt.Println("Usage: simple configure [flags]")
				fmt.Println()
				fmt.Println("Create default configuration file")
				fmt.Println()
				fmt.Println("Flags:")
				fmt.Println("  -h, --help     Show help")
				fmt.Println("  -f, --force    Overwrite existing configuration file")
				return
			}
		}

		configureCmd := &cmd.ConfigureCmd{Force: force}
		err := configureCmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Get config file path
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		os.Exit(1)
	}

	// Parse command line with config file support
	ctx := kong.Parse(&CLI,
		kong.Name("simple"),
		kong.Description("A CLI tool for Plain API with terminal UI"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Configuration(kongyaml.Loader, configPath),
		kong.Vars{
			"config_file": configPath,
		},
	)

	// Load configuration
	cfg, err := config.Load(CLI.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Run the selected command
	err = ctx.Run(cfg)
	ctx.FatalIfErrorf(err)
}
