# ChromeDriver_AutoInstall

*Porting of https://github.com/yeongbin-jo/python-chromedriver-autoinstaller to GoLang*

# Example

```
package main

import (
	"log"

	chromedriverautoinstall "github.com/nicolasrod/chromedriver_autoinstall"
)

func main() {
    // Check current Chrome version and if chromedriver does not exists or the version y older
    // than the version of Chrome, a new compatible chromedriver will be downloaded as
	// ./chromedriver
	err := chromedriverautoinstall.InstallChromeDriver("./chromedriver")
	if err != nil {
		log.Fatal(err)
	}
}
```
