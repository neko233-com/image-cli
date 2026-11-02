package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/neko233-com/image-cli/internal/version"
)

const defaultRepo = "neko233-com/image-cli"

type Options struct {
	Repo           string
	CurrentVersion string
	BinaryName     string
}

type Result struct {
	CurrentVersion string
	LatestVersion  string
	Route          string
	Updated        bool
}

type release struct {
	TagName string `json:"tag_name"`
}

// Run 下载并替换当前可执行文件（需校验 SHA256SUMS）。
func Run(ctx context.Context, options Options) (Result, error) {
	if options.Repo == "" {
		options.Repo = defaultRepo
	}
	if options.CurrentVersion == "" {
		options.CurrentVersion = "dev"
	}
	if options.BinaryName == "" {
		options.BinaryName = version.Binary
	}

	latest, route, err := latestVersion(ctx, options.Repo)
	if err != nil {
		return Result{CurrentVersion: options.CurrentVersion}, err
	}
	result := Result{CurrentVersion: options.CurrentVersion, LatestVersion: latest, Route: route}
	if compareVersions(options.CurrentVersion, latest) >= 0 {
		return result, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return result, fmt.Errorf("locate executable: %w", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return result, fmt.Errorf("resolve executable: %w", err)
	}

	asset := assetFileName(options.BinaryName, latest, runtime.GOOS, runtime.GOARCH)
	if asset == "" {
		return result, fmt.Errorf("unsupported update target: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	baseURL := "https://github.com/" + options.Repo + "/releases/download/" + latest
	routes := candidateURLs(baseURL)
	selectedAssetURL, err := fastestURL(ctx, appendSuffix(routes, asset))
	if err != nil {
		return result, fmt.Errorf("find reachable update route: %w", err)
	}
	selectedBase := strings.TrimSuffix(selectedAssetURL, "/"+asset)
	routes = prioritize(routes, selectedBase)

	tempDir, err := os.MkdirTemp("", "image-cli-update-")
	if err != nil {
		return result, fmt.Errorf("create update directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, asset)
	sumsPath := filepath.Join(tempDir, "SHA256SUMS")
	var lastRouteErr error
	for _, routeBase := range routes {
		_ = os.Remove(binaryPath)
		_ = os.Remove(sumsPath)
		if err := download(ctx, routeBase+"/"+asset, binaryPath); err != nil {
			lastRouteErr = err
			continue
		}
		if err := download(ctx, routeBase+"/SHA256SUMS", sumsPath); err != nil {
			lastRouteErr = err
			continue
		}
		sums, err := os.ReadFile(sumsPath)
		if err != nil {
			lastRouteErr = err
			continue
		}
		expected, err := checksumFor(string(sums), asset)
		if err != nil {
			lastRouteErr = err
			continue
		}
		if err := verifySHA256(binaryPath, expected); err != nil {
			lastRouteErr = err
			continue
		}
		result.Route = routeBase
		lastRouteErr = nil
		break
	}
	if lastRouteErr != nil {
		return result, fmt.Errorf("download and verify release: %w", lastRouteErr)
	}

	if err := installBinary(binaryPath, executable); err != nil {
		return result, fmt.Errorf("replace executable: %w", err)
	}
	result.Updated = true
	return result, nil
}

// NotifyOptions 控制懒检查（仅提示，不自动替换）。
type NotifyOptions struct {
	Repo           string
	CurrentVersion string
	CachePath      string
	MinInterval    time.Duration
}

// NotifyIfNewer 在后台节流检查新版本；若有更新则返回提示文案（空表示无需提示）。
func NotifyIfNewer(ctx context.Context, opts NotifyOptions) (string, error) {
	if opts.Repo == "" {
		opts.Repo = defaultRepo
	}
	if opts.CurrentVersion == "" || opts.CurrentVersion == "dev" {
		return "", nil
	}
	if opts.MinInterval <= 0 {
		opts.MinInterval = 24 * time.Hour
	}
	cachePath, err := resolveCachePath(opts.CachePath)
	if err != nil {
		return "", err
	}
	if shouldSkipCheck(cachePath, opts.MinInterval) {
		return "", nil
	}
	latest, route, err := latestVersion(ctx, opts.Repo)
	if err != nil {
		_ = writeCache(cachePath, cacheEntry{CheckedAt: time.Now().UTC(), LastError: err.Error()})
		return "", err
	}
	_ = writeCache(cachePath, cacheEntry{CheckedAt: time.Now().UTC(), Latest: latest, Route: route})
	if compareVersions(opts.CurrentVersion, latest) >= 0 {
		return "", nil
	}
	return fmt.Sprintf("image-cli %s 可用（当前 %s）。运行: image-cli update", latest, opts.CurrentVersion), nil
}

type cacheEntry struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest,omitempty"`
	Route     string    `json:"route,omitempty"`
	LastError string    `json:"last_error,omitempty"`
}

func resolveCachePath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "image-cli", "update-check.json"), nil
}

func shouldSkipCheck(path string, minInterval time.Duration) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var entry cacheEntry
	if json.Unmarshal(data, &entry) != nil {
		return false
	}
	return time.Since(entry.CheckedAt) < minInterval
}

func writeCache(path string, entry cacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func latestVersion(ctx context.Context, repo string) (string, string, error) {
	apiURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	candidates := candidateURLs(apiURL)
	selected, err := fastestURL(ctx, candidates)
	if err != nil {
		return "", "", err
	}
	var lastErr error
	for _, candidate := range prioritize(candidates, selected) {
		current, err := fetchLatest(ctx, candidate)
		if err == nil {
			return current, candidate, nil
		}
		lastErr = err
	}
	return "", selected, lastErr
}

func fetchLatest(ctx context.Context, selected string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, selected, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "image-cli-updater")
	response, err := client().Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("latest release returned %s", response.Status)
	}
	var current release
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&current); err != nil {
		return "", fmt.Errorf("decode latest release: %w", err)
	}
	if !strings.HasPrefix(current.TagName, "v") || strings.TrimPrefix(current.TagName, "v") == "" {
		return "", errors.New("latest release has invalid version tag")
	}
	return current.TagName, nil
}

