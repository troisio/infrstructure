package main

import (
  "fmt"
  "os"
  "os/exec"
  "io/ioutil"
  "time"
  "log"
  "net/http"
  "encoding/json"
  "github.com/digitalocean/godo"
  "golang.org/x/oauth2"
)

type GithubWebHook struct {
  Ref string
  Before string
  After string

  Head_commit struct {
    Id string
  }

  Repository struct {
    Clone_url string
  }
}

func main() {
  http.HandleFunc("/github", func(writer http.ResponseWriter, request *http.Request) {
    defer request.Body.Close()
    decoder := json.NewDecoder(request.Body)
    webhook := new(GithubWebHook)
    decoder.Decode(&webhook)

    token := oauth2.Token{AccessToken: os.Args[1]}
    source := oauth2.StaticTokenSource(&token)
    oauthClient := oauth2.NewClient(oauth2.NoContext, source)
    client := godo.NewClient(oauthClient)

    currenttime := int32(time.Now().Unix())

    cmd := exec.Command(
      "ssh-keygen",
      "-b", "4048",
      "-f", fmt.Sprintf("%v", currenttime),
      "-t", "rsa",
      "-N", "",
      "-C", fmt.Sprintf("\"jivecakeapi-%v\"", currenttime),
    )

    cmd.Run()

    publicBytes, err := ioutil.ReadFile(fmt.Sprintf("%v.pub", currenttime))

    if err != nil {
      panic(err)
    }

    sshKey, _, _ := client.Keys.Create(&godo.KeyCreateRequest{
      Name: fmt.Sprintf("%v.pub", currenttime),
      PublicKey: string(publicBytes),
    })

    createRequest := godo.DropletCreateRequest{
      Name: fmt.Sprintf("jivecakeapi-%v", currenttime),
      Region: "nyc3",
      Size: "512mb",
      Image: godo.DropletCreateImage{
        Slug: "centos-7-x64",
      },
      SSHKeys: []godo.DropletCreateSSHKey{
        godo.DropletCreateSSHKey{ID: sshKey.ID, Fingerprint: sshKey.Fingerprint},
      },
      IPv6: true,
    }

    newDroplet, _, err := client.Droplets.Create(&createRequest)

    if err != nil {
      panic(err)
    }

    fmt.Fprintf(writer, "%+v\n", newDroplet)
  })

  log.Fatal(http.ListenAndServe(":8080", nil))
}