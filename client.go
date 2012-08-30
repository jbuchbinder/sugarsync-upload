// SUGARSYNC UPLOAD
// https://github.com/jbuchbinder/sugarsync-upload

package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ACCESS_KEY_ID      = "MzI1MzA4MjEzNDI4MTA0NzczNjI"
	APP_ID             = "/sc/3253082/350_101900713"
	PRIVATE_ACCESS_KEY = "OWFlMmI0MWI0MjljNGNkMGJiNzFlNjM1NjUwNTZlODU"

	AUTHORIZATION_URL = "https://api.sugarsync.com/authorization"
	REFRESH_URL       = "https://api.sugarsync.com/app-authorization"
)

var (
	DEBUG    = flag.Bool("debug", false, "debug mode")
	username = flag.String("username", "", "sugarsync email/user name")
	password = flag.String("password", "", "sugarsync password")
	action   = flag.String("action", "upload", "upload|list")
	dest     = flag.String("dest", "", "destination folder (or 'mb' for magic briefcase, 'wa' for web archive, etc)")
)

var transmitFiles []string

type AuthorizationResponse struct {
	Expiration string `xml:"expiration"`
	User       string `xml:"user"`
}

type CollectionContents struct {
	Collection []SugarsyncCollection `xml:"collection"`
	File       []SugarsyncFile       `xml:"file"`
}

type SugarsyncCollection struct {
	Type        string `xml:"type,attr"`
	DisplayName string `xml:"displayName"`
	Ref         string `xml:"ref"`
	Contents    string `xml:"contents"`
}

type SugarsyncFile struct {
	DisplayName     string `xml:"displayName"`
	Ref             string `xml:"ref"`
	Size            int64  `xml:"size"`
	LastModified    string `xml:"lastModified"`
	MediaType       string `xml:"mediaType"`
	PresentOnServer bool   `xml:"presentOnServer"`
	FileData        string `xml:"fileData"`
}

type UserInfo struct {
	Username         string `xml:"username" json:"username"`
	Nickname         string `xml:"nickname" json:"nickname"`
	Workspaces       string `xml:"workspaces" json:"workspaces"`
	SyncFolders      string `xml:"syncfolders" json:"syncfolders"`
	Deleted          string `xml:"deleted" json:"deleted"`
	MagicBriefcase   string `xml:"magicBriefcase" json:"magicBriefcase"`
	WebArchive       string `xml:"webArchive" json:"webArchive"`
	MobilePhotos     string `xml:"mobilePhotos" json:"mobilePhotos"`
	ReceivedShares   string `xml:"receivedShares" json:"receivedShares"`
	Contacts         string `xml:"contacts" json:"contacts"`
	Albums           string `xml:"albums" json:"albums"`
	RecentActivities string `xml:"recentActivities" json:"recentActivities"`
	PublicLinks      string `xml:"publicLinks" json:"publicLinks"`
}

func main() {
	flag.Parse()
	transmitFiles = flag.Args()

	if *username == "" || *password == "" {
		panic("Username and password must be set (-h for more details)")
	}

	switch *action {
	case "upload":
		{
			if transmitFiles == nil || len(transmitFiles) < 1 {
				panic("Files must be specified in upload mode (-h for more details)")
			}
		}
	case "list":
		{
		}
	default:
		{
			panic("Invalid action (-h for more details)")
		}
	}

	r := refresh(*username, *password)
	if *DEBUG {
		fmt.Println("refresh token = " + r)
	}
	a, ua := auth(r)
	if *DEBUG {
		fmt.Println("auth token = " + a)
		fmt.Println("user url = " + ua)
	}

	var myDest string

	switch *dest {
	case "mb":
		{
			ui := getUserInfo(a, ua)
			if *DEBUG {
				fmt.Println("magic briefcase = " + ui.MagicBriefcase)
			}
			myDest = ui.MagicBriefcase
		}
	case "wa":
		{
			ui := getUserInfo(a, ua)
			if *DEBUG {
				fmt.Println("web archive = " + ui.WebArchive)
			}
			myDest = ui.WebArchive
		}
	default:
		{
			myDest = *dest
			if *DEBUG {
				fmt.Println("Do nothing, dest already set")
			}
		}
	}

	switch *action {
	case "list":
		{
			if *DEBUG {
				fmt.Printf("myDest = '" + myDest + "'\n")
			}
			i := getLocationInfo(a, myDest)
			for j := 0; j < len(i.Collection); j++ {
				fmt.Printf("DIR: %s : %s\n", i.Collection[j].DisplayName, i.Collection[j].Ref)
			}
			for j := 0; j < len(i.File); j++ {
				fmt.Printf("FILE: %s (%d bytes) : %s\n", i.File[j].DisplayName, i.File[j].Size, i.File[j].Ref)
			}
		}
	case "upload":
		{
			// Post new file, get file info first
			for i := 0; i < len(transmitFiles); i++ {
				fl := getNewFileLocation(a, myDest, filepath.Base(transmitFiles[i]))
				fmt.Println("Uploading " + transmitFiles[i] + " to " + fl)
				uploadFile(a, fl, transmitFiles[i])
			}
		}
	}
}

