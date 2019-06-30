# HFRD PTE Test Package for IBP Starter Plan

## Description
pte-hfrd is a simple POC test case execution package for HFRD and PTE.

## Build Requirements
- docker

## Build Process
1. Create an IBP Starter Plan Instance
2. Copy your IBP Starter Plan org1 and org2 Connection Profile, and the service keys into a new folder named `creds`
  - `creds/apikeys.json`
  - `creds/org1ConnectionProfile.json`
  - `creds/org2ConnectionProfile.json`
3. Build the docker image
  - `docker build --rm -f Dockerfile -t pte-hfrd:latest .`

## Usage
Start the run with
  - `docker run pte-hfrd ./docker-entrypoint.sh`
