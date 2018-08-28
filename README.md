# listchallenges-exporter

Simple exporter for lists found on listchallenges.com, using Go and Agouti.

## Getting started

### Installing chromedriver

System-wide install:
```bash
brew tap homebrew/cask
brew cask install chromedriver
chromedriver --version # should work!
```

Local install:
Download the latest chromedriver: http://chromedriver.chromium.org/
Put it in your homedir: ```~/chromedriver```


### Installing dependencies
```bash
# golang < 1.11
go get github.com/sclevine/agouti
# golang >= 1.11:
go get    # all dependencies from go.mod will be installed
```

### Developing

Running/building code:
```bash
# During development
go run exporter.go

# Building final artifact
go build -o bin/exporter exporter.go
```
