package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/formatters"
	"github.com/mattneto928/deadhead/models"
	"github.com/mattneto928/deadhead/search"
)

var (
	proxy         = flag.String("proxy", "", "HTTP proxy URL")
	toCity        = flag.String("to", "", "destination city or airport (optional; omit for worldwide search)")
	skipWorldwide = flag.Bool("skip-worldwide", false, "skip worldwide fare computation")
	travelers     = flag.Int("travelers", 1, "number of travelers")
	maxPrice      = flag.Int("max-price", 0, "maximum total trip price")
	leaveAfter    = flag.Int("leave-after", 0, "outbound must depart after this hour (0-23)")
	leaveBefore   = flag.Int("leave-before", 0, "outbound must depart before this hour (0-23)")
	returnAfter   = flag.Int("return-after", 0, "return must depart after this hour (0-23)")
	returnBefore  = flag.Int("return-before", 0, "return must depart before this hour (0-23)")
	exclude       = flag.String("exclude", "", "comma-separated airports to exclude")
	outputJSON    = flag.String("out-json", "", "save results as JSON to this file")
	outputMD      = flag.String("out-md", "", "save results as markdown to this file")
	overwrite     = flag.Bool("overwrite", false, "overwrite existing output file")
	help          = flag.Bool("help", false, "show help")
	batchFile     = flag.String("batch", "", "path to JSON file containing array of queries")
)

type BatchQuery struct {
	From          string `json:"from"`
	To            string `json:"to,omitempty"`
	Depart        string `json:"depart"`
	Return        string `json:"return,omitempty"`
	Travelers     int    `json:"travelers,omitempty"`
	MaxPrice      int    `json:"max_price,omitempty"`
	LeaveAfter    int    `json:"leave_after,omitempty"`
	LeaveBefore   int    `json:"leave_before,omitempty"`
	ReturnAfter   int    `json:"return_after,omitempty"`
	ReturnBefore  int    `json:"return_before,omitempty"`
	Exclude       string `json:"exclude,omitempty"`
	SkipWorldwide bool   `json:"skip_worldwide,omitempty"`
}

// logger writes flight results to stdout (silenced when saving to a file).
// status writes progress and summary lines to stderr (always visible).
var (
	logger *log.Logger
	status *log.Logger
)

func init() {
	log.SetFlags(0) // suppress timestamps from the default logger used in sub-packages
	logger = log.New(os.Stdout, "", 0)
	status = log.New(os.Stderr, "", 0)
	search.StatusLogger = status.Printf
}

// boolFlag is the interface Go's flag package uses internally to detect bool flags.
type boolFlag interface {
	IsBoolFlag() bool
}

// parsePositional separates flag tokens from positional tokens so that flags
// can appear anywhere in the command (before or after positional args).
// Returns the positional args in order.
func parsePositional() []string {
	var flagTokens []string
	var positional []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		// Strip leading dashes and any =value suffix to look up the flag name.
		name := strings.TrimLeft(arg, "-")
		if eq := strings.Index(name, "="); eq != -1 {
			// --flag=value form: single token, no lookahead needed.
			name = name[:eq]
			flagTokens = append(flagTokens, arg)
			continue
		}
		flagTokens = append(flagTokens, arg)
		// Check if this flag takes a value (i.e. is not a bool flag).
		if f := flag.Lookup(name); f != nil {
			if bf, ok := f.Value.(boolFlag); !ok || !bf.IsBoolFlag() {
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					flagTokens = append(flagTokens, args[i])
				}
			}
		}
	}

	flag.CommandLine.Parse(flagTokens) //nolint:errcheck
	return positional
}

func usage() {
	fmt.Println(`Usage: deadhead [OPTIONS] FROM [TO] DEPART [RETURN]

Arguments:
  FROM    origin city or airport code (e.g. NYC, JFK)
  TO      destination city or airport code (e.g. LON, LHR) - optional, omit for worldwide search
  DEPART  departure date (2026-05-12)
  RETURN  return date (2026-05-20) - optional, omit for one-way

Examples:
  deadhead NYC 2026-05-12                          one-way, worldwide
  deadhead NYC 2026-05-12 2026-05-20               round-trip, worldwide
  deadhead NYC LON 2026-05-12                      one-way to London
  deadhead NYC LON 2026-05-12 2026-05-20           round-trip to London
  deadhead NYC LON 2026-05-12 --max-price 400      with price filter

Options:`)
	flag.PrintDefaults()
}

