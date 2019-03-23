GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=Ventus

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

deps:
	go get -u "github.com/PuerkitoBio/goquery"
	go get -u "github.com/kelseyhightower/envconfig"
	go get -u "github.com/nlopes/slack"
	go get -u "github.com/Kotaro7750/Ventus/wind"

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

