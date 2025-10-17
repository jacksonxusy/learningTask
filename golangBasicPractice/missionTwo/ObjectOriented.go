package main

import "fmt"

var _ Shape = (*Rectangle)(nil)

type Shape interface {
	Area() int
	Perimeter() int
}

type Rectangle struct {
	Width, Height int
}

func (r *Rectangle) Area() int {
	return r.Width * r.Height
}

func (r *Rectangle) Perimeter() int {
	return 2*r.Width + 2*r.Height
}

type Person struct {
	Name string
	Age  int
}

type Employee struct {
	Person
	EmployID int
}

func (e Employee) PrintInfo() {
	fmt.Println("Employee:", e)
}

func main() {
	var rectangle Rectangle
	rectangle.Width = 100
	rectangle.Height = 200
	fmt.Println(rectangle.Area())
	fmt.Println(rectangle.Perimeter())

	var employee Employee
	employee.EmployID = 10
	employee.Name = "Jack"
	employee.Age = 25
	employee.PrintInfo()
}
