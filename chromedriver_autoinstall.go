// Package chromedriver_autoinstall enables an easy way to check the installed version
// of Google Chrome/Chromium and installs, if it is not already, a compatible
// *chromedriver* binary.
package chromedriverautoinstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicolasrod/chromedriver_autoinstall/internal/cdriver"
	"github.com/nicolasrod/chromedriver_autoinstall/internal/utils"
)

// Checks and installs the *ChromeDriver* binary in *cdriver_fullpath*
func InstallChromeDriver(cdriver_fullpath string) error {
	if cdriver_fullpath == "" {
		return errors.New("chromedriver path cannot be empty")
	}

	vChrome, err := cdriver.InstalledChromeVersion("")
	if err != nil {
		return err
	}

	vChromeDriver, err := cdriver.InstalledChromeDriverVersion(cdriver_fullpath)
	if err != nil {
		vChromeDriver = "1.0.0.0"
	}

	if cdriver.NeedsUpdating(vChrome, vChromeDriver) {
		dl_url, err := cdriver.GetChromeDriverURL(vChrome, false)
		if err != nil {
			return err
		}

		tmpname, err := os.MkdirTemp(".", "cdrv")
		if err != nil {
			return err
		}

		zipfile, err := utils.DownloadTo(dl_url, tmpname)
		if err != nil {
			return err
		}

		bin := fmt.Sprintf("%s/%s/%s",
			tmpname,
			utils.RemoveExt(filepath.Base(zipfile)),
			cdriver.GetChromeDriverFilename())

		utils.Unzip(zipfile, tmpname)

		if err := os.Rename(bin, cdriver_fullpath); err != nil {
			return err
		}

		if err := os.RemoveAll(tmpname); err != nil {
			return err
		}

		if err := os.Chmod(cdriver_fullpath, 0777); err != nil {
			return err
		}
	}
	return nil
}
