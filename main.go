package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gregdel/pushover"
	"github.com/pkg/errors"
)

const userAgent = "golang:webpage-monitor:v0.0.0 (by /u/Poo-Poo-Kitty)"

var (
	url            string
	element        string
	pushoverKey    string
	pushoverTarget string
)

func init() {
	flag.StringVar(&url, "url", "", "url of webpage")
	flag.StringVar(&element, "element", "", "CSS selector for element")
	flag.StringVar(&pushoverKey, "pushoverkey", "", "pushover API key")
	flag.StringVar(&pushoverTarget, "pushovertarget", "", "pushover user to target")
	flag.Parse()
}

func main() {
	r, err := pageHTML(url)
	if err != nil {
		log.Printf("Error while getting page HTML: %v", err)
		return
	}
	defer r.Close()

	val, err := elementValue(r, element)
	if err != nil {
		log.Printf("Error while getting element value: %v", err)
		sendNotification(pushoverKey, pushoverTarget, "Webpage Monitor error", err.Error())
		return
	}

	log.Printf("Element %q = %s", element, val)
	err = sendNotification(pushoverKey, pushoverTarget, "Webpage Monitor change", val)
	if err != nil {
		log.Printf("Error while sending notification: %v", err)
	}
}

func pageHTML(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request to %q", url)
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make request to %q", url)
	}
	return resp.Body, nil
}

func elementValue(r io.Reader, element string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", errors.Wrap(err, "failed to create document")
	}

	selection := doc.Find(element).First()
	if selection.Length() == 0 {
		return "", errors.Errorf("unable to find %q in document", element)
	}

	return selection.Text(), nil
}

func sendNotification(pushoverKey, pushoverTarget, title, body string) error {
	app := pushover.New(pushoverKey)
	recipient := pushover.NewRecipient(pushoverTarget)
	message := pushover.NewMessageWithTitle(body, title)
	_, err := app.SendMessage(message, recipient)
	return err
}
