package azure

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

var PAT = os.Getenv("AZURE_PAT")

type Package struct {
    Name     string
    Versions []string
}

func getFeedID(org string) string {
    apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/packaging/feeds?api-version=6.0-preview.1", org)

    req, _ := http.NewRequest("GET", apiUrl, nil)
    req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(":" + PAT)))
    req.Header.Set("Accept", "*/*")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        panic(fmt.Sprintf("Failed to get feed ID (%d): %s", resp.StatusCode, body))
    }

    var result struct {
        Value []struct {
            ID string `json:"id"`
        } `json:"value"`
    }
    body, _ := io.ReadAll(resp.Body)
    json.Unmarshal(body, &result)

    if len(result.Value) == 0 {
        panic("No feeds found.")
    }

    return result.Value[0].ID
}

func FetchPackages(feedUrl string) []Package {
    if PAT == "" {
        panic("❌ AZURE_PAT is not set!")
    }

    fmt.Printf("🔐 Using AZURE_PAT starting with: %s...\n", PAT[:5])
    fmt.Printf("🔍 Fetching packages using feed discovery via: %s\n", feedUrl)

    org := extractOrg(feedUrl)
    feedID := getFeedID(org)

    apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/Packaging/Feeds/%s/Packages?api-version=6.0-preview.1", org, feedID)
    req, _ := http.NewRequest("GET", apiUrl, nil)
    req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(":" + PAT)))
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

    var pkgs []Package
    for _, item := range pkgList.Value {
        versions := fetchPackageVersions(org, feedID, item.ID)
        pkgs = append(pkgs, Package{
            Name:     item.Name,
            Versions: versions,
        })
    }

    return pkgs
}

func fetchPackageVersions(org, feedID, packageID string) []string {
    apiUrl := fmt.Sprintf("https://feeds.dev.azure.com/%s/_apis/Packaging/Feeds/%s/Packages/%s/Versions?api-version=6.0-preview.1", org, feedID, packageID)
    req, _ := http.NewRequest("GET", apiUrl, nil)
    req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(":" + PAT)))
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

func DownloadPackage(_, name, version string) string {
    fmt.Printf("📦 Skipping actual download of %s@%s\n", name, version)
    return fmt.Sprintf("%s.%s.nupkg", name, version)
}

func extractOrg(feedUrl string) string {
    parts := []rune(feedUrl)
    idx := 0
    for i := 0; i < len(parts); i++ {
        if string(parts[i:i+1]) == "/" {
            idx++
            if idx == 4 {
                return string(parts[8:i])
            }
        }
    }
    return "sp880706"
}
