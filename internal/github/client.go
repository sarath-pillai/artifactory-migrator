package github

import (
    "fmt"
    "os"
    "os/exec"
)

var (
    GithubPAT    = os.Getenv("GITHUB_TOKEN")
    GithubUser   = os.Getenv("GITHUB_USER")
    GithubSource = "https://nuget.pkg.github.com/" + GithubUser + "/index.json"
)

func PushToGitHub(nupkgPath string) {
    if GithubPAT == "" || GithubUser == "" {
        fmt.Println("❌ GITHUB_TOKEN and GITHUB_USER environment variables must be set.")
        return
    }

    cmd := exec.Command("dotnet", "nuget", "push", nupkgPath,
        "--source", GithubSource,
        "--api-key", GithubPAT,
        "--skip-duplicate",
    )

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        fmt.Printf("❌ Failed to push %s: %v\n", nupkgPath, err)
    }
}