func auth(refreshToken string) (authToken string, userResource string) {
	client := http.Client{}
	payload := "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n" +
		"<tokenAuthRequest>\n" +
		"<accessKeyId>" + ACCESS_KEY_ID + "</accessKeyId>\n" +
		"<privateAccessKey>" + PRIVATE_ACCESS_KEY + "</privateAccessKey>\n" +
		"<refreshToken>" + refreshToken + "</refreshToken>\n" +
		"</tokenAuthRequest>\n"
	req, _ := http.NewRequest("POST", AUTHORIZATION_URL, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/xml; charset=UTF-8")

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	// Extract user resource from body
	//io.Copy(os.Stderr, res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf(err.Error())
	}
	var obj AuthorizationResponse
	err = xml.Unmarshal(body, &obj)
	if err != nil {
		fmt.Printf(err.Error())
	}
	userResource = obj.User

	authToken = res.Header.Get("Location")
	return
}

func refresh(user, pass string) string {
	client := http.Client{}
	payload := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<appAuthorization>\n" +
		"<username>" + user + "</username>\n" +
		"<password>" + pass + "</password>\n" +
		"<application>" + APP_ID + "</application>\n" +
		"<accessKeyId>" + ACCESS_KEY_ID + "</accessKeyId>\n" +
		"<privateAccessKey>" + PRIVATE_ACCESS_KEY + "</privateAccessKey>\n" +
		"</appAuthorization>\n"
	if *DEBUG {
		fmt.Println("Posting to " + REFRESH_URL + " with:\n" + payload)
	}
	req, err := http.NewRequest("POST", REFRESH_URL, strings.NewReader(string(payload)))
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/xml; charset=UTF-8")
	req.SetBasicAuth(user, pass)

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	io.Copy(os.Stderr, res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("IO: ")
		fmt.Println(err)
		fmt.Println(body)
	}
	return res.Header.Get("Location")
}

func getNewFileLocation(authToken string, folder string, fileName string) string {
	client := http.Client{}
	payload := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<file>\n" +
		"<displayName>" + fileName + "</displayName>\n" +
		"<mediaType>application/octet-stream</mediaType>\n" +
		"</file>\n"
	if *DEBUG {
		fmt.Println("Posting to " + REFRESH_URL + " with:\n" + payload)
	}
	req, err := http.NewRequest("POST", folder, strings.NewReader(string(payload)))
	req.Header.Set("Authorization", authToken)

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	return res.Header.Get("Location")
}

func getUserInfo(authToken string, userResource string) UserInfo {
	client := http.Client{}
	req, _ := http.NewRequest("GET", userResource, nil)
	req.Header.Set("Authorization", authToken)

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	// Extract user resource from body
	//io.Copy(os.Stderr, res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf(err.Error())
	}
	var obj UserInfo
	err = xml.Unmarshal(body, &obj)
	if err != nil {
		fmt.Printf(err.Error())
	}

	return obj
}

func getLocationInfo(authToken string, resource string) CollectionContents {
	client := http.Client{}
	req, _ := http.NewRequest("GET", resource+"/contents", nil)
	req.Header.Set("Authorization", authToken)

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	// Extract user resource from body
	//io.Copy(os.Stderr, res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf(err.Error())
	}
	var obj CollectionContents
	err = xml.Unmarshal(body, &obj)
	if err != nil {
		fmt.Printf(err.Error())
	}

	return obj
}

func uploadFile(authToken string, fileLocation string, file string) {
	client := http.Client{}
	fData, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err.Error())
	}
	req, err := http.NewRequest("PUT", fileLocation+"/data", strings.NewReader(string(fData)))
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprint(len(fData)))

	if *DEBUG {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ")
		fmt.Println(err)
	}
	defer res.Body.Close()

	if *DEBUG {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}
}
