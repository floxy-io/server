# Floxy

Floxy is a SSH private end-to-end service that allows developers to use your own native command line to start a high secure tunneling communication between client-server or http-server. No need of any installations. Floxy offer's two different kind of services:

**client-server:** High secure tunneling connection between two computers, using the common SSH command.

**http-server:** Expose a specific PORT on your desktop or server through a floxy's subdomain (e.g.: https://{subdomain}.web.floxy.io).

### Feature
* No need of installations
* Use of own computer ssh command line (linux and macos), or SSH window's app
* 2 steps to use floxy's services
* High secure ssh tunneling using a random 32 bits password or RSA key

### Usages:
* IoT, allowing communication between one computer and a device
* Access remote computers, using all SSH features
* Direct Transfer of files
* Expose a specific PORT over the internet inside a Floxy's domain

## Configuration's steps:

First you have to generate a captcha's key to use floxy's APIs.

[linux/macos]

1. You have to have installed: curl, ssh.
2. Open your terminal and type:

````
$ curl curl --location --request POST 'http://floxy.io/api/s2s' \
--header 'Content-Type: application/json'
````
response:
````
{
	"server_command":  "ssh -N -R 52443:localhost:$PORT ssh.floxy.io",
	"client_command":  "ssh -N -L localhost:$PORT:52443 ssh.floxy.io",
	"password": "jjTDbbI87hsf345ujnd89odmjyyhbdaq"
}
````
3. Save this payload content, you will use it in the server and local machine. The $PORT variable means: in server machine the port that you will forward request to (e.g.: SSH port, a http service). In the local machine means the port that you will allocate to send a connection to server side.

4. type the response line command on the same terminal and after type the password received on the last request:
````
(server machine)
$ ssh -N -R 52443:localhost:$PORT ssh.floxy.io
password: ***************
````

Now your server is connected on floxy's service. On the client Machine, do the same first and third step, install curl, ssh if needed and run the command sent on **client_command** field:

````
(client machine)
$ ssh -N -L localhost:$PORT:52443 ssh.floxy.io
password: ***************
````

Ready for use! Your local machine is connected to your server machine. 