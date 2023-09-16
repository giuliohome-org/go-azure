docker build -t giuliohome/go-azure:v1.1.6 .
docker run --env-file ./env.secret -it giuliohome/go-azure:v1.1.6 