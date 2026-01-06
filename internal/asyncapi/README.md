# Asyncapi

## Getting started

1. Add comments to your API source code server

```go
// @title Notifyer
// @version 1.0
// @protocol nats
// @url nats:://localhost:4222
func main() {
	flag.Parse()
	var command = flag.Arg(0)

	if command != "" {
		if command == "doc" {
			_, filename, _, _ := runtime.Caller(0)
			asyncapi.Gen(filename, OutFileDoc)
		} else {
			systemd.RunCommand(command, ServiceName, ConfFile)
		}
	}
}
```

2. Add coments pub, sub

```go
// PublishUserCreated publishes a user created event
// @type pub
// @name user.created
// @summary User Created Event
// @description Publishes an event when a new user is created
// @payload UserCreatedEvent
return s.nc.Publish("user.created", data)
```

3. Generate command

```sh
go run ./cmd/notifyer doc -o ./doc/notifyer.yaml
```

## General API Info

| annotation | description                                                | example                        |
| ---------- | ---------------------------------------------------------- | ------------------------------ |
| title      | **Required.** The title of the application.                | // \@title Swagger Example API  |
| version    | **Required.** Provides the version of the application API. | // \@version 1.0                |
| protocol   | protocol                                                   | // \@protocol nats              |
| url        | url server                                                 | // \@url nats:://localhost:4222 |

## API Operation

| annotation  | description                                 | example                                                |
| ----------- | ------------------------------------------- | ------------------------------------------------------ |
| Name        | Name topic                                  | // \@name {ownerId}.notify.get                          |
| Type        | Type channel - Pub, sub                     | // \@type Pub                                           |
| Description | A short description of the application.     | // \@description This is a sample server celler server. |
| Summary     | A short summary of what the operation does. | // \@summary This is a sample                           |
| Payload     | Payload data                                | // \@payload notifyer.GetPayload                        |
| Response    | As same as `success` and `failure`          | // \@response notifyer.Notify                           |
