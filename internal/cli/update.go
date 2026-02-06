package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dulait/grit/internal/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update grit to the latest version",
	Long:  "Check for a newer release on GitHub and replace the current binary.",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	u := updater.New(Version)

	if u.IsDev() {
		fmt.Println("You are running a dev build — the current version cannot be determined.")
		if !confirmAction("Update to the latest release anyway?") {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	fmt.Println("Checking for updates...")

	release, err := u.FetchLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")

	if !u.IsDev() && u.IsUpToDate(release) {
		fmt.Printf("Already up to date (v%s).\n", latestVer)
		return nil
	}

	if !u.IsDev() {
		fmt.Printf("New version available: v%s (current: v%s)\n", latestVer, strings.TrimPrefix(Version, "v"))
	} else {
		fmt.Printf("Latest version: v%s\n", latestVer)
	}

	asset, err := u.FindAsset(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	archivePath, err := u.DownloadAsset(ctx, asset)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	fmt.Println("Applying update...")

	if err := u.Apply(archivePath); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	fmt.Printf("Successfully updated to v%s!\n", latestVer)
	return nil
}
