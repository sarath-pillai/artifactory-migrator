package main

import (
	"flag"
	"fmt"
	"os"

	"nuget-migrator/internal/azure"
	"nuget-migrator/internal/github"
)

const version = "0.1.0"

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `
Artifactory Migrator - Migrate NuGet packages from Azure DevOps to GitHub Packages

Usage:
  artifactory-migrator <azure_feed_url>

Environment variables:
  AZURE_PAT       Azure DevOps personal access token
  GITHUB_TOKEN    GitHub personal access token
  GITHUB_USER     GitHub username or org

Example:
  $ export AZURE_PAT=xxxx
  $ export GITHUB_TOKEN=yyyy
  $ export GITHUB_USER=myuser
  $ artifactory-migrator https://pkgs.dev.azure.com/orgname

`)
}

func main() {
	showVersion := flag.Bool("version", false, "Print version")
	help := flag.Bool("help", false, "Show help message")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *showVersion {
		fmt.Println("Artifactory Migrator version", version)
		return
	}

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	feedUrl := flag.Arg(0)
	pkgs := azure.FetchPackages(feedUrl)

	for _, pkg := range pkgs {
		fmt.Printf("📦 %s\n", pkg.Name)
		for _, version := range pkg.Versions {
			fmt.Printf("  └─ %s\n", version)
			file := azure.DownloadPackage(feedUrl, pkg.Name, version)
			github.PushToGitHub(file)
			if err := os.Remove(file); err != nil {
				fmt.Printf("⚠️  Failed to delete %s: %v\n", file, err)
			} else {
				fmt.Printf("🧹 Deleted local file: %s\n", file)
			}
		}
	}
}

