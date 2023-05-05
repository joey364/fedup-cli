/*
Copyright © 2023 Joel Owusu-Ansah
*/
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Run command that installs the dependencies
func install(deps []string) {
	checkIfIsFedoraWorkstation()
	for _, v := range deps {
		cmd := exec.Command("sudo", "dnf", "install", v)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkIfIsFedoraWorkstation() {
	re := regexp.MustCompile(`(?i)fedora`)

	grepCmd := exec.Command("grep", "-iE", "^id=", "/etc/os-release")

	out, err := grepCmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}

	stringOut := string(out)
	stringOut = strings.Split(stringOut, "=")[1]

	if isFedora := re.MatchString(stringOut); !isFedora {
		log.Fatal("This tool is for fedora workstations only!")
	}
}

func getCurrentReleaseVersion() int {

	if _, err := os.Stat("/etc/os-release"); errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	checkIfIsFedoraWorkstation()

	grepCmd := exec.Command("grep", "-i", "version_id", "/etc/os-release")
	cutCmd := exec.Command("cut", "-d", "=", "-f", "2")

	r, w := io.Pipe()

	grepCmd.Stdout = w
	cutCmd.Stdin = r

	var result bytes.Buffer
	cutCmd.Stdout = &result

	grepCmd.Start()
	cutCmd.Start()
	grepCmd.Wait()
	w.Close()
	cutCmd.Wait()

	currentVer, _ := strconv.Atoi(strings.TrimSpace(result.String()))
	return currentVer
}

func getLatestReleaseVersion() int {

	cmdString := "curl -s \"https://getfedora.org/releases.json\" | jq -r '[.[].version | select(test(\"^\\\\d*$\"))] | max'"

	cmd := exec.Command("/bin/bash", "-c", cmdString)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}

	latestVersion, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return latestVersion
}

func upgradeCurrentSystem() {
	upgradeCmd := exec.Command("sudo", "dnf", "upgrade", "--refresh", "-y")

	upgradeCmd.Stdout = os.Stdout
	upgradeCmd.Stderr = os.Stderr

	err := upgradeCmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func installUpgradeDependencies() {
	cmd := exec.Command("sudo", "dnf", "install", "dnf-plugin-system-upgrade", "-y")

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func installUpgradedPackages(version int) {
	release := fmt.Sprintf("--releasever=%d", version)
	cmd := exec.Command("sudo", "dnf", "system-upgrade", "download", release, "-y")

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func triggerUpgrade() {
	cmd := exec.Command("sudo", "dnf", "system-upgrade", "reboot", "-y")

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fedup",
	Short: "Upgrade fedora workstation",
	Long:  `A cli to upgrade your fedora workstation installation.`,
	Run: func(cmd *cobra.Command, args []string) {

		var deps = []string{"jq", "curl"}
		install(deps)
		currentVersion := getCurrentReleaseVersion()
		latestVersion := getLatestReleaseVersion()

		if currentVersion == latestVersion {
			fmt.Println("you are on the latest release of fedora workstation ")
			os.Exit(0)
		}

		upgradeCurrentSystem()
		installUpgradeDependencies()
		installUpgradedPackages(latestVersion)
		triggerUpgrade()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
