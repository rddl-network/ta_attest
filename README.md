# TrustAnchor attestation service

This service receives http(s)://localhost:8080/firmware requests and responds with a patched firmware of type for the RDDL Network Tasmota ESP32 based solutions. Each response contains another random private key within the firmware. 
The corresponding public key is registered as a TrustAnchor machine ID at RDDL Network.

The firmware is expected to be located at ./tasmota32-rddl.bin.
The latest firmware can be found at https://github.com/rddl-network/Tasmota/releases.

## Building
The service can be build with

```
go build -v ./cmd/ta
```

## Execution
A build service can be executed via ```./ta``` or be run via the following go command without having it previously built
```
go run cmd/ta/main.go
```

## Configuration
The service needs to be configured via the ```./app.env``` file or environment variables. The defaults are
```
PLANETMINT_GO=planetmint-god
PLANETMINT_ACTOR=plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5
PLANETMINT_KEYRING=                                           #is optional
FIRMWARE_ESP32=./tasmota32-rddl.bin
FIRMWARE_ESP32C3=./tasmota32c3-rddl.bin
SERIVE_PORT=8080
SERVICE_BIND=localhost
```
A sample ```./app.env``` file is at ```./app.env.template```