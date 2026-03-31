# deadhead

Search for cheap flights by scraping Skiplagged. Supports one-way, round-trip, worldwide, and batch multi-leg queries. Automatically detects hidden-city tickets.

**Hidden-city ticketing:** Sometimes booking a flight to city C with a layover in city B is cheaper than flying directly to B. You book the full trip, get off at B, and skip the rest. This tool finds and flags those options.

**Why "deadhead"?** In aviation, a deadhead is a crew member riding a flight as a passenger without working. Same idea.

---

## Forked from

Inspired by and forked from [go-skiplagged](https://github.com/minormending/go-skiplagged) by [@minormending](https://github.com/minormending).

The original repo is outdated and no longer works. Skiplagged added Cloudflare bot detection that blocks the plain HTTP requests it used. This fork fixes that by using a real Chrome session to clear the challenge before making requests.

Other changes:
- **Batch mode** for running multi-leg itineraries from a single JSON file, useful for agentic/automated workflows
- **Hidden-city detection** flagged in output
- **One-way pricing fix:** the original passed a return date even for one-way queries, causing the endpoint to return round-trip prices
- **Improved CLI:** flags work anywhere in the command
- **Comprehensive test suite:** 65 unit tests with full mock coverage across all packages

---

## Install

### 1. Install Go

Requires **Go 1.26+**.

| Platform | Instructions |
|----------|-------------|
| macOS | `brew install go` or download from [go.dev/dl](https://go.dev/dl) |
| Windows | Download the `.msi` installer from [go.dev/dl](https://go.dev/dl) and run it |
| Linux | Download from [go.dev/dl](https://go.dev/dl) or use your package manager (`sudo apt install golang-go`, `sudo dnf install golang`, etc.) |

Verify: `go version`

### 2. Install Google Chrome

deadhead launches a real Chrome window to pass Skiplagged's Cloudflare bot detection. Chrome must be installed - it does not need to be your default browser.

**macOS**

```bash
brew install --cask google-chrome
```
Or download from [google.com/chrome](https://www.google.com/chrome). Chrome is found automatically via its standard app path.

**Windows**

Download and install from [google.com/chrome](https://www.google.com/chrome). Chrome is found automatically via its standard install path.

**Linux**

```bash
# Debian / Ubuntu
sudo apt update && sudo apt install -y google-chrome-stable

# If google-chrome-stable is not in your repos, add the Google repo first:
wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo apt install ./google-chrome-stable_current_amd64.deb

# Fedora / RHEL
sudo dnf install google-chrome-stable

# Arch
yay -S google-chrome

# Chromium also works (lighter alternative)
sudo apt install chromium-browser     # Debian/Ubuntu
sudo dnf install chromium             # Fedora
```

### 3. Install deadhead

**Option A - go install (recommended, no clone needed)**

```bash
go install github.com/mattneto928/deadhead/cmd/deadhead@latest
```

This places `deadhead` (or `deadhead.exe` on Windows) in `$HOME/go/bin`. That directory must be in your `PATH` or the command won't be found after install.

Add it if needed, then restart your terminal:

```bash
# macOS / Linux - add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$HOME/go/bin"

# Windows (PowerShell, permanent)
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\go\bin", "User")
```

Verify:
```bash
which deadhead      # macOS / Linux
where deadhead      # Windows
```

**Option B - build from source**

```bash
git clone https://github.com/MattNeto928/deadhead
cd deadhead

# macOS / Linux
make install          # builds and copies to /usr/local/bin/deadhead

# macOS / Linux (without make)
go build -o deadhead ./cmd/deadhead/
sudo mv deadhead /usr/local/bin/

# Windows (PowerShell)
go build -o deadhead.exe .\cmd\deadhead\
# Then move deadhead.exe to a folder already in your PATH, e.g.:
Move-Item deadhead.exe "$env:USERPROFILE\go\bin\deadhead.exe"
```

---

## Usage

```
deadhead [OPTIONS] FROM [TO] DEPART [RETURN]

  FROM    origin airport or city code (e.g. NYC, JFK)
  TO      destination code, optional (omit for worldwide search)
  DEPART  departure date (yyyy-MM-dd)
  RETURN  return date, optional (omit for one-way)

Options:
  -batch string         JSON file with array of queries
  -exclude string       comma-separated airports to exclude
  -leave-after  int     outbound departs after this hour (0-23)
  -leave-before int     outbound departs before this hour (0-23)
  -return-after  int    return departs after this hour (0-23)
  -return-before int    return departs before this hour (0-23)
  -max-price int        max total price in USD
  -travelers int        number of travelers (default 1)
  -out-json string      save results to a JSON file
  -out-md   string      save results to a Markdown file
  -overwrite            overwrite existing output file
  -skip-worldwide       skip the worldwide fare scan
  -proxy string         HTTP proxy URL
  -help                 show help
```

---

## Examples

```bash
# One-way
deadhead NYC LGW 2026-05-12

# Round-trip with filters
deadhead NYC AUS 2026-03-04 2026-03-08 --max-price 200 --leave-before 13

# Worldwide cheapest destinations from NYC
deadhead NYC 2026-05-12

# Save to file
deadhead NYC LON 2026-05-12 --out-json results.json
```

**Output:**
```
London, England (LGW)  $279 one-way
  JFK -> LGW  $279  Norse Atlantic UK       6:20 PM -> 6:20 AM
  min outbound: $279
```

---

## Batch Mode

Run multiple queries at once using a shared browser session. Useful for checking every leg of a multi-city trip in one go.

**Batch file (JSON array):**
```json
[
  { "from": "NYC", "to": "LGW", "depart": "2026-05-12" },
  { "from": "LGW", "to": "ZRH", "depart": "2026-05-16" },
  { "from": "ZRH", "to": "PRG", "depart": "2026-05-19" },
  { "from": "PRG", "to": "IST", "depart": "2026-05-22" }
]
```

**Run it:**
```bash
deadhead -batch my_legs.json
```

All CLI options are available per-query in the JSON object:

| Field | Type | Description |
|-------|------|-------------|
| `from` | string | Required. Origin code |
| `to` | string | Destination code (omit for worldwide) |
| `depart` | string | Required. Date as `yyyy-MM-dd` |
| `return` | string | Return date (omit for one-way) |
| `travelers` | int | Default 1 |
| `max_price` | int | Price ceiling |
| `leave_after` | int | Outbound departs after this hour |
| `leave_before` | int | Outbound departs before this hour |
| `return_after` | int | Return departs after this hour |
| `return_before` | int | Return departs before this hour |
| `exclude` | string | Comma-separated airports to exclude |
| `skip_worldwide` | bool | Skip worldwide scan for this query |

---

## Hidden-City Detection

Multi-segment flights where the final airport does not match your destination are flagged as potential hidden-city tickets:

```
IST -> BKK  $630  Turkish Airlines    1:55 AM -> 3:25 PM  [HIDDEN-CITY to CNX]
PVG -> KIX  $312  Japan Airlines      1:15 PM -> 4:35 PM  [HIDDEN-CITY to HND]
```

**Warning:** Hidden-city ticketing violates most airline terms of service. Do not check bags. Any return segments on the same booking will be cancelled if you skip the final leg.

---

## Testing

### Unit tests

All packages have full unit test coverage. No browser or network access required - HTTP calls are mocked with `httptest`.

```bash
go test ./...
```

### Integration tests

Integration tests run against the real Skiplagged site and require Chrome and a network connection. They always use a departure date one month from today so they never query the past.

```bash
go test -v -tags integration ./integration/
```

---

## Notes

- On startup, Chrome opens a visible window to load `skiplagged.com` and pass the Cloudflare challenge. Headless mode trips the bot detection, so it runs with a real window. Once the challenge clears, the session cookies are extracted and all subsequent requests are plain HTTP.
- There is no public Skiplagged API. This tool hits their internal search endpoint (`skiplagged.com/api/search.php`) that their own frontend uses. It returns JSON but is undocumented and subject to change.
- Always omits the return date parameter. Passing one causes the endpoint to return round-trip pricing even for one-way queries.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT
