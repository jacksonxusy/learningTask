###Package concepts###
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



###make and chan keywords###
make was only used to initiate three in-built referrence type:  
1. slice `slice := make([]int, length, capacity)`
2. map `m := make(map[string]int, 20)`
3. channel ` headers := make(chan *types.Header)`


**chan is used to transmit data safely between goroutine.**
1. can read and write.  `c := make(chan int)`.  
2. only can write.
```
var sendOnly chan<- int
// write into 1
sendOnly <- 1
```
3. only can read.  
```
var recvOnly <-chan int
// receive 1
x := <-recOnly
```
select is used to listen multiple channels, and executes the case that becomes ready first.
```
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


###Basic knowleadges###
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
```
var x interface{} = 123

y, ok := x.(int)
fmt.Println(y, ok) // 123 true

z, ok := x.(string)
fmt.Println(z, ok) // "" false
```

