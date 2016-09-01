package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

func getAlbumsList(user *User) (albumsList []Album, err error) {
	if _, err = os.Stat(albumJSONPath); os.IsNotExist(err) {
		fmt.Print("\nFetching list of albums...")

		var allAlbums []Album

		offset := 1
		limit := 200
		total := -1

		formValues := url.Values{
			"_verbosity": {"1"},
			"_filteruri": {"AlbumImages"},
			"_filter":    {"Uri,Name,Privacy,Protected,OriginalSizes,TotalSizes,ImageCount"},
			"start":      {strconv.Itoa(offset)},
			"count":      {strconv.Itoa(limit)},
		}

		for total == -1 || offset <= total {
			formValues.Set("start", strconv.Itoa(offset))

			var src []byte

			userAlbumsURL := API_BASE + user.Uri + "!albums"
			if src, err = apiGet(userAlbumsURL, formValues); err != nil {
				return
			}

			var apiResult APIResult
			var albumResponse APIAlbumsResponse

			if err = json.Unmarshal(src, &apiResult); err != nil {
				log.Println(err)
				return
			}

			if err = json.Unmarshal(apiResult.Response, &albumResponse); err != nil {
				log.Println(err)
				return
			}

			allAlbums = append(allAlbums, albumResponse.Album...)
			total = albumResponse.Pages.Total

			if total == 0 {
				fmt.Println("\n\nERROR! It seems like there is no albums! Sorry...")
				fmt.Println(openIssueMsg)
				os.Exit(0)
			}

			offset += limit
		}
		fmt.Println(" Done!")

		fmt.Print("Writing albums list file...")
		albumsJson, err := json.Marshal(allAlbums)

		err = ioutil.WriteFile(albumJSONPath, albumsJson, filePerm)
		fmt.Println(" Done!")

		if err != nil {
			fmt.Println("ERROR! Unable to write albums JSON file to disk. Sorry...")
			fmt.Println(openIssueMsg)
			os.Exit(0)
		}
	}

	if _, err := os.Stat(albumJSONPath); os.IsNotExist(err) {
		fmt.Println("ERROR! Unable to find the albums JSON file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	src, err := ioutil.ReadFile(albumJSONPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the albums JSON file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	err = json.Unmarshal(src, &albumsList)

	return
}

func getImageList(albums []Album) (mediaList []Media, err error) {
	if _, err = os.Stat(mediaJSONPath); os.IsNotExist(err) {
		fmt.Println("\nFetching list of images...")

		totalImages := 0

		for _, album := range albums {
			totalImages += album.ImageCount
		}

		limit := 200

		progress := pb.New(totalImages)
		progress.ShowCounters = true
		progress.ShowTimeLeft = true
		progress.Start()

		formValues := url.Values{
			"_filteruri": {""},
			"_filter":    {"FileName,Processing,Format,OriginalHeight,OriginalWidth,Collectable,IsArchive,IsVideo,ImageKey,ArchivedUri,ArchivedSize"},
			"_verbosity": {"1"},
			"start":      {"1"},
			"count":      {strconv.Itoa(limit)},
		}

		for _, album := range albums {
			offset := 1
			total := album.ImageCount

			curURL := API_BASE + album.Uris.AlbumImages

			for offset <= total {
				formValues.Set("start", strconv.Itoa(offset))

				var src []byte
				if src, err = apiGet(curURL, formValues); err != nil {
					return
				}

				var apiResult APIResult
				var imagesResponse APIAlbumImagesResponse

				if err = json.Unmarshal(src, &apiResult); err != nil {
					log.Println(err)
					return
				}

				if err = json.Unmarshal(apiResult.Response, &imagesResponse); err != nil {
					log.Println(err)
					return
				}

				mediaList = append(mediaList, imagesResponse.AlbumImage...)

				progress.Add(imagesResponse.Pages.Count)

				offset += limit
			}
		}

		progress.FinishPrint("Done fetching JSON index")

		fmt.Print("Writing status file...")
		mediaJSON, err := json.Marshal(mediaList)

		err = ioutil.WriteFile(mediaJSONPath, mediaJSON, filePerm)
		fmt.Println(" Done!")

		if err != nil {
			fmt.Println("ERROR! Unable to write JSON index file to disk. Sorry...")
			fmt.Println(openIssueMsg)
			os.Exit(0)
		}
	}

	if _, err := os.Stat(mediaJSONPath); os.IsNotExist(err) {
		fmt.Println("ERROR! Unable to find the JSON index file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	src, err := ioutil.ReadFile(mediaJSONPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the JSON index file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	err = json.Unmarshal(src, &mediaList)

	return
}

func fetchFiles(mediaList []Media) {
	var err error

	fmt.Println("\nTrying to extract pictures and videos...")

	oauthClient.Header.Del("Accept")

	progressCount := len(mediaList)
	for i, media := range mediaList {
		filePath := mediaPath + "/" + media.ImageKey + "." + media.Format

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			mediaList[i].Status = ""
		} else if media.Status == "started" {
			mediaList[i].Status = ""
		} else if media.Status == "done" {
			progressCount--
		} else if !retryFlag && media.Status == "failed" {
			progressCount--
		}
	}

	mediaJSON, _ := json.Marshal(mediaList)
	err = ioutil.WriteFile(mediaJSONPath, mediaJSON, filePerm)

	ch := make(chan bool, 10)
	mediaLock := sync.Mutex{}

	progress := pb.New(progressCount)
	progress.ShowCounters = true
	progress.ShowTimeLeft = true
	progress.Start()

	split := false
	if len(mediaList) > SPLIT_LIMIT {
		split = true
	}

	fails := 0
	success := 0

	for i, media := range mediaList {
		if mediaList[i].Status == "done" {
			success += 1
			continue
		}

		if !retryFlag && mediaList[i].Status == "failed" {
			fails += 1
			continue
		}

		if i > 0 && i%100 == 0 {
			mediaLock.Lock()
			mediaJSON, _ := json.Marshal(mediaList)
			err = ioutil.WriteFile(mediaJSONPath, mediaJSON, filePerm)
			mediaLock.Unlock()

			if err != nil {
				fmt.Println("ERROR! Unable to write update JSON index file to disk. Sorry...")
				fmt.Println(openIssueMsg)
				os.Exit(0)
			}
		}

		ch <- true

		go func(ch chan bool, index int, media Media) {
			mediaLock.Lock()
			mediaList[index].Status = "started"
			mediaLock.Unlock()
			extraPath := ""
			if split {
				extraPath = strconv.Itoa((index / SPLIT_LIMIT) + 1)
			}
			fetchMedia(&media, extraPath)
			mediaLock.Lock()
			mediaList[index] = media
			if media.Status == "done" {
				success += 1
			} else {
				fails += 1
			}
			progress.Increment()
			mediaLock.Unlock()
			<-ch
		}(ch, i, media)
	}

	// Wait for the last routines to be done
	for len(ch) > 0 {
		time.Sleep(500 * time.Millisecond)
	}

	mediaLock.Lock()
	mediaJSON, _ = json.Marshal(mediaList)
	err = ioutil.WriteFile(mediaJSONPath, mediaJSON, filePerm)
	mediaLock.Unlock()

	if err != nil {
		fmt.Println("ERROR! Unable to write update JSON index file to disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	progress.Finish()

	fmt.Println("Done trying to fetch all pictures and videos.")
	fmt.Println("Result:")
	fmt.Println("\tSuccess:", success)
	fmt.Println("\tFailed: ", fails)
}

func fetchMedia(media *Media, extraPath string) {
	media.Retries += 1
	media.Status = "started"

	filePath := mediaPath
	if extraPath != "" {
		filePath += "/" + extraPath
		err := os.MkdirAll(filePath, pathPerm)
		if err != nil {
			return
		}
	}
	filePath += "/" + media.ImageKey + "." + media.Format

	out, err := os.Create(filePath)
	if err != nil {
		media.Status = "failed"
		out.Close()
		os.Remove(filePath)
		return
	}

	res, err := oauthClient.Get(nil, tokenCred, media.ArchivedUri, nil)
	if err != nil || res.StatusCode != 200 {
		media.Status = "failed"
		out.Close()
		if res != nil {
			res.Body.Close()
		}
		os.Remove(filePath)
		return
	}

	n, err := io.Copy(out, res.Body)
	if err != nil {
		media.Status = "failed"
		out.Close()
		res.Body.Close()
		os.Remove(filePath)
		return
	}

	if n < 1000 {
		media.Status = "failed"
		out.Close()
		res.Body.Close()
		os.Remove(filePath)

	} else {
		media.Status = "done"
		out.Close()
		res.Body.Close()
	}
}
