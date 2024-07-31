FROM golang:1.21.1-bookworm
WORKDIR /server
COPY go.mod ./
COPY go.sum ./
RUN go mod download 
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . /server
RUN CGO_ENABLED=1 go build -o bin/server cmd/server/main.go
CMD ./bin/server

# run this command to build the image
# docker build -t racepicks-backend-app .
# run this command to test the container
# docker run -p 8888:8080 racepicks-backend-app
# netstat -aon | findstr 8080
# taskkill /PID xxxx /F


#deloyment to GCP
# gcloud auth login
# -- Tag the image with the registry name
# docker tag app gcr.io/racepicks-backend/app
# -- give docker access to the registry
# gcloud auth configure-docker
# -- Push the image to the registry
# docker push gcr.io/racepicks-backend/app

# -- Googles Cloud Run service

# -------------------------------------------------------------------------------- Start Instructions Backend --------------------------------------------------------------------------------
# Run these commands from ther roots of backend project
# gcloud config set project racepicks-backend

# docker build -t racepicks-backend-app .
# docker run -p 8888:80 racepicks-backend-app
# docker tag racepicks-backend-app gcr.io/racepicks-backend/racepicks-backend-app
# gcloud auth configure-docker
# docker push  gcr.io/racepicks-backend/racepicks-backend-app

# -------------------------------------------------------------------------------- Start Instructions Frontend --------------------------------------------------------------------------------
# # Run these commands from ther roots of frontend project
# docker build -t racepicks-frontend-app .
# docker tag racepicks-frontend-app gcr.io/racepicks-frontend/racepicks-frontend-app
# gcloud auth configure-docker
# docker push  gcr.io/racepicks-frontend/racepicks-frontend-app