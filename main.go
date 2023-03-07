package main

import (
	"fmt"
	"gopress/commands"
	"gopress/shared"
	"os"
)

func usage() {
	fmt.Println("Usage: gopress <input-file> (-d/-e)")
	fmt.Println("-e - encode mode")
	fmt.Println("-d - decode mode")
	os.Exit(1)
}

func parse_args() (config shared.Config) {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Input file is missing.")
		usage()
	}

	config.Filename = args[0]
	args = args[1:]
	var is_encode, is_decode, is_output bool
	for _, arg := range args {
		if is_output {
			config.Output = arg
			is_output = false
		}
		if arg == "-o" {
			if config.Output != "" {
				fmt.Println("Argument -o provided multiple times.")
				usage()
			}
			is_output = true
		}
		if arg == "-e" {
			is_encode = true
		}
		if arg == "-d" {
			is_decode = true
		}
	}
	if !((is_encode || is_decode) && !(is_encode && is_decode)) { // NOT XOR
		fmt.Println("Can't provide -d and -e at the same time.")
		usage()
	}
	if is_encode {
		config.Mode = shared.ENCODE_MODE
	} else {
		config.Mode = shared.DECODE_MODE
	}

	return
}

func main() {
	config := parse_args()

	input, err := os.ReadFile(config.Filename)
	if err != nil {
		panic(err)
	}

	if config.Mode == shared.ENCODE_MODE {
		commands.Encode(config, input)
	} else {
		commands.Decode(config, input)
	}
}
