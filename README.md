# Ladon
> Ladon was the serpent-like dragon that twined and twisted around the tree in the Garden of the Hesperides and guarded the golden apples

![Actions](https://github.com/grupokindynos/ladon/workflows/Ladon/badge.svg)
[![codecov](https://codecov.io/gh/grupokindynos/ladon/branch/master/graph/badge.svg)](https://codecov.io/gh/grupokindynos/ladon)
[![Go Report](https://goreportcard.com/badge/github.com/grupokindynos/ladon)](https://goreportcard.com/report/github.com/grupokindynos/ladon) 
[![GoDocs](https://godoc.org/github.com/grupokindynos/ladon?status.svg)](http://godoc.org/github.com/grupokindynos/ladon)

Ladon Microservice API to purchase vouchers from PolisPay

## Deploy

#### Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/grupokindynos/ladon)

#### Docker

To deploy to docker, simply pull the image
```
docker pull kindynos/ladon:latest
```
Create a new `.env` file with all the necessary environment variables defined on `app.json`

Run the docker image
```
docker run -p 8080:8080 --env-file .env kindynos/ladon:latest 
```

## Building

To run Ladon from the source code, first you need to install golang, follow this guide:
```
https://golang.org/doc/install
```

To run Ladon simply clone de repository:

```
git clone https://github.com/grupokindynos/ladon 
```

Install dependencies
```
go mod download
```

Make sure the port is configured under en enviroment variable `PORT=8080`


## API Reference

@TODO

## Testing

Simply run:
```
go test ./...
```

## Contributing

To contribute to this repository, please fork it, create a new branch and submit a pull request.
