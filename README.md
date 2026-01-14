# Introduction 
This repository holds a Golang prometheus exporter based on the Habanalabs ML library.


# Golang prometheus exporter
## Testing locally
Note requires access to the lamatriz dockerhub org
```
********
go-hlml and habana-metric exporter should reside in the same directory !!!
********
$ docker run -it --privileged --network=host -v /dev:/dev habana-metric-exporter:0.6.1 --port <PORT_NUMBER> (default port number 41611)

# In other window
$ curl localhost:41611/metrics
```


## Building Golang Docker container

### Environment set up
* Install Docker v20+
    - https://docs.docker.com/engine/install/ubuntu/ 

### Build metric-exporter
* Clone metric-exporter repo and move to the dev branch
    - `git clone --recursive https://github.com/habanaai/metric-exporter.git`
    - `cd metric-exporter`
    - `git checkout dev`
    
* Modify metric-exporter/scripts/setup-env.sh to contain your github token and your installed Golang version.
		
        #!/usr/bin/env bash
		set -eu
		
		export SYNAPSE_VERSION="1.1.0" 		# Replace this value with your desired Synapse version
		export SYNAPSE_BUILD="614"		# Replace this value with your desired Synapse build number
		export SYNAPSE_DIST="ubuntu18.04"	# Replace this with your desired dist version
		
		export GITHUB_BRANCH="dev"		# Replace this value with your desired metric-exporter branch
		export GITHUB_TOKEN=${GITHUB_TOKEN:-""} # Add your key here
		export GOHLML_BRANCH="dev"		# Replace this value with your desired gohlml branch
		
		export DOCKER_IMG_NAME="habanaai/metric-exporter"
		export DOCKER_IMG_TAG="0.6.1"		# This is the user defined tag for the generated container
		
		export GO_VERSION="1.17.3" # Replace with version of Golang currently installed on host
		
* Build gohlml module
	- `go env -w GOPRIVATE=github.com/habanaai`
	- `GO111MODULE=on go get github.com/habanaai/gohlml@dev` 
	- `GO111MODULE=on go build -mod=readonly` 
* Build docker image
    - `cd build/docker`
    - `bash docker-build.sh local  # build using a local copy`
    - `bash docker-build.sh remote # build using a copy from github`



## Running Golang exporter 
The kubernetes daemonset is setup to run in the monitoring namespace.  To run the daemonset, service and serviceMonitor you can run:

    cd go
    kubectl apply -f manifests

## WARNING
Note that the Prometheus-operator must be configured with the correct namespaces or else metrics from other namespaces won't be collected.  Note the kube-prometheus in La Matriz github properly has the habana namespace set

