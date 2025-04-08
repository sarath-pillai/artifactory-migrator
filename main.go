package main

import (
    "fmt"
    "os"

    "nuget-migrator/internal/azure"
    "nuget-migrator/internal/github"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: nuget-migrator <azure_feed_url>")
        os.Exit(1)
    }

    feedUrl := os.Args[1]
    pkgs := azure.FetchPackages(feedUrl)

    for _, pkg := range pkgs {
        fmt.Printf("📦 %s\n", pkg.Name)
        for _, version := range pkg.Versions {
            fmt.Printf("  └─ %s\n", version)
            file := azure.DownloadPackage(feedUrl, pkg.Name, version)
            github.PushToGitHub(file)

            // Clean up local file
            err := os.Remove(file)
            if err != nil {
                fmt.Printf("⚠️  Failed to delete %s: %v\n", file, err)
            } else {
                fmt.Printf("🧹 Deleted local file: %s\n", file)
            }
        }
    }
}
