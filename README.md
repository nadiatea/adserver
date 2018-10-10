## Synopsis

The goal of this project is to propose an implementation of a small adserver. The adserver search the campaign with the best price (upper ?) that match with the device type, the placement and the country.
An ad server log allow to log all successed search and fetch the sum of price of matching results by placement and campaign or in total.

## Code Example

```bash
docker-compose up
```
## Installation

Provide code examples and explanations of how to get the project.

### Build the project

`cd adserver && GOOS=linux CGO_ENABLED=0 go build -a -ldflags "-s -w -X 'main.buildTag=' -X 'main.buildDate=2018/10/09 12:05:35'" -v -o adserver`

`cd adserverlogs && GOOS=linux CGO_ENABLED=0 go build -a -ldflags "-s -w -X 'main.buildTag=' -X 'main.buildDate=2018/10/09 12:05:35'" -v -o adserverlogs`


## API Reference

### Ad server

#### POST /ad?placement=[placementID]

Search the best campaign (with the best price) that match with the placement id, the country and the device type

##### Parameters

Name |	Type  |	Type param  |	Description | Example value
-----|--------|-------------|-------------|--------------
placement | string | query | ID of a placement | 3946ca64ff78d93ca61090a437cbb6b3
search request | json objet | body | search request | ```json {  "country": "FRA",  "device": "MOBILE"}```

##### Responses

Code | Description | Example Value
-----|-------------|--------------
200	| Search is successfull, please find the best in the result. | ```json {
    "campaign":"9c0abe51c6e6655d81de2d044d4fb194",
    "content":{
        "title":"Fumer tue",
        "description":"Fumer c'est mal",
        "landing":"http://www.tabac-info-service.fr/"
    }
}```
400	| The provided data do not meet requirements | ```json {
  "error": "string"
}```
404	| Data not found | ```json {
  "error": "string"
}```


### Ad server logs

#### GET /sum

Fetch the sum of price foreach resulting campaign match

##### Responses

Code | Description | Example Value
-----|-------------|--------------
200	| Sum of matching result campaigns. | 1248.3
404	| Data not found | ```json {
  "error": "string"
}```

#### GET /sum_placement?placement=[placementID]
Fetch the sum of price foreach resulting campaign match for a placement ID
##### Parameters
Name |	Type  |	Type param  |	Description | Example value
-----|--------|-------------|-------------|--------------
placement | string | query | ID of a placement | 3946ca64ff78d93ca61090a437cbb6b3

##### Responses
Code | Description | Example Value
-----|-------------|--------------
200	| Sum of matching result campaigns for the given placement. | ```json {
"d0f631ca1ddba8db3bcfcb9e057cdc98":10,
"9c0abe51c6e6655d81de2d044d4fb194":6.9
} ```
400	| The provided data do not meet requirements | ```json {
  "error": "string"
}```
404	| Data not found | ```json {
  "error": "string"
}```

## Tests

```bash
go test adServer_test.go
```
