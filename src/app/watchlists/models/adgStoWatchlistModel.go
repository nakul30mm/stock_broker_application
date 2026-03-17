package models

import "watchlists/commons/constants"

type BffAdgStoWatchlistRequest struct {
	Action       constants.Actiontype `json:"action"` //ADD or  DEL or GET
	ScripId      string               `json:"scripId"`
	WatchlistIds []uint64             `json:"watchlistIds"`
}

type WatchlistWithId struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type BffAdgStoWatchlistResponse struct {
	Status          string               `json:"message"`        //success or failure, if succeded - descriptive message, if failed - failed adding/ deleting/ getting scrip to wathclist...
	Action          constants.Actiontype `json:"action"`         //the one from the request
	WatchlistWithId []WatchlistWithId    `json:"watchlistNames"` //
	Warnings        []string             `json:"warnings"`       //only while adding if a scrip already exists in a watchlist
}

type AdgStoWatchlistResult struct {
	AddedTo     []uint64
	ExistsIn    []uint64
	DeletedFrom []uint64
	FoundIn     []uint64
}
