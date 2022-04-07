SHELL = /bin/sh

buildAndRun: build run

#runs server bin
run: 
	/home/daniel/Documents/gocms/bin/main

#builds server
build:
	go build -o bin /home/daniel/Documents/gocms/cmd/main.go

#runs clear table util
clear:
	/home/daniel/Documents/gocms/bin/clearTable

#makes sure db is working
test:
	go test ./database

#opens up server config file in text editor
config:
	gedit ./serverConfig.yaml

#connect to cmsdb database using pg.sh
db:
	./pg.sh