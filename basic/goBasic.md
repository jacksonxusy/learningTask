### Package concepts

1. all file must use one package in one directory.
2. if two go file belongs to package main, then it must compile together. `go run *.go` or `go run A.go B.go`
3. `go mod init <module-name>`  this is to create a go module, which includes go code, plus a go.mod file, plus dependencies managed automatically.  
4. go function naming, exported(visible outside the package): start with uppercase, unexported- start with lowercase.

```
myproject/
├── go.mod        ← ONE module
├── main.go       ← package main
├── utils/
│   └── crypto.go ← package utils
└── db/
    └── mysql.go  ← package db
```

### make and chan keywords

make was only used to initiate three in-built referrence type:  

1. slice `slice := make([]int, length, capacity)`
2. map `m := make(map[string]int, 20)`
3. channel ` headers := make(chan *types.Header)`

**chan is used to transmit data safely between goroutine.**

1. can read and write.  `c := make(chan int)`.  

2. only can write.go
   
   ```go
   var sendOnly chan<- int
   // write into 1
   sendOnly <- 1
   ```

3. only can read.  
   
   ```go
   var recvOnly <-chan int
   // receive 1
   x := <-recOnly
   ```
   
   select is used to listen multiple channels, and executes the case that becomes ready first.
   
   ```gol
   select {
   case msg := <-ch1:
    fmt.Println("Received from ch1:", msg)
   case msg := <-ch2:
    fmt.Println("Received from ch2:", msg)
   }
   ```
   
   whichever channel receives a value first is executed.

**Unbuffered channel**
`c := make(chan int)`

1. send blocks until someone receives.  
2. receive blocks until someone sends.  
3. perfect for enforcing synchronization.  

**Buffered channel**
`c := make(chan int, 3)` 

1. send does not block until the buffer is full.  
2. receive block only when buffer is empty.  
3. good for queue-lick behavior.  

### Basic knowleadges

1. in Go, the backtick symbol ` … ` does represent a string, but a special kind called a raw string literal.
   used for JSON, SQL, Regex.
2. := (short variable declaration) requires at least one new variable must be declared on the left side.

```
event := struct {
    key   [32]byte
    value [32]byte
}{}
```

3. creating an anonymous concrete struct type and an instance of it.

4. `publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)` this is an type assertion, it ask go if the value of publickKey is type of.
   
```go
   var x interface{} = 123
```
``` go
y, ok := x.(int)
fmt.Println(y, ok) // 123 true

z, ok := x.(string)
fmt.Println(z, ok) // "" false
```
#### go path issue
This project uses replace because:

- It's a learning/development project
- EasySwapBase isn't published as a versioned package
- Convenience for local development
- In production, you'd:

Publish EasySwapBase to GitHub with version tags
- Remove the replace directive
- Use specific versions: require github.com/... v1.2.3

#### goroutine
In Go, sync.WaitGroup is the standard way to wait for a collection of goroutines to finish executing. Without it, the main function might finish and exit the program before your concurrent workers have a chance to complete their tasks.

Think of it as a counter: you increment it when a task starts and decrement it when a task finishes. The program "waits" until that counter hits zero.

1. The Core Methods
To use a WaitGroup, you only need to understand three methods:

Add(int): Increases the counter by the given integer. Usually called before starting a goroutine.

Done(): Decreases the counter by 1. Usually called inside the goroutine via defer.

Wait(): Blocks the execution of the current thread (usually main) until the counter becomes 0.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	// 2. Signal that this goroutine is done when the function exits
	defer wg.Done()

	fmt.Printf("Worker %d starting...\n", id)
	time.Sleep(time.Second) // Simulating work
	fmt.Printf("Worker %d finished!\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		// 1. Increment the counter before launching the goroutine
		wg.Add(1)
		go worker(i, &wg)
	}

	// 3. Block until the counter returns to 0
	fmt.Println("Main: Waiting for workers...")
	wg.Wait()
	fmt.Println("Main: All workers finished. Exiting.")
}
```
3. Best Practices & Common Pitfalls
Pass by Pointer
If you pass a sync.WaitGroup to a function, you must pass it as a pointer (*sync.WaitGroup). In Go, everything is passed by value; if you pass it directly, the function receives a copy, the Done() call won't affect the original counter, and Wait() will block forever (deadlock).*

Use defer for Done()
Always use defer wg.Done() at the very top of your goroutine function. This ensures that even if the function panics or returns early, the counter is decremented, preventing the main function from hanging.

Call Add() Outside the Goroutine
Always call wg.Add() in the parent thread before the go statement. If you put wg.Add() inside the goroutine, there is a "race condition": the Wait() in the main thread might execute before the goroutine has even started, making the program think there's nothing to wait for.

#### defer keywords

The defer keyword in Go (Golang) is used to schedule a function call to be executed later, specifically just before the surrounding function returns. This ensures cleanup actions (like closing files or unlocking mutexes) happen reliably, regardless of how the function exits—whether normally, via an explicit return, or due to a panic.
Key Behaviors

- Delayed Execution: The deferred function runs after the surrounding function's body completes (including after any return statement sets values), but before control returns to the caller.
- Argument Evaluation: Arguments are evaluated immediately when the defer statement is encountered, not when the deferred function runs.
- LIFO Order: Multiple defer statements in the same function are executed in last-in, first-out (reverse) order.
- Even on Panic: Deferred calls still run if the function panics.
Modifying Returns: Deferred anonymous functions can access and modify named return values.

``` go
package main
import "fmt"

func main() {
    a()
}

func a() {
    i := 0
    defer fmt.Println("Deferred value:", i)  // i evaluated now (0)
    i++
    fmt.Println("Immediate value:", i)       // Prints 1
}
```
Immediate value: 1
Deferred value: 0


```
func (s *Service) Start() {
    // ...
}
```
In Go, this is called a method with a receiver:

(s *Service) is the receiver - this makes Start() a method that belongs to the Service type
s is the receiver variable name - it's like this or self in other languages. Inside the method, you use s to access the instance's fields and call other methods
*Service indicates this is a pointer receiver - the method receives a pointer to a Service instance, which allows it to modify the original instance
So to be precise:*

The method belongs to the Service type (not to the variable s)
s is just the variable name you use inside the method to refer to the instance
When you call someServiceInstance.Start(), inside the method, s will refer to someServiceInstance
```
func (s *Service) Start() {
    threading.GoSafe(s.SyncOrderBookEventLoop)
    threading.GoSafe(s.UpKeepingCollectionFloorChangeLoop)
}
```
s is accessing two methods/functions of the Service instance: SyncOrderBookEventLoop and UpKeepingCollectionFloorChangeLoop
Since it's a pointer receiver (*Service), any modifications made to s inside the method will affect the original instance*



