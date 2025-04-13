# Container Bootstrapper
The goal of this project is to create a small app that mocks the necessary dependencies for xfsc services.
It currently supports mocking nats, redis, OPA, Postgres, and Hydra.

## Table of Contents
1. [Installation](#installation)
2. [Usage](#usage)

## Installation
Pull the project and use `go run .` to start the server. Docker daemon needs to be started prior starting this server.  
Use `--config path\to\config.yaml` to load your config. When a config.yaml is loaded, all mentioned images are auto starting.

## Usage
There are three rest endpoints.
>GET /v1/services  
>returns a json body with all containers and their status
>```json
>{
>  "name": container_name,
>  "id": container_id,
>  "image": container_image,
>  "status": container_status
>}
>```

>POST /v1/services  
>expects a json body with the containers to start, if false can be omitted
>```json
>{
>	"nats": true,
>	"redis": true,
>	"opa": true,
>	"postgres": false,
>	"hydra": false
>}
>```

>DEL /v1/services  
>expects a json body with the containers to stop, if false can be omitted
>```json
>{
>	"nats": true,
>	"redis": true,
>	"opa": true,
>	"postgres": false,
>	"hydra": false
>}
>```