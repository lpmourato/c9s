# Project Structure

This document describes the refactored project structure for c9s.

## Directory Layout

```
c9s/
├── cmd/c9s/                     # Application entrypoint
│   └── main.go                  # Main function (moved from root)
├── internal/                    # Private application code
│   ├── app/                     # Application orchestration
│   │   └── app.go              # Main app logic and coordination
│   ├── cli/                     # CLI parsing and commands
│   │   └── cli.go              # Kong CLI definitions and parsing
│   ├── config/                  # Configuration management
│   │   ├── cloud_run.go        # Cloud Run specific config
│   │   └── gcp_config.go       # GCP configuration
│   ├── datasource/              # Data access layer
│   │   ├── cloudrun.go         # Cloud Run datasource implementation
│   │   ├── datasource.go       # Interface and factory
│   │   └── mock.go             # Mock datasource for testing
│   ├── domain/                  # Business logic and entities
│   │   ├── cloudrun/           # Cloud Run domain objects
│   │   │   └── cloudrun_service.go
│   │   └── monitoring/         # Monitoring domain logic (future)
│   ├── infrastructure/          # External service integrations
│   │   ├── gcp/                # Google Cloud Platform integration
│   │   │   ├── provider.go     # GCP service provider
│   │   │   └── service_details.go
│   │   └── filesystem/         # File system operations (future)
│   ├── ui/                     # User interface layer
│   │   ├── ui.go              # UI package wrapper
│   │   └── tui/               # Terminal UI components
│   │       ├── app.go         # TUI application
│   │       ├── command_*.go   # Command handling
│   │       ├── styled_table.go # Table styling
│   │       └── table.go       # Table component
│   └── views/                  # UI views and screens
│       ├── cloud_run.go       # Main Cloud Run services view
│       ├── deployment_view.go # Deployment details view
│       └── log_view.go        # Log streaming view
├── tools/                      # Development and debugging tools
│   └── debug/                  # Debug utilities (future)
├── scripts/                    # Build and deployment scripts
│   └── build.sh               # Build script
├── docs/                       # Documentation
├── bin/                        # Build output (created by make)
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── README.md                   # Project documentation
└── LICENSE                     # License file
```

## Key Changes

### 1. Standard Go Project Layout
- Moved `main.go` to `cmd/c9s/` following Go conventions
- Created proper separation between internal packages

### 2. Layered Architecture
- **CLI Layer** (`internal/cli/`): Command-line interface handling
- **Application Layer** (`internal/app/`): Application orchestration
- **Domain Layer** (`internal/domain/`): Business logic
- **Infrastructure Layer** (`internal/infrastructure/`): External integrations
- **UI Layer** (`internal/ui/`): User interface components

### 3. Build System
- Added `Makefile` with common development tasks
- Created build scripts in `scripts/` directory
- Build output goes to `bin/` directory

### 4. Import Path Updates
- Updated all import paths to reflect new structure
- UI components moved from `internal/ui` to `internal/ui/tui`
- GCP integration moved to `internal/infrastructure/gcp`

## Development Workflow

### Building
```bash
make build           # Build the application
make run            # Build and run with GCP
make run-mock       # Build and run with mock data
make test           # Run tests
make clean          # Clean build artifacts
```

### Adding New Features
1. **Domain logic**: Add to `internal/domain/`
2. **External integrations**: Add to `internal/infrastructure/`
3. **UI components**: Add to `internal/ui/tui/`
4. **Views**: Add to `internal/views/`

## Benefits

1. **Maintainability**: Clear separation of concerns
2. **Testability**: Isolated layers for easier testing
3. **Scalability**: Easy to add new features and components
4. **Standards Compliance**: Follows Go project layout conventions
5. **Developer Experience**: Makefile simplifies common tasks

## Migration Notes

- Old import paths have been updated throughout the codebase
- The application builds and runs successfully with the new structure
- All existing functionality is preserved
- Build output is now in `bin/` directory instead of root