func candidateURLs(original string) []string {
	urls := []string{original}
	if os.Getenv("IMAGE_CLI_DIRECT_ONLY") == "1" {
		return urls
	}
	mirrors, configured := os.LookupEnv("IMAGE_CLI_UPDATE_MIRRORS")
	if !configured {
		mirrors = "https://gh-proxy.com/,https://ghfast.top/,https://ghproxy.net/"
	}
	for _, mirror := range strings.Split(mirrors, ",") {
		mirror = strings.TrimSpace(mirror)
		if mirror == "" {
			continue
		}
		if strings.Contains(mirror, "{url}") {
			urls = append(urls, strings.ReplaceAll(mirror, "{url}", original))
		} else {
			urls = append(urls, strings.TrimRight(mirror, "/")+"/"+original)
		}
	}
	return unique(urls)
}

func appendSuffix(routes []string, suffix string) []string {
	urls := make([]string, 0, len(routes))
	for _, route := range routes {
		urls = append(urls, strings.TrimRight(route, "/")+"/"+suffix)
	}
	return urls
}

func fastestURL(ctx context.Context, urls []string) (string, error) {
	type result struct {
		url string
		err error
	}
	probeContext, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan result, len(urls))
	for _, url := range urls {
		go func(url string) {
			request, err := http.NewRequestWithContext(probeContext, http.MethodGet, url, nil)
			if err != nil {
				results <- result{url: url, err: err}
				return
			}
			request.Header.Set("Range", "bytes=0-65535")
			request.Header.Set("User-Agent", "image-cli-updater")
			response, err := client().Do(request)
			if err != nil {
				results <- result{url: url, err: err}
				return
			}
			_, readErr := io.CopyN(io.Discard, response.Body, 64<<10)
			_ = response.Body.Close()
			if response.StatusCode < 200 || response.StatusCode >= 400 {
				results <- result{url: url, err: fmt.Errorf("status %s", response.Status)}
				return
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				results <- result{url: url, err: readErr}
				return
			}
			results <- result{url: url}
		}(url)
	}
	var lastErr error
	for range urls {
		select {
		case found := <-results:
			if found.err == nil {
				return found.url, nil
			}
			lastErr = found.err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no candidate route")
	}
	return "", lastErr
}

func prioritize(values []string, first string) []string {
	ordered := []string{}
	for _, value := range append([]string{first}, values...) {
		if value == "" || contains(ordered, value) {
			continue
		}
		ordered = append(ordered, value)
	}
	return ordered
}

func unique(values []string) []string {
	result := []string{}
	for _, value := range values {
		if !contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func download(ctx context.Context, url, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "image-cli-updater")
	response, err := client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", response.Status)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, io.LimitReader(response.Body, 256<<20))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func client() *http.Client {
	return &http.Client{
		Timeout:   45 * time.Second,
		Transport: &http.Transport{Proxy: nil, ForceAttemptHTTP2: true},
	}
}

func assetFileName(binary, tag, goos, goarch string) string {
	_ = strings.TrimPrefix(tag, "v")
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	switch goos {
	case "linux", "darwin", "windows":
		if goarch == "amd64" || goarch == "arm64" {
			return fmt.Sprintf("%s-%s-%s%s", binary, goos, goarch, ext)
		}
	}
	return ""
}

func checksumFor(contents, asset string) (string, error) {
	for _, line := range strings.Split(contents, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := filepath.Base(fields[len(fields)-1])
		if name != asset && name != "dist/"+asset {
			continue
		}
		checksum := strings.ToLower(fields[0])
		if len(checksum) == sha256.Size*2 {
			if _, err := hex.DecodeString(checksum); err == nil {
				return checksum, nil
			}
		}
	}
	return "", fmt.Errorf("checksum missing for %s", asset)
}

func verifySHA256(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("checksum verification failed")
	}
	return nil
}

func installBinary(source, destination string) error {
	if runtime.GOOS != "windows" {
		newPath := destination + ".new"
		if err := os.Remove(newPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := copyFile(source, newPath); err != nil {
			return err
		}
		if err := os.Chmod(newPath, 0o755); err != nil {
			_ = os.Remove(newPath)
			return err
		}
		if err := os.Rename(newPath, destination); err != nil {
			_ = os.Remove(newPath)
			return err
		}
		return nil
	}
	newPath := destination + ".new"
	if err := os.Remove(newPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := copyFile(source, newPath); err != nil {
		return err
	}
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	script := "$new=" + quote(newPath) + ";$old=" + quote(destination) + ";for($i=0;$i -lt 20;$i++){try{Move-Item -LiteralPath $new -Destination $old -Force;exit 0}catch{Start-Sleep -Milliseconds 250}};exit 1"
	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	if err := command.Start(); err != nil {
		return err
	}
	return nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func compareVersions(current, latest string) int {
	left := versionParts(current)
	right := versionParts(latest)
	for index := 0; index < 3; index++ {
		if left[index] < right[index] {
			return -1
		}
		if left[index] > right[index] {
			return 1
		}
	}
	return 0
}

func versionParts(version string) [3]int {
	var parts [3]int
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	for index, value := range strings.Split(version, ".") {
		if index >= len(parts) {
			break
		}
		parts[index], _ = strconv.Atoi(value)
	}
	return parts
}
