package main

import (
	"io"
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Tokens Tokens
}

type Tokens struct {
	Discord string
}

func GetConfig() Config {
    var config Config

    file, err := os.Open("config.toml")
    byteToml, _ := io.ReadAll(file)
    if err != nil {
        log.Fatalln("Could not load config:", err)
    }
    err = toml.Unmarshal(byteToml, &config)
    if err != nil {
        log.Fatalln("Could not Unmarshal config:", err)
    }

    return config
}
