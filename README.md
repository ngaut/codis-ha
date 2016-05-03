
Get & Compile:
---------------

	go get github.com/left2right/codis-ha


	cd codis-ha


	go build

Usage:
---------------

	Usage:

		codis-ha sentinel   [--server=S]  [--logLevel=L]

		codis-ha latency  	[--server=S]  [--logLevel=L] [--quiet]

	Options:

		-s S, --server=S                 Set api server address, default is "localhost:18087".

		-l L, --logLevel=L               Set loglevel, default is "info".

		-q, --quiet            			 Set latency output less information without slot latency.

Example:
---------------

	codis-ha sentinel -s localhost:18087

	codis-ha latency -s localhost:18087


