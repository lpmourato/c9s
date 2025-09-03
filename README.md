# c9s

A terminal UI to navigate and view logs from Google Cloud Run services, inspired by [k9s](https://github.com/derailed/k9s).

## About

This project is a fork and adaptation of [k9s](https://github.com/derailed/k9s), originally designed for Kubernetes clusters. `c9s` reuses the UI and navigation concepts from k9s, but is focused on Google Cloud Run:
- List Cloud Run services in a table view
- View service details and traffic splits
- Stream logs for selected services

## Reference
- Original k9s repository: https://github.com/derailed/k9s
- k9s authors: https://github.com/derailed/k9s/graphs/contributors

## Features
- Fast, keyboard-driven navigation
- Live updates of Cloud Run services
- View logs with `Ctrl+L`
- Simple configuration via flags or environment variables

## Usage

## Build

### Using Make (recommended)
```bash
make build
```

### Manual build
```bash
# build first (optional)
go build -o bin/c9s ./cmd/c9s
```

### Build for Apple Silicon (arm64)
```bash
GOARCH=arm64 GOOS=darwin go build -o bin/c9s ./cmd/c9s
```

## Usage

### Quick examples

- Run against your Cloud Run services (GCP):
```bash
# Using make
make run

# Or manually
./bin/c9s gcp --project=my-project --region=us-central1
```

- Run in test mode (uses the bundled mock datasource):
```bash
# Using make
make run-mock

# Or manually - subcommand (preferred):
./bin/c9s mock

# or equivalent flag:
./bin/c9s gcp --datasource=mock
```

## License

This project inherits the [Apache 2.0 License](https://github.com/derailed/k9s/blob/master/LICENSE) from k9s.

## Attribution

This project is based on k9s by Fernand Galiana and contributors. See the original repo for more details and credits.
