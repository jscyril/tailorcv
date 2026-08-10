// Command packagetectonic assembles TailorCV's pinned, offline Tectonic runtime.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	tectonicVersion = "0.16.9"
	bundleSourceURL = "https://data1.fullyjustified.net/tlextras-2022.0r0.tar"
	bundleDigest    = "6ffe055852f8faf66c0acbe1a7fb27f87b869a90bad1204f3bf4d9683f597c7c"
	bundleFilename  = "tectonic-resources.zip"
	maxArchiveBytes = 64 << 20
	commandTimeout  = 4 * time.Minute
)

type releaseAsset struct {
	url        string
	sha256     string
	archive    string
	executable string
}

var releaseAssets = map[string]releaseAsset{
	"linux/amd64": {
		url:        "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.16.9/tectonic-0.16.9-x86_64-unknown-linux-musl.tar.gz",
		sha256:     "60b13a0826ae7ad9ce34b4a2df06bff2cfcfa6dda8a915477c0cbb84e1a4a902",
		archive:    "tar.gz",
		executable: "tectonic",
	},
	"darwin/arm64": {
		url:        "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.16.9/tectonic-0.16.9-aarch64-apple-darwin.tar.gz",
		sha256:     "edb67c61aba768289f6da441c9e6f523cfaff4f8b2a5708523ef29c543f8e88e",
		archive:    "tar.gz",
		executable: "tectonic",
	},
	"windows/amd64": {
		url:        "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.16.9/tectonic-0.16.9-x86_64-pc-windows-msvc.zip",
		sha256:     "131a24604785a9600989a3d91225f597df52ac06f00aeffe86fd529f99ee5cdd",
		archive:    "zip",
		executable: "tectonic.exe",
	},
}

func main() {
	targetOS := flag.String("goos", runtime.GOOS, "target operating system")
	targetArch := flag.String("goarch", runtime.GOARCH, "target architecture")
	destination := flag.String("destination", "", "directory placed beside the packaged application executable")
	fixtures := flag.String("fixtures", filepath.FromSlash("build/tectonic/fixtures"), "directory containing offline LaTeX fixtures")
	flag.Parse()

	if strings.TrimSpace(*destination) == "" {
		fatalf("-destination is required")
	}
	asset, ok := releaseAssets[*targetOS+"/"+*targetArch]
	if !ok {
		fatalf("no pinned Tectonic asset for %s/%s", *targetOS, *targetArch)
	}
	fixturePaths, err := findFixtures(*fixtures)
	if err != nil {
		fatalf("load offline fixtures: %v", err)
	}
	if err := packageRuntime(asset, *destination, fixturePaths); err != nil {
		fatalf("package Tectonic: %v", err)
	}
	fmt.Printf("Packaged Tectonic %s and %s for %s/%s in %s\n", tectonicVersion, bundleFilename, *targetOS, *targetArch, *destination)
}

func packageRuntime(asset releaseAsset, destination string, fixtures []string) error {
	temporary, err := os.MkdirTemp("", "tailorcv-tectonic-package-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	archivePath := filepath.Join(temporary, "tectonic-archive")
	if err := downloadVerified(asset.url, asset.sha256, archivePath); err != nil {
		return err
	}
	extractedPath := filepath.Join(temporary, asset.executable)
	if err := extractExecutable(archivePath, asset, extractedPath); err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	packagedExecutable := filepath.Join(destination, asset.executable)
	if err := copyFile(extractedPath, packagedExecutable, 0o755); err != nil {
		return fmt.Errorf("install executable: %w", err)
	}

	resourceDirectory, err := hydrateResources(packagedExecutable, temporary, fixtures)
	if err != nil {
		return err
	}
	packagedBundle := filepath.Join(destination, bundleFilename)
	if err := writeResourceBundle(resourceDirectory, packagedBundle); err != nil {
		return err
	}
	if err := verifyOffline(packagedExecutable, packagedBundle, temporary, fixtures); err != nil {
		return err
	}
	return nil
}

func findFixtures(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".tex") {
			continue
		}
		absolute, err := filepath.Abs(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		paths = append(paths, absolute)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no .tex fixtures found in %s", directory)
	}
	return paths, nil
}

func downloadVerified(url, expectedDigest, destination string) error {
	client := &http.Client{Timeout: commandTimeout}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected HTTP status %s", url, response.Status)
	}

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(output, hash), io.LimitReader(response.Body, maxArchiveBytes+1))
	closeErr := output.Close()
	if copyErr != nil {
		return fmt.Errorf("download %s: %w", url, copyErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if written > maxArchiveBytes {
		return fmt.Errorf("download %s exceeded the %d MiB limit", url, maxArchiveBytes>>20)
	}
	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if actualDigest != expectedDigest {
		return fmt.Errorf("download %s failed SHA-256 verification: got %s", url, actualDigest)
	}
	return nil
}

func extractExecutable(archivePath string, asset releaseAsset, destination string) error {
	switch asset.archive {
	case "tar.gz":
		return extractExecutableFromTarGZ(archivePath, asset.executable, destination)
	case "zip":
		return extractExecutableFromZIP(archivePath, asset.executable, destination)
	default:
		return fmt.Errorf("unsupported archive format %q", asset.archive)
	}
}

func extractExecutableFromTarGZ(archivePath, executableName, destination string) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	compressed, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(filepath.Clean(header.Name)) == executableName {
			return writeReader(destination, reader, 0o755)
		}
	}
	return fmt.Errorf("%s is missing from the release archive", executableName)
}

