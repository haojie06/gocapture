package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
)

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

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func sortIPs(bandwidthMap map[string]*IPStruct) PairList {
	pl := make(PairList, len(bandwidthMap))
	i := 0
	for k, v := range bandwidthMap {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	// sort.Sort(pl)
	return pl
}
