
## opentracing-go-rpc-example

install zipkin
```
docker run --name zipkin -d -p 9411:9411 openzipkin/zipkin
```

run server
```
cd server 
go run main.go
```

run client
```
cd cliet 
go run main.go
```


