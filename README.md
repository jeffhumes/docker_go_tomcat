DOCKER_GO

This project is intended to make setting up a docker instance easier by adding configuration
files and settings that can be changed at build or run time.

1. Edit go.conf
	- update the instance name variable
	- update the port numbers to meet your requirements

2. Run update-ports
	- ./go update-ports
		-- this will update all of the necessary files with the updated port numbers

3. Run the docker instance build
	- ./go build
		- this will run all of the commands in Dockerfile to create the instance

4. Run the 'go' script
	use this command: ./go start

