 HFRD api server
 ===============

* [Introduction](#introduction)
* [How to build and run api server](#how-to-build-and-run-api-server)
    * [Prerequisites](#prerequisites)
    * [Build and run hfrd api server binary](#build-and-run-hfrd-api-server-binary)
    * [Build and run hfrd api server locally with Docker](#build-and-run-hfrd-api-server-locally-with-docker)
        * [Build hfrd docker image](#build-hfrd-docker-image)
        * [Run hfrd api server docker container](#run-hfrd-api-server-docker-container)
* [api server authentication with JWT](#api-server-authentication-with-jwt)
    * [Steps to get JWT token from IBM Cloud](#steps-to-get-jwt-token-from-ibm-cloud)

## Introduction
HFRD api server is implemented in Go language. It exposes RESTful API
to testers. Testers can call the RESTful APIs to create blockchain
networks in Bluemix staging/producation and on premise IBM Blockchain
Platform. After a network is created, tester should be able to retrieve
all organizations' connection profiles of the blockchain network. With
connection profiles, testers are able to create channels/submit transactions
to the blockchain network, no matter where the blockchain network is
created.

Testers can also submit test scripts to api server, api server will trigger
Jenkins jobs to execute tests.

Swagger doc and api server endpoint are as below. The two endpoints serve with
latest code in master branch of this repository

- hfrd api server swagger doc:  http://<<server ip or name>>:9595/
- hfrd api server endpoint:     http://<<server ip or name>>:8080/

## How to build and run api server
### Prerequisites
api server should work on MacOS and Linux

To run the golang api server, clone this project under
&lt;somedir&gt;/src,

        git clone https://github.com/litong01/hfrd

then include &lt;somedir&gt; in your GOPATH
environment, then change directory to hfrd

        cd $GOPATH/src/hfrd

### Build and run hfrd api server binary
From hfrd root directory, issue

        make api
will generate the api binary in **.build/bin/hfrdserver**

To config the server, please use the hfrd/api/var/config.json as an example,
create a subdirectory named var in where the api binary is. Make
changes to that file according to your configuration. Then run the
following command:

        ./hfrdserver

The program is configured by config.json file. You should change the
baseUrl to reflect your own environment.

> Note: run "make api-clean" to delete old api binary file

### Build and run hfrd api server locally with Docker

#### Build hfrd docker image

From hfrd root directory, issue

        make api-docker
will generate the docker image *hfrd/server:latest*

> Note: run "make api-docker-clean" to delete old api docker image

#### Run hfrd api server docker container

Make changes to config.json according to your configuration.
Then, from hfrd root directory, issue

        docker run -idt -v $(pwd)/api/var:/opt/hfrd/var hfrd/server

## api server authentication with [JWT](https://jwt.io/)
If auth is enabled for api server and auth type is set to "jwt", then hfrd users
should get a jwt token from IBM Cloud and put the jwt token in http request header
like below

        Authorization: Bearer <jwt token>
Otherwise hfrd will return 401 Unauthorized

### Steps to get JWT token from IBM Cloud
1. Login https://console.bluemix.net/ with your IBM id
2. Open menu Manage->Security->[Platform API](https://console.bluemix.net/iam/#/apikeys)
3. Create a new API key and copy the API key secret which will be used later
4. Get JWT from IBM Cloud with
```bash
curl -X POST "https://iam.bluemix.net/oidc/token" \
-H "Content-Type:application/x-www-form-urlencoded" \
-H "Accept: application/json" \
-d "grant_type=urn:ibm:params:oauth:grant-type:apikey\
&apikey=<YOUR_APIKEY>"
```
Replace `<YOUR_APIKEY>` with the api key you received from step 3.

The response body of step 4 is a JSON string. Extract `access_token` from the JSON string
and take it as jwt token. The default expiration for the jwt token is 1 hour
