// main.go
package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
  "sync"

	"github.com/PuerkitoBio/goquery"
)

type XmlSettings struct {
	XmlName            xml.Name
	WebsiteUrl         string
	BaseWebsiteUrl     string
	FileTypeToDownload string
	DownloadLocation   string
}

func main() {
	xmlSettings := getSettings()

	fmt.Println("Parsing HTML from", xmlSettings.WebsiteUrl)

	linkScrape(xmlSettings.WebsiteUrl, xmlSettings.BaseWebsiteUrl, xmlSettings.FileTypeToDownload, xmlSettings.DownloadLocation)
}

func linkScrape(websiteUrl string, baseWebsiteUrl string, fileType string, downloadLocation string) {
	doc, err := goquery.NewDocument(websiteUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var wg sync.WaitGroup

	// use CSS selector found with the browser inspector
	// for each, use index and item
	doc.Find("body a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		linkHref := strings.TrimSpace(link)

		if strings.HasSuffix(linkHref, fileType) {
			if strings.HasPrefix(linkHref, "/") {
				linkHref = baseWebsiteUrl + linkHref
			} else if !strings.HasPrefix(linkHref, websiteUrl) {
				linkHref = websiteUrl + "/" + linkHref
			}

			fileName := path.Join(downloadLocation, linkHref[strings.LastIndex(linkHref, "/")+1:])

			//don't download the file if it already exists
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				fmt.Println("Downloading", linkHref)

        wg.Add(1)
				go downloadFile(fileName, linkHref, &wg)
			}
		}
	})

  fmt.Println("Please wait, as each file completes you will be notified.")
	wg.Wait()

	fmt.Println("Done.")
}

func downloadFile(filepath string, url string, wg *sync.WaitGroup) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		fmt.Println("Failed download", filepath)
		fmt.Println(err)
		wg.Done()
		return
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed download", filepath)
		fmt.Println(err)
		wg.Done()
		return
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Failed download", filepath)
		fmt.Println(err)
		wg.Done()
		return
	}

  fmt.Println("Downloaded", filepath)
	wg.Done()
}

func getSettings() XmlSettings {
	settingsFilePath, err := filepath.Abs("settings.xml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Open the Settings.xml file
	file, err := os.Open(settingsFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	xmlSettings, err := readSettingsFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return xmlSettings
}

func readSettingsFile(reader io.Reader) (XmlSettings, error) {
	var xmlSettings XmlSettings
	if err := xml.NewDecoder(reader).Decode(&xmlSettings); err != nil {
		return xmlSettings, err
	}

	return xmlSettings, nil
}
