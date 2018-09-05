package main

import (
	"./ddns"
	"fmt"
	"math/rand"
	"time"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
)

const (
	interval = time.Hour //interval between update required check
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

func main() {
	args := os.Args
	if len(args) == 1 {
		panic("No configuration file provided -- exiting...")
	}

	config := loadConfiguration(args[1])
	ddns := ddns.NewDdnsUpdater(config.Hostname, config.Username, config.Password, nil)

	res, err := ddns.RequireUpdate()

	if err != nil {
		panic(err)
	}

	if res {
		ch := scheduleUpdate(ddns, 10*time.Second)
		err = <-ch
		if err != nil {
			fmt.Print(err)
		}
	}

	for {
		time.Sleep(interval)
		res, err := ddns.RequireUpdate()
		if err != nil {
			fmt.Print(err)
		} else {
			if res {
				ch := scheduleUpdate(ddns, 15*time.Minute)
				err := <-ch
				if err != nil {
					fmt.Print(err)
				}
			} else {
				fmt.Print("%v: %v\n", time.Now(), "No dns record update required")
			}
		}
	}

}

func scheduleUpdate(ddns DdnsUpdater, maxSleep time.Duration) chan error {
	ch := make(chan error)
	wait := time.Duration(rand.Int63n(int64(maxSleep)))
	go updateDyRecords(ddns, wait, ch)
	return ch
}

func updateDyRecords(ddns DdnsUpdater, t time.Duration, ch chan error) {
	time.Sleep(t)
	ch <- ddns.Update()
}

func loadConfiguration(filename string) *DyConfiguration {
	data, err := ioutil.ReadFile(filename)
	if (err != nil) {
		panic("Could not read configuration file -- exiting...")
	}

	var config DyConfiguration
	err = yaml.UnmarshalStrict(data, &config)
	if (err != nil) {
		panic(fmt.Sprintf("Invalid configuration file -- exiting..."))
	}

	return &config

}