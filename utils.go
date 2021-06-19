package main

import "log"

func handleErr(err error, info string) {
	if err != nil {
		log.Panic(info, err.Error())
	}
}

func logErr(err error, info string) {
	if err != nil {
		log.Println(info, err.Error())
	}
}
