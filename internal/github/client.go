package github

import (
    "fmt"
    "os"
    "os/exec"
)

var (
    GitHubPAT  = os.Getenv("GITHUB_TOKEN")
    GitHubUser = os.Getenv("GITHUB_USER")
)

func PushToGitHub(nupkgPath string) {
    if GitHubPAT == "" || GitHubUser == "" {
        fmt.Println("❌ GITHUB_TOKEN and GITHUB_USER must be set")
        return
    }

    githubSource := fmt.Sprintf("https://nuget.pkg.github.com/%s/index.json", GitHubUser)

    cmd := exec.Command("dotnet", "nuget", "push", nupkgPath,
        "--source", githubSource,
        "--api-key", GitHubPAT,
        "--skip-duplicate",
    )

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        fmt.Printf("❌ Failed to push %s: %v\n", nupkgPath, err)
    }
}
