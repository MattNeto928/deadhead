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

If `deadhead` is not found after install, `$HOME/go/bin` is not in your `PATH`. Add it permanently:

**macOS / Linux (zsh)**
```bash
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc
```

**macOS / Linux (bash)**
```bash
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc
```

> Running `export PATH=...` directly in the terminal only applies to the current session. Writing it to `~/.zshrc` or `~/.bashrc` makes it permanent.

**Windows (PowerShell, permanent)**
```powershell
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\go\bin", "User")
```
Restart PowerShell after running this.

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

### One-way to a specific city

```bash
deadhead NYC LGW 2026-05-12
```

```
Launching browser...
Browser ready.
Searching NYC -> LGW on 2026-05-12...

London, England (LGW)  $279 one-way
  JFK -> LGW  $279  Norse Atlantic UK       6:20 PM -> 6:20 AM
  EWR -> LGW  $349  Norse Atlantic UK       5:45 PM -> 5:55 AM  (+1 layovers)
  min outbound: $279

Search complete. 1 destination with available flights.
```

### Round-trip with departure time filter

```bash
deadhead NYC AUS 2026-05-12 2026-05-19 --leave-before 13
```

```
Launching browser...
Browser ready.
Searching NYC -> AUS  (2026-05-12 -> 2026-05-19)...

Austin, Texas (AUS)  $198 round trip
  JFK -> AUS  $99   American Airlines       8:00 AM -> 11:42 AM
  JFK -> AUS  $109  Spirit Airlines         10:15 AM -> 2:01 PM
  min outbound: $99

  AUS -> JFK  $99   American Airlines       6:30 AM -> 2:45 PM
  AUS -> JFK  $149  Delta Air Lines         8:00 AM -> 4:30 PM
  min return:   $99

Search complete. 1 destination with available flights.
```

### Round-trip with price cap

```bash
deadhead NYC AUS 2026-05-12 2026-05-19 --max-price 200
```

### Worldwide cheapest destinations (one-way)

```bash
deadhead NYC 2026-05-12
```

```
Launching browser...
Browser ready.
Searching worldwide from NYC on 2026-05-12...
Found 42 candidate destinations. Fetching flight details...
[1/42] MCO -- 6 flight(s) from $88
[2/42] CHS -- 3 flight(s) from $91
[3/42] MIA -- 8 flight(s) from $104
[4/42] BOS -- no qualifying flights
...

Orlando, Florida (MCO)  $88 one-way
  JFK -> MCO  $88   Spirit Airlines         6:15 AM -> 9:02 AM
  JFK -> MCO  $103  Frontier Airlines       7:40 AM -> 10:20 AM
  min outbound: $88

Charleston, South Carolina (CHS)  $91 one-way
  EWR -> CHS  $91   Breeze Airways          7:25 AM -> 9:45 AM
  min outbound: $91

...

Search complete. 11 destinations with available flights.
```

### Filter worldwide results by price and departure time

```bash
deadhead NYC 2026-05-12 --max-price 150 --leave-after 8 --leave-before 16
```

### Multiple travelers

```bash
deadhead NYC LAX 2026-05-12 2026-05-19 --travelers 2
```

### Exclude specific airports

```bash
deadhead NYC LON 2026-05-12 --exclude LHR,LGW
```

### Save results to JSON and Markdown

```bash
deadhead NYC LON 2026-05-12 --out-json results.json --out-md results.md
```

### Batch: check every leg of a multi-city trip

Create a file `legs.json`:

```json
[
  { "from": "NYC", "to": "LGW", "depart": "2026-05-12" },
  { "from": "LGW", "to": "ZRH", "depart": "2026-05-16" },
  { "from": "ZRH", "to": "PRG", "depart": "2026-05-19" },
  { "from": "PRG", "to": "NYC", "depart": "2026-05-22" }
]
```

```bash
deadhead --batch legs.json --out-json trip.json
```

```
Running 4 batch queries...

[1/4] NYC -> LGW  (2026-05-12)
  1 destination(s) with available flights.
London, England (LGW)  $279 one-way
  JFK -> LGW  $279  Norse Atlantic UK  6:20 PM -> 6:20 AM
  min outbound: $279

[2/4] LGW -> ZRH  (2026-05-16)
  1 destination(s) with available flights.
Zurich, Switzerland (ZRH)  $89 one-way
  LGW -> ZRH  $89   easyJet            8:30 AM -> 11:40 AM
  min outbound: $89

...

Batch complete. 4 total result(s) across all queries.
Results saved to trip.json
```

---

## Batch Mode

All CLI options are available per-query in the batch JSON object:

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
