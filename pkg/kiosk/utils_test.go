package kiosk

import (
	"net/url"
	"path"
	"testing"
)

func TestChangeIDtoUID(t *testing.T) {
	anyURL := "https://mygrafana.com/playlists/play/1234"
	uid := "myRandomUID"
	idURL, err := ChangeIDtoUID(anyURL, uid)
	if err != nil {
		t.Fatalf("Unable to parse url %v", err)
	}
	parsedIdURL, err := url.Parse(idURL)
	if err != nil {
		t.Fatalf("Unable to parse url %v", err)
	}

	baseUID := path.Base(parsedIdURL.Path)
	if baseUID != "myRandomUID" {
		t.Fatalf("baseUID should be myRandomUID but returned %v", baseUID)
	}
	if idURL != "https://mygrafana.com/playlists/play/myRandomUID" {
		t.Fatalf("idURL should be https://mygrafana.com/playlists/play/myRandomUID but returned %v", idURL)
	}
}

func TestURLChangeIDtoUID(t *testing.T) {
	anyURL := "https://mygrafana.com/playlists/play/1234"
	uid := "myRandomUID"

	urlA, err := url.Parse(anyURL)
	if err != nil {
		t.Fatalf("Unable to parse URL")
	}
	uidURL := UrlChangeIDtoUID(urlA, uid)

	baseUID := path.Base(uidURL.Path)
	if baseUID != "myRandomUID" {
		t.Fatalf("baseUID should be myRandomUID but returned %v", baseUID)
	}
	if uidURL.String() != "https://mygrafana.com/playlists/play/myRandomUID" {
		t.Fatalf("idURL should be https://mygrafana.com/playlists/play/myRandomUID but returned %v", uidURL.String())
	}
}
