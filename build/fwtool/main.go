package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	flagFWIn      string
	flagFWOut     string
	flagSquashIn  string
	flagSquashOut string
)

func usage() {
	fmt.Printf(`fwtool, a draytek vigor firmware manipulation tool. Usage:

Extract squashfs:
	%s -fw_in v167_50.all -squash_out squash.bin

Update squashfs:
	%s -fw_in v167_50.all -squash_in squash-custom.bin -fw_out v167_50-custom.all
`, os.Args[0], os.Args[0])
	os.Exit(1)
}

func main() {
	flag.StringVar(&flagFWIn, "fw_in", "", "Path to firmware file to read")
	flag.StringVar(&flagFWOut, "fw_out", "", "Path to firmware file to write")
	flag.StringVar(&flagSquashIn, "squash_in", "", "Path to read squashfs blob from")
	flag.StringVar(&flagSquashOut, "squash_out", "", "Path to write squashfs blob to")
	flag.Parse()

	if flagFWIn != "" && flagFWOut == "" && flagSquashIn == "" && flagSquashOut != "" {
		squashExtract()
		return
	}
	if flagFWIn != "" && flagFWOut != "" && flagSquashIn != "" && flagSquashOut == "" {
		squashUpdate()
		return
	}
	usage()
}

func squashExtract() {
	fw, err := parseFWFile(flagFWIn)
	if err != nil {
		log.Fatalf("Could not read firmware file: %v", err)
	}
	if err := ioutil.WriteFile(flagSquashOut, fw.squash, 0600); err != nil {
		log.Fatalf("Could not write squashfs: %v", err)
	}
	log.Printf("SquashFS blob written to %s.", flagSquashOut)
}

func squashUpdate() {
	fw, err := parseFWFile(flagFWIn)
	if err != nil {
		log.Fatalf("Could not read firmware file: %v", err)
	}

	squash, err := ioutil.ReadFile(flagSquashIn)
	if err != nil {
		log.Fatalf("Could not read squashfs file: %v", err)
	}

	fw.squash = squash

	if err := fw.write(flagFWOut); err != nil {
		log.Fatalf("Could not write firmware file: %v", err)
	}
}
