package vectordb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ArticleClassName = "Article"
	VectorDimensions = 1536
	WeaviateVersion  = "1.32.3"
	WeaviatePort     = "8080"
	WeaviateHost     = "localhost"
)

type WeaviateDB struct {
	process *exec.Cmd
	baseURL string
	client  *http.Client
	dataDir string
}

type Article struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Summary         string    `json:"summary"`
	Link            string    `json:"link"`
	PublicationDate time.Time `json:"publicationDate"`
	Vector          []float32 `json:"vector,omitempty"`
}

type WeaviateClass struct {
	Class      string              `json:"class"`
	Properties []WeaviateProperty  `json:"properties"`
	Vectorizer string              `json:"vectorizer"`
}

type WeaviateProperty struct {
	Name     string   `json:"name"`
	DataType []string `json:"dataType"`
}

type WeaviateObject struct {
	Class      string                 `json:"class"`
	Properties map[string]interface{} `json:"properties"`
	Vector     []float32              `json:"vector,omitempty"`
}

type WeaviateSearchResponse struct {
	Data struct {
		Get map[string][]map[string]interface{} `json:"Get"`
	} `json:"data"`
}

// ProgressReader wraps an io.Reader and displays download progress
type ProgressReader struct {
	Reader io.Reader
	Total  int64
	read   int64
	lastPercent int
}

// Read implements io.Reader interface with progress tracking
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.read += int64(n)
	
	// Calculate percentage
	percent := int(float64(pr.read) / float64(pr.Total) * 100)
	
	// Only update progress bar every 5% to avoid spam
	if percent != pr.lastPercent && percent%5 == 0 {
		pr.lastPercent = percent
		pr.displayProgress(percent)
	}
	
	return n, err
}

// displayProgress shows a simple progress bar
func (pr *ProgressReader) displayProgress(percent int) {
	barWidth := 40
	filledWidth := percent * barWidth / 100
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"
	
	fmt.Printf("\r📥 Downloading: %s %d%% (%.1f MB / %.1f MB)", 
		bar, percent, 
		float64(pr.read)/1024/1024, 
		float64(pr.Total)/1024/1024)
	
	if percent >= 100 {
		fmt.Println() // New line when complete
	}
}

func NewWeaviateDB(ctx context.Context) (*WeaviateDB, error) {
	// Create data directory
	dataDir := filepath.Join(".", "weaviate-data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Download Weaviate binary if not exists
	binaryPath, err := downloadWeaviateBinary(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download Weaviate binary: %w", err)
	}

	// Convert to absolute path for execution
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for binary: %w", err)
	}

	// Verify binary exists and is executable
	if _, err := os.Stat(absBinaryPath); err != nil {
		return nil, fmt.Errorf("binary not found at %s: %w", absBinaryPath, err)
	}

	// Ensure binary is executable (in case permissions were lost)
	if err := os.Chmod(absBinaryPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to make binary executable: %w", err)
	}

	fmt.Printf("🔧 Using Weaviate binary: %s\n", absBinaryPath)

	// Start Weaviate process
	cmd := exec.CommandContext(ctx, absBinaryPath,
		"--host", WeaviateHost,
		"--port", WeaviatePort,
		"--scheme", "http",
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true",
		"PERSISTENCE_DATA_PATH=data", // Relative to working directory
		"QUERY_DEFAULTS_LIMIT=25",
		"CLUSTER_HOSTNAME=node1",
	)

	// Set working directory
	cmd.Dir = dataDir

	fmt.Println("🚀 Starting Weaviate binary...")

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Weaviate: %w", err)
	}

	baseURL := fmt.Sprintf("http://%s:%s/v1", WeaviateHost, WeaviatePort)

	db := &WeaviateDB{
		process: cmd,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		dataDir: dataDir,
	}

	// Wait for Weaviate to be ready
	if err := db.waitForReady(ctx); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("Weaviate failed to start: %w", err)
	}

	// Create schema
	if err := db.createSchema(ctx); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

