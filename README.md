## adm

a command-line file download accelerator with support for concurrent segmented downloads.

### Installation

install from source

```bash
git clone https://github.com/arpnghosh/adm.git
cd adm
go build -o adm main.go
```

Or install directly:

```bash
go install github.com/arpnghosh/adm@latest
```

### Usage

#### Basic Download

```bash
adm <URL>
```

#### With Custom Segments

```bash
adm -s <number> <URL>
adm --segment <number> <URL>
```

### Examples

Download a file with default 4 segments:

```bash
adm https://example.com/file.zip
```

Download a file using 8 parallel segments:

```bash
adm -s 8 https://example.com/file.zip
adm --segment 8 https://example.com/file.zip
```

### Options

- `-s, --segment` - Number of segments for parallel download (default: 4)

