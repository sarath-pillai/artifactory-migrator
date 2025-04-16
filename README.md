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
- `GITHUB_TOKEN` — GitHub Personal Access Token (with `write:packages` scope)
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
export GITHUB_TOKEN=yyyy
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

---

## 📝 Notes

- **Flags must come before the feed URL** (this is a Go convention)
- Package name matching is **case-insensitive**
- If you specify `--pkg-version`, you **must** also specify `--package`
- Local `.nupkg` files are deleted after successful upload to GitHub Packages
- The tool outputs progress and errors to the console

---

## 🔗 Related

- [Azure Artifacts Documentation](https://learn.microsoft.com/en-us/azure/devops/artifacts/)
- [GitHub Packages Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-nuget-registry)

---

## 🏷️ Version

Current version: **v0.1.0**

---

## 🧑‍💻 Contributing

Pull requests and issues are welcome!  
For feature requests or bug reports, please open an [issue](https://github.com/your-repo/issues).

---

## 📄 License

MIT License (see [LICENSE](LICENSE) for details)

