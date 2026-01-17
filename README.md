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

### Options

- `-s, --segment` - Number of segments for parallel download (default: 4)
- `-o, --output` - Output filename for the downloaded file

### Usage

#### Basic Download

```bash
adm <URL>
```

#### With optional segment count flag

```bash
adm -s <number> <URL>
adm --segment <number> <URL>
```

#### With optional output filename flag

```bash
adm -o <filename> <URL>
adm --output <filename> <URL>
```

#### With both flags
```bash
adm -s <number> -o <filename> <URL>
adm --segment <number> --output <filename> <URL>
```

### Options

- `-s, --segment` - Number of segments for parallel download (default: 4)

