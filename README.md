### jivecake infrastructure files


#### Install

```sh
docker build -t jivecakeinfrastructure .
```

#### Run

```sh
docker run \
  -it \
  -d \
  -p 80:80 \
  --rm \
  --name jivecakeinfrastructure \
  -v /root/.ssh/id_rsa:/root/.ssh/id_rsa \
  -v /root/.ssh/id_rsa.pub:/root/.ssh/id_rsa.pub \
  jivecakeinfrastructure
```