# RESTful API Sample
http://localhost:8080/

## REST API endpoints:
```
GET http://localhost:8080/users/
POST http://localhost:8080/users/
GET http://localhost:8080/user/{id:[0-9]+}
```

### Examples
http://localhost:8080/

http://localhost:8080/users/

http://localhost:8080/user/2

Test REST API with curl
```
curl -H "Content-Type: application/json" -v http://localhost:8080/users/

[{"ID":2,"Name":"Doe","Age":22,"City":"Vancouver"},{"ID":1,"Name":"John","Age":31,"City":"New York"}]
```

```
curl -H "Content-Type: application/json" -v http://localhost:8080/user/2

{"id":2,"name":"Doe","age":22,"city":"Vancouver"}
```

## Installation
```
  go get github.com/rnidev/rest-api-sample
```

## Run in Docker
- Docker Compose
```
docker-compose up
```

## Makefile
- Build binary to ./build/
```
make build
```
- Run tests
```
make test
```
- Clean up tests and binary files
```
make clean
```
- Build For Linux
```
make build-linux
```

## Package Used
[miniredis](https://github.com/alicebob/miniredis): Pure Go Redis test server, used in Go unittests.

[redigo](github.com/gomodule/redigo): Redigo is a Go client for the Redis database.

[gorilla/mux](https://github.com/gorilla/mux) A powerful HTTP router and URL matcher for building Go web servers
## Go version
```1.12.17```

## License
This project is licensed under the terms of the MIT license.

