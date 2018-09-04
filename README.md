# listchallenges-exporter

Simple exporter for lists found on listchallenges.com, using Go and [Agouti](https://agouti.org/).

Outputs json, for easy further manipulation with e.g. [jq](https://stedolan.github.io/jq/).

# Getting started

## Installing chromedriver

System-wide install:
```bash
brew tap homebrew/cask
brew cask install chromedriver
chromedriver --version # should work!
```

Local install:
Download the latest chromedriver: http://chromedriver.chromium.org/
Put it in your homedir: ```~/chromedriver```

## Running

```bash
go run exporter.go --list-url https://www.listchallenges.com/the-european-capitals-of-culture

# Use --debug for some output in between
go run exporter.go --debug --list-url https://www.listchallenges.com/the-european-capitals-of-culture

# Just get all the items using jq
go run exporter.go --list-url https://www.listchallenges.com/reddit-top-250-movies | jq -r ".items[].name"

```

The tool can also fetch completion of lists by logging in:

```bash
# Set username and password
export LC_USERNAME=""; export LC_PASSWORD="";
go run exporter.go --debug --username "$LC_USERNAME" --password "$LC_PASSWORD"

# Print both name and whether the item is checked or not
go run exporter.go --debug --list-url https://www.listchallenges.com/reddit-top-250-movies --username "$LC_USERNAME" --password "$LC_PASSWORD" | jq -r '.items[] | "\(.name), \(.checked)"'
```

## Developing

### Installing dependencies
```bash
# golang < 1.11
go get github.com/sclevine/agouti
# golang >= 1.11:
go get    # all dependencies from go.mod will be installed
```

### Getting your hands dirty
Running/building code:
```bash
# During development
go run exporter.go

# Building final artifact
go build -o bin/exporter exporter.go
```

# TODO
- Use of proper logger and ```--debug``` mode to be able to supress verbose output by default
- Code clean up
- Support for scraping list completion by logging into account
- Parallel fetching of pages