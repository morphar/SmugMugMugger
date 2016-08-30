package main

import "encoding/json"

type Pages struct {
	Total          int
	Start          int
	Count          int
	RequestedCount int
}

type APIResponse struct {
	Uri            string
	UriDescription string
	EndpointType   string
	Locator        string
	LocatorType    string
	Pages          Pages
}

type User struct {
	Uri        string
	WebUri     string
	Name       string
	FirstName  string
	LastName   string
	NickName   string
	ImageCount int
}

// Only used for holding uris to album images
type AlbumUris struct {
	AlbumImages string
}

type Album struct {
	Uri           string
	Name          string
	Privacy       string
	Protected     bool
	OriginalSizes int
	TotalSizes    int
	ImageCount    int
	Uris          AlbumUris
}

type Media struct {
	FileName       string
	Processing     bool
	Format         string
	OriginalHeight int
	OriginalWidth  int
	Collectable    bool
	IsArchive      bool
	IsVideo        bool
	ImageKey       string
	ArchivedUri    string
	ArchivedSize   int
	Status         string
	Retries        int
}

type APIUserResponse struct {
	APIResponse
	User User
}

type APIAlbumImagesResponse struct {
	APIResponse
	AlbumImage []Media
}

type APIAlbumsResponse struct {
	APIResponse
	Album []Album
}

type APIResult struct {
	Response json.RawMessage
	Code     int
	Message  string
}
