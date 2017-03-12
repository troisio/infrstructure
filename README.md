### jivecake infrastructure files

A few details:

You need to fill out `settings.json` file

You need to load an `id_rsa` as a volume file, since this server will need ssh access to newly created droplets. Naturally, those droplets which are newly created will need to have a public key entry in their respective `~/.ssh/authorized_keys` file.

```sh
docker run \
  -it \
  -p 80:80 \
  --rm \
  -v ~/.ssh/id_rsa:/root/id_rsa \
  -v $(pwd):/root/jivecakeinfrastructure \
  golang:1.8.0 /bin/bash -c \
  "go get golang.org/x/oauth2 github.com/digitalocean/godo github.com/google/go-github/github && go run /root/jivecakeinfrastructure/continuousdelivery.go /root/jivecakeinfrastructure/settings.json"
```