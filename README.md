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

### Build for Apple Silicon (arm64)
```bash
GOARCH=arm64 GOOS=darwin go build -o c9s
```

### Run
```bash
./c9s --project=<your-gcp-project> --region=<your-region>
```
Or set environment variables:
```bash
export GCP_PROJECT=<your-gcp-project>
export GCP_REGION=<your-region>
./c9s
```

## License

This project inherits the [Apache 2.0 License](https://github.com/derailed/k9s/blob/master/LICENSE) from k9s.

## Attribution

This project is based on k9s by Fernand Galiana and contributors. See the original repo for more details and credits.
