package database

import "sync"

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) *DB {
	return &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
}
