package azure

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var PAT = os.Getenv("AZURE_PAT")

type Package struct {
	Name     string
	Versions []string
}

func getFeedID(org string) (feedID, projectID string) {
	apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/packaging/feeds?api-version=6.0-preview.1", org)
	fmt.Printf("🌐 GET %s\n", apiUrl)

	req, _ := http.NewRequest("GET", apiUrl, nil)
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+PAT))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "*/*")

	fmt.Println("🧪 Sending request to Azure DevOps...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("💥 HTTP request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("📥 HTTP %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("❌ Failed to get feed ID (%d): %s", resp.StatusCode, string(body)))
	}

	var result struct {
		Value []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Visibility string `json:"visibility"`
			Project    *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project,omitempty"`
			UpstreamSources []struct {
				Protocol string `json:"protocol"`
			} `json:"upstreamSources"`
		} `json:"value"`
	}
	json.Unmarshal(body, &result)

	if len(result.Value) == 0 {
		panic("❌ No feeds found in Azure DevOps response.")
	}

	targetFeed := os.Getenv("AZURE_FEED")
	fmt.Printf("🔍 Looking for NuGet feed matching: '%s'\n\n", targetFeed)
	fmt.Println("📋 Feeds discovered:")

	var nugetFeeds []struct {
		ID         string
		Name       string
		ProjectID  string
		Visibility string
	}

	for _, feed := range result.Value {
		hasNuGet := false
		for _, source := range feed.UpstreamSources {
			if strings.EqualFold(source.Protocol, "nuget") {
				hasNuGet = true
				break
			}
		}

		projectName := "-"
		if feed.Project != nil {
			projectName = feed.Project.Name
		}

		visibility := feed.Visibility
		if visibility == "" {
			visibility = "unknown"
		}

		fmt.Printf("• %s (ID: %s) [%s, nuget %s, project: %s]\n",
			feed.Name,
			feed.ID,
			strings.ToLower(visibility),
			map[bool]string{true: "✅", false: "❌"}[hasNuGet],
			projectName,
		)

		if hasNuGet {
			nugetFeeds = append(nugetFeeds, struct {
				ID         string
				Name       string
				ProjectID  string
				Visibility string
			}{
				ID:   feed.ID,
				Name: feed.Name,
				ProjectID: func() string {
					if feed.Project != nil {
						return feed.Project.ID
					}
					return ""
				}(),
				Visibility: visibility,
			})
		}
	}

	if len(nugetFeeds) == 0 {
		panic("❌ No NuGet-compatible feeds found.")
	}

	for _, feed := range nugetFeeds {
		if targetFeed == "" || strings.EqualFold(feed.Name, targetFeed) {
			fmt.Printf("\n✅ Selected NuGet feed: %s (ID: %s)\n", feed.Name, feed.ID)
			return feed.ID, feed.ProjectID
		}
	}

	panic(fmt.Sprintf("❌ NuGet feed '%s' not found in organization '%s'.", targetFeed, org))
}

