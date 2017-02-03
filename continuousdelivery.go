package main

import (
  "fmt"
  "os"
  "bytes"
  "strings"
  "os/exec"
  "io/ioutil"
  "text/template"
  "time"
  "log"
  "net/http"
  "encoding/json"
  "github.com/digitalocean/godo"
  "golang.org/x/oauth2"
)

type GithubWebHook struct {
  After string

  Repository struct {
    Clone_url string
    Name string
    Full_name string
  }
}

func createApiDroplet(client *godo.Client, installScript string, currentTime int32) *godo.Droplet {
  createRequest := godo.DropletCreateRequest{
    Name: fmt.Sprintf("jivecakeapi-%v", currentTime),
    Region: "nyc1",
    Size: "2gb",
    Image: godo.DropletCreateImage{
      Slug: "centos-7-x64",
    },
    SSHKeys: []godo.DropletCreateSSHKey{
      godo.DropletCreateSSHKey{ID: 1726114, Fingerprint: "65:d9:a9:c6:a9:0d:32:84:25:6b:18:98:51:bd:45:ba"},
    },
    IPv6: true,
  }

  newDroplet, _, _ := client.Droplets.Create(&createRequest)
  droplet, _, _ := client.Droplets.Get(newDroplet.ID)

  for len(droplet.Networks.V4) < 1 {
    droplet, _, _ = client.Droplets.Get(newDroplet.ID)
  }

  ip, _ := droplet.PublicIPv4()

  errors := [6]error{}

  errors[0] = exec.Command(
    "ssh",
    "-o", "StrictHostKeyChecking=no",
    fmt.Sprintf("root@%s", ip),
    "bash", "-s", string(installScript),
  ).Run()
  errors[1] = exec.Command("scp", "jivecakeapi/ssh/id_rsa", fmt.Sprintf("root@%s:/root/.ssh/id_rsa", ip)).Run()
  errors[2] = exec.Command("scp", "jivecakeapi/settings.yml", fmt.Sprintf("root@%s:/root/settings.yml", ip)).Run()
  errors[3] = exec.Command("scp", "jivecakeapi/docker-compose.yml", fmt.Sprintf("root@%s:/root/docker-compose.yml", ip)).Run()
  errors[4] = exec.Command("scp", "-r", "jivecakeapi/tls", fmt.Sprintf("root@%s:/root/tls", ip),).Run()
  errors[5] = exec.Command(
    "ssh",
    "-o", "StrictHostKeyChecking=no",
    fmt.Sprintf("root@%s", ip),
    "docker-compose",
    "--project-name", "jivecakeapi",
    "--file", "/root/docker-compose.yml",
    "up",
    "-d",
  ).Run()

  hasError := false

  for _, err := range errors {
    if err != nil {
      fmt.Printf("%+v\n", err)
      hasError = true
    }
  }

  if !hasError {
    /*block until droplet ready*/

    assign, _, _ := client.FloatingIPActions.Assign("45.55.105.84", droplet.ID)

    fmt.Printf("%+v\n", assign)
  }

  return droplet
}

func main() {
  token := oauth2.Token{AccessToken: os.Args[1]}
  source := oauth2.StaticTokenSource(&token)
  oauthClient := oauth2.NewClient(oauth2.NoContext, source)
  client := godo.NewClient(oauthClient)

  installScriptBytes, _ := ioutil.ReadFile("centos7/install.sh")
  centosInstallScript := string(installScriptBytes)

  http.HandleFunc("/github", func(writer http.ResponseWriter, request *http.Request) {
    defer request.Body.Close()

    //Secure this endpoint
    //https://developer.github.com/webhooks/securing/
    //https://gist.github.com/rjz/b51dc03061dbcff1c521

    decoder := json.NewDecoder(request.Body)
    webhook := new(GithubWebHook)
    decoder.Decode(&webhook)

    if webhook.Repository.Full_name == "troisio/jivecakeapi" {
      template, _ := template.ParseFiles("jivecakeapi/install-template.sh")
      var bytes bytes.Buffer
      template.Execute(&bytes, webhook)

      installScript := strings.Join([]string{centosInstallScript, bytes.String()}, "\n\n")

      fmt.Printf("%v: Installing %s %s\n", time.Now(), webhook.Repository.Full_name, webhook.After)

      droplet := createApiDroplet(client, installScript, int32(time.Now().Unix()))
      ip, _ := droplet.PublicIPv4()

      fmt.Printf("%v: Installation complete %v\n\n", time.Now(), ip)
    }
  })

  log.Fatal(http.ListenAndServe(":8080", nil))
}