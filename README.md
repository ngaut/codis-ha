[![Build Status](https://travis-ci.org/ngaut/codis-ha.svg?branch=master)](https://travis-ci.org/ngaut/codis-ha)


Get & Compile:
---------------

	go get github.com/ngaut/codis-ha


	cd codis-ha


	go build

Usage:
---------------

	Usage:

		codis-ha sentinel   [--server=S]  [--logLevel=L]

		codis-ha latency  	[--server=S]  [--logLevel=L]

	Options:

		-s S, --server=S                 Set api server address, default is "localhost:18087".

		-l L, --logLevel=L               Set loglevel, default is "info".

Example:
---------------

	codis-ha sentinel -s localhost:18087

	codis-ha latency -s localhost:18087


