# artifactory-migrator
Migrate Artifactory Packages from ADO to Github Packages



Artifactory Migrator - Migrate NuGet packages from Azure DevOps to GitHub Packages

Usage:
  artifactory-migrator <azure_feed_url>

Environment variables:
```bash
  AZURE_PAT       Azure DevOps personal access token
  GITHUB_TOKEN    GitHub personal access token
  GITHUB_USER     GitHub username or org
```

Example:

```bash
  $ export AZURE_PAT=xxxx
  $ export GITHUB_TOKEN=yyyy
  $ export GITHUB_USER=myuser
  $ artifactory-migrator https://pkgs.dev.azure.com/orgname
```