func extractExecutableFromZIP(archivePath, executableName, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if !file.FileInfo().Mode().IsRegular() || filepath.Base(filepath.Clean(file.Name)) != executableName {
			continue
		}
		input, err := file.Open()
		if err != nil {
			return err
		}
		err = writeReader(destination, input, 0o755)
		closeErr := input.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	return fmt.Errorf("%s is missing from the release archive", executableName)
}

func hydrateResources(executable, temporary string, fixtures []string) (string, error) {
	cacheRoot := filepath.Join(temporary, "online-cache")
	for index, fixture := range fixtures {
		outputDirectory := filepath.Join(temporary, fmt.Sprintf("online-output-%d", index))
		if err := os.MkdirAll(outputDirectory, 0o700); err != nil {
			return "", err
		}
		arguments := compileArguments(bundleSourceURL, outputDirectory, fixture, false)
		if output, err := runTectonic(executable, arguments, cacheRoot, false); err != nil {
			return "", fmt.Errorf("hydrate resources with %s: %w\n%s", filepath.Base(fixture), err, output)
		}
	}

	var resourceDirectory string
	err := filepath.WalkDir(cacheRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && entry.Name() == bundleDigest {
			resourceDirectory = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if resourceDirectory == "" {
		return "", fmt.Errorf("Tectonic cache did not contain pinned bundle %s", bundleDigest)
	}
	return resourceDirectory, nil
}

func writeResourceBundle(resourceDirectory, destination string) error {
	entries, err := os.ReadDir(resourceDirectory)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() != "SHA256SUM" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return fmt.Errorf("pinned resource cache is empty")
	}

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(output)
	for _, name := range names {
		if err := addFileToZIP(archive, filepath.Join(resourceDirectory, name), name); err != nil {
			_ = archive.Close()
			_ = output.Close()
			return err
		}
	}
	if err := addBytesToZIP(archive, "SHA256SUM", []byte(bundleDigest)); err != nil {
		_ = archive.Close()
		_ = output.Close()
		return err
	}
	if err := archive.Close(); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

func addFileToZIP(archive *zip.Writer, path, name string) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()
	writer, err := archive.CreateHeader(zipHeader(name))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, input)
	return err
}

func addBytesToZIP(archive *zip.Writer, name string, content []byte) error {
	writer, err := archive.CreateHeader(zipHeader(name))
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	return err
}

func zipHeader(name string) *zip.FileHeader {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(0o644)
	header.SetModTime(time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC))
	return header
}

func verifyOffline(executable, bundle, temporary string, fixtures []string) error {
	cacheRoot := filepath.Join(temporary, "offline-cache")
	for index, fixture := range fixtures {
		outputDirectory := filepath.Join(temporary, fmt.Sprintf("offline-output-%d", index))
		if err := os.MkdirAll(outputDirectory, 0o700); err != nil {
			return err
		}
		arguments := compileArguments(bundle, outputDirectory, fixture, true)
		if output, err := runTectonic(executable, arguments, cacheRoot, true); err != nil {
			return fmt.Errorf("offline verification with %s: %w\n%s", filepath.Base(fixture), err, output)
		}
		pdfPath := filepath.Join(outputDirectory, strings.TrimSuffix(filepath.Base(fixture), filepath.Ext(fixture))+".pdf")
		pdf, err := os.ReadFile(pdfPath)
		if err != nil {
			return fmt.Errorf("read offline fixture PDF: %w", err)
		}
		if len(pdf) < 5 || string(pdf[:5]) != "%PDF-" {
			return fmt.Errorf("offline fixture %s did not produce a PDF", filepath.Base(fixture))
		}
	}
	return nil
}

func compileArguments(bundle, outputDirectory, fixture string, onlyCached bool) []string {
	arguments := []string{"--untrusted", "--keep-logs", "--color", "never", "--bundle", bundle, "--outdir", outputDirectory}
	if onlyCached {
		arguments = append([]string{"--only-cached"}, arguments...)
	}
	return append(arguments, fixture)
}

func runTectonic(executable string, arguments []string, cacheRoot string, blockNetwork bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, executable, arguments...)
	environment := setEnvironment(os.Environ(), "XDG_CACHE_HOME", cacheRoot)
	environment = setEnvironment(environment, "LOCALAPPDATA", cacheRoot)
	environment = setEnvironment(environment, "TECTONIC_CACHE_DIR", cacheRoot)
	environment = setEnvironment(environment, "TECTONIC_UNTRUSTED_MODE", "1")
	if blockNetwork {
		for _, key := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"} {
			environment = setEnvironment(environment, key, "http://127.0.0.1:9")
		}
		environment = setEnvironment(environment, "NO_PROXY", "")
		environment = setEnvironment(environment, "no_proxy", "")
	}
	command.Env = environment
	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("timed out after %s", commandTimeout)
	}
	return string(output), err
}

func setEnvironment(environment []string, key, value string) []string {
	prefix := strings.ToUpper(key) + "="
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if !strings.HasPrefix(strings.ToUpper(item), prefix) {
			result = append(result, item)
		}
	}
	return append(result, key+"="+value)
}

func copyFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	return writeReader(destination, input, mode)
}

func writeReader(destination string, reader io.Reader, mode os.FileMode) error {
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, reader)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "packagetectonic: "+format+"\n", arguments...)
	os.Exit(1)
}
