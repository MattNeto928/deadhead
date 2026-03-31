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
- **URL encoding:** city codes are properly encoded in API requests
- **Comprehensive test suite:** 65 unit tests with full mock coverage across all packages

---

## Install

Requires Go 1.26+ and Google Chrome.

```bash
git clone https://github.com/MattNeto928/deadhead
cd deadhead
go build -o deadhead ./cmd/...
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

All packages have full unit test coverage. No browser or network access required — HTTP calls are mocked with `httptest`.

```bash
go test ./...
```

### Integration tests

Integration tests run against the real Skiplagged site and require Chrome and a network connection. They use a departure date one month from the current date so they never query the past.

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
