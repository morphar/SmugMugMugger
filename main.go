package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/garyburd/go-oauth/oauth"
)

const (
	OAUTH_ORIGIN = "https://secure.smugmug.com"
	API_BASE     = "https://api.smugmug.com"
	API_ORIGIN   = API_BASE + "/api/v2"
	SPLIT_LIMIT  = 10000
)

var (
	mediaPath     string = "smugmug"
	albumJSONPath string = "smug_albums.json"
	mediaJSONPath string = "smug_images.json"
	tokenCredPath string = "credentials.json"

	pathPerm os.FileMode = 0770
	filePerm os.FileMode = 0770

	oauthClient oauth.Client

	tokenCred *oauth.Credentials

	appName    string = "SmugMug Mugger"
	githubLink string = "https://github.com/morphar/smugmugmugger"
	issuesLink string = githubLink + "/issues"

	openIssueMsg string = "Please try again later or go to GitHub and open an issue: " + issuesLink + "\n"

	// Flags
	retryFlag  bool = false // Retry failed images and videos?
	helpFlag   bool = false // Print help text
	statusFlag bool = false // Print current status
	resetFlag  bool = false // Reset everything and start over
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	oauthClient = oauth.Client{
		TemporaryCredentialRequestURI: OAUTH_ORIGIN + "/services/oauth/1.0a/getRequestToken",
		TokenRequestURI:               OAUTH_ORIGIN + "/services/oauth/1.0a/getAccessToken",
		ResourceOwnerAuthorizationURI: OAUTH_ORIGIN + "/services/oauth/1.0a/authorize",
	}

	// Make sure we get json back
	oauthClient.Header = http.Header{}
	oauthClient.Header.Add("Accept", "application/json")

	// Set the credentials
	oauthClient.Credentials.Token = API_KEY
	oauthClient.Credentials.Secret = API_SECRET

	flag.BoolVar(&retryFlag, "retry", retryFlag, "Retry failed images and videos?")
	flag.BoolVar(&helpFlag, "help", helpFlag, "Print help text")
	flag.BoolVar(&statusFlag, "status", statusFlag, "Print current status")
	flag.BoolVar(&resetFlag, "reset", resetFlag, "Reset everything and start over")

	err := os.MkdirAll(mediaPath, pathPerm)
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error

	flag.Parse()

	if helpFlag {
		printHelp()
		return
	}

	if statusFlag {
		printStatus()
		return
	}

	if resetFlag {
		fmt.Print("Resetting...")

		if _, err := os.Stat(albumJSONPath); err == nil {
			os.Remove(albumJSONPath)
		}

		if _, err := os.Stat(mediaJSONPath); err == nil {
			os.Remove(mediaJSONPath)
		}

		if _, err := os.Stat(mediaPath); err == nil {
			os.RemoveAll(mediaPath)
		}
		fmt.Println(" Done!")
		return
	}

	var user *User
	if user, err = authorize(); err != nil {
		log.Fatal(err)
	}

	if user == nil {
		fmt.Println("ERROR! Unable to get a user. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}
	if user.ImageCount == 0 {
		fmt.Println("ERROR! It seems like there is no images! Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	var albumsList []Album
	if albumsList, err = getAlbumsList(user); err != nil {
		fmt.Println("ERROR! Unable to fetch list of albums. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	var mediaList []Media
	if mediaList, err = getImageList(albumsList); err != nil {
		fmt.Println("ERROR! Unable to fetch list of images. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	fetchFiles(mediaList)
}

func apiGet(url string, form url.Values) (body []byte, err error) {
	res, err := oauthClient.Get(nil, tokenCred, url, form)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func printHelp() {
	fmt.Println("You can retry failed fetches, see current status or reset everything")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println(`./rescuelife -retry`)
	fmt.Println("")
}

func printStatus() {
	if _, err := os.Stat(mediaJSONPath); os.IsNotExist(err) {
		fmt.Println("ERROR! Unable to find the media JSON file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	src, err := ioutil.ReadFile(mediaJSONPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the media JSON file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	var mediaList []Media

	json.Unmarshal(src, &mediaList)

	var failed, started, done, waiting int
	total := len(mediaList)
	for _, media := range mediaList {
		switch media.Status {
		case "done":
			done++
		case "started":
			started++
		case "failed":
			failed++
		default:
			waiting++
		}
	}

	fmt.Println("\nStatus for fetching")
	fmt.Println("-----------------------------")
	fmt.Println("Succeeded:", done)
	fmt.Println("Failed:   ", failed)
	fmt.Println("Fetching: ", started)
	fmt.Println("Waiting:  ", waiting)
	fmt.Println("Total:    ", total)
	fmt.Println("")
}
