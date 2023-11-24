// Package cdriver contains auxiliary function
package cdriver

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nicolasrod/chromedriver_autoinstall/internal/utils"
)

const KGV_NEW_JSON = "googlechromelabs.github.io/chrome-for-testing/known-good-versions-with-downloads.json"
const CD_OLD_URL = "chromedriver.storage.googleapis.com/"

type versions struct {
	Timestamp time.Time `json:"timestamp"`
	Versions  []struct {
		Version   string `json:"version"`
		Revision  string `json:"revision"`
		Downloads struct {
			ChromeDriver []struct {
				Platform string `json:"platform"`
				URL      string `json:"url"`
			} `json:"chromedriver"`
		} `json:"downloads"`
	} `json:"versions"`
}

func toint(v string) int {
	no, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}

	return no
}
func versionIsGTEQ(v1 string, v2 string) bool {
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")

	return toint(p1[0]) >= toint(p2[0]) && toint(p1[1]) >= toint(p2[1]) &&
		toint(p1[2]) >= toint(p2[2]) && toint(p1[3]) >= toint(p2[3])
}

func platformArch(chrome_version string) (string, string) {
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "amd64" {
			return "linux", "64"
		} else {
			return "linux", "32"
		}
	}

	if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" {
			return "win", "64"
		} else {
			return "win", "32"
		}
	}

	if runtime.GOOS == "darwin" {
		pf := "mac"
		arch := ""

		if runtime.GOARCH == "arm" {
			if chrome_version != "" && versionIsGTEQ(chrome_version, "115.0.5763.0") {
				arch = "-arm64"
			} else if chrome_version != "" && versionIsGTEQ(chrome_version, "106.0.5249.21") {
				arch = "64_m1"
			} else {
				arch = "_arm64"
			}
		} else if runtime.GOARCH == "i386" {
			if chrome_version != "" && chrome_version >= "115.0.5763.0" {
				arch = "-x64"
			} else {
				arch = "mac64"
			}
		} else {
			return "", ""
		}

		return pf, arch
	}

	return "", ""
}

// Returns the Download URL for a ChromeDriver binary compatible with Chrome version = chromeVersion
func GetChromeDriverURL(chromeVersion string, no_ssl bool) (string, error) {
	pf, arch := platformArch(chromeVersion)
	platform := pf + arch

	proto := "https://"
	if no_ssl {
		proto = "http://"
	}

	vChrome := strings.Split(chromeVersion, ".")
	vToTest := strings.Join(vChrome[0:3], ".")

	if vChrome[0] >= "115" {
		versions_url := proto + KGV_NEW_JSON

		data, err := utils.CurlContent(versions_url)
		if err != nil {
			return "", err
		}

		var f versions
		err = json.Unmarshal([]byte(data), &f)
		if err != nil {
			return "", err
		}

		var latest string

		for _, it := range f.Versions {
			tmp := strings.Join(strings.Split(it.Version, ".")[0:3], ".")

			if tmp != vToTest {
				continue
			}

			for _, url := range it.Downloads.ChromeDriver {
				if url.Platform == platform {
					latest = url.URL
				}
			}
		}

		return latest, nil
	} else {
		return proto + CD_OLD_URL + chromeVersion + "/chromedriver_" + platform + ".zip", nil
	}
}

// Gets the installed ChromeDriver version. If *binary* is empty, we assume the binary is
// in the current directory. If you want to specify a binary that is in the OS PATH,
// *binary* = "chromedriver"
func InstalledChromeDriverVersion(binary string) (string, error) {
	if binary == "" {
		binary = GetChromeDriverFilename()
	}

	cmd := exec.Command(binary, "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.Split(strings.TrimSpace(
		strings.ReplaceAll(string(out), "ChromeDriver", "")), " ")[0], nil
}

// Return is Chrome version and ChromeDriver version are compatible. If not, it is needed to
// download a new ChromeDriver matching the Chrome version
func NeedsUpdating(vChrome string, vChromeDriver string) bool {
	v1 := strings.Join(strings.Split(vChrome, ".")[0:3], ".")
	v2 := strings.Join(strings.Split(vChromeDriver, ".")[0:3], ".")

	return v1 != v2
}

// Get the installed Chrome version. Based on the OS, the path is inferred (hit and miss sometimes)
// if the parameter *path* is empty.
func InstalledChromeVersion(path string) (string, error) {
	if path == "" {
		if runtime.GOOS == "darwin" {
			path = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		} else if runtime.GOOS == "linux" {
			for _, it := range []string{"google-chrome", "google-chrome-stable",
				"google-chrome-beta", "google-chrome-dev", "chromium-browser", "chromium"} {

				cmd := exec.Command("which", it)
				out, err := cmd.Output()
				if err != nil {
					continue
				}
				path = strings.TrimSpace(string(out))
			}
		} else if runtime.GOOS == "windows" {
			gc_folder := "\\Google\\Chrome\\Application"
			path = ""

			for _, it := range []string{os.Getenv("ProgramFiles") + gc_folder,
				os.Getenv("ProgramFiles(x86)") + gc_folder,
				os.Getenv("ProgramW6432") + gc_folder} {

				if _, err := os.Stat(it); os.IsNotExist(err) {
					continue
				}

				path = strings.TrimSpace(it)
			}

			if path != "" {
				var numlatest uint64 = 0
				var latest_ver string = ""

				files, err := os.ReadDir(path)
				if err != nil {
					return "", err
				}

				for _, file := range files {
					if !file.IsDir() {
						continue
					}

					num, err := strconv.ParseUint(strings.ReplaceAll(file.Name(), ".", ""), 10, 0)
					if err == nil {
						if num > numlatest {
							latest_ver = file.Name()
						}
					}
				}

				path = path + "\\" + latest_ver
			}
		}
	}

	cmd := exec.Command(path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	v := strings.ReplaceAll(string(out), "Chromium", "")
	v = strings.ReplaceAll(v, "Google Chrome", "")
	return strings.TrimSpace(v), nil
}

// Tentative location for the ChromeDriver binary
func GetChromeDriverFilename() string {
	if runtime.GOOS == "windows" {
		return "chromedriver.exe"
	} else {
		return "./chromedriver"
	}
}
