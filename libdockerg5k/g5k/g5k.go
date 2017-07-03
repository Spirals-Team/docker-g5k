package g5k

import (
	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
)

// G5K stores all informations needed to use the Grid5000 API
type G5K struct {
	username string
	password string
	sitesAPI map[string]*api.Client
}

// Init initialize a new G5K struct with the given parameters
func Init(username string, password string) *G5K {
	return &G5K{
		username: username,
		password: password,
		sitesAPI: map[string]*api.Client{},
	}
}

// createSiteAPI create a new Grid5000 API client for the given site
func (g *G5K) createSiteAPI(site string) {
	g.sitesAPI[site] = api.NewClient(g.username, g.password, site)
}

// getSiteAPI returns the API client for the given site (create it if not exist)
func (g *G5K) getSiteAPI(site string) *api.Client {
	// create API client for the site if it does not exist
	if _, ok := g.sitesAPI[site]; !ok {
		g.createSiteAPI(site)
	}

	return g.sitesAPI[site]
}
