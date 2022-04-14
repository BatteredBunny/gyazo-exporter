package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

type ImageAPIResponse struct {
	URL     string `json:"url"`
	ImageID string `json:"image_id"`

	Metadata struct {
		App string `json:"app"`
	} `json:"metadata"`

	CreatedAt string `json:"created_at"`
	Type      string `json:"type"`
}

func main() {
	var accessToken string
	flag.StringVar(&accessToken, "access_token", "", "gyazo api access token")
	flag.Parse()

	if accessToken == "" {
		log.Fatal("access_token is required")
	}

	if err := os.Mkdir("images", 0777); err != nil {
		if err.Error() != "mkdir images: file exists" { // If the folder exists already its nothing to be worried about
			log.Fatal(err)
		}
	}

	downloadClient := grab.NewClient()
	httpClient := &http.Client{}

	images := requestImages(httpClient, &accessToken)
	fmt.Println("Found", len(images), "images")

	for len(images) != 0 {
		for _, image := range images {
			if image.URL == "" { // For some reason non premium API will give empty responses
				continue
			}

			fileName := getNewFileName(&image)

			fmt.Println("Processing", fileName)
			req, err := grab.NewRequest("./images/"+fileName, image.URL)
			if err != nil {
				log.Fatal(err)
			}

			resp := downloadClient.Do(req)

			t := time.NewTicker(500 * time.Millisecond)
			defer t.Stop()

		Loop:
			for {
				select {
				case <-t.C:
					fmt.Printf("Transferred %v / %v bytes (%.2f%%)\n", resp.BytesComplete(), resp.Size, 100*resp.Progress())
				case <-resp.Done:
					break Loop
				}
			}

			// check for errors
			if err = resp.Err(); err != nil {
				fmt.Println("Download failed ❌")
				log.Panic(err)
			}

			fmt.Println("Successfully downloaded ✅")
			deleteImage(httpClient, &accessToken, &image.ImageID)
			fmt.Println("Successfully deleted from gyazo ✅")
		}

		images = requestImages(httpClient, &accessToken)
		fmt.Println("Found", len(images), "images")
	}

	fmt.Println("Finished, have a nice day! :)")
}

// Requests image api
func requestImages(client *http.Client, accessToken *string) []ImageAPIResponse {
	resp, err := client.Get("https://api.gyazo.com/api/images?per_page=100&access_token=" + *accessToken)
	if err != nil {
		log.Fatal(err)
	}

	jsonRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var images []ImageAPIResponse
	json.Unmarshal(jsonRaw, &images)

	return images
}

// Requests image deletion api
func deleteImage(client *http.Client, accessToken *string, imageID *string) {
	req, err := http.NewRequest("DELETE", "https://api.gyazo.com/api/images/"+*imageID+"?access_token="+*accessToken, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal("Error deleting image ❌")
	}
}

// Creates a new nice filename from the metadata
func getNewFileName(image *ImageAPIResponse) string {
	if image.Metadata.App != "" && image.Metadata.App != " " {
		return strings.ReplaceAll(" ", "_", image.Metadata.App) + "_" + image.CreatedAt[:len(image.CreatedAt)-5] + "." + image.Type
	}

	return image.CreatedAt[:len(image.CreatedAt)-5] + "." + image.Type
}
