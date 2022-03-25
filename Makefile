# Docker
NAME := czaring/image-upload
TAG := 0.0.1
IMG := ${NAME}:${TAG}
LATEST := ${NAME}:latest
# postman
POSTMAN_DIR = ./docs
POSTMAN_FILE = upload.postman_collection.json

default: build push

all: clean_rights clean build push

build:
	@echo "Build the docker image"
	@docker build -t ${IMG} .

push:
	@echo "Uploading Containerized Image Uploader Service ========== "
	@docker push ${IMG}

clean_rights:
	@echo "Add access rights to the image cleaner"
	chmod a+x ./clean.sh

clean:
	@echo "Remove any existing image"
	./clean.sh

token:
	@echo "Set token in environment variable"
	@export TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDgxMjUwMTAsImV4cCI6MTY0ODY0MzQ0MCwiYXVkIjoid3d3LmJyYW5rYXMuY29tIiwic3ViIjoiaW1hZ2V1cGxvYWRlcnNlcnZpY2UiLCJGaXJzdE5hbWUiOiJDemFyIiwiTGFzdE5hbWUiOiJNYXBhbG8iLCJFbWFpbCI6Im1jemFybWF5bmVAZ21haWwuY29tIiwiUHVycG9zZSI6IkdvIEFzc2Vzc21lbnQgRXhhbSJ9.7kmFTDoujeQ8EhaC4crAdpAVvsUducMjcc1CEXIYLak
	@echo "Done!"
	# echo $TOKEN #you can use this to confirm if variable has been set

run:
	@echo "Start the application server"
	go run main.go

run_postman:
	@echo "Run the postman collection"
	@newman run ${POSTMAN_DIR}/${POSTMAN_FILE}

