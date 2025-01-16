package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"time"
)

// characterized as the number of success and number of failures
//
//	for a given probe (i.e. all that is needed to compute availability)
type availabilityPair struct {
	successes int
	failures  int
}

// Decided to keep the url and name as separate (and differentiating properties)
// for future use cases where we want to see the availability on a per unique URI basis
type probe struct {
	url     string
	name    string
	uri     string
	method  string            // optional
	headers map[string]string // optional
	body    string            // optional

	availability availabilityPair
}

var probeDB = make(map[string]probe)

func (p *probe) initialize() {
	var err error
	p.uri, err = url.JoinPath(p.url, p.name)
	if err != nil {
		fmt.Printf("Error joining URL and name: %v\n", err)
		return
	}
}

func parseConfig(configFilePath string) error {

	// read local yaml file
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return err
	}

	// parse contents into an array of probes and store them directly in probeDB
	var probes []probe
	err = yaml.Unmarshal(data, &probes)
	if err != nil {
		fmt.Printf("Error unmarshalling config file: %v\n", err)
		return err
	}
	for _, p := range probes {
		p.initialize()
		probeDB[p.uri] = p
	}
	return nil
}

func executeProbe(probe_uri string, p probe) error {
	// TODO: implement the actual probe execution
	// build the request based on the probe definition
	// req, err := http.NewRequest(p.method, p.uri, strings.NewReader(p.body))
	return nil
}

func performHealthChecks() error {
	// iterate through all probes
	for probe_uri, probe := range probeDB {
		err := executeProbe(probe_uri, probe)
		if err != nil {
			fmt.Printf("Error executing probe: %v\n", err)
			return err
		}
	}
}

func computeAvailability() map[string]availabilityPair {
	// (re-)init avaialbility stats
	var availabilityStatsByDomain = make(map[string]availabilityPair)
	// get each probe's stats and coalesce them by domain
	for _, probe := range probeDB {
		availabilityStatsByDomain[probe.url] = availabilityPair{
			successes: availabilityStatsByDomain[probe.url].successes + probe.availability.successes,
			failures:  availabilityStatsByDomain[probe.url].failures + probe.availability.failures,
		}
	}
	return availabilityStatsByDomain
}

func reportAvailability(stats map[string]availabilityPair) {
	// log the availability for each domain
	for domain, pair := range stats {
		availability := float64(pair.successes) / float64(pair.successes+pair.failures) * 100
		fmt.Printf("%v has %v%% availability percentage\n", domain, int(availability+0.5))
	}
}

func main() {

	configFilePath := "./monitors.yaml"

	if len(os.Args) > 1 {
		configFilePath = os.Args[1]
	} else {
		fmt.Println("Using default config file path (./monitors.yaml)")
	}

	err := parseConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return
	}

	for {
		performHealthChecks()
		availabilityStats := computeAvailability()
		if err != nil {
			fmt.Printf("Error computing availability: %v\n", err)
			return
		}
		reportAvailability(availabilityStats)
		time.Sleep(15 * time.Second)
	}
}
