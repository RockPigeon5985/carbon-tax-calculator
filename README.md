# carbon-tax-calculator

### Protobuf
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```
### GRPC
```
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
NOTE: you need to set the /go/bin directory in your path.
Like this or whatever your go directory lives.
```
PATH="${PATH}:${HOME}/go/bin"
```

## Install the package dependencies

### protobuffer package
```
go get google.golang.org/protobuf
```

### grpc package
```
go get google.golang.org/grpc
```

## Installing Prometheus
Installing Prometheus golang client
```
go get github.com/prometheus/client_golang/prometheus
```

Install Prometheus natively on your system
```
GO111MODULE=on
go install github.com/prometheus/prometheus/cmd/...
prometheus --config.file=your_config.yml
```

## Installing grafana
```
docker run -d -p 3000:3000 --name=grafana grafana/grafana-enterprise
```