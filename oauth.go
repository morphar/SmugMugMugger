package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/skratchdot/open-golang/open"
)

func authorize() (user *User, err error) {
	if _, err = os.Stat(tokenCredPath); os.IsNotExist(err) {
		if err = authorizeApp(); err != nil {
			return
		}
	}

	tokenCredJSON, err := ioutil.ReadFile(tokenCredPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the JSON credentials file from disk. Sorry...")
		fmt.Println(openIssueMsg)
		os.Exit(0)
	}

	if err = json.Unmarshal(tokenCredJSON, &tokenCred); err != nil {
		return
	}

	if user, err = getUser(); err != nil {
		if err = authorizeApp(); err != nil {
			return
		}
		if user, err = getUser(); err != nil {
			return
		}
	}

	return
}

func authorizeApp() (err error) {
	if _, err = os.Stat(tokenCredPath); os.IsNotExist(err) {
		fmt.Printf("Hi and welcome to %s!\n\n", appName)
		fmt.Println("To download your pictures, we need you to authorize this app with SmugMug.\n")
		fmt.Println("When you have pressed \"enter\" a browser will open up.")
		fmt.Printf("You will be asked to authorize %s.\n\n", appName)
		fmt.Println("When you have clicked the \"Authorize\" button, you will see a number (PIN).")
		fmt.Println("Please come back here and enter PIN code, then the download will begin.\n")
		fmt.Println("Now please hit the \"enter\" key...")
		var empty string
		fmt.Scanln(&empty)
	}

	// Request the temporary credentials
	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, "oob", nil)
	if err != nil {
		return errors.New("Failed while trying to request temporary credentials: " + err.Error())
	}

	formValues := url.Values{
		"Access": {"Full"},
	}
	u := oauthClient.AuthorizationURL(tempCred, formValues)

	// Ask the user for the verification PIN
	fmt.Printf("Please enter the PIN here: ")

	// Automatically open the URL in a browser
	open.Run(u)

	var pin string
	fmt.Scanln(&pin)

	// Request token credentials
	tokenCred, _, err = oauthClient.RequestToken(nil, tempCred, pin)
	if err != nil {
		return
	}

	// Save the credentials for later
	tokenCredJSON, err := json.Marshal(tokenCred)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(tokenCredPath, tokenCredJSON, filePerm)
	if err != nil {
		os.Remove(tokenCredPath)
	}

	return
}

func getUser() (user *User, err error) {
	var src []byte

	if src, err = apiGet(API_ORIGIN+"!authuser", nil); err != nil {
		return
	}

	var apiResult APIResult
	var userResponse APIUserResponse

	if err = json.Unmarshal(src, &apiResult); err != nil {
		return
	}

	if err = json.Unmarshal(apiResult.Response, &userResponse); err != nil {
		return
	}

	user = &userResponse.User

	return
}
