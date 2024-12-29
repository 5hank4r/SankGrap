package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Command-line flags
	domainFlag := flag.String("d", "", "Domain to filter subdomains (e.g., example.com)")
	urlsFile := flag.String("f", "", "File containing URLs to process")
	outputFile := flag.String("o", "", "File to save extracted subdomains")
	workers := flag.Int("w", 10, "Number of concurrent workers")
	modeFlag := flag.String("m", "both", "Mode: 'rb' (response body), 'rh' (response header), or 'both'")
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	if *domainFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: Domain (-d) is required")
		os.Exit(1)
	}

	if *modeFlag != "rb" && *modeFlag != "rh" && *modeFlag != "both" {
		fmt.Fprintln(os.Stderr, "Error: Invalid mode (-m). Use 'rb', 'rh', or 'both'")
		os.Exit(1)
	}

	// Read URLs from file
	urls, err := readLines(*urlsFile)
	if err != nil {
		log.Fatalf("Failed to read URLs: %v", err)
	}

	// Prepare regex for subdomains
	subdomainRegex := regexp.MustCompile(`([a-zA-Z0-9-]+\.)+` + regexp.QuoteMeta(*domainFlag))

	// Progress bar
	bar := progressbar.Default(int64(len(urls)))

	// Results collection
	results := make(map[string]struct{})
	var mu sync.Mutex

	// Process URLs concurrently
	var wg sync.WaitGroup
	urlsChan := make(chan string, len(urls))
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlsChan {
				processURL(url, subdomainRegex, results, &mu, bar, *verboseFlag, *modeFlag)
			}
		}()
	}

	// Feed URLs into the channel
	for _, url := range urls {
		urlsChan <- url
	}
	close(urlsChan)

	wg.Wait()

	// Write results to output file or stdout
	if err := writeResults(results, *outputFile); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}

	fmt.Println("Extraction complete.")
}

func processURL(url string, subdomainRegex *regexp.Regexp, results map[string]struct{}, mu *sync.Mutex, bar *progressbar.ProgressBar, verbose bool, mode string) {
	defer bar.Add(1)

	if verbose {
		log.Printf("Processing: %s", url)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if verbose {
			log.Printf("Failed to create request for %s: %v", url, err)
		}
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			log.Printf("Failed to fetch URL %s: %v", url, err)
		}
		return
	}
	defer resp.Body.Close()

	// Extract subdomains based on mode
	if mode == "rh" || mode == "both" {
		extractFromHeaders(resp.Header, subdomainRegex, results, mu)
	}

	if mode == "rb" || mode == "both" {
		extractFromBody(resp.Body, subdomainRegex, results, mu)
	}
}

func extractFromHeaders(headers http.Header, subdomainRegex *regexp.Regexp, results map[string]struct{}, mu *sync.Mutex) {
	for _, values := range headers {
		for _, value := range values {
			matches := subdomainRegex.FindAllString(value, -1)
			mu.Lock()
			for _, match := range matches {
				results[match] = struct{}{}
			}
			mu.Unlock()
		}
	}
}

func extractFromBody(body io.ReadCloser, subdomainRegex *regexp.Regexp, results map[string]struct{}, mu *sync.Mutex) {
	bodyContent, err := io.ReadAll(body)
	if err == nil {
		matches := subdomainRegex.FindAllString(string(bodyContent), -1)
		mu.Lock()
		for _, match := range matches {
			results[match] = struct{}{}
		}
		mu.Unlock()
	}
}

// Utility functions for reading lines and writing results
func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	return lines, scanner.Err()
}

func writeResults(results map[string]struct{}, outputFile string) error {
	if outputFile == "" {
		for subdomain := range results {
			fmt.Println(subdomain)
		}
		return nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for subdomain := range results {
		if _, err := file.WriteString(subdomain + "\n"); err != nil {
			return err
		}
	}
	return nil
}
