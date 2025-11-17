# Nexus Retention Policy Tool 🚀

A production-ready Go application to manage Docker image retention policies in Sonatype Nexus Repository Manager 3. Since Nexus doesn't natively support retention policies for container image tags, this tool provides automated cleanup based on configurable rules.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why This Tool?

Nexus Repository Manager 3 lacks built-in retention policies for Docker images. This leads to:
- 💾 Wasted storage space
- 🐌 Slower repository operations
- 💰 Increased infrastructure costs
- 🔍 Difficulty finding relevant images

This tool solves these problems by providing flexible, automated cleanup based on your specific needs.

## Features

- 🐳 **Docker Repository Support**: Automatically discovers and processes all Docker hosted repositories
- 📋 **Flexible Rules**: Define retention policies using regex patterns
- 🔒 **Protected Tags**: Prevent deletion of specific tags (e.g., `latest`, `stable`)
- 🔍 **Dry Run Mode**: Preview deletions before executing
- 📊 **CSV Logging**: Human-readable deletion logs with timestamps
- ⏰ **Scheduling**: Run once or as a recurring job using cron syntax
- 🔄 **Pagination Handling**: Automatically handles Nexus API pagination

## Quick Start

```bash
# 1. Download or build the binary
go build -o nexus-retention-policy ./cmd

# 2. Create configuration
cp config.example.yaml config.yaml
nano config.yaml

# 3. Test with dry run (no deletions, default mode)
./nexus-retention-policy --config config.yaml

# 4. Test with verbose output (shows all images including unmatched)
./nexus-retention-policy --config config.yaml --verbose

# 5. Run for real (execute deletions)
./nexus-retention-policy --config config.yaml --exec
```


## Configuration

Create a `config.yaml` file:

```yaml
nexus:
  url: "https://nexus.example.com"
  username: "admin"
  password: "secret"
  timeout: 30 # seconds

rules:
  - name: "production images"
    regex: "^prod-.*"
    keep: 10
  - name: "development images"
    regex: "^dev-.*"
    keep: 5
  - name: "feature branches"
    regex: "^feature-.*"
    keep: 3

protected_tags:
  - "latest"
  - "stable"
  - "main"

# Scheduling configuration
# Leave empty or set to "" for one-time execution
# Use cron format for recurring jobs (e.g., "0 2 * * *" for daily at 2 AM)
schedule: ""

# CSV log file path
log_file: "deletion_log.csv"
```

### Configuration Options

#### Nexus Settings
- `url`: Base URL of your Nexus instance
- `username`: Nexus username with delete permissions
- `password`: Nexus password
- `timeout`: HTTP request timeout in seconds

#### Retention Rules
Rules are evaluated in order. The first matching rule determines the retention count.

- `name`: Descriptive name for the rule
- `regex`: Regular expression to match image names
- `keep`: Number of most recent tags to keep

**Important:** Only images matching at least one rule will be processed. Images that don't match any rule are skipped entirely. To process all images, add a catch-all rule at the end:

```yaml
rules:
  - name: "specific images"
    regex: "^prod-.*"
    keep: 10
  - name: "all other images"
    regex: ".*"
    keep: 5
```

#### Other Settings
- `protected_tags`: List of tags that should never be deleted
- `schedule`: Cron expression for scheduled execution (empty = one-time)
- `log_file`: Path to CSV log file

### Cron Schedule Examples

```yaml
# Every day at 2 AM
schedule: "0 2 * * *"

# Every Sunday at midnight
schedule: "0 0 * * 0"

# Every 6 hours
schedule: "0 */6 * * *"

# Every weekday at 3 AM
schedule: "0 3 * * 1-5"
```

## Usage

### Command Line Flags

- `--config`: Path to configuration file (default: `config.yaml`)
- `--exec`: Execute deletions (default is dry-run mode)
- `--verbose`: Show all images including unmatched ones

### One-time Execution

```bash
# Dry run (default, no deletions)
./nexus-retention-policy --config config.yaml

# Dry run with verbose output
./nexus-retention-policy --config config.yaml --verbose

# Execute deletions
./nexus-retention-policy --config config.yaml --exec

# Execute with verbose output
./nexus-retention-policy --config config.yaml --exec --verbose
```

### Scheduled Execution

Set the `schedule` field in `config.yaml` and run:

```bash
# Scheduled dry run
./nexus-retention-policy --config config.yaml

# Scheduled execution
./nexus-retention-policy --config config.yaml --exec
```

The tool will run continuously and execute at the specified intervals. Press `Ctrl+C` to stop.

### Output Modes

**Normal mode (default):**
- Shows matched images and their retention rules
- Shows tags being kept and deleted
- Hides unmatched images

