## jivecake infrastructure files


#### Start the continuous delivery server

```sh
docker run -it -d -p 9000:8080 -v $(pwd)/continuousdelivery:/root/continuousdelivery golang:1.6.4 /bin/bash -c "go build -o /root/continuousdelivery/continuousdelivery /root/continuousdelivery/continuousdelivery.go && /root/continuousdelivery/continuousdelivery $DIGITALOCEAN_API_KEY"
```