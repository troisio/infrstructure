### jivecake infrastructure files

#### Start the continuous delivery server

```sh
docker run -it -d -p 9000:8080 -v $(pwd):/root/jivecakeinfrastructure golang:1.6.4 /bin/bash -c "go build /root/jivecakeinfrastructure/continuousdelivery.go && /root/jivecakeinfrastructure/continuousdelivery $DIGITALOCEAN_API_KEY"
```