**Verbose mode (`--verbose`):**
- Shows all images including unmatched ones
- Useful for debugging rule patterns

## How It Works

1. **Discovery**: Fetches all Docker hosted repositories from Nexus
2. **Component Retrieval**: Gets all components (images) from each repository with pagination
3. **Grouping**: Groups components by image name
4. **Rule Matching**: Applies retention rules based on regex patterns
5. **Sorting**: Sorts tags by last modified date (most recent first)
6. **Protection**: Excludes protected tags from deletion
7. **Cleanup**: Deletes components exceeding the retention count
8. **Logging**: Records all deletions to CSV file

## Deletion Log Format

The CSV log includes:

| Column | Description |
|--------|-------------|
| Timestamp | When the deletion occurred (RFC3339 format) |
| Repository | Nexus repository name |
| Image Name | Docker image name |
| Tag | Image tag/version |
| Component ID | Nexus component ID |
| Rule | Which rule triggered the deletion |
| Dry Run | Whether this was a dry run |

Example:
```csv
Timestamp,Repository,Image Name,Tag,Component ID,Rule,Dry Run
2024-01-15T10:30:00Z,docker-hosted,myapp,v1.0.0,abc123,production images,false
```

## Best Practices

1. **Start with Dry Run**: Always test without `--exec` flag first
2. **Use Verbose Mode**: Run with `--verbose` to see all images and verify rules
3. **Protect Important Tags**: Add critical tags to `protected_tags`
4. **Conservative Retention**: Start with higher `keep` values
5. **Monitor Logs**: Review `deletion_log.csv` regularly
6. **Backup**: Ensure you have backups before running with `--exec`
7. **Test Regex**: Verify your regex patterns match expected images

## Troubleshooting

### Authentication Errors
- Verify username and password in config
- Ensure user has delete permissions in Nexus

### No Components Found
- Check that repositories are type "hosted" and format "docker"
- Verify repository names in Nexus UI

### Regex Not Matching
- Test regex patterns at https://regex101.com/
- Remember: patterns match against image names, not tags

## Security Considerations

- Store `config.yaml` securely (contains credentials)
- Use environment variables for sensitive data in production
- Limit Nexus user permissions to only what's needed
- Review deletion logs for unexpected behavior

## License

MIT License - See LICENSE file for details

## Project Structure

```
nexus-retention-policy/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── logger/
│   │   └── logger.go        # CSV logging
│   ├── nexus/
│   │   └── client.go        # Nexus API client
│   └── retention/
│       └── policy.go        # Retention policy engine
├── config.yaml              # Configuration file
├── Dockerfile               # Docker image
├── docker-compose.yml       # Docker Compose setup
├── Makefile                 # Build automation
├── README.md                # This file
├── QUICKSTART.md            # Quick start guide
├── INSTALL.md               # Installation guide
├── EXAMPLES.md              # Configuration examples
└── LICENSE                  # MIT License
```

## Contributing

Contributions are welcome! Here's how you can help:

1. 🐛 Report bugs by opening an issue
2. 💡 Suggest features or improvements
3. 🔧 Submit pull requests
4. 📖 Improve documentation
5. ⭐ Star the project if you find it useful

### Development

```bash
# Clone the repository
git clone <repository-url>
cd nexus-retention-policy

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
make build

# Format code
make fmt

# Run linter
make lint
```

## Roadmap

- [ ] Support for additional repository formats (Maven, npm, etc.)
- [ ] Web UI for configuration and monitoring
- [ ] Prometheus metrics export
- [ ] Slack/Email notifications
- [ ] Advanced filtering (by size, download count, etc.)
- [ ] Backup before deletion option
- [ ] Multi-Nexus instance support

## FAQ

**Q: Will this delete images that are currently in use?**  
A: The tool only deletes based on age and count. Use `protected_tags` to prevent deletion of critical images.

**Q: Can I undo deletions?**  
A: No, deletions are permanent. Always test without `--exec` flag first and review the logs.

**Q: Does this work with Nexus 2?**  
A: No, this tool is designed for Nexus Repository Manager 3 only.

**Q: How do I handle multiple Nexus instances?**  
A: Run separate instances of the tool with different config files.

**Q: What happens if the tool crashes during deletion?**  
A: The tool processes components one at a time. Partial deletions are logged in the CSV file.

## Support

- 📧 Open an issue for bug reports or feature requests
- 💬 Discussions for questions and community support
- 📖 Check the documentation files for detailed information

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [robfig/cron](https://github.com/robfig/cron) for scheduling
- Integrates with [Sonatype Nexus Repository Manager](https://www.sonatype.com/products/repository-oss)

---

Made with ❤️ for the DevOps community
