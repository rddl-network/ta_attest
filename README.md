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

The following command will attest a given newline seperated file of Trust Wallet machine IDs to the configured network:
```
./ta --attest-machine-ids-by-file keys.yaml
```

## Configuration
The service needs to be configured via the ```./app.env``` file or environment variables.
A default configuration file is created at first run.
Please adapt it and rerun the application.
