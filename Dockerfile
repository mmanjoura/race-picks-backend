FROM golang:1.21.6-bookworm
WORKDIR /server
COPY go.mod ./
COPY go.sum ./
RUN go mod download 
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . /server
RUN CGO_ENABLED=1 go build -o bin/server cmd/server/main.go
CMD ./bin/server

# run this command to build the image
# docker build -t race-picks-backend-app .
# run this command to test the container
# docker run -p 8888:8080 race-picks-backend-app
# netstat -aon | findstr 8080
# taskkill /PID xxxx /F


#deloyment to GCP
# gcloud auth login
# -- Tag the image with the registry name
# docker tag app gcr.io/race-picks-backend/app
# -- give docker access to the registry
# gcloud auth configure-docker
# -- Push the image to the registry
# docker push gcr.io/race-picks-backend/app

# -- Googles Cloud Run service

# -------------------------------------------------------------------------------- Start Instructions Backend --------------------------------------------------------------------------------
# Run these commands from ther roots of backend project
# gcloud config set project race-picks-backend

# docker build -t race-picks-backend-app .
# docker run -p 8888:80 race-picks-backend-app
# docker tag race-picks-backend-app gcr.io/race-picks-backend/race-picks-backend-app
# gcloud auth configure-docker
# docker push  gcr.io/race-picks-backend/race-picks-backend-app

# -------------------------------------------------------------------------------- Start Instructions Frontend --------------------------------------------------------------------------------
# # Run these commands from ther roots of frontend project
# docker build -t race-picks-frontend-app .
# docker tag race-picks-frontend-app gcr.io/race-picks-frontend/race-picks-frontend-app
# gcloud auth configure-docker
# docker push  gcr.io/race-picks-frontend/race-picks-frontend-app