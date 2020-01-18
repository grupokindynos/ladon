# Ladon
> Ladon was the serpent-like dragon that twined and twisted around the tree in the Garden of the Hesperides and guarded the golden apples

![Actions](https://github.com/grupokindynos/ladon/workflows/Ladon/badge.svg)
[![codecov](https://codecov.io/gh/grupokindynos/ladon/branch/master/graph/badge.svg)](https://codecov.io/gh/grupokindynos/ladon)
[![Go Report](https://goreportcard.com/badge/github.com/grupokindynos/ladon)](https://goreportcard.com/report/github.com/grupokindynos/ladon) 
[![GoDocs](https://godoc.org/github.com/grupokindynos/ladon?status.svg)](http://godoc.org/github.com/grupokindynos/ladon)

Ladon Microservice API to purchase vouchers from PolisPay

## Deploy

#### Docker

To deploy to docker, simply pull the image
```
docker pull kindynos/ladon:latest
```
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

## Running flags
```
-local
```

Set this flag to run Ladon using the testing Hestia database. Default is false (production mode).
When using this flag you must be running hestia locally on port 8080.

```
-port=xxxx
```

Specifies the running port. Default is 8080.

```
-stop-proc
```

Set this flag to run Ladon without the *processor*.

```
-no-txs
```

Set this flag to avoid publishing txs on the Blockchain but store them on the database.
WARNING: -local flag must be set in order to use this flag.

```
-skip-val
```

Set this flag to skip validations on txs (currently just skipping the minimum amount of confirmations required to process a tx)
WARNING: -local flag must be set in order to use this flag.


## API Reference

Documentation: [API Reference](https://documenter.getpostman.com/view/4345063/SVmySJBd?version=latest)

## Testing

Simply run:
```
go test ./...
```

## Contributing

To contribute to this repository, please fork it, create a new branch and submit a pull request.
