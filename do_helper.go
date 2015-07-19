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
var listDropletsNetworkVar bool
var publicDropletIPVar bool

func init() {
	flag.BoolVar(&listDropletsVar, "listDroplets", false, "List basic info on all Droplets")
	flag.BoolVar(&listDropletsNetworkVar, "listDropletsNetwork", false, "List network info for individual Droplet")
	flag.BoolVar(&publicDropletIPVar, "publicDropletIP", false, "Get Public IP address for specific Droplet")
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
	for droplet := 0; droplet < len(myDroplets); droplet++ {
		fmt.Printf("%s (%d) Created: %s.\n", myDroplets[droplet].Name, myDroplets[droplet].ID, myDroplets[droplet].Created)
	}
}

func listDropletsNetwork(myDroplets []godo.Droplet) {
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

func publicDropletIP(dropletName string, myDroplets []godo.Droplet) {
	for droplet := 0; droplet < len(myDroplets); droplet++ {
		if dropletName == myDroplets[droplet].Name {
			netinfo := myDroplets[droplet].Networks.V4
			for neti := 0; neti < len(netinfo); neti++ {
				if netinfo[neti].Type == "public" {
					fmt.Printf("%s", netinfo[neti].IPAddress)
				}
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
	if len(myDroplets) == 0 {
		log.Fatal("No droplets")
	}
	// TODO: no more hard coding
	dropletName := "Data01"
	// command branching
	switch {

	case listDropletsVar == true:
		listDroplets(myDroplets)

	case listDropletsNetworkVar == true:
		listDropletsNetwork(myDroplets)

	case publicDropletIPVar == true:
		publicDropletIP(dropletName, myDroplets)

	default:
		flag.PrintDefaults()
	}
}