func FetchPackages(feedUrl string, filterPkg string, filterVersion string) []Package {
	if PAT == "" {
		panic("❌ AZURE_PAT is not set!")
	}

	fmt.Printf("🔐 Using AZURE_PAT starting with: %s...\n", PAT[:5])
	fmt.Printf("🔍 Fetching packages using feed discovery via: %s\n", feedUrl)

	org := extractOrg(feedUrl)
	feedID, projectID := getFeedID(org)

	pathPrefix := org
	if projectID != "" {
		pathPrefix = fmt.Sprintf("%s/%s", org, projectID)
	}
	cachedOnly := !includeUpstreamPackages()
	apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/Packaging/Feeds/%s/Packages?api-version=6.0-preview.1", pathPrefix, feedID)
	if cachedOnly {
		apiUrl += "&isCached=true"
	}
	fmt.Printf("📦 Fetching package list from: %s\n", apiUrl)

	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+PAT)))
	req.Header.Set("Accept", "*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Package list query failed (%d): %s", resp.StatusCode, body))
	}

	var pkgList struct {
		Value []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"value"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &pkgList)

	regex := os.Getenv("AZURE_PACKAGE_FILTER")
	var pkgPattern *regexp.Regexp
	if regex != "" {
		var err error
		pkgPattern, err = regexp.Compile(regex)
		if err != nil {
			panic(fmt.Sprintf("❌ Invalid AZURE_PACKAGE_FILTER regex: %s", err))
		}
	}

	var pkgs []Package
	for _, item := range pkgList.Value {
		if filterPkg != "" && !strings.EqualFold(item.Name, filterPkg) {
			continue
		}

		if pkgPattern != nil && !pkgPattern.MatchString(item.Name) {
			continue
		}

		protocolType := getPackageProtocolType(pathPrefix, feedID, item.ID)
		if protocolType != "nuget" {
			if filterPkg == "" {
				fmt.Printf("⚠️  Skipping %s (protocol: %s)\n", item.Name, protocolType)
			}
			continue
		}

		fmt.Printf("📌 Found package: %s\n", item.Name)
		versions := fetchPackageVersions(pathPrefix, feedID, item.ID)
		if filterVersion != "" {
			found := false
			for _, v := range versions {
				if strings.EqualFold(v, filterVersion) {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("⚠️  Version %s not found in package %s. Skipping.\n", filterVersion, item.Name)
				continue
			}
			versions = []string{filterVersion}
		}

		pkgs = append(pkgs, Package{
			Name:     item.Name,
			Versions: versions,
		})

		if filterPkg != "" {
			break
		}
	}

	return pkgs
}

func fetchPackageVersions(pathPrefix, feedID, packageID string) []string {
	apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/Packaging/Feeds/%s/Packages/%s/Versions?api-version=6.0-preview.1", pathPrefix, feedID, packageID)
	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+PAT)))
	req.Header.Set("Accept", "*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Version fetch failed (%d): %s", resp.StatusCode, body))
	}

	var versionList struct {
		Value []struct {
			Version string `json:"version"`
		} `json:"value"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &versionList)

	var versions []string
	for _, v := range versionList.Value {
		versions = append(versions, v.Version)
	}

	return versions
}

func DownloadPackage(feedUrl, name, version string) string {
	org := extractOrg(feedUrl)
	feedID, projectID := getFeedID(org)

	pathPrefix := org
	if projectID != "" {
		pathPrefix = fmt.Sprintf("%s/%s", org, projectID)
	}

	url := fmt.Sprintf("https://pkgs.dev.azure.com/%s/_apis/packaging/feeds/%s/nuget/packages/%s/versions/%s/content?api-version=6.0-preview.1",
		pathPrefix, feedID, name, version)

	fmt.Printf("⬇️  Downloading %s@%s from Azure...\n", name, version)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+PAT)))
	req.Header.Set("Accept", "*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ HTTP request error while downloading %s@%s: %v\n", name, version, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("⚠️  Skipping %s@%s — Download failed (HTTP %d): %s\n", name, version, resp.StatusCode, body)
		return ""
	}

	safeVersion := regexp.MustCompile(`[\/:"*?<>|]`).ReplaceAllString(version, "-")
	fileName := fmt.Sprintf("%s.%s.nupkg", name, safeVersion)

	out, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("❌ Failed to create file %s: %v\n", fileName, err)
		return ""
	}
	defer out.Close()
	io.Copy(out, resp.Body)

	return fileName
}

func extractOrg(feedUrl string) string {
	parsed, err := url.Parse(feedUrl)
	if err != nil {
		panic("❌ Invalid feed URL")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 1 {
		panic("❌ Could not extract organization from URL")
	}

	return parts[0]
}

func includeUpstreamPackages() bool {
	val := strings.ToLower(os.Getenv("AZURE_INCLUDE_UPSTREAM"))
	return val == "true" || val == "1" || val == "yes"
}

func getPackageProtocolType(pathPrefix, feedID, packageID string) string {
	apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/packaging/feeds/%s/packages/%s?api-version=6.0-preview.1",
		pathPrefix, feedID, packageID)

	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+PAT)))
	req.Header.Set("Accept", "*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Failed to get package type (%d): %s", resp.StatusCode, body))
	}

	var result struct {
		ProtocolType string `json:"protocolType"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	return strings.ToLower(result.ProtocolType)
}
