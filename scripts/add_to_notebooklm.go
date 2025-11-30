package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	notebookLMURL    = "https://notebooklm.google.com"
	urlsFile         = "urls.txt"
	processedFile    = ".processed_urls.txt"
	defaultTimeout   = 30 * time.Second
)

// Config holds configuration for NotebookLM automation
type Config struct {
	AccessToken  string
	RefreshToken string
}

func main() {
	log.Println("Starting NotebookLM automation...")

	// Load configuration from environment
	config := Config{
		AccessToken:  os.Getenv("GOOGLE_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN"),
	}

	if config.AccessToken == "" {
		log.Fatal("Error: GOOGLE_ACCESS_TOKEN not found in environment")
	}

	// Get new URLs to process
	newURLs, err := getNewURLs()
	if err != nil {
		log.Fatalf("Error getting new URLs: %v", err)
	}

	if len(newURLs) == 0 {
		log.Println("No new URLs to process")
		return
	}

	log.Printf("Found %d new URLs to process\n", len(newURLs))

	// Setup Chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Login to NotebookLM
	if err := loginWithOAuth(ctx, config); err != nil {
		log.Fatalf("Error during login: %v", err)
	}

	// Process each URL
	for _, url := range newURLs {
		log.Printf("Processing: %s\n", url)
		if err := addURLToNotebookLM(ctx, url); err != nil {
			log.Printf("Error adding %s: %v\n", url, err)
			continue
		}

		if err := markAsProcessed(url); err != nil {
			log.Printf("Warning: Could not mark URL as processed: %v\n", err)
		}

		time.Sleep(2 * time.Second)
	}

	// Generate audio guide
	if err := generateAudioGuide(ctx); err != nil {
		log.Printf("Error generating audio: %v\n", err)
	}

	log.Println("All URLs processed successfully")
}

// loginWithOAuth authenticates using OAuth tokens
func loginWithOAuth(ctx context.Context, config Config) error {
	log.Println("Logging in to NotebookLM...")

	return chromedp.Run(ctx,
		chromedp.Navigate(notebookLMURL),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(fmt.Sprintf(`
			localStorage.setItem('access_token', '%s');
			localStorage.setItem('refresh_token', '%s');
		`, config.AccessToken, config.RefreshToken), nil),
		chromedp.Reload(),
		chromedp.Sleep(3*time.Second),
	)
}

// addURLToNotebookLM adds a URL as a source to NotebookLM
func addURLToNotebookLM(ctx context.Context, url string) error {
	log.Printf("Adding URL to NotebookLM: %s\n", url)

	err := chromedp.Run(ctx,
		// Click "Add source" button
		chromedp.WaitVisible(`//button[contains(., 'Add source')]`, chromedp.BySearch),
		chromedp.Click(`//button[contains(., 'Add source')]`, chromedp.BySearch),
		chromedp.Sleep(2*time.Second),

		// Enter URL
		chromedp.WaitVisible(`input[type='url']`, chromedp.ByQuery),
		chromedp.SendKeys(`input[type='url']`, url, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),

		// Click Add button
		chromedp.Click(`//button[contains(., 'Add')]`, chromedp.BySearch),
		chromedp.Sleep(5*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}

	log.Printf("Successfully added: %s\n", url)
	return nil
}

// generateAudioGuide triggers audio guide generation
func generateAudioGuide(ctx context.Context) error {
	log.Println("Generating audio guide...")

	err := chromedp.Run(ctx,
		// Click Studio button
		chromedp.WaitVisible(`//button[contains(., 'Studio')]`, chromedp.BySearch),
		chromedp.Click(`//button[contains(., 'Studio')]`, chromedp.BySearch),
		chromedp.Sleep(2*time.Second),

		// Click Generate button
		chromedp.WaitVisible(`//button[contains(., 'Generate')]`, chromedp.BySearch),
		chromedp.Click(`//button[contains(., 'Generate')]`, chromedp.BySearch),
		chromedp.Sleep(10*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to generate audio: %w", err)
	}

	log.Println("Audio generation started")
	return nil
}

// getNewURLs returns URLs that haven't been processed yet
func getNewURLs() ([]string, error) {
	// Read all URLs
	allURLs, err := readLines(urlsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading URLs file: %w", err)
	}

	// Filter out comments and empty lines
	var validURLs []string
	for _, url := range allURLs {
		url = strings.TrimSpace(url)
		if url != "" && !strings.HasPrefix(url, "#") {
			validURLs = append(validURLs, url)
		}
	}

	// Read processed URLs
	processedURLs := make(map[string]bool)
	if _, err := os.Stat(processedFile); err == nil {
		processed, err := readLines(processedFile)
		if err != nil {
			return nil, fmt.Errorf("error reading processed URLs file: %w", err)
		}
		for _, url := range processed {
			processedURLs[strings.TrimSpace(url)] = true
		}
	}

	// Filter out already processed URLs
	var newURLs []string
	for _, url := range validURLs {
		if !processedURLs[url] {
			newURLs = append(newURLs, url)
		}
	}

	return newURLs, nil
}

// markAsProcessed marks a URL as processed
func markAsProcessed(url string) error {
	f, err := os.OpenFile(processedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening processed file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(url + "\n"); err != nil {
		return fmt.Errorf("error writing to processed file: %w", err)
	}

	return nil
}

// readLines reads a file and returns its lines
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
