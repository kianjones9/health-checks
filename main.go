package main

import (
	"fmt"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// characterized as the number of success and number of failures
//
//	for a given probe (i.e. all that is needed to compute availability)
type AvailabilityPair struct {
	Successes int
	Failures  int
}

// Decided to keep the url and uri as separate (and differentiating properties)
// for future use cases where we want to see the availability on a per unique URI basis
type Probe struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`  // optional
	Headers map[string]string `yaml:"headers,omitempty"` // optional
	Body    string            `yaml:"body,omitempty"`    // optional

	Domain       string
	Availability AvailabilityPair
}

var probeDB = make(map[string]Probe)

var client = &http.Client{
	Timeout: time.Second * 10,
}

func (p *Probe) initialize() {
	var err error
	parsedURL, err := url.Parse(p.URL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
	}
	p.Domain = parsedURL.Hostname()
}

func (p Probe) String() string {
	return fmt.Sprintf("Probe %v %v %v %v %v %v\n", p.Name, p.Domain, p.URL, p.Method, p.Headers, p.Body)
}

func parseConfig(configFilePath string) error {

	// read local yaml file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return err
	}

	// parse contents into an array of probes and store them directly in probeDB
	var probes []Probe
	err = yaml.Unmarshal(data, &probes)
	if err != nil {
		fmt.Printf("Error unmarshalling config file: %v\n", err)
		return err
	}
	for _, p := range probes {
		p.initialize()
		probeDB[p.URL] = p
	}
	return nil
}

func executeProbe(p *Probe) error {
	// build the request based on the probe definition
	req, err := http.NewRequest(p.Method, p.URL, strings.NewReader(p.Body))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return err
	}
	// add headers
	for key, value := range p.Headers {
		req.Header.Add(key, value)
	}

	// use httptrace to get highly performant and accurate latency metric
	var start time.Time
	var duration time.Duration
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			duration = time.Since(start)
		},
	}

	// instrument request with tracing
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// execute the request
	start = time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		// don't return err bc error is still signal we cannot reach the endpoint
	}
	// update the probe's availability stats
	// check for not nil in case there is an issue reaching the site
	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && duration < 500*time.Millisecond {
		p.Availability.Successes++
	} else {
		p.Availability.Failures++
	}
	probeDB[p.URL] = *p

	return err
}

func performHealthChecks() error {
	// iterate through all probes
	for _, probe := range probeDB {
		err := executeProbe(&probe)
		if err != nil {
			fmt.Printf("Error executing probe: %v\n", err)
			return err
		}
	}
	return nil
}

func computeAvailability() map[string]AvailabilityPair {
	// (re-)init availability stats. This is done in case we want to support altering monitors.yaml on the fly
	var availabilityStatsByDomain = make(map[string]AvailabilityPair)

	// get each probe's stats and coalesce them by domain
	for _, probe := range probeDB {

		if availabilityStatsByDomain[probe.Domain] == (AvailabilityPair{}) {
			availabilityStatsByDomain[probe.Domain] = AvailabilityPair{0, 0}
		}
		availabilityStatsByDomain[probe.Domain] = AvailabilityPair{
			Successes: availabilityStatsByDomain[probe.Domain].Successes + probe.Availability.Successes,
			Failures:  availabilityStatsByDomain[probe.Domain].Failures + probe.Availability.Failures,
		}
	}
	return availabilityStatsByDomain
}

func reportAvailability(stats map[string]AvailabilityPair) {
	// log the availability for each domain
	for domain, pair := range stats {
		availability := float64(pair.Successes) / float64(pair.Successes+pair.Failures) * 100
		fmt.Printf("%v has %v%% availability percentage\n", domain, int(availability+0.5))
	}
}

func main() {

	var configFilePath string
	if len(os.Args) > 1 {
		configFilePath = os.Args[1]
	} else {
		configFilePath = "./monitors.yaml"
	}

	err := parseConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return
	}

	for {
		performHealthChecks() // wrapper for executing all probes
		availabilityStats := computeAvailability()
		reportAvailability(availabilityStats)
		time.Sleep(15 * time.Second)
	}
}
