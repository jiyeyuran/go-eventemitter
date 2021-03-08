# EventEmitter in Golang

## Synopsis

EventEmitter is an implementation of the Event-based architecture in Golang.

## Installation
```golang
import "github.com/jiyeyuran/go-eventemitter"
```

## Examples

##### Creating an instance.
```
em := eventemitter.NewEventEmitter()
```

##### An usual example.
```golang
em.On("foo", func(){
    // Some code...
})
em.Emit("foo")
```

##### It will be triggered only once and then callbacks will be removed.
```golang
em.Once("foo", func() {
  // some code...
});

em.Emit("foo");
// Nothing happend.
em.Emit("foo");
```

##### Callback with parameters.
```golang
em.On("foo", func(bar, baz string) {
  // some code...
})

em.Emit("foo", "var 1 for bar", "var 2 for baz")
// Missing parameters
ee.Emit("foo")
// With extra parameters
ee.Emit("foo", "bar", "baz", "extra parameter")
```

##### Callback's call can be ordered by "weight" parameter.
```golang
em.On("foo", func() {
    fmt.Println("3")
})

em.On("foo", func() {
    fmt.Println("2")
})

em.On("foo", func() {
    fmt.Println("1")
})

em.Emit("foo")
// 3
// 2
// 1
```

##### Set maxNumberOfListeners as a parameter when creating new instance.
```golang
em := eventemitter.NewEventEmitter(eventemitter.WithMaxListeners(1))


em.On("foo", func(){
    // Some code...
})

// Note: it will show warn in console.
em.On("foo", func(){
    // Some other code...
})
```

##### Asynchronous vs. Synchronous
```golang
wait := make(chan struct{})

em.On("foo", func(bar int){
    // Some block code...
    <-wait
})

// synchronous and panic
em.Emit("foo", "bad parameter")
// Emit asynchronously
em.SaftEmit("foo")
// Note: it will show error in console and no panic will be trigged.
em.SaftEmit("foo", "bad parameter")
// Wait result
em.SafeEmit("foo").Wait()
```
##### Avoid goroutine leak by removing listener
```golang
em := eventemitter.NewEventEmitter()
foo := func(){}

em.On("foo", foo)
em.On("foo", func(){})
em.On("bar", func(){})

// Remove the specified listen of event "foo"
em.Off("foo", foo)
// Remove all listeners of event "foo"
em.RemoveAllListeners("foo")
// Remove all listeners in EventEmitter object
em.RemoveAllListeners()
```