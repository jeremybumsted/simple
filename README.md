# Simple - Plain in your terminal

A TUI and command-line tool written in Go that wraps [Plain](https://plain.com)'s GraphQL API.

## Prerequisites

- Go `1.21` or later (tested on `1.24.2`)
- A Plain API key (get one from your Plain workspace settings)

### Build from Source

```bash
git clone <repository-url>
cd simple
go mod tidy
go build -o simple .
```

### Quick Setup

After building or installing, run the configure command to create a default configuration file:

```bash
simple configure
```

This will create a configuration file at `~/.simple/config.yaml` with default settings.

## Configuration

The CLI looks for configuration in `~/.simple/config.yaml`. You can create a default configuration file using:

```bash
simple configure
```

This creates a configuration file at `~/.simple/config.yaml` with placeholder values that you can customize.

### Configuration File Example

```yaml
# Plain API configuration
plain:
  workspace_id: "plain-workspace-id"
  # Your Plain API key (can also be set via PLAIN_API_KEY environment variable)
  api_key: "your-api-key-here"

  # Plain API endpoint (usually no need to change this)
  endpoint: "https://core-api.uk.plain.com/graphql/v1"

# UI configuration
ui:
  # Theme for the terminal UI
  theme: "default"

  # Number of items to show per page
  page_size: 20

  # Show debug information
  show_debug: false
```

### Environment Variables

You can also set configuration via environment variables:

- `PLAIN_API_KEY`: Your Plain API key (required)

### Initial Setup

1. **Create configuration file**:
   ```bash
   simple configure
   ```

2. **Get your API key**:
   - Log into your Plain workspace
   - Go to Settings → API Keys
   - Create a new API key with appropriate permissions:
     - `customer:read` - for customer operations
     - `thread:read` - for thread operations
     - `label:read` and `label:create` - for label operations

3. **Configure the API key** (choose one method):
   - Set environment variable: `export PLAIN_API_KEY="your-api-key"`
   - Edit `~/.simple/config.yaml` and update the `api_key` field

## Usage

### Terminal UI (Default)

Launch the interactive terminal UI:

```bash
simple
# or explicitly
simple tui
```

#### Navigation

- **Arrow keys** or **j/k**: Navigate up/down in thread list
- **Enter**: View thread details
- **q** or **Esc**: Go back/quit (from detail view or quit application)
- **r**: Refresh current view
- **n**: Next page (when available)
- **1**: Filter to TODO threads only
- **2**: Filter to SNOOZED threads only
- **3**: Show all threads
- **d**: Open dashboard view
- **t**: Switch to threads view (from dashboard)
- **b**: Open selected thread in browser (from detail view)
- **/**: Search/filter threads (built-in list filtering)

### Command Line Interface

You can also use the CLI without the TUI for scripting and automation:

#### CLI Commands (Alternative)

For scripting and automation, you can use CLI commands:

##### Threads

```bash
# List threads
simple threads list

# List with pagination
simple threads list --limit 30 --cursor "cursor-string"

# Get thread by ID
simple threads get th_1234567890
```

### Global Options

```bash
# Create or recreate configuration file
simple configure
simple configure --force  # Overwrite existing config

# Use custom config file
simple --config /path/to/config.yaml <command>

# Show help
simple --help
simple <command> --help
```

## API Reference

This CLI wraps the Plain GraphQL API. Key operations include:


### Threads
- List threads with pagination
- Get thread details with messages
- Filter by status and priority

## Development

### Built with

- **[Kong](https://github.com/alecthomas/kong)**: Command-line parser
- **[Kong-YAML](https://github.com/alecthomas/kong-yaml)**: YAML configuration support
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: Terminal UI framework
- **[Bubbles](https://github.com/charmbracelet/bubbles)**: TUI components (list, spinner, etc.)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Style definitions for TUI
- **[GraphQL](https://github.com/machinebox/graphql)**: GraphQL client

### Building

```bash
# Build for current platform
go build -o simple .

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o simple-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o simple-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o simple-windows-amd64.exe .
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- 🐛 **Issues**: Report bugs via GitHub Issues
- 💬 **Discussions**: Ask questions in GitHub Discussions
- 📖 **Docs**: Check the Plain API documentation for API details
