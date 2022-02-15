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

type ImageApiResponse struct {
	Url      string `json:"url"`
	Image_id string `json:"image_id"`

	Metadata struct {
		App string `json:"app"`
	} `json:"metadata"`

	Created_at string `json:"created_at"`
	Type       string `json:"type"`
}

func main() {
	var access_token string
	flag.StringVar(&access_token, "access_token", "", "gyazo api access token")
	flag.Parse()

	if access_token == "" {
		log.Fatal("access_token is required")
	}

	if err := os.Mkdir("images", 0777); err != nil {
		if err.Error() != "mkdir images: file exists" { // If the folder exists already its nothing to be worried about
			log.Fatal(err)
		}
	}

	downloadClient := grab.NewClient()
	httpClient := &http.Client{}

	var images []ImageApiResponse = requestImages(httpClient, &access_token)
	fmt.Println("Found", len(images), "images")

	for len(images) != 0 {
		for _, image := range images {
			if image.Url == "" { // For some reason non premium API will give empty responses
				continue
			}

			fileName := getNewFileName(&image)

			fmt.Println("Processing", fileName)
			req, err := grab.NewRequest("./images/"+fileName, image.Url)
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
			if err := resp.Err(); err != nil {
				fmt.Println("Download failed ❌")
				log.Fatal(err)
			}

			fmt.Println("Successfully downloaded ✅")
			deleteImage(httpClient, &access_token, &image.Image_id)
			fmt.Println("Successfully deleted from gyazo ✅")
		}

		images = requestImages(httpClient, &access_token)
		fmt.Println("Found", len(images), "images")
	}

	fmt.Println("Finished, have a nice day! :)")
}

// Requests image api
func requestImages(client *http.Client, access_token *string) []ImageApiResponse {
	resp, err := client.Get("https://api.gyazo.com/api/images?per_page=100&access_token=" + *access_token)
	if err != nil {
		log.Fatal(err)
	}

	json_raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var images []ImageApiResponse
	json.Unmarshal(json_raw, &images)
	return images
}

// Requests image deletion api
func deleteImage(client *http.Client, access_token *string, image_id *string) {
	req, err := http.NewRequest("DELETE", "https://api.gyazo.com/api/images/"+*image_id+"?access_token="+*access_token, nil)
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
func getNewFileName(image *ImageApiResponse) string {
	if image.Metadata.App != "" && image.Metadata.App != " " {
		return strings.ReplaceAll(" ", "_", image.Metadata.App) + "_" + image.Created_at[:len(image.Created_at)-5] + "." + image.Type
	}

	return image.Created_at[:len(image.Created_at)-5] + "." + image.Type
}
