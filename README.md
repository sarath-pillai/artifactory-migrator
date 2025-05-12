# artifactory-migrator

**Migrate NuGet packages from Azure DevOps (ADO) Artifacts to GitHub Packages.**

---

## 🚀 Features

- **Migrate all packages** and all their versions from Azure Artifacts to GitHub Packages
- **Migrate only a specific package** (all its versions)
- **Migrate only a specific version** of a package
- Cleans up local `.nupkg` files after push

---

## 🛠️ Requirements

Set these **environment variables** before running the tool:

- `AZURE_PAT` — Azure DevOps Personal Access Token (with read access to Azure Artifacts)
- `GH_TKN` — GitHub Personal Access Token (with `write:packages` scope)
- `GITHUB_USER` — GitHub username or organization (target for publishing)

---

## ⚡ Usage

```bash
artifactory-migrator [--package <name>] [--pkg-version <version>] <azure_feed_url>
```

- `--package <name>`: (Optional) Migrate only the named package (case-insensitive)
- `--pkg-version <version>`: (Optional, requires `--package`) Migrate only the specified version of the named package
- `<azure_feed_url>`: The Azure DevOps Artifacts feed URL (e.g., `https://pkgs.dev.azure.com/orgname`)

---

## 📦 Examples

### 1. Migrate all packages (all versions)

```bash
export AZURE_PAT=xxxx
export GH_TKN=yyyy
export GITHUB_USER=myuser

artifactory-migrator https://pkgs.dev.azure.com/orgname
```

### 2. Migrate all versions of a specific package

```bash
artifactory-migrator --package SampleNugetPackage https://pkgs.dev.azure.com/orgname
```

### 3. Migrate a specific version of a specific package

```bash
artifactory-migrator --package SampleNugetPackage --pkg-version 1.2.3 https://pkgs.dev.azure.com/orgname
```

### 4. Package Filtering with Regex

Match packages that start with a prefix:

```bash
export AZURE_PACKAGE_FILTER="^MyLib"
```
Matches: MyLib, MyLibrary.Utils, MyLib123

Match packages that end with a suffix:

```bash
export AZURE_PACKAGE_FILTER="Client$"
```
Matches: AzureClient, GitClient, Http.Client

Match packages that contain a specific word:

```bash
export AZURE_PACKAGE_FILTER="Analytics"
```
Matches: UserAnalytics, My.Analytics.Core, AnalyticsData

Match exact package name:
```bash
export AZURE_PACKAGE_FILTER="^Exact.Package.Name$"
```
Only matches Exact.Package.Name

Match packages with numeric suffix:

```bash
export AZURE_PACKAGE_FILTER=".*[0-9]+$"
```
Matches: MyLib2, Toolkit123, Package1

### 5. Using a specific feed in azure
`AZURE_FEED` Specifies the name of a specific Azure DevOps feed to use.
If unset, the tool will attempt to auto-discover the first NuGet-compatible feed available.

Use when your organization has multiple feeds.

Example: export AZURE_FEED=MyFeed

### 6. Upstream Control

`AZURE_INCLUDE_UPSTREAM`
Controls whether packages from upstream sources (e.g., NuGet.org, Maven Central) are included.

true, 1, or yes: include upstream/public packages.
Any other value or unset: only include packages directly uploaded to the feed.

Example: `export AZURE_INCLUDE_UPSTREAM=true`
---

## 📝 Notes

- **Flags must come before the feed URL** (this is a Go convention)
- Package name matching is **case-insensitive**
- If you specify `--pkg-version`, you **must** also specify `--package`
- Local `.nupkg` files are deleted after successful upload to GitHub Packages
- If you use AZURE_PACKAGE_FILTER and --package together, AZURE_PACKAGE_FILTER will be ignored.
- The tool outputs progress and errors to the console

---

## 🔗 Related

- [Azure Artifacts Documentation](https://learn.microsoft.com/en-us/azure/devops/artifacts/)
- [GitHub Packages Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-nuget-registry)

---

## 🏷️ Version

Current version: **v0.1.0**

---
