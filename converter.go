package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/go-shiori/go-readability"
)

type ConversionResult struct {
	Markdown    string
	Title       string
	Description string
	URL         string
}

func Convert(html, originalURL string) (*ConversionResult, error) {
	// 1. Clean the HTML using readability (akin to what MCP Fetch does)
	parsedURL, _ := url.Parse(originalURL)
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		// If readability fails, fallback to raw HTML conversion
		fmt.Printf("Warning: readability failed: %v, falling back to raw conversion\n", err)
	}

	contentToConvert := html
	title := ""
	description := ""

	if err == nil {
		contentToConvert = article.Content
		title = article.Title
		description = article.Excerpt
	}

	// 2. Convert HTML to Markdown
	// Use v2 of html-to-markdown with base and commonmark plugins registered
	mdConverter := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
		),
	)
	markdown, err := mdConverter.ConvertString(contentToConvert)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return &ConversionResult{
		Markdown:    markdown,
		Title:       title,
		Description: description,
		URL:         originalURL,
	}, nil
}
