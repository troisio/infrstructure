package main

import (
  "context"
  "fmt"
  "os"
  "os/exec"
  "log"
  "io/ioutil"
  "time"
  "net/http"
  "crypto/tls"
  "encoding/json"
  "github.com/digitalocean/godo"
  "golang.org/x/oauth2"
  "github.com/google/go-github/github"
)

type Settings struct {
  Digitalocean struct {
    AccessToken string
  }

  Github struct {
    Secret string
  }
}

type GithubWebHook struct {
  After string
  Ref string

  Repository struct {
    Clone_url string
    Name string
    Full_name string
  }
}

func swapJiveCakeApiDroplet(client *godo.Client, hook *GithubWebHook) *godo.Droplet {
  images, _, _ := client.Images.ListUser(context.TODO(), nil)
  imageId := -1

  for _, image := range images {
      if image.Name == "centos-2gb-nyc1-01-jivecakeapi" {
          imageId = image.ID
          break
      }
  }

  if imageId == -1 {
    fmt.Errorf("Unable to find image centos-2gb-nyc1-01-jivecakeapi in %+v", images)
    return nil
  }

  createRequest := godo.DropletCreateRequest{
    Name: fmt.Sprintf("jivecakeapi-%v", (*hook).After),
    Region: "nyc1",
    Size: "2gb",
    Image: godo.DropletCreateImage{
      ID: imageId,
    },
    SSHKeys: []godo.DropletCreateSSHKey{
      godo.DropletCreateSSHKey{ID: 1726114, Fingerprint: "65:d9:a9:c6:a9:0d:32:84:25:6b:18:98:51:bd:45:ba"},
      godo.DropletCreateSSHKey{ID: 7151393, Fingerprint: "d4:fe:af:89:22:bf:f9:ab:ad:a5:ea:99:20:9d:ed:6d"},
    },
    IPv6: true,
    Monitoring: true,
  }

  newDroplet, _, err := client.Droplets.Create(context.TODO(), &createRequest)

  if err != nil {
    panic(err)
  }

  droplet, _, _ := client.Droplets.Get(context.TODO(), newDroplet.ID)

  for len(droplet.Networks.V4) == 0 || droplet.Status != "active" {
    droplet, _, _ = client.Droplets.Get(context.TODO(), newDroplet.ID)
  }

  ip, _ := droplet.PublicIPv4()

  cmd := exec.Command(
    "ssh",
    "-o", "StrictHostKeyChecking=no",
    "root@" + ip,
    "docker-compose --project-name jivecakeapi --file /root/docker-compose.yml up -d",
  )

  err = cmd.Run()

  if err != nil {
    client.Droplets.Delete(context.TODO(), newDroplet.ID)
    panic(err)
  }

  httpClient := &http.Client{Transport: &http.Transport{
      TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }}

  for i := 0; i < 50; i++ {
    _, err = httpClient.Get("https://" + ip)

    if err == nil {
      break
    }

    time.Sleep(5 * time.Second)
  }

  if err != nil {
    client.Droplets.Delete(context.TODO(), newDroplet.ID)
    panic(err)
  }

  _, _, err = client.FloatingIPActions.Assign(context.TODO(), "45.55.105.84", droplet.ID)

  if err != nil {
    client.Droplets.Delete(context.TODO(), newDroplet.ID)
    panic(err)
  }

  return droplet
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
    defer request.Body.Close()

    if request.Method == "POST" {
      payload, err := github.ValidatePayload(request, []byte(settings.Github.Secret))

      if err == nil {
        webhook := new(GithubWebHook)
        json.Unmarshal(payload, &webhook)

        if webhook.Repository.Full_name == "troisio/jivecakeapi" && webhook.Ref == "refs/heads/master" {
          droplet := swapJiveCakeApiDroplet(client, webhook)

          if droplet != nil {
            fmt.Printf("\n\nJiveCakeAPI Task Complete\n\n%+v\n\n", droplet)
          }
        }
      }
    }
  })

  log.Fatal(http.ListenAndServe(":80", nil))
}