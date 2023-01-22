Tor Clone (A very, very simple version of Tor)
=============================================


Description
===========
This was a simple project to see if I could recreate a simplified version of tor with its directory and nodes.

It uses Diffie-Hellman to create and share the secret to use with AES encryption. The client has all the keys to encrypt and decrypt each layer. On the other hand, the nodes will only have their respective private key to decrypt and encrypt a part of the encrypted message.

It all works by creating a circuit of TCP connections, client to entry, entry to relay, relay to exit, exit to external source.

To get the all the node's info the client will query the directory, it contains all nodes, and it will randomly select one.

Each node when started will tell the directory, its listening IP and port for the client to use, so it automatically sets up the whole network logistics. 

How to use
==========
You need [docker](https://www.docker.com/products/docker-desktop/) and [docker-compose](https://docs.docker.com/compose/install/) for this to run.

To check if you have them installed:
```bash
docker version
docker-compose version
```

When you have both of them installed, you are good to go! To run type
```bash
docker-compose up -d
```
And the compose will set everything up.

To access it:
- http://localhost:8080 for the client
- http://localhost:8081 for the directory
- http://localhost:8082 for the sample end server


### Client

The client is very straight-forward, it has 3 available modes:
1. Check circuit - The message is optional, it will query an end server in a docker container and tell you where your query came from
2. Search Google - Simply type something to search and press submit
3. Raw URL or Text - You can type a raw URL starting with http or https and watch the magic happen, if the text does not start with http it will simply tell you that nothing was queried

The client sometimes does not show you your result, if that happens you can simply press `refresh response` and everything should work fine, if not refresh the webpage.

### Directory

It is an API server that you can query
- /      → Will tell you if it's alive, used for docker purposes
- /entry → Will return information of a random entry node connection
- /relay → Will return information of a random relay node connection
- /exit  → Will return information of a random exit node connection

### Sample end server

You can access it and will tell you what IP is being used to enter the docker network

### Docker containers

You can see what data is passing through each node by going into the logs of each container. 
Inside the client container you can see what the query is, what the encrypted message is and the decrypted response.
Inside the nodes you will see the encrypted data coming back

In the docker-compose.yml this could be hard to see as you have to go node by node checking which one it went through, that is why I created a docker-compose2.yml, it only has 1 node of each, so you can easily see what is going on.

```bash
docker-compose -f docker-compose2.yml up -d
```


I want to try more nodes
========================
So you want more nodes well, that is actually pretty simple to do, just copy the node you want to add and paste it in the compose

### Entry node


```yaml
  entry-nnn: # Replace nnn with the next number in line, or you can call it whatever you want it is up to you 
    build:
      context: .
      dockerfile: entry/entry.Dockerfile
    environment:
      - PORT=9999  # Optional, port number to listen on you can set it to whatever you want or just remove it, just don't leave it blank just in case
    depends_on:
      directory:
        condition: service_healthy
```

### Relay node 


```yaml
  # This node does not have a set port unlike the one above and it works perfectly

  relay-nnn: # Replace nnn with the next number in line, or you can call it whatever you want it is up to you 
    build:
      context: .
      dockerfile: relay/relay.Dockerfile
    depends_on:
      directory:
        condition: service_healthy
```

### Exit node 


```yaml

  exit-nnn: # Replace nnn with the next number in line, or you can call it whatever you want it is up to you 
    build:
      context: .
      dockerfile: exit/exit.Dockerfile
    environment:
      - PORT=3454   # Optional, port number to listen on you can set it to whatever you want or just remove it, just don't leave it blank just in case
    depends_on:
      directory:
        condition: service_healthy

```

Problems you could have
=======================
### Port collision

In the compose, for you to be able to access the client, directory and end server from your local machine we have to "expose" ports, if you already have something running on ports 8080, 8081 or 8082 (Or all of them at once) you have to change it in the compose

So let's say I have another web server on my local machine running on port 8080, I would open the compose and edit the client definition like so

Original:
```yaml
# ... Client definition 

  ports:
      - "8080:8080" # This is forwarding the connection to port 8080 in local machine, if there was any port collision just change it to "yourport:8080"

# Client definition ...
```

After port change from 8080 to 1234:
```yaml
# ... Client definition 

  ports:
      - "1234:8080" # This is forwarding the connection to port 8080 in local machine, if there was any port collision just change it to "yourport:8080"

# Client definition ...
```

### Dead nodes

When any node dies (they don't have the greatest error checking yet) it does not notify the directory, so that means it will leave dead nodes in the system. When the client requests a new node, neither the directory nor the client checks if this node is alive, which means that the client and all other nodes will be left hanging, waiting for a response, you will notice this has happened because the client stays loading infinitely.

This has a simple solution, which is to simply restart the compose:

```bash
docker-compose restart
```

or if you are using docker-compose2.yml

```bash
docker-compose -f docker-compose2.yml restart 
```