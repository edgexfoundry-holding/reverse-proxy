# EdgeX Foundry Security Services Implemented with Go
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security services.
The security service will need KONG ( https://konghq.com/) and Vault (https://www.vaultproject.io/) to be started first. Make sure they are running and the edgexsecurity will check their status///. 


## Install and Deploy

1. Make sure KONG is up and running
2. Make sure Vault is up and running
3. Build edgexsecurity service with the command below
```
go get github.com/edgexfoundry/edgexsecurity
cd edgexsecurity/core
go buld -o edgexsecurity
```
4. Create res folder in the same folder as executable and copy configuration.toml
5. Modify the parameters in the configuration.toml file. Make sure the information for the KONG service, Vault service and Edgex microservices are correct
6. Run the edgexsecurity service with the command below
```
./edgexsecurity init=true
```
7. Use command below for more options
```
./edgexsecurity -h
```

## Features
- Reverse proxy for the existing edgex microservices
- Account creation & JWT authentication for existing services


## Usage
```
# initialize reverse proxy 
./edgexsecurity init=true

#reset reverse proxy
./edgexsecurity reset=true

#create account and return JWT for the account 
./edgexsecurity userddd=guest

#delete account
./edgexsecurity userdel=guest

# to access exisitng microservices API like ping service of command microservice
```
use JWT as query string 
curl -k -v https://kong-ip:8443/command/api/v1/ping?jwt= <JWT from account creation>
or use JWT in HEADER
curl -k -v https://kong-ip:8443/command/api/v1/ping -H "Authorization: Bearer <JWT from account creation>"

``` 

## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)