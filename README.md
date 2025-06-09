# go-eventemitter

A simple, lightweight event emitter implementation for Go, inspired by Node.js's EventEmitter.

## Features

- Simple and intuitive API
- Asynchronous event emission
- Support for multiple event listeners
- Support for once listeners
- Zero dependencies

## Installation

```bash
go get github.com/jiyeyuran/go-eventemitter
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/jiyeyuran/go-eventemitter"
)

func main() {
	emitter := eventemitter.New()

	// Add event listener
	emitter.On("greeting", func(data interface{}) {
		fmt.Println("Hello ", data)
	})

	emitter.Once("helloOnce", func(data interface{}) {
		fmt.Println("Hello once ", data)
	})

	// Emit event
	emitter.Emit("greeting", "World")

	emitter.Emit("helloOnce", "World")
	emitter.Emit("helloOnce", "World")

	emitter.On("greeting", func(data interface{}) {
		time.Sleep(time.Second)
		fmt.Println("Hello ", data)
	})

	// Emit event asynchronously
	result := emitter.AsyncEmit("greeting", "World")

	fmt.Println("Waiting for all greeting listeners to complete...")

	// Wait for all listeners to complete
	if err := result.Wait(); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("All greeting listeners have completed.")
}
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.