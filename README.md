# Simple - Plain API CLI

A command-line tool that wraps the Plain API with a beautiful terminal UI, built using Go and the Charm Bracelet terminal UI libraries.

## Features

- 🚀 **Terminal UI**: Beautiful interactive terminal interface using Bubble Tea and Bubbles
- 🎯 **Thread Management**: Browse and manage Plain threads with filtering (TODO, SNOOZED, All)
- 📊 **Dashboard**: Interactive dashboard with modular components showing thread metrics
- 📈 **Thread Analytics**: View counts for TODO threads, snoozed threads, today's threads, and unassigned threads
- 🎨 **Rich Display**: View thread details, customer info, and company information
- ⚡ **GraphQL Client**: Efficient API communication using machinebox/graphql
- 🎭 **Styled Output**: Rich text styling with Lipgloss and color-coded components
- ⚙️ **Configuration**: YAML-based configuration with environment variable support
- 🔍 **Filtering**: Filter threads by status with real-time updates

## Installation

### Prerequisites

- Go 1.21 or later
- A Plain API key (get one from your Plain workspace settings)

### Build from Source

```bash
git clone <repository-url>
cd simple
go mod tidy
go build -o simple .
```

### Install Binary

```bash
# Install directly with go install
go install github.com/yourusername/simple@latest
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

##### Customers

```bash
# List customers
simple customers list

# List with pagination
simple customers list --limit 50 --cursor "cursor-string"

# Get customer by email
simple customers get customer@example.com

# Search customers
simple customers search "John Doe"
```

##### Threads

```bash
# List threads
simple threads list

# List with pagination
simple threads list --limit 30 --cursor "cursor-string"

# Get thread by ID
simple threads get th_1234567890
```

##### Labels

```bash
# List all labels
simple labels list

# Create a new label
simple labels create "Bug Report" --color "#FF0000"
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

### Customers
- List customers with pagination
- Get customer by email
- Search customers by name

### Threads
- List threads with pagination
- Get thread details with messages
- Filter by status and priority

### Labels
- List all labels
- Create new labels with colors

## Development

### Project Structure

```
simple/
├── main.go           # Entry point and CLI setup
├── cmd/              # Command definitions
│   ├── tui.go        # Terminal UI command
│   ├── customers.go  # Customer commands
│   ├── threads.go    # Thread commands
│   └── labels.go     # Label commands
├── client/           # Plain API GraphQL client
│   └── plain.go      # API client implementation
├── config/           # Configuration handling
│   └── config.go     # Config loading and validation
├── types/            # Type definitions
│   └── types.go      # Plain API data types
└── ui/               # Terminal UI components
    ├── app.go        # Main TUI application
    ├── dashboard.go  # Dashboard view with modular components
    ├── customers.go  # Customers view
    ├── threads.go    # Threads view
    └── labels.go     # Labels view
└── go.mod            # Go module definition
```

### Dependencies

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

## Examples

### Basic Workflow

1. **Setup**: Configure your API key
   ```bash
   export PLAIN_API_KEY="your-api-key"
   ```

2. **Launch TUI**: Start the interactive terminal interface
   ```bash
   simple
   ```

3. **View Dashboard**: Press 'd' to see the dashboard with:
   - 📋 TODO Threads count
   - 😴 Snoozed Threads count  
   - 📅 Threads Created Today
   - 👤 Unassigned Threads (estimated)

4. **Browse Threads**: Press 't' to switch to threads view and explore

3. **Explore and Analyze**: Use the TUI to:
   - View the dashboard (press 'd') for thread metrics and analytics
   - Filter threads by status (1: TODO, 2: SNOOZED, 3: All)
   - View detailed thread information
   - See customer and company details
   - Open threads in your browser

4. **Script Access**: Use CLI commands for automation
   ```bash
   # Get customer details for scripting
   simple customers get customer@example.com

   # List recent threads
   simple threads list --limit 10
   ```

### Integration Examples

```bash
# Export customer list to CSV (with jq)
simple customers list --format json | jq -r '.[] | [.id, .fullName, .email] | @csv'

# Check thread status
THREAD_ID="th_1234567890"
STATUS=$(simple threads get $THREAD_ID --format json | jq -r '.status')
echo "Thread $THREAD_ID status: $STATUS"
```

## Troubleshooting

### Common Issues

1. **API Key Issues**
   ```
   Error: Plain API key is required
   ```
   - Set `PLAIN_API_KEY` environment variable
   - Or configure in `~/.simple/config.yaml`

2. **Connection Issues**
   ```
   Error: failed to get customers: connection refused
   ```
   - Check your internet connection
   - Verify the API endpoint in configuration
   - Ensure your API key has proper permissions

3. **Configuration Issues**
   ```
   Error: failed to parse config file
   ```
   - Check YAML syntax in `~/.simple/config.yaml`
   - Use the example configuration as reference

### Debug Mode

Enable debug mode for troubleshooting:

```yaml
ui:
  show_debug: true
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