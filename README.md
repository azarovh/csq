# csq
Client-Server application with a message queue

RabbitMQ was chosen as a message broker. To configure the application see config.yaml

### Building

Working Go environment 1.18 and Docker are required to run the app

```shell
git clone https://github.com/azarovh/csq.git
cd csq/
go build
```

### Running

Running instance of RabbitMQ is required for the app to communicate. The easiest way is to run prebuild official Docker image in a separate shell

```shell
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.10-management
```

To run the server use:
```shell
./csq -mode=server
```
By default the result will be written to result.txt file

Any number of clients can be run in a separate shells:
```shell
./csq -mode=client
```
Client will read the data from input.txt in a special format
