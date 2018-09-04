package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jawher/mow.cli"
	"github.com/sclevine/agouti"
)

type PageCrawler struct {
	PageDriver *agouti.WebDriver
	LoggedIn   bool
}

// Datastructures to encapsulate list items that are part of a list challenge
type ListChallenge struct {
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	ListItems []ListItem `json:"items"`
}

type ListItem struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	Checked bool   `json:"checked"`
}

func main() {
	cp := cli.App("exporter", "Simple exporter for listchallenges.com")
	cp.Version("v version", "exporter 0.1")

	// CLI args
	debug := cp.BoolOpt("debug", false, "Enable debug mode")
	listURL := cp.StringOpt("list-url", "", "URL of the list on listchallenges.com")
	username := cp.StringOpt("username", "", "Username to login to listchallenges.com")
	password := cp.StringOpt("password", "", "Password to login to listchallenges.com")

	cp.Action = func() {

		// setup logger if debug is enabled, disable otherwise
		if *debug {
			log.SetOutput(os.Stderr)
			log.Println("Debug mode enabled")
		} else {
			// as per https://stackoverflow.com/a/34457930/381010
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard)
		}

		log.Println("Starting chromedriver...")

		chromeOptions := agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu"})
		agoutiDriver := agouti.ChromeDriver(chromeOptions, agouti.Timeout(10))

		// agouti.ChromeDriver() assumes the 'chromedriver' binary is available systemwide:
		// https://github.com/sclevine/agouti/blob/5e39ce136dd6bafe76094887339bed399ad23a9d/agouti.go#L36-L42
		// If that's not the case, we need to be able to specify the path of the chromedriver, which requires a create a
		// custom NewWebDriver object.
		// usr, _ := user.Current()
		// homedir := usr.HomeDir
		// command := []string{filepath.Join(homedir + "/chromedriver"), "--port={{.Port}}"}
		// agoutiDriver := agouti.NewWebDriver("http://{{.Address}}", command, chromeOptions)
		agoutiDriver.Start()

		crawler := PageCrawler{agoutiDriver, false}
		// Use Go's defer to call agoutiDriver.Stop() after exiting this function
		// Note that this is required to avoid dangling chromedriver processes
		defer func() {
			log.Println("Stopping chromedriver...")
			agoutiDriver.Stop()
			log.Println("DONE")
		}()

		// Open new page in browser, navigate to list URL
		page, _ := agoutiDriver.NewPage()

		// Login (optional) and navigate to listURL
		if *username != "" && *password != "" {
			crawler.login(page, *username, *password)
		}
		page.Navigate(*listURL)

		// Scrape list name
		el := page.Find("#MainContent_panelListName h2")
		listName, _ := el.Text()
		log.Printf("Crawling list \"%s\" on %s\n", listName, *listURL)

		// Walk all the pages in the list and collect their items
		items := crawler.walkAndCollectList(page, *listURL)
		listChallenge := ListChallenge{listName, *listURL, items}

		// Marshall result to json and print to stdout
		jsonBytes, err := json.Marshal(listChallenge)
		if err != nil {
			log.Fatal("Something went wrong while marshalling to json")
		}
		fmt.Println(string(jsonBytes))
	}

	cp.Run(os.Args)

}

func (crawler *PageCrawler) login(page *agouti.Page, username string, password string) {
	// Navigate to login page, populate username/password and click "Login"
	log.Println("Logging in...")
	page.Navigate("https://www.listchallenges.com/login-email")
	page.Find("#MainContent_textBoxEmailLogIn").Fill(username)
	page.Find("#MainContent_textBoxPasswordLogIn").Fill(password)
	page.Find("#MainContent_buttonLogIn").Click()

	log.Println("Waiting for login to occur...")
	time.Sleep(1 * time.Second)

	// Login is successful if the we get redirected to the profile page
	pageURL, _ := page.URL()
	if strings.HasPrefix(pageURL, "https://www.listchallenges.com/profile") {
		log.Println("Login successful!")
		crawler.LoggedIn = true
	} else {
		log.Println("Login failed :(")
		crawler.LoggedIn = false
	}

}

// Walks all the pages in a list and collects the items on each page
func (crawler *PageCrawler) walkAndCollectList(page *agouti.Page, listURL string) []ListItem {
	var allItems []ListItem

	// determine page count
	pageCount, err := page.All("#pagerChecklist li").Count()
	if err != nil {
		log.Fatal("Unable to determine page count", err)
	}
	// substract 2 from count to account for "Prev" and "Next" buttons
	// Use math.Max() to deal with single page lists
	pageCount = int(math.Max(1, float64(pageCount-2)))
	log.Printf("Discovered %d page(s) in list", pageCount)

	// Navigate to each page and collect items on page
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {

		// messages := make(chan string, 1)
		nextPageURL := listURL + "/checklist/" + strconv.Itoa(pageIndex)

		log.Println("Crawling", nextPageURL)
		page.Navigate(nextPageURL)

		pageItems := crawler.collectItems(page)
		allItems = append(allItems, pageItems...)
	}
	return allItems
}

// Collects all list items on a single given page
func (crawler *PageCrawler) collectItems(page *agouti.Page) []ListItem {
	items := page.All("#repeaterListItems .list-item")
	itemCount, err := items.Count()
	if err != nil {
		log.Fatal("Cannot determine number of items on page")
	}

	var listItems []ListItem
	for i := 0; i < itemCount; i++ {
		item := items.At(i)
		itemClasses, _ := item.Attribute("class")
		itemName, _ := item.First(".item-name").Text()
		pageURL, _ := page.URL()
		var checked bool
		if crawler.LoggedIn {
			checked = strings.Contains(itemClasses, "checked")
		} else {
			checked = false
		}
		listItems = append(listItems, ListItem{itemName, pageURL, checked})
	}
	return listItems
}
