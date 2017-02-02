git clone {{.Repository.Clone_url}} ~/{{.Repository.Name}}
cd ~/{{.Repository.Name}}
git reset --hard {{.After}}

cd docker
docker build -t {{.Repository.Name}} .