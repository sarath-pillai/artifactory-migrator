package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	args := []string{
		"nuget", "push", nupkgPath,
		"--source", githubSource,
		"--api-key", GitHubPAT,
		"--skip-duplicate",
	}

	fmt.Printf("🚀 Running: dotnet %s\n", strings.Join(args, " "))

	cmd := exec.Command("dotnet", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("❌ dotnet exited with code %d\n", exitErr.ExitCode())
		}
		fmt.Printf("❌ Failed to push %s: %v\n", nupkgPath, err)
	} else {
		fmt.Printf("✅ Successfully pushed: %s\n", nupkgPath)
	}
}
