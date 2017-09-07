# thetooth.name
[![GoDoc](https://godoc.org/github.com/thetooth/thetooth.name?status.svg)](https://godoc.org/github.com/thetooth/thetooth.name) [![Build Status](https://ci.ameoto.com/api/badges/thetooth/thetooth.name/status.svg)](https://ci.ameoto.com/thetooth/thetooth.name) 💩

Open source 12 factor personal blog

## Building
```
go get github.com/thetooth/thetooth.name
cd $GOPATH/github.com/thetooth/thetooth.name
dep ensure
go build -o server main.go
```
## Usage
```
Usage of ./server:
  -image_dir="images/": Where you keep images
```