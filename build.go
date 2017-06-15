package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/digitalocean/godo"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

//Settings to run server
type Settings struct {
	Digitalocean struct {
		AccessToken string
	}

	Github struct {
		Secret string
	}
}

func swapHTMLDroplet(client *godo.Client, event *github.PushEvent) (*godo.Droplet, error) {
	images, _, _ := client.Images.ListUser(context.TODO(), nil)
	imageID := -1

	for _, image := range images {
		if image.Name == "jivecakehtml" {
			imageID = image.ID
			break
		}
	}

	if imageID == -1 {
		return nil, errors.New("Unable to find image jivecakehtml")
	}

	createRequest := godo.DropletCreateRequest{
		Name:   fmt.Sprintf("jivecakehtml-%v", event.After),
		Region: "nyc3",
		Size:   "2gb",
		Image: godo.DropletCreateImage{
			ID: imageID,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			godo.DropletCreateSSHKey{ID: 1726114, Fingerprint: ""},
			godo.DropletCreateSSHKey{ID: 7151393, Fingerprint: ""},
		},
		IPv6:       true,
		Monitoring: true,
	}

	newDroplet, _, err := client.Droplets.Create(context.TODO(), &createRequest)

	if err != nil {
		return nil, err
	}

	droplet, _, _ := client.Droplets.Get(context.TODO(), newDroplet.ID)

	for len(droplet.Networks.V4) == 0 || droplet.Status != "active" {
		droplet, _, _ = client.Droplets.Get(context.TODO(), newDroplet.ID)
	}

	ip, _ := droplet.PublicIPv4()

	for i := 0; i < 10; i++ {
		sshError := exec.Command(
			"ssh",
			"-o", "StrictHostKeyChecking=no",
			"root@"+ip,
			"bash restart.sh",
		).Run()

		if sshError == nil {
			break
		} else if i == 9 {
			return newDroplet, sshError
		}

		time.Sleep(1 * time.Second)
	}

	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	for i := 0; i < 20; i++ {
		_, err = httpClient.Get("https://" + ip)

		if err == nil {
			break
		}

		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return newDroplet, err
	}

	_, _, err = client.FloatingIPActions.Assign(context.TODO(), "138.197.59.43", droplet.ID)

	if err != nil {
		return newDroplet, err
	}

	return newDroplet, nil
}

func swapAPIDroplet(client *godo.Client, event *github.PushEvent) (*godo.Droplet, error) {
	images, _, _ := client.Images.ListUser(context.TODO(), nil)
	imageID := -1

	for _, image := range images {
		if image.Name == "jivecakeapi" {
			imageID = image.ID
			break
		}
	}

	if imageID == -1 {
		return nil, errors.New("Unable to find image jivecakeapi")
	}

	createRequest := godo.DropletCreateRequest{
		Name:   fmt.Sprintf("jivecakeapi-%v", *event.After),
		Region: "nyc3",
		Size:   "2gb",
		Image: godo.DropletCreateImage{
			ID: imageID,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			godo.DropletCreateSSHKey{ID: 1726114, Fingerprint: ""},
			godo.DropletCreateSSHKey{ID: 7151393, Fingerprint: ""},
		},
		IPv6:       true,
		Monitoring: true,
	}

	newDroplet, _, err := client.Droplets.Create(context.TODO(), &createRequest)

	if err != nil {
		return nil, err
	}

	droplet, _, _ := client.Droplets.Get(context.TODO(), newDroplet.ID)

	for len(droplet.Networks.V4) == 0 || droplet.Status != "active" {
		droplet, _, _ = client.Droplets.Get(context.TODO(), newDroplet.ID)
	}

	ip, _ := droplet.PublicIPv4()

	for i := 0; i < 10; i++ {
		sshError := exec.Command(
			"ssh",
			"-o", "StrictHostKeyChecking=no",
			"root@"+ip,
			"docker-compose --project-name jivecakeapi up -d",
		).Run()

		if sshError == nil {
			break
		} else if i == 9 {
			return newDroplet, sshError
		}

		time.Sleep(1 * time.Second)
	}

	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	for i := 0; i < 20; i++ {
		_, err = httpClient.Get("https://" + ip)

		if err == nil {
			break
		}

		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return newDroplet, err
	}

	_, _, err = client.FloatingIPActions.Assign(context.TODO(), "159.203.149.210", droplet.ID)

	if err != nil {
		return newDroplet, err
	}

	return newDroplet, nil
}

func main() {
	bytes, _ := ioutil.ReadFile(os.Args[1])

	settings := new(Settings)
	json.Unmarshal(bytes, &settings)

	token := oauth2.Token{AccessToken: settings.Digitalocean.AccessToken}
	source := oauth2.StaticTokenSource(&token)
	oauthClient := oauth2.NewClient(oauth2.NoContext, source)
	client := godo.NewClient(oauthClient)

	http.HandleFunc("/github", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		defer request.Body.Close()

		fmt.Printf("recieved request\n")

		if request.Method == "POST" {
			payload, err := github.ValidatePayload(request, []byte(settings.Github.Secret))

			if err == nil {
				githubEvent, parseErr := github.ParseWebHook("push", payload)

				if parseErr != nil {
					panic(parseErr)
				}

				switch event := githubEvent.(type) {
				case *github.PushEvent:
					if *(event.Repo.FullName) == "troisio/jivecakeapi" && *(event.Ref) == "refs/heads/master" {
						fmt.Printf("start troisio/jivecakeapi %v+v\n", time.Now())

						go func() {
							droplet, apiErr := swapAPIDroplet(client, event)

							if apiErr == nil {
								fmt.Printf("jivecakeapi built %v\n", time.Now())
							} else {
								if droplet == nil {
								} else {
									client.Droplets.Delete(context.TODO(), droplet.ID)
								}

								fmt.Printf("%+v\n", apiErr)
							}

							fmt.Printf("end troisio/jivecakeapi %v+v\n", time.Now())
						}()
					} else if *(event.Repo.FullName) == "troisio/jivecakehtml" && *(event.Ref) == "refs/heads/master" {
						fmt.Printf("start troisio/jivecakehtml %v+v\n", time.Now())

						go func() {
							droplet, swapError := swapHTMLDroplet(client, event)

							if swapError == nil {
								fmt.Printf("jivecakehtml built %v\n", time.Now())
							} else {
								if droplet != nil {
									client.Droplets.Delete(context.TODO(), droplet.ID)
								}

								fmt.Printf("%+v\n", swapError)
							}
						}()
					}
				}
			}
		}
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}
