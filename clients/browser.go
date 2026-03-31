package clients

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type headerTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("User-Agent", t.userAgent)
	req2.Header.Set("Accept", "application/json, text/plain, */*")
	req2.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req2.Header.Set("Referer", "https://skiplagged.com/")
	return t.base.RoundTrip(req2)
}

// Init launches a headless browser to solve the Cloudflare challenge,
// then builds the shared httpClient with the resulting session cookies.
func Init() error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("headless", false),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()

	var ua string
	var rawCookies []*network.Cookie

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://skiplagged.com"),
		// Wait until CF challenge clears — title changes away from "Just a moment..."
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Poll(
				`document.title.indexOf('Just a moment') === -1 && document.readyState === 'complete'`,
				nil,
				chromedp.WithPollingInterval(500*time.Millisecond),
				chromedp.WithPollingTimeout(50*time.Second),
			).Do(ctx)
		}),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`navigator.userAgent`, &ua),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			rawCookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)
	if err != nil {
		return fmt.Errorf("browser init failed: %w", err)
	}

	// cookiejar.New only errors when options are non-nil and invalid; nil is always safe.
	jar, _ := cookiejar.New(nil)
	// url.Parse only errors on invalid UTF-8; this literal is always valid.
	u, _ := url.Parse("https://skiplagged.com")

	httpCookies := make([]*http.Cookie, 0, len(rawCookies))
	for _, c := range rawCookies {
		// Strip double-quotes from cookie values — some CF cookies are JSON-encoded
		// strings where the raw value includes surrounding/embedded quotes, which
		// Go's net/http rejects as invalid per RFC 6265.
		value := strings.ReplaceAll(c.Value, `"`, "")
		httpCookies = append(httpCookies, &http.Cookie{
			Name:  c.Name,
			Value: value,
		})
	}
	jar.SetCookies(u, httpCookies)

	HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
		Transport: &headerTransport{
			base:      http.DefaultTransport,
			userAgent: ua,
		},
	}
	return nil
}