func downloadWeaviateBinary(dataDir string) (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	binaryName := "weaviate"
	if runtime.GOOS == "windows" {
		binaryName = "weaviate.exe"
	}

	binaryPath := filepath.Join(dataDir, binaryName)

	// Check if binary already exists
	if _, err := os.Stat(binaryPath); err == nil {
		fmt.Println("✅ Weaviate binary already exists")
		return binaryPath, nil
	}

	fmt.Printf("📥 Downloading Weaviate binary v%s for %s/%s...\n", WeaviateVersion, runtime.GOOS, arch)

	// Download URL for Weaviate binary
	downloadURL := fmt.Sprintf("https://github.com/weaviate/weaviate/releases/download/v%s/weaviate-v%s-linux-%s.tar.gz", 
		WeaviateVersion, WeaviateVersion, arch)

	// Download the tar.gz file with progress bar
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download Weaviate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download Weaviate: HTTP %d", resp.StatusCode)
	}

	// Get content length for progress bar
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		contentLength = 50 * 1024 * 1024 // Default to 50MB if unknown
	}

	// Create progress reader
	progressReader := &ProgressReader{
		Reader: resp.Body,
		Total:  contentLength,
	}

	// Extract the binary from tar.gz with progress
	if err := extractWeaviateBinary(progressReader, dataDir, binaryName); err != nil {
		return "", fmt.Errorf("failed to extract Weaviate binary: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	fmt.Println("✅ Weaviate binary downloaded and extracted")
	return binaryPath, nil
}

func extractWeaviateBinary(reader io.Reader, dataDir, binaryName string) error {
	// Create gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract the weaviate binary
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Look for the weaviate binary (it might be in a subdirectory)
		if strings.HasSuffix(header.Name, "/weaviate") || header.Name == "weaviate" {
			// Create the binary file
			binaryPath := filepath.Join(dataDir, binaryName)
			outFile, err := os.Create(binaryPath)
			if err != nil {
				return fmt.Errorf("failed to create binary file: %w", err)
			}
			defer outFile.Close()

			// Copy the binary content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to copy binary content: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("weaviate binary not found in archive")
}

func (w *WeaviateDB) waitForReady(ctx context.Context) error {
	fmt.Println("⏳ Waiting for Weaviate to be ready...")
	
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Weaviate to be ready")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"/meta", nil)
			if err != nil {
				continue
			}

			resp, err := w.client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("✅ Weaviate is ready!")
				return nil
			}
		}
	}
}

func (w *WeaviateDB) createSchema(ctx context.Context) error {
	class := WeaviateClass{
		Class: ArticleClassName,
		Properties: []WeaviateProperty{
			{Name: "title", DataType: []string{"text"}},
			{Name: "description", DataType: []string{"text"}},
			{Name: "summary", DataType: []string{"text"}},
			{Name: "link", DataType: []string{"text"}},
			{Name: "publicationDate", DataType: []string{"date"}},
		},
		Vectorizer: "none",
	}

	jsonData, err := json.Marshal(class)
	if err != nil {
		return fmt.Errorf("failed to marshal class: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/schema", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create schema, status: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("✅ Weaviate schema created")
	return nil
}

func (w *WeaviateDB) StoreArticle(ctx context.Context, article *Article) error {
	obj := WeaviateObject{
		Class: ArticleClassName,
		Properties: map[string]interface{}{
			"title":           article.Title,
			"description":     article.Description,
			"summary":         article.Summary,
			"link":            article.Link,
			"publicationDate": article.PublicationDate.Format(time.RFC3339),
		},
		Vector: article.Vector,
	}

	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/objects", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to store object, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (w *WeaviateDB) SearchSimilar(ctx context.Context, queryVector []float32, limit int) ([]*Article, error) {
	query := fmt.Sprintf(`{
		Get {
			%s(
				nearVector: {
					vector: %s
					certainty: 0.7
				}
				limit: %d
			) {
				title
				description
				summary
				link
				publicationDate
			}
		}
	}`, ArticleClassName, vectorToString(queryVector), limit)

	reqBody := map[string]string{"query": query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var searchResp WeaviateSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	articles := make([]*Article, 0)
	if articleList, ok := searchResp.Data.Get[ArticleClassName]; ok {
		for _, item := range articleList {
			article := &Article{}
			
			if title, ok := item["title"].(string); ok {
				article.Title = title
			}
			if description, ok := item["description"].(string); ok {
				article.Description = description
			}
			if summary, ok := item["summary"].(string); ok {
				article.Summary = summary
			}
			if link, ok := item["link"].(string); ok {
				article.Link = link
			}
			if pubDate, ok := item["publicationDate"].(string); ok {
				if parsedDate, err := time.Parse(time.RFC3339, pubDate); err == nil {
					article.PublicationDate = parsedDate
				}
			}
			
			articles = append(articles, article)
		}
	}

	return articles, nil
}

func (w *WeaviateDB) Close(ctx context.Context) error {
	if w.process != nil {
		fmt.Println("🛑 Stopping Weaviate...")
		if err := w.process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill Weaviate process: %w", err)
		}
		w.process.Wait()
		fmt.Println("✅ Weaviate stopped")
	}
	return nil
}

func vectorToString(vector []float32) string {
	result := "["
	for i, v := range vector {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%.6f", v)
	}
	result += "]"
	return result
}