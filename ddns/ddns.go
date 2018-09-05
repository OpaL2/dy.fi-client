package ddns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

//ip address storgae
type ipAddress string

// Ddns client information
type Ddns struct {
	Hostname   string    //dy.fi hostname
	Username   string    //dy.fi username
	Password   string    //dy.fi plaintext password
	lastUpdate time.Time //last time ip addtess has been udated
	ip         ipAddress // ip address after last update
}

//Errors from ddns client
type DdnsError struct {
	Timestamp   time.Time //time when error was triggered
	Description string    //description of error
}

const (
	maxUpdateInterval = 7 * 24 * time.Hour
	minUpdateInterval = 5 * 24 * time.Hour
)

//Stores pointer to http.Client used
var client *http.Client

//Creates new ddns client and returns pointer to it
//iinitializes http client if not provided or set before
func NewDdnsUpdater(hostname, username, password string, c *http.Client) *Ddns {

	if c != nil {
		SetClient(c)
	}

	if client == nil {
		SetClient(newClient())
	}

	return &Ddns{
		Hostname:   hostname,
		Username:   username,
		Password:   password,
		ip:         ipAddress("0.0.0.0"),
		lastUpdate: time.Now(),
	}

}

// Set http client
func SetClient(c *http.Client) {
	client = c
}

func newClient() *http.Client {

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

}

//Update dy.fi dns database
//returns nil when succesful error otherwise
func (ddns *Ddns) Update() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.dy.fi/nic/update?hostname=%s.dy.fi", ddns.Hostname), nil)
	if err != nil {
		return newError(fmt.Sprintf("Could not create http request -- %s", err))
	}

	req.SetBasicAuth(ddns.Username, ddns.Password)
	req.Header.Add("User-agent", "palojarvi-golang-dy/0.1 (oskari@palojarvi.fi)")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "www.dy.fi")

	resp, err := client.Do(req)
	if err != nil {
		return newError(fmt.Sprintf("Update request failed -- %s", err))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return newError(fmt.Sprintf("Could not read update request body -- %s", err))
	}

	re := regexp.MustCompile("(nochg|good)")
	if !re.Match(body) {
		return newError(fmt.Sprintf("Updating dy.fi dns records failed -- Return status: %s", body))
	}

	// If ip address where updated, extracting and storing new IP address from response
	if string(re.Find(body)) == "good" {
		re = regexp.MustCompile("(\\d{1,3}\\.?){4}")
		result := re.Find(body)
		ddns.ip = ipAddress(result)
	}

	if ddns.ip == "0.0.0.0" {
		ip, err := getipAddress()
		if err != nil {
			return err
		} else {
			ddns.ip = ip
		}
	}

	ddns.lastUpdate = time.Now()

	return newError(fmt.Sprintf("Updating dy.fi dns records succesful -- Return status: %s", body))
}

//Checks if dy.fi dns database needs updating
//returns true if update is needed false otherwise
func (ddns *Ddns) RequireUpdate() (bool, error) {
	ip, err := getipAddress()

	if err != nil {
		return false, err
	}

	if hasipChanged(ip, ddns.ip) || hasTimePassed(ddns.lastUpdate) {
		return true, nil
	}

	return false, nil
}

func getipAddress() (ipAddress, error) {

	ip := ipAddress("0.0.0.0")

	if client == nil {
		return ip, newError("No http client provided")
	}

	resp, err := client.Get("http://checkip.dy.fi")

	if err != nil {
		return ip, newError(fmt.Sprintf("ip address check connection error -- %s", err))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ip, newError("Could not read response from checkip request")
	}

	re := regexp.MustCompile("(\\d{1,3}\\.?){4}")
	result := re.Find(body)
	if result == nil {
		return ip, newError("Invalid response from checkip, could not parse IP address")
	}

	return ipAddress(result), nil
}

func hasipChanged(addr1, addr2 ipAddress) bool {
	return addr1 != addr2
}

func hasTimePassed(lastUpdate time.Time) bool {
	currentTime := time.Now()
	minUpdateTime := lastUpdate.Add(minUpdateInterval)
	return currentTime.After(minUpdateTime)
}

func (e DdnsError) Error() string {
	return fmt.Sprintf("%v: %v\n", e.Timestamp, e.Description)
}

func newError(description string) DdnsError {
	return DdnsError{
		time.Now(),
		description,
	}
}
