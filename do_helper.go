package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

import (
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type TokenSource struct {
	AccessToken string
}

var listDropletsVar bool
var dropletNetworkVar bool

func init() {
	flag.BoolVar(&listDropletsVar, "listDroplets", false, "List basic info on all Droplets")
	flag.BoolVar(&dropletNetworkVar, "dropletNetwork", false, "List network info for individual Droplet")
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// Pass Environment variable name to retrieve
// Return token value / error
func getToken(varName string) (string, error) {
	if len(varName) == 0 {
		return "", errors.New("getToken: No Variable key passed")
	} else {
		tokenValue := os.Getenv(varName)
		if len(tokenValue) == 0 {
			return "", errors.New("getToken: No token value for key")
		} else {
			return tokenValue, nil
		}
	}
}

// Pass token value to get a client
func getAuthClient(tokenValue string) *godo.Client {
	tokenSource := &TokenSource{
		AccessToken: tokenValue,
	}

	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	return client
}

func DropletList(client *godo.Client) ([]godo.Droplet, error) {
	list := []godo.Droplet{}

	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(opt)
		if err != nil {
			return nil, err
		}

		for _, d := range droplets {
			list = append(list, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}
	return list, nil
}

func prettyFullOutput(input godo.Droplet) {
	response, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Print(string(response))
}

func listDroplets(myDroplets []godo.Droplet) {
	if len(myDroplets) == 0 {
		fmt.Println("listDroplets: No droplets")
	} else {
		for droplet := 0; droplet < len(myDroplets); droplet++ {
			fmt.Printf("%s (%d) Created: %s.\n", myDroplets[droplet].Name, myDroplets[droplet].ID, myDroplets[droplet].Created)
		}
	}
}

func dropletNetwork(myDroplets []godo.Droplet) {
	if len(myDroplets) == 0 {
		fmt.Println("dropletNetwork: No droplets")
	} else {
		for droplet := 0; droplet < len(myDroplets); droplet++ {
			// Individual droplet header
			fmt.Printf("%s (%d): \n", myDroplets[droplet].Name, myDroplets[droplet].ID)
			netinfo := myDroplets[droplet].Networks.V4
			// Each droplets network info for each interface
			for neti := 0; neti < len(netinfo); neti++ {
				fmt.Printf("\t%s: %s\n", netinfo[neti].Type, netinfo[neti].IPAddress)
			}
		}
	}
}

func main() {
	// Parse flags for later execution
	flag.Parse()

	// Construct client
	tokenValue, err := getToken("DO_TOKEN")
	if err != nil {
		log.Fatal(err)
	}
	client := getAuthClient(tokenValue)

	// Assuming that API will need to happen. Running once
	myDroplets, _ := DropletList(client)

	// command branching
	switch {

	case listDropletsVar == true:
		listDroplets(myDroplets)

	case dropletNetworkVar == true:
		dropletNetwork(myDroplets)

	default:
		flag.PrintDefaults()
	}
}
