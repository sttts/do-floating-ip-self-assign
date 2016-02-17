package main

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"strconv"
)

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	tokenFile := flag.String("token-file", "", "A file path with a DigitalOcean API token.")
	token := flag.String("token", "", "A DigitalOcean API token.")
	updatePeriod := flag.Duration("update-period", 60 * time.Second, "The time between floating IP update tries, 0 for only initial assignment.")
	floatingIP := flag.String("floating-ip", "", "The floating IP address to self-assign.")
	retries := flag.Int64("retries", 5, "The number of retries when self-assignment fails, negative values for forever.")
	retryBackoff := flag.Duration("backoff", time.Second, "Initial backoff time after a failure.")
	retryBackoffFactor := flag.Float64("backoff-factor", 1.2, "Backoff time multiplier after each failure.")
	retryBackoffMax := flag.Duration("backoff-max", time.Second*30, "Maximum backoff time after a failure.")

	flag.Parse()

	// check for flag consistency
	if *tokenFile == "" && *token == "" {
		glog.Fatal("token or tokenfile is required")
	}
	if *tokenFile != "" && *token != "" {
		glog.Fatal("token and tokenfile cannot be specified both")
	}

	if *floatingIP == "" {
		glog.Fatal("floating-ip is required")
	}

	// process flags
	if *tokenFile != "" {
		bs, err := ioutil.ReadFile(*tokenFile)
		if err != nil {
			glog.Fatalf("error reading token file %s: %v", *tokenFile, err)
		}
		*token = string(bs)
	}

	// read droplet id
	glog.V(2).Info("Getting this droplet's id")
	resp, err := http.Get("http://169.254.169.254/metadata/v1/id")
	if err != nil {
		glog.Fatalf("cannot get droplet id: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Fatalf("cannot get droplet id: %v", err)
		}
		glog.Fatalf("status code %d getting droplet id: %v", string(body))
	}
	body, _, err := bufio.NewReader(resp.Body).ReadLine()
	if err != nil && err != io.EOF {
		glog.Fatal(err)
	}
	dropLetId64, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		glog.Fatalf("droplet-id format error: %v", err)
	}
	dropLetId := int(dropLetId64)
	glog.V(4).Infof("Got this droplet's id: %d", dropLetId)

	// create client for API
	tokenSource := &tokenSource{
		AccessToken: *token,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	first := true
	for {
		if !first {
			glog.V(4).Infof("Waiting for %v for next floating IP self-assignment", *updatePeriod)
			time.Sleep(*updatePeriod)
		}
		first = false

		fip, resp, err := client.FloatingIPs.Get(*floatingIP)
		if err != nil {
			glog.Error(err)
			continue
		}
		if resp.StatusCode != 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Errorf("cannot get floating ip %s: %v", *floatingIP, err)
			} else {
				glog.Fatalf("cannot get floating ip %s: %v", *floatingIP, body)
			}
			continue
		}
		if fip.Droplet != nil && fip.Droplet.ID == dropLetId {
			glog.V(3).Infof("Floating ip %s is already assigned to droplet %d", *floatingIP, dropLetId)
			continue
		}

		backoff := *retryBackoff
		actionId := int(-1)
		for r := *retries; r >= 0; r-- {
			glog.V(2).Infof("Trying to assign the floating ip %s to droplet %d", *floatingIP, dropLetId)

			// try to assign the float-up to this droplet
			action, _, err := client.FloatingIPActions.Assign(*floatingIP, dropLetId)
			if err == nil {
				actionId = action.ID
				break
			}

			glog.Errorf("FloatingIPsActions.Assign returned error: %v", err)

			glog.V(4).Infof("Waiting backoff time %v before next float ip self-assignment try", backoff)
			time.Sleep(backoff)

			// update backoff
			backoff = time.Duration(float64(backoff.Nanoseconds()) * *retryBackoffFactor)
			if backoff > *retryBackoffMax {
				backoff = *retryBackoffMax
			}
		}
		if actionId != -1 {
			// wait for action to finish
			timeout := time.Now().Add(30 * time.Second)
			waitForAction:
			for {
				if time.Now().After(timeout) {
					glog.Error("Timeout waiting for assignment to finish")
					break
				}

				// waiting until event is finished
				action, _, err := client.FloatingIPActions.Get(*floatingIP, actionId)
				if err != nil {
					glog.Error(err)
				} else {
					switch action.Status {
					case "completed":
						glog.Infof("Floating ip %s successfully assigned to droplet %d", *floatingIP, dropLetId)
						break waitForAction
					case "errored":
						glog.Infof("Assignment failed")
						break waitForAction
					}
				}

				time.Sleep(5 * time.Second)
			}
		}

		if *updatePeriod == 0 {
			os.Exit(0)
		}
	}
}
