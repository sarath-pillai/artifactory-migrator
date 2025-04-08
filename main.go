package main

import (
    "fmt"
    "os"

    "nuget-migrator/internal/azure"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: nuget-migrator <any_azure_devops_url_for_org>")
        os.Exit(1)
    }

    feedUrl := os.Args[1]
    pkgs := azure.FetchPackages(feedUrl)

    for _, pkg := range pkgs {
        fmt.Printf("📦 %s\n", pkg.Name)
        for _, version := range pkg.Versions {
            fmt.Printf("  └─ %s\n", version)
        }
    }
}