func saveJSON(req *models.Request, summaries []*search.CitySummary) error {
	if *outputJSON == "" {
		return nil
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	f, err := os.OpenFile(*outputJSON, flags, 0666)
	if err != nil {
		if os.IsExist(err) && *overwrite {
			os.Remove(*outputJSON)
			return saveJSON(req, summaries)
		}
		return err
	}
	defer f.Close()
	return formatters.ToJSON(f, req, summaries)
}

func saveMarkdown(summaries []*search.CitySummary) error {
	if *outputMD == "" {
		return nil
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	f, err := os.OpenFile(*outputMD, flags, 0666)
	if err != nil {
		if os.IsExist(err) && *overwrite {
			os.Remove(*outputMD)
			return saveMarkdown(summaries)
		}
		return err
	}
	defer f.Close()
	return formatters.ToMarkdown(f, summaries)
}

func printSummaries(summaries []*search.CitySummary, oneWay bool) {
	if len(summaries) == 0 {
		status.Println("No flights found matching your criteria. Try relaxing your filters.")
		return
	}

	for _, s := range summaries {
		if oneWay {
			logger.Printf("%s (%s)  $%d one-way\n", s.FullName, s.Name, s.MinLeavingPrice)
		} else {
			logger.Printf("%s (%s)  $%d round trip\n", s.FullName, s.Name, s.MinRoundTripPrice)
		}

		for _, f := range s.Leaving {
			tag := ""
			if f.IsHiddenCity {
				tag = fmt.Sprintf(" [HIDDEN-CITY to %s]", f.HiddenDestination)
			} else if f.Layovers > 0 {
				tag = fmt.Sprintf(" (+%d layovers)", f.Layovers)
			}

			logger.Printf("  %s -> %s  $%d  %-22s  %s -> %s%s\n",
				f.Departure.Airport,
				f.Arrival.Airport,
				f.Price,
				f.Airline,
				f.Departure.Time.Format("3:04 PM"),
				f.Arrival.Time.Format("3:04 PM"),
				tag,
			)
		}
		logger.Printf("  min outbound: $%d\n", s.MinLeavingPrice)

		if !oneWay && len(s.Returning) > 0 {
			logger.Println()
			for _, f := range s.Returning {
				tag := ""
				if f.IsHiddenCity {
					tag = fmt.Sprintf(" [HIDDEN-CITY to %s]", f.HiddenDestination)
				} else if f.Layovers > 0 {
					tag = fmt.Sprintf(" (+%d layovers)", f.Layovers)
				}

				logger.Printf("  %s -> %s  $%d  %-22s  %s -> %s%s\n",
					f.Departure.Airport,
					f.Arrival.Airport,
					f.Price,
					f.Airline,
					f.Departure.Time.Format("3:04 PM"),
					f.Arrival.Time.Format("3:04 PM"),
					tag,
				)
			}
			logger.Printf("  min return:   $%d\n", s.MinReturningPrice)
		}
		logger.Println()
	}

	if len(summaries) == 1 {
		status.Printf("Search complete. 1 destination with available flights.\n")
	} else {
		status.Printf("Search complete. %d destinations with available flights.\n", len(summaries))
	}
}

func main() {
	flag.Usage = usage
	positional := parsePositional()

	if *help || (len(positional) == 0 && *batchFile == "") {
		usage()
		return
	}

	if *proxy != "" {
		os.Setenv("HTTP_PROXY", *proxy)
	}

	// Silence the flight-data logger when saving to a file so results don't
	// appear on stdout AND in the file. Status messages still go to stderr.
	if *outputJSON != "" || *outputMD != "" {
		logger.SetOutput(io.Discard)
	}

	status.Println("Launching browser...")
	if err := clients.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	status.Println("Browser ready.")

	if *batchFile != "" {
		handleBatch()
		return
	}

	// Detect dates vs city codes in positional args.
	var cities, dates []string
	for _, p := range positional {
		if isDate(p) {
			dates = append(dates, p)
		} else {
			cities = append(cities, p)
		}
	}

	if len(cities) == 0 {
		fmt.Fprintln(os.Stderr, "error: origin city is required")
		usage()
		os.Exit(1)
	}
	if len(dates) == 0 {
		fmt.Fprintln(os.Stderr, "error: departure date is required")
		usage()
		os.Exit(1)
	}

	fromCity := cities[0]
	dest := *toCity
	if dest == "" && len(cities) > 1 {
		dest = cities[1]
	}

	depart := dates[0]
	returnDate := ""
	oneWay := true
	if len(dates) > 1 {
		returnDate = dates[1]
		oneWay = false
	}

	req, err := models.NewRequest(fromCity, dest, depart, returnDate, *travelers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	req.WithMaxPrice(*maxPrice).
		WithLeavingCriteria(*leaveAfter, *leaveBefore).
		WithReturningCriteria(*returnAfter, *returnBefore).
		WithExcludeAirportsCriteria(strings.Split(*exclude, ","))

	var summaries []*search.CitySummary
	if dest != "" {
		if oneWay {
			status.Printf("Searching %s -> %s on %s...\n", fromCity, dest, depart)
		} else {
			status.Printf("Searching %s -> %s  (%s -> %s)...\n", fromCity, dest, depart, returnDate)
		}
		summaries = []*search.CitySummary{{Name: dest}}
		summaries = search.GetAllFlightSummariesToCity(req, summaries)
	} else {
		if oneWay {
			status.Printf("Searching worldwide from %s on %s...\n", fromCity, depart)
		} else {
			status.Printf("Searching worldwide from %s  (%s -> %s)...\n", fromCity, depart, returnDate)
		}
		summaries, err = search.GetCitySummaryLeavingCity(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(summaries) == 0 {
			status.Println("No destinations found. The API returned no results for this route.")
			return
		}
		status.Printf("Found %d candidate destinations. Fetching flight details...\n", len(summaries))
		if !*skipWorldwide {
			summaries = search.GetAllFlightSummariesToCity(req, summaries)
		}
	}

	printSummaries(summaries, oneWay)

	if *outputJSON != "" {
		if err := saveJSON(req, summaries); err != nil {
			fmt.Fprintf(os.Stderr, "error saving JSON: %v\n", err)
			os.Exit(1)
		}
		status.Printf("Results saved to %s\n", *outputJSON)
	}
	if *outputMD != "" {
		if err := saveMarkdown(summaries); err != nil {
			fmt.Fprintf(os.Stderr, "error saving markdown: %v\n", err)
			os.Exit(1)
		}
		status.Printf("Results saved to %s\n", *outputMD)
	}
}

func handleBatch() {
	b, err := os.ReadFile(*batchFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading batch file: %v\n", err)
		os.Exit(1)
	}

	var queries []BatchQuery
	if err := json.Unmarshal(b, &queries); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing batch file: %v\n", err)
		os.Exit(1)
	}

	status.Printf("Running %d batch queries...\n", len(queries))

	var allSummaries []*search.CitySummary
	for i, q := range queries {
		dest := q.To
		if dest == "" {
			dest = "worldwide"
		}
		status.Printf("\n[%d/%d] %s -> %s  (%s)\n", i+1, len(queries), q.From, dest, q.Depart)

		travelersCount := q.Travelers
		if travelersCount == 0 {
			travelersCount = 1
		}

		req, err := models.NewRequest(q.From, q.To, q.Depart, q.Return, travelersCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error creating request: %v\n", err)
			continue
		}
		var excludes []string
		if q.Exclude != "" {
			excludes = strings.Split(q.Exclude, ",")
		}
		req.WithMaxPrice(q.MaxPrice).
			WithLeavingCriteria(q.LeaveAfter, q.LeaveBefore).
			WithReturningCriteria(q.ReturnAfter, q.ReturnBefore).
			WithExcludeAirportsCriteria(excludes)

		var summaries []*search.CitySummary
		if q.To != "" {
			summaries = []*search.CitySummary{{Name: q.To}}
			summaries = search.GetAllFlightSummariesToCity(req, summaries)
		} else {
			summaries, err = search.GetCitySummaryLeavingCity(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error fetching destinations: %v\n", err)
				continue
			}
			status.Printf("  Found %d candidate destinations.\n", len(summaries))
			if !q.SkipWorldwide {
				summaries = search.GetAllFlightSummariesToCity(req, summaries)
			}
		}

		oneWay := q.Return == ""
		if len(summaries) == 0 {
			status.Println("  No flights found matching your criteria.")
		} else {
			status.Printf("  %d destination(s) with available flights.\n", len(summaries))
		}
		printSummaries(summaries, oneWay)
		allSummaries = append(allSummaries, summaries...)
	}

	status.Printf("\nBatch complete. %d total result(s) across all queries.\n", len(allSummaries))

	// Batch results are saved with an empty request envelope; the per-query details
	// are visible in the printed output above.
	if len(allSummaries) > 0 {
		dummyReq := &models.Request{}
		if err := saveJSON(dummyReq, allSummaries); err != nil {
			fmt.Fprintf(os.Stderr, "error saving JSON: %v\n", err)
		} else if *outputJSON != "" {
			status.Printf("Results saved to %s\n", *outputJSON)
		}
		if err := saveMarkdown(allSummaries); err != nil {
			fmt.Fprintf(os.Stderr, "error saving markdown: %v\n", err)
		} else if *outputMD != "" {
			status.Printf("Results saved to %s\n", *outputMD)
		}
	}
}

func isDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	for i, c := range s {
		if i == 4 || i == 7 {
			if c != '-' {
				return false
			}
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
