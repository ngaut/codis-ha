
Get & Compile:
---------------

	go get github.com/left2right/codis-ha


	cd codis-ha


	go build

Usage:
---------------

	Usage:

		codis-ha sentinel   [--server=S]

		codis-ha latency  	[--server=S]  [--quiet]

		codis-ha migrate    [--server=S] [--zookeeper=Z] [--kill] [--num=N] [--from=F] [--target=T]
		
		codis-ha version

	Options:

		-s S, --server=S                 Set api server address, default is "localhost:18087".

		-q, --quiet            			 Set latency output less information without slot latency.

		-z Z, --zookeeper=Z          Set zookeeper address, default is "localhost:2181".

		-k, --kill                   Kill the already running migrate tasks.

		-n N, --num=N                The number slot want to move, if the specified number bigger than slot number in from group, then move all slot in from group.
	
		-f F, --from=F 				 Specify the from group id ,where move the slots from.
		
		-t T, --target=T             Specify the target group id ,where move the slots to.

Example:
---------------

	codis-ha sentinel -s localhost:18087

	codis-ha latency -s localhost:18087

	codis-ha migrate    -s localhost:18087 -z localhost:2181 -k 

	codis-ha migrate    -s localhost:18087 -z localhost:2181  -n 3 -f 2 -t 4


