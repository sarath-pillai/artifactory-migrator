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
  artifactory-migrator [--package <name>] [--pkg-version <version>] <azure_feed_url>

Environment variables:
  AZURE_PAT       Azure DevOps personal access token
  GITHUB_TOKEN    GitHub personal access token
  GITHUB_USER     GitHub username or org

Examples:
  $ export AZURE_PAT=xxxx
  $ export GITHUB_TOKEN=yyyy
  $ export GITHUB_USER=myuser

  # Migrate everything
  $ artifactory-migrator https://pkgs.dev.azure.com/orgname

  # Migrate all versions of one package
  $ artifactory-migrator --package SampleNugetPackage https://pkgs.dev.azure.com/orgname

  # Migrate a specific version of a package
  $ artifactory-migrator --package SampleNugetPackage --pkg-version 1.0.0 https://pkgs.dev.azure.com/orgname
`)
}

func main() {
	showVersion := flag.Bool("version", false, "Print tool version")
	help := flag.Bool("help", false, "Show help message")
	pkgName := flag.String("package", "", "Specific NuGet package name to migrate")
	pkgVersion := flag.String("pkg-version", "", "Specific version to migrate (requires --package)")

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
		fmt.Println("❌ You must provide the Azure feed URL.")
		flag.Usage()
		os.Exit(1)
	}

	if *pkgVersion != "" && *pkgName == "" {
		fmt.Println("❌ --pkg-version requires --package to be set.")
		os.Exit(1)
	}

	feedUrl := flag.Arg(0)
	pkgs := azure.FetchPackages(feedUrl, *pkgName, *pkgVersion)

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

