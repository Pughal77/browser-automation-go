package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

type ScrapeResult struct {
	HTML        string
	Title       string
	Description string
}

func Scrape(ctx context.Context, targetURL string, timeout time.Duration) (*ScrapeResult, error) {
	// Create a launcher
	l := launcher.New().
		Headless(true).
		Set("headless", "new") // Ensure we use the new headless mode as requested

	launchURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.MustClose()

	// Use stealth
	page := stealth.MustPage(browser)

	// Set timeout on the page context if provided
	page = page.Context(ctx)

	// Navigate to the URL
	err = page.Navigate(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", targetURL, err)
	}

	// Wait for the page load event (if not already loaded)
	_ = page.Timeout(3 * time.Second).WaitLoad()

	// Wait for the page to be stable (network idle) with a short timeout
	err = page.Timeout(5 * time.Second).WaitStable(time.Millisecond * 500)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Info: Page networking didn't fully idle within 5s, proceeding with extraction anyway.")
		} else {
			fmt.Printf("Warning: WaitStable encountered an issue: %v\n", err)
		}
	}

	// Sleep just a little bit to let JS render final elements
	time.Sleep(1 * time.Second)

	// Get page content
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("failed to get page HTML: %w", err)
	}

	// Get metadata without blocking indefinitely if elements don't exist
	title := ""
	if titleElem, err := page.Timeout(1 * time.Second).Element("title"); err == nil {
		title, _ = titleElem.Text()
	}

	// Get description from meta tag
	description := ""
	if descElem, err := page.Timeout(1 * time.Second).Element(`meta[name="description"]`); err == nil {
		if content, err := descElem.Attribute("content"); err == nil && content != nil {
			description = *content
		}
	}

	if description == "" {
		// Try og:description
		if descElem, err := page.Timeout(1 * time.Second).Element(`meta[property="og:description"]`); err == nil {
			if content, err := descElem.Attribute("content"); err == nil && content != nil {
				description = *content
			}
		}
	}

	return &ScrapeResult{
		HTML:        html,
		Title:       title,
		Description: fmt.Sprint(description),
	}, nil
}
