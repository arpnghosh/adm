## adm

adm is a CLI utility for downloading files from the internet. Currently, it only supports HTTP and HTTPS.

## Prerequisites

- [Go](https://go.dev/dl/) 1.23 or later

## Installation

```bash
go install github.com/arpnghosh/adm@latest
```

Or build from source:

```bash
git clone https://github.com/arpnghosh/adm.git
cd adm
go build -o adm .
./adm  # Run the binary
sudo mv adm /usr/local/bin/  # Optional: install system-wide
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-s, --segment` | Parallel segment count (1-16) | 4 |
| `-o, --output` | Output filename (extension auto-detected) | From URL |


## Usage

```bash
adm <URL>                      # Default: 4 segments
adm -s 8 <URL>                 # 8 parallel segments
adm -o filename <URL>          # Custom output name
adm -s 8 -o filename <URL>     # Both options
```

