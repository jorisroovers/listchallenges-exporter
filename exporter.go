package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jawher/mow.cli"
	"github.com/sclevine/agouti"
)

type PageCrawler struct {
	PageDriver *agouti.WebDriver
}

// Datastructures to encapsulate list items that are part of a list challenge
type ListChallenge struct {
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	ListItems []ListItem `json:"items"`
}

type ListItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func main() {
	cp := cli.App("exporter", "Simple exporter for listchallenges.com")
	cp.Version("v version", "exporter 0.1")
	listURL := cp.StringOpt("list-url", "", "URL of the list on listchallenges.com")

	cp.Action = func() {
		fmt.Fprintln(os.Stderr, "Starting chromedriver...")

		chromeOptions := agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu"})
		agoutiDriver := agouti.ChromeDriver(chromeOptions)
		crawler := PageCrawler{agoutiDriver}

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
		// Note that this is required to avoid dangling chromedriver processes
		defer func() {
			fmt.Fprint(os.Stderr, "Stopping chromedriver...")
			agoutiDriver.Stop()
			fmt.Fprintln(os.Stderr, "DONE")
		}()

		// Open new page in browser, navigate to list URL
		page, err := agoutiDriver.NewPage()
		if err != nil {
			log.Fatal(err)
		}
		page.Navigate(*listURL)

		// Scrape list name
		el := page.Find("#MainContent_panelListName h2")
		listName, _ := el.Text()
		fmt.Fprintf(os.Stderr, "Crawling list \"%s\" on %s\n", listName, *listURL)

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

// Walks all the pages in a list and collects the items on each page
func (crawler *PageCrawler) walkAndCollectList(page *agouti.Page, listURL string) []ListItem {
	pageIndex := 1
	var allItems []ListItem
	for true {
		nextPageURL := listURL + "/checklist/" + strconv.Itoa(pageIndex)

		page.Navigate(nextPageURL)
		actualURL, _ := page.URL()
		// listchallenges.com will redirect to page 1 when navigating to a page that is out of range
		// hence, we keep trying to navigate to the next page, until we get redirected (= next page url is different from actual page)
		if actualURL != nextPageURL {
			break
		}
		fmt.Fprintf(os.Stderr, "Crawling %s\n", nextPageURL)
		pageIndex += 1

		pageItems := crawler.collectItems(page)
		allItems = append(allItems, pageItems...)
	}
	return allItems
}

// Collects all list items on a single given page
func (*PageCrawler) collectItems(page *agouti.Page) []ListItem {
	items := page.All("#repeaterListItems .item-name")
	i := 0
	var listItems []ListItem
	for true {
		itemName, err := items.At(i).Text()
		if err != nil {
			break
		}
		pageURL, _ := page.URL()
		listItems = append(listItems, ListItem{itemName, pageURL})
		i += 1
	}
	return listItems
}
