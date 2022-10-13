package kiosk

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	grapi "github.com/grafana/grafana-api-golang-client"
)

// GenerateURL constructs URL with appropriate parameters for kiosk mode
func GenerateURL(anURL string, kioskMode string, autoFit bool, isPlayList bool) string {
	u, _ := url.ParseRequestURI(anURL)
	q, _ := url.ParseQuery(u.RawQuery)

	switch kioskMode {
	case "tv": // TV
		q.Set("kiosk", "tv") // no sidebar, topnav without buttons
		log.Printf("KioskMode: TV")
	case "full": // FULLSCREEN
		q.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	case "disabled": // FULLSCREEN
		log.Printf("KioskMode: Disabled")
	default: // disabled
		q.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	}
	// a playlist should also go inactive immediately
	if isPlayList {
		q.Set("inactive", "1")
	}
	u.RawQuery = q.Encode()
	if autoFit {
		u.RawQuery += "&autofitpanels"
	}
	return u.String()
}

func NewGrafanaClient(anURL, username, password string, ignoreCertErrors bool) (*grapi.Client, error) {
	userinfo := url.UserPassword(username, password)
	clientConfig := grapi.Config{
		APIKey:    "",
		BasicAuth: userinfo,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: ignoreCertErrors,
				},
			},
		},
		OrgID:      0,
		NumRetries: 0,
	}

	u, err := url.Parse(anURL)
	if err != nil {
		return nil, err
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	grafanaClient, err := grapi.New(u.String(), clientConfig)
	if err != nil {
		return nil, err
	}

	return grafanaClient, nil
}

// getPlayListUID, get the UID of a playlist from an id
func GetPlayListUID(anURL string, client *grapi.Client) (string, error) {
	idURL, err := url.Parse(anURL)
	if err != nil {
		return "", err
	}
	id := path.Base(idURL.Path)

	log.Println("Playlist ID:", id)
	playLists, err := client.Playlists(url.Values{})
	if err != nil {
		return "", err
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}

	playList := &grapi.Playlist{}
	for _, avaliablePlayList := range *playLists {
		if avaliablePlayList.ID == intID {
			playList, err = client.Playlist(avaliablePlayList.UID)
			if err != nil {
				return "", err
			}
			break
		}
	}
	// TODO what to return if it's not a hit

	return playList.UID, nil
}

func ChangeIDtoUID(anURL, uid string) (string, error) {
	urlA, err := url.Parse(anURL)
	if err != nil {
		return "", err
	}
	urlPATH := urlA.Path

	// Make the url path in to a list
	urlPATH = strings.TrimSpace(urlPATH)
	//Cut off the leading and trailing forward slashes, if they exist.
	//This cuts off the leading forward slash.
	urlPATH = strings.TrimPrefix(urlPATH, "/")

	//This cuts off the trailing forward slash.
	if strings.HasSuffix(urlPATH, "/") {
		cut_off_last_char_len := len(urlPATH) - 1
		urlPATH = urlPATH[:cut_off_last_char_len]
	}
	//We need to isolate the individual components of the path.
	splitURLpath := strings.Split(urlPATH, "/")
	// delete the last item in the list
	splitURLpath = splitURLpath[:len(splitURLpath)-1]
	splitURLpath = append(splitURLpath, uid)

	// make in to string again
	fixedURL := strings.Join(splitURLpath, "/")
	// make a full URL again
	fixedURL = urlA.Scheme + "://" + urlA.Host + "/" + fixedURL
	return fixedURL, nil
}

func UrlChangeIDtoUID(anURL *url.URL, uid string) *url.URL {
	splitURLpath := strings.Split(anURL.Path, "/")
	// delete the last item in the list
	splitURLpath = splitURLpath[:len(splitURLpath)-1]
	splitURLpath = append(splitURLpath, uid)
	// make in to string again
	fixedURL := strings.Join(splitURLpath, "/")
	anURL.Path = fixedURL

	return anURL
}
