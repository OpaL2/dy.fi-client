package main

import (
	"fmt"
	"github.com/OpaL2/dy.fi-client/ddns"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

const (
	interval = time.Hour  //interval between update required check
)

type DyConfiguration struct {
	Hostname string
	Username string
	Password string
}

type DdnsUpdater interface {
	RequireUpdate() (bool, error)
	Update() error
}

type updater struct {
	Updater DdnsUpdater
}

func main() {
	fmt.Print("Starting...\n")
	args := os.Args
	if len(args) == 1 {
		panic("No configuration file provided -- exiting...")
	}
	fmt.Print("Loading configuration...\n")
	config := loadConfiguration(args[1])
	fmt.Print("Configuration loaded\n")
	ddns := ddns.NewDdnsUpdater(config.Hostname, config.Username, config.Password, nil)

	fmt.Print("Testing connection...\n")
	//connection testing
	_, err := ddns.RequireUpdate()

	if err != nil {
		panic(err)
	}
	fmt.Print("Connection test completed\n")

	cErr := make(chan error)
	update := make(chan struct{})
	testUpdate := make(chan chan bool)

	go func() {
		for ch := range testUpdate {
			res, err := ddns.RequireUpdate()
			if err != nil {
				cErr <- err
			}
			ch <- res
		}
	}()

	go func() {
		for range update {
			err := ddns.Update()
			if err != nil {
				cErr <- err
			}
		}
	}()

	var updateTimer <- chan time.Time
	var testReturn chan bool
	updateTimer = nil
	testTimeout := time.NewTimer(interval)
	testReturn = nil

	for {
		select {
		case <-testTimeout.C:
			testTimeout.Reset(interval)
			testReturn = make(chan bool)
			testUpdate <- testReturn

		case <-updateTimer:
			update <- struct{}{}

		case err := <-cErr:
			fmt.Print(err)

		case res := <-testReturn:
			if res {
				updateTimer = time.After(time.Duration(rand.Int63n(int64(time.Minute*15))))
			}
		}
	}
}


func loadConfiguration(filename string) *DyConfiguration {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic("Could not read configuration file -- exiting...")
	}

	var config DyConfiguration
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		panic(fmt.Sprintf("Invalid configuration file -- exiting..."))
	}

	return &config

}
