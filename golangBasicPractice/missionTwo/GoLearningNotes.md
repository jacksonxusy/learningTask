# Go Learning Notes

## Basic Concepts

### Pointers

#### Basic Usage
- Use `*` to declare pointer types
- Use `&` to get variable address
- Use `*` to dereference pointers

#### Example Code
```go
func increaseInt(x *int) int {
    *x += 10  // Modify original variable through pointer
    return *x
}

func multiElement(arr *[]int) []int {
    for index, element := range *arr {
        (*arr)[index] = element * 2  // Modify slice elements
    }
    return *arr
}
```

#### Key Points
- Pointers in Go are safer; no pointer arithmetic supported
- Pointers often used for function parameters to avoid value copying
- `nil` is the zero value for pointer types

---

### Object-Oriented Programming

#### Structs
```go
type Rectangle struct {
    Width, Height int
}

type Person struct {
    Name string
    Age  int
}
```

#### Interfaces
```go
type Shape interface {
    Area() int
    Perimeter() int
}

func (r *Rectangle) Area() int {
    return r.Width * r.Height
}

func (r *Rectangle) Perimeter() int {
    return 2*r.Width + 2*r.Height
}
```

#### Composition
```go
type Employee struct {
    Person      // Anonymous embedding for inheritance-like behavior
    EmployID int
}
```

#### Interface Check
```go
var _ Shape = (*Rectangle)(nil)  // Compile-time interface implementation check
```

#### Key Points
- Go has no classes; uses structs and methods for OOP
- Composition over inheritance for code reuse
- Interfaces are implemented implicitly, no explicit declaration needed

---

### Concurrency Programming

#### Goroutine Basics
```go
func printOddNumber() {
    for i := 0; i < 10; i++ {
        if i%2 != 0 {
            fmt.Println("odd: " + strconv.Itoa(i))
        }
    }
}

func main() {
    go printOddNumber()  // Start goroutine
    go printEvenNumber() // Start another goroutine
    time.Sleep(1 * time.Second) // Wait for goroutines to complete
}
```

#### Goroutine Features
- Lightweight threads managed by Go runtime
- Small memory footprint (few KB)
- Started with `go` keyword

#### Synchronization Primitives

##### Mutex (Mutual Exclusion)
```go
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()    // Lock
    c.count++
    c.mu.Unlock()  // Unlock
}
```

##### Channels
```go
// Unbuffered channel
ch := make(chan int)
ch <- 1    // Send
value := <-ch  // Receive

// Buffered channel
bufferedCh := make(chan int, 10)
```

#### Concurrency Patterns
- **Producer-Consumer**: Use channels for data passing
- **Worker Pool**: Multiple goroutines process task queue
- **Pipeline**: Data processing pipeline

---

## Core Features

### 1. Memory Management
- Automatic garbage collection (GC)
- Automatic stack and heap management
- Escape analysis

### 2. Error Handling
- Explicit error handling
- `error` interface
- `panic` and `recover`

### 3. Package Management
- `go mod` module management
- Semantic versioning
- Dependency management

### 4. Compilation Features
- Static compilation
- Cross-compilation support
- Fast compilation

---

## Best Practices

### 1. Code Style
- Use `gofmt` for code formatting
- Camel case naming
- Simplicity and readability first

### 2. Error Handling
- Always check errors
- Early return for errors
- Meaningful error messages

### 3. Concurrency Programming
- Prefer channels for communication
- Avoid shared memory
- Proper resource cleanup

### 4. Performance Optimization
- Reduce memory allocation
- Use slices instead of arrays
- Proper use of concurrency

---

## Learning Path

### Fundamentals ✅
- [x] Basic syntax and control structures
- [x] Functions and methods
- [x] Structs and interfaces
- [x] Pointers and memory management

### Intermediate 🔄
- [ ] Advanced concurrency programming
- [ ] Package management and module system
- [ ] Reflection and type system
- [ ] Testing and benchmarking

### Advanced ⏳
- [ ] Web development (Gin/Echo frameworks)
- [ ] Database operations (GORM/sqlx)
- [ ] Microservices architecture
- [ ] Performance tuning

---

## Practice Examples from Mission Two

### 1. Pointer Operations
- Modifying variables through pointers
- Working with slice pointers
- Memory-efficient parameter passing

### 2. OOP Concepts
- Interface implementation and checking
- Struct composition for inheritance
- Method receivers (value vs pointer)

### 3. Concurrency Examples
- Basic goroutine usage
- Synchronization with sleep (simple example)
- Multiple goroutine coordination

### 4. Synchronization Patterns
- Mutex for protecting shared data
- Channel-based communication
- Producer-consumer patterns

---

## References

- [Official Go Documentation](https://golang.org/doc/)
- [Go Tour](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go by Example](https://gobyexample.com/)

---

*Last Updated: October 2024*