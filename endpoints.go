package main

type endpoints string

const (
	endpointRoot endpoints = "/"
	endpointGui  endpoints = "/KaoriGui/"
)

func (e endpoints) String() string {
	return string(e)
}