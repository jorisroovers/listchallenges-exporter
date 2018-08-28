package main

import (
	"fmt"
	"log"

	"github.com/sclevine/agouti"
)

func main() {
	fmt.Println("Starting chromedriver...")

	chromeOptions := agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu"})
	agoutiDriver := agouti.ChromeDriver(chromeOptions)

	// agouti.ChromeDriver() assumes the 'chromedriver' binary is available systemwide:
	// https://github.com/sclevine/agouti/blob/5e39ce136dd6bafe76094887339bed399ad23a9d/agouti.go#L36-L42
	// If that's not the case, we need to be able to specify the path of the chromedriver, which requires a create a
	// custom NewWebDriver object.
	// usr, _ := user.Current()
	// homedir := usr.HomeDir
	// command := []string{filepath.Join(homedir + "/chromedriver"), "--port={{.Port}}"}
	// agoutiDriver := agouti.NewWebDriver("http://{{.Address}}", command, chromeOptions)

	agoutiDriver.Start()
	// Use Go's defer to call agoutiDriver.Stop() after exiting this function
	// Note that this is required to not have dangling chromedriver processes
	defer agoutiDriver.Stop()

	page, err := agoutiDriver.NewPage()
	if err != nil {
		log.Fatal(err)
	}
	page.Navigate("https://www.listchallenges.com/reddit-top-250-movies")
	el := page.Find("#MainContent_panelListName h2")
	listName, _ := el.Text()
	fmt.Println(listName)
}
