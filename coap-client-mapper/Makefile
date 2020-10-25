.PHONY: all clean

all: coap-time-client
	docker build . -t bruk/coap-time-client
	docker push bruk/coap-time-client

# for testing build only the binary
test: coap-client-mapper
	scp coap-client-mapper ubuntu@kubeedge_edge-1:/home/ubuntu

coap-client-mapper: main.go clean
	go build -o coap-client-mapper main.go
	
clean:
	rm -f coap-client-mapper
	docker rmi -f bruk/coap-time-client