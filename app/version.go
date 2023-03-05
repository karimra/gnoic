package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	gitURL  = ""
)

var downloadURL = "https://github.com/karimra/gnoic/raw/main/install.sh"

func (a *App) RunEVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("version : %s\n", version)
	fmt.Printf(" commit : %s\n", commit)
	fmt.Printf("   date : %s\n", date)
	fmt.Printf(" gitURL : %s\n", gitURL)
	fmt.Printf("   docs : https://gnoic.kmrd.dev\n")
	return nil
}

func (a *App) VersionUpgradeRun(cmd *cobra.Command, args []string) error {
	f, err := os.CreateTemp("", "gnmic")
	defer os.Remove(f.Name())
	if err != nil {
		return err
	}
	err = downloadFile(cmd.Context(), downloadURL, f)
	if err != nil {
		return err
	}

	var c *exec.Cmd
	switch a.Config.LocalFlags.UpgradeUsePkg {
	case true:
		c = exec.Command("bash", f.Name(), "--use-pkg")
	case false:
		c = exec.Command("bash", f.Name())
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}
	return nil
}

// downloadFile will download a file from a URL and write its content to a file
func downloadFile(ctx context.Context, url string, file *os.File) error {
	client := http.Client{Timeout: 30 * time.Second}
	// Get the data
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) InitVersionUpgradeFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&a.Config.LocalFlags.UpgradeUsePkg, "use-pkg", false, "upgrade using package")
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}
