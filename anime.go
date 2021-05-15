package main

type Anime struct {
	Id string
	Name string
	Episodes []*Episode
}

type Episode struct {
	Number string
	Links map[EpLanguage]map[EpQuality]map[string]StreamLink
}

type EpLanguage string
type EpQuality string

type StreamLink struct{
	Link string
	Width int
	Height int
	Duration float64
	Bitrate int
}

func NewAnime() *Anime {
	var a Anime
	return &a
}

func NewEpisode() *Episode {
	var ep Episode
	ep.Links = make(map[EpLanguage]map[EpQuality]map[string]StreamLink)
	return &ep
}