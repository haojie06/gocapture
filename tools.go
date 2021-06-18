package main

import "log"

func handleErr(err error, info string) {
	if err != nil {
		log.Fatal(info, err.Error())
	}
}
