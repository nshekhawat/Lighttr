# Lighttr

Lighttr is a lightweight HTTP request debugger with an interactive TUI interface. It provides a more user-friendly alternative to raw curl commands, allowing you to build and execute HTTP requests step-by-step.

## Features

- Interactive TUI for building HTTP requests
- Support for all common HTTP methods (GET, POST, PUT, DELETE, etc.)
- Custom headers and query parameters
- Request body support (JSON, form data, raw text)
- Request preview before sending
- Response viewing with formatted output
- Command-line interface for quick requests

## Installation

To install Lighttr, you need Go 1.16 or later. Then run:

```bash
go install github.com/nshekhawat/lighttr/cmd/lighttr@latest
```

Or clone the repository and build from source:

```bash
git clone https://github.com/nshekhawat/lighttr.git
cd lighttr
go build -o lighttr ./cmd/lighttr
```

## Usage

### Interactive Mode

Simply run `lighttr` without any arguments to launch the interactive TUI:

```bash
lighttr
```

In the TUI:
1. Navigate between fields using Tab/Shift+Tab or Up/Down arrows
2. Fill in the request details:
   - URL (e.g., https://api.example.com/path)
   - Method (GET, POST, PUT, DELETE, etc.)
   - Headers (format: key:value,key2:value2)
   - Query Parameters (format: key=value&key2=value2)
   - Request Body (JSON, form data, or raw text)
3. Press Enter to preview the request
4. Press Enter again to send the request
5. View the response details
6. Press ESC to go back or Ctrl+C to quit

### Command-line Mode

You can also use Lighttr directly from the command line:

```bash
lighttr --method POST \
        --url "https://api.example.com/data" \
        --headers "Content-Type:application/json,Authorization:Bearer token" \
        --body '{"key": "value"}'
```

Available flags:
- `--method`: HTTP method (default: GET)
- `--url`: Target URL (required in command-line mode)
- `--headers`: Request headers in key:value,key2:value2 format
- `--body`: Request body