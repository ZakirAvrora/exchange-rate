# Plata Exchange Rate Assignment

## Prerequisites
Requires `docker-compose`

## Launching the application
- Clone the repository
- To start the application, write `docker-composer up -d` on the command line

## Service
- By default app runs on port: `8080`

## Endpoints
To get information regarding API endpoints and their specifications, do following:
- visit `localhost:8080/swagger/index.html`
- lookup Swagger (OpenAPI) API specification file in `./docs`

## Constraints:
There are several constaraints regarding currencies that are supported, since I am using `https://exchangeratesapi.io/` 
for external service for getting exchange rates and its free pricing option:
- Base currency that only supported is `EUR`
- There is also list of supported of secondary currencies for free pricing. However I have limited it to the following ten currencies:
`BTC, MXN, USD, BYR, AED, KZT, RUB, XAU, XAG, LYD` to make it more deterministic.
