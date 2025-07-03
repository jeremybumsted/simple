package cmd

import (
	"fmt"
	"os"

	"simple/config"
)

// ConfigureCmd represents the configure command
type ConfigureCmd struct {
	Force bool `help:"Overwrite existing configuration file" short:"f"`
}

// Run executes the configure command
func (c *ConfigureCmd) Run() error {
	// Get the config file path
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil && !c.Force {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Println("Use --force to overwrite the existing configuration file.")
		return nil
	}

	// Create the default configuration file
	if err := config.CreateDefaultConfig(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("Configuration file created at: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file and set your Plain API key")
	fmt.Println("2. Optionally set your workspace ID")
	fmt.Println("3. Set the PLAIN_API_KEY environment variable, or")
	fmt.Println("4. Update the api_key field in the configuration file")
	fmt.Println("\nExample:")
	fmt.Printf("  export PLAIN_API_KEY=your-api-key-here\n")
	fmt.Printf("  %s\n", configPath)

	return nil
}
