package main

import (
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
	// Init command line Flags (boolean, flagName, defaultValue, description
	flag.BoolVar(&listDropletsVar, "listDroplets", false, "List basic info on all Droplets")
	flag.BoolVar(&listDropletsNetworkVar, "listDropletsNetwork", false, "List network info for individual Droplet")
	flag.BoolVar(&publicDropletIPVar, "publicDropletIP", false, "Get Public IP address for specific Droplet")
}

/*
	Token method interface
*/
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

/*
	Return an DO API Key value from Environment variable
*/
func getAPIKey(varName string) (string, error) {
	if len(varName) == 0 {
		return "", errors.New("getAPIKey: No Variable key passed")
	} else {
		tokenValue := os.Getenv(varName)
		if len(tokenValue) == 0 {
			return "", errors.New("getAPIKey: No token value for key")
		} else {
			return tokenValue, nil
		}
	}
}

/*
	Return a DO authenticated client from an API Key Token value
*/
func getAuthClient(tokenValue string) *godo.Client {
	tokenSource := &TokenSource{
		AccessToken: tokenValue,
	}

	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	return client
}

/*
	Return a list of all Digital Ocean droplets information
	Example from https://github.com/digitalocean/godo @ Pagination section
*/
func getDropletList(client *godo.Client) ([]godo.Droplet, error) {
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

/*
	Print general info (name, id, creation time) of each Droplet
*/
func listDroplets(myDroplets []godo.Droplet) {
	for droplet := 0; droplet < len(myDroplets); droplet++ {
		fmt.Printf("%s (%d) Created: %s.\n", myDroplets[droplet].Name, myDroplets[droplet].ID, myDroplets[droplet].Created)
	}
}

/*
	Print all network information (private/public interfaces) about all Droplets
*/
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

/*
	Print IP address for Public network interface for a specific Droplet (by name)
*/
func publicDropletIP(dropletName string, myDroplets []godo.Droplet) {
	for droplet := 0; droplet < len(myDroplets); droplet++ {
		if myDroplets[droplet].Name == dropletName {
			// Get network interface info for this droplet
			netinfo := myDroplets[droplet].Networks.V4
			for neti := 0; neti < len(netinfo); neti++ {
				// Only retrive the public interface
				if netinfo[neti].Type == "public" {
					fmt.Printf("%s", netinfo[neti].IPAddress)
				}
			}
		}
	}
}

func main() {
	// Parse command line flags defined in init()
	flag.Parse()

	// Retrieve API Key and construct a DO client
	tokenValue, err := getAPIKey("DO_TOKEN")
	if err != nil {
		log.Fatal(err)
	}
	client := getAuthClient(tokenValue)

	// Assuming that API will need to happen. Running once
	myDroplets, _ := getDropletList(client)
	if len(myDroplets) == 0 {
		log.Fatal("No droplets")
	}
	// TODO: no more hard coding
	dropletName := "Data01"

	// Command line flag processing
	switch {

	case listDropletsVar == true:
		listDroplets(myDroplets)

	case listDropletsNetworkVar == true:
		listDropletsNetwork(myDroplets)

	case publicDropletIPVar == true:
		publicDropletIP(dropletName, myDroplets)

	// If none, print defaults from flag definitions
	default:
		flag.PrintDefaults()
	}
}
