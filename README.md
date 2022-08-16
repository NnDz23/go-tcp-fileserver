
# go-tcp-fileserver

A simple fileserver in which clients subscribe to channels on the server and are available to receive files sent to it by other clients.




## Build the server

To build the server, from the main directory

```bash
  cd server
  go build
```

## Start the server

To start the server, from the main directory

```bash
  server/server start
```
This should start the tcp fileserver and the fileserver admin API.

## Build the client

To build the client, from the main directory

```bash
  cd client
  go build
```

## Subscribe to a channel with a client

To subscribe to a channel on the server, from the main directory

```bash
  client/client subscribe -c channelName
```
This should subscribe the client to a channel with name channelName, if channel does not exists it gets created.
Client gets prompted for a confirmation on whether allow incoming files to overwrite existing ones or not.
Incoming files get stored in a dir named files/channelName.

## Send files to a channel

To send files to a channel on the server, from the main directory

```bash
  client/client send -c channelName -f file/route.ext
```
This should send the file specified to the channel with name channelName. If file specified is not valid(route specified is a dir or does not have the file extension) it prevents any request from happening.
    