package main

import (
	"flag"
	"log"
	"strings"
	"time"

	api "github.com/go-routeros/routeros"
)

var (
	command  = flag.String("command", "/ip/firewall/address-list/listen", "RouterOS command")
	address  = flag.String("address", "127.0.0.1:8728", "RouterOS address and port")
	username = flag.String("username", "admin", "User name")
	password = flag.String("password", "admin", "Password")
	timeout  = flag.Duration("timeout", 10*time.Second, "Cancel after")
	useTLS   = flag.Bool("tls", false, "Use TLS")
)

func main() {
	flag.Parse()

	c := &api.Client{
		Address:  *address,
		Username: *username,
		Password: *password,
		Queue:    100,
	}
	var err error
	if *useTLS {
		err = c.ConnectTLS(nil)
	} else {
		err = c.Connect()
	}
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	errC := c.Async()

	go func() {
		l, err := c.ListenArgs(strings.Split(*command, " "))
		if err != nil {
			log.Print(err)
		}

		go func() {
			time.Sleep(*timeout)

			log.Print("Cancelling the RouterOS command...")
			_, err := l.Cancel()
			if err != nil {
				log.Fatal(err)
			}
		}()

		log.Print("Waiting for !re...")
		for sen := range l.Chan() {
			log.Printf("Update: %s", sen)
		}

		err = l.Err()
		if err != nil {
			log.Fatal(err)
		}

		log.Print("Done!")
		c.Close()
	}()

	err = <-errC
	if err != nil {
		log.Fatal(err)
	}
}