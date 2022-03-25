- [Requirements](#requirements)
- [Value Proposition](#value-proposition)
- [Overview](#overview)
- [Logical Architecture](#logical-architecture)
- [Use Case](#use-case)
- [Components](#components)
- [Security](#security)
  - [Data at Rest & Data in Transit](#data-at-rest--data-in-transit)
  - [Certificate, Password, & Key Management](#certificate-password--key-management)
  - [IAM](#iam)
- [Cost](#cost)
- [Availability Model](#availability-model)
- [API Reference](#api-reference)
  - [Get Auth Token](#get-auth-token)
  - [Upload Image](#upload-image)
    - [Case: Happy Path](#case-happy-path)
    - [Case: No file has been selected](#case-no-file-has-been-selected)
    - [Case: Uploaded file > 8MB](#case-uploaded-file--8mb)
    - [Case: Invalid Auth](#case-invalid-auth)
    - [Case: Upload unsupported file type](#case-upload-unsupported-file-type)
    - [Case: Unable to save the record](#case-unable-to-save-the-record)

## Requirements

[Technical Assessment](https://docs.google.com/document/d/1DC9HO2NRzqivgLP1Tl2YMaZGnIZbv9ERZd0H1oi4b0Y/)

## Value Proposition

- Can be used as image sharing platform 

## Overview

Solutions aims to provide users a handy web app server to easily share images 

## Logical Architecture

_see /docs/design.drawio_
![high level design](out/docs/design/logicaldesign.png)

## Use Case

_see /docs/design.puml_
![sequence diagram](out/docs/design/usecasedesign.png)

## Components
_Question: What is the current app lifecycle strategy (change management), patching?_

Current:
| Application | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `Image Upload Service` | `MICROSERVICE` | Contains FE and BE items for the app |

Roadmap:
| Application | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `Web App` | `MICROFRONTEND` | Web interface accessible by users |
| `API Gateway` | `PROXY` | NGINX that routes request made by web app |
| `Image Upload Service` | `BE MICROSERVICE` | Will contain only backend items |

Optional:
| Application | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `ELK` | `Log Management` | Platform (ElasticSearch - logging, LogStash - storage & forwarding, and Kibana - visualization) |
| ` - ` | `Metrics` | What to use for the metrics? |
| `S3` | `Storage` | To replace the hot/temporary folder |
| `Azure KeyVault/Vault Hashicorp` | `Secrets Management` | Will be used for keeping the API Keys |

## Security
### Data at Rest & Data in Transit
| Aspect | Item     | Description | Strategy |
| :-------- | :------- | :------------------------- | :------ |
| Data at Rest| MongoDB | Need to assess the data retention policy and CIA Rating | X |
| Data in Transit | Web App & Image Upload Service | Endpoint Security | Auth Tokens |

### Certificate, Password, & Key Management 
_need to decide on technology to be used_

### IAM
_this is more on the web app access; NPAs as well if data will be stored in the future_

## Cost
_add here a cost breakdown if decided to use S3 and Openshift_

## Availability Model
| System | SLA     | Latency Performance Level | Strategy in Event of Service Loss | Strategy in Event of Site Loss| Replication Strategy|
| :-------- | :------- | :------------------------- | :------ | :------ | :------ |

_to be defined per system; also, consider the operational resilience: recoverability_
- do we have any audit compliance for the max outage or any SLO and SLA with the clients for consideration here?



## API Reference

### Get Auth Token
Set the TOKEN variable `export TOKEN=<some_value>`
```http
  GET /token
```
`RESPONSE: 200 | SUCCESS`
```
"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDgxMjUwMTAsImV4cCI6MTY0ODY0MzQ0MCwiYXVkIjoid3d3LmJyYW5rYXMuY29tIiwic3ViIjoiaW1hZ2V1cGxvYWRlcnNlcnZpY2UiLCJGaXJzdE5hbWUiOiJDemFyIiwiTGFzdE5hbWUiOiJNYXBhbG8iLCJFbWFpbCI6Im1jemFybWF5bmVAZ21haWwuY29tIiwiUHVycG9zZSI6IkdvIEFzc2Vzc21lbnQgRXhhbSJ9.7kmFTDoujeQ8EhaC4crAdpAVvsUducMjcc1CEXIYLak"
```

### Upload Image

```http
  POST /upload
```

#### Case: Happy Path
`REQUEST`
```
--form 'auth="eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDgxMjUwMTAsImV4cCI6MTY0ODY0MzQ0MCwiYXVkIjoid3d3LmJyYW5rYXMuY29tIiwic3ViIjoiaW1hZ2V1cGxvYWRlcnNlcnZpY2UiLCJGaXJzdE5hbWUiOiJDemFyIiwiTGFzdE5hbWUiOiJNYXBhbG8iLCJFbWFpbCI6Im1jemFybWF5bmVAZ21haWwuY29tIiwiUHVycG9zZSI6IkdvIEFzc2Vzc21lbnQgRXhhbSJ9.7kmFTDoujeQ8EhaC4crAdpAVvsUducMjcc1CEXIYLak"' \
--form 'file=@"/Users/czarmayne/Documents/Professional.jpeg"'
```

`RESPONSE: 200 | SUCCESS`
```
{
    "InsertedID": "623db1e8d3b9343d2af41301"
}
```

#### Case: No file has been selected

`RESPONSE: 400 | BAD REQUEST`

```
Error Retrieving the File
```

#### Case: Uploaded file > 8MB

`RESPONSE: 400 | BAD REQUEST`
```
File exceeded the size limit of 8 megabytes
```

#### Case: Invalid Auth 

`RESPONSE: 403 | FORBIDDEN`
```
Access not allowed
```

#### Case: Upload unsupported file type

`RESPONSE: 415 | UNSUPPORTED MEDIA TYPE`

```
Invalid File Type
```

#### Case: Unable to save the record

`RESPONSE: 500 | INTERNAL SERVER ERROR`
```
Error occurred while uploading the file
```