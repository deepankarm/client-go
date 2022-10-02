# Jina Golang Client

### Install

```bash
go get github.com/deepankarm/client-go
```

### Basic Usage

```go
package main

import (
	"fmt"

	"github.com/deepankarm/client-go"
	"github.com/deepankarm/client-go/docarray"
	"github.com/deepankarm/client-go/jina"
)

// Create a Document
func getDoc(id string) *docarray.DocumentProto {
	return &docarray.DocumentProto{
		Id: id,
		Content: &docarray.DocumentProto_Text{
			Text: "Hello world. This is a test document with id:" + id,
		},
	}
}

// Create a DocumentArray with 3 Documents
func getDocarray() *docarray.DocumentArrayProto {
	return &docarray.DocumentArrayProto{
		Docs: []*docarray.DocumentProto{getDoc("1"), getDoc("2"), getDoc("3")},
	}
}

// Create DataRequest with a DocumentArray
func getDataRequest() *jina.DataRequestProto {
	return &jina.DataRequestProto{
		Data: &jina.DataRequestProto_DataContentProto{
			Documents: &jina.DataRequestProto_DataContentProto_Docs{
				Docs: getDocarray(),
			},
		},
	}
}

// Generate `DataRequest`s with random DocumentArrays
func generateDataRequests() <-chan *jina.DataRequestProto {
	requests := make(chan *jina.DataRequestProto)
	go func() {
		// Generate 10 requests
		for i := 0; i < 10; i++ {
			requests <- getDataRequest()
		}
		defer close(requests)
	}()
	return requests
}

// Custom OnDone callback
func OnDone(resp *jina.DataRequestProto) {
	fmt.Println("Got a successful response!")
}

// Custom OnError callback
func OnError(resp *jina.DataRequestProto) {
	fmt.Println("Got an error in response!")
}

func main() {
    	// Create a HTTP client (expects a Jina Flow with http protocol running on localhost:12345)
	HTTPClient, err := client.NewHTTPClient("http://localhost:12345")
	if err != nil {
		panic(err)
	}
    
    	// Send requests to the Flow
	HTTPClient.POST(generateDataRequests(), OnDone, OnError, nil)
}

```



### Examples


| Example |  |
| :---   | ---:  |
| [gRPC](examples/grpc/README.md) | Stream requests using gRPC Client |
| [HTTP](examples/http/README.md) | Stream requests using HTTP Client |
| [WebSocket](examples/websocket/README.md) | Stream requests using WebSocket Client |
| DocArray usage | Example usage of DocArray (TODO) |


### Gotchas

##### Directory structure

```bash
.
├── client.go                   # Client interface
├── docarray                    # docarray package
│   ├── docarray.pb.go          # generated from docarray.proto  
│   └── json.go                 # custom json (un)marshaler for few fields in docarray.proto
├── grpc.go                     # gRPC client
├── http.go                     # HTTP client
├── jina                        # jina package
│   ├── jina_grpc.pb.go         # generated from jina.proto
│   ├── jina.pb.go              # generated from jina.proto
│   └── json.go                 # custom json (un)marshaler for few fields in jina.proto
├── protos
│   ├── docarray.proto          # proto file for DocArray
│   └── jina.proto              # proto file for Jina
├── scripts
│   └── protogen.sh             # script to Golang code from proto files
└── websocket.go                # WebSocket client
```

- `scripts/protogen.sh` generates the Golang code from the protos. Each proto generates code in a separate package. This is to avoid name clashes.

- `jina/json.go` and `docarray/json.go` are custom json (un)marshalers for few fields in `jina.proto` and `docarray.proto` respectively. 

- `client.go` defines the `Client` interface. This is implemented by `grpc.go`, `http.go` and `websocket.go`.


#### Jina version 

TODO

