# Copier

  I am a copier, I copy everything from one to another

## Features

* Copies from field to field with same name
* Copies from method to field with same name
* Copies from field to method with same name
* Copies from slice to slice
* Copies from structure to slice
* Copies structure the methods described above\
with use mapping function 

## Usage

```go
package main

import (
	"fmt"
	"github.com/massalex/copier"
	"strconv"
)

type User struct {
	Name string
	Role string
	Age  int32
	Box  string
}

func (user *User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	Name      string
	Age       int32
	DoubleAge int32
	EmployeId int64
	SuperRule string
	Box       int
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

func (employee *Employee) BoxMap(box string) {
	num := 0
	num, _ = strconv.Atoi(box)
	employee.Box = num
}

func main() {
	var (
		user      = User{Name: "Jinzhu", Age: 18, Role: "Admin", Box: "34"}
		users     = []User{{Name: "Jinzhu", Age: 18, Role: "Admin", Box: "24"}, {Name: "jinzhu 2", Age: 30, Role: "Dev", Box: "55"}}
		employee  = Employee{}
		employees = []Employee{}
	)

	copier.New(&user, &employee, "").Copy()

	fmt.Printf("%#v \n", employee)
	// Employee{
	//    Name: "Jinzhu",           // Copy from field
	//    Age: 18,                  // Copy from field
	//    DoubleAge: 36,            // Copy from method
	//    EmployeeId: 0,            // Ignored
	//    SuperRule: "Super Admin", // Copy to method
	//    Box: 0,					// Ignored
	// }

	// Copy struct to slice
	copier.New(&user, &employees, "").Copy()

	fmt.Printf("%#v \n", employees)
	// []Employee{
	//   {Name: "Jinzhu", Age: 18, DoubleAge: 36, EmployeId: 0, SuperRule: "Super Admin", Box: 0}
	// }

	// Copy slice to slice
	employees = []Employee{}
	copier.New(&users, &employees, "").Copy()

	fmt.Printf("%#v \n", employees)
	// []Employee{
	//   {Name: "Jinzhu", Age: 18, DoubleAge: 36, EmployeId: 0, SuperRule: "Super Admin", Box: 0},
	//   {Name: "jinzhu 2", Age: 30, DoubleAge: 60, EmployeId: 0, SuperRule: "Super Dev", Box: 0},
	// }


	// Copy with map method use
	copier.New(&user, &employee, "Map").Copy()

	fmt.Printf("%#v \n", employee)
	// Employee{
	//    Name: "Jinzhu",           // Copy from field
	//    Age: 18,                  // Copy from field
	//    DoubleAge: 36,            // Copy from method
	//    EmployeeId: 0,            // Ignored
	//    SuperRule: "Super Admin", // Copy to method
	//    Box: 34,					// Set by Employee.BoxMap
	// }
}
```

## Contributing

You can help to make the project better, check out [http://gorm.io/contribute.html](http://gorm.io/contribute.html) for things you can do.

# Author

**jinzhu**

* <http://github.com/jinzhu>
* <wosmvp@gmail.com>
* <http://twitter.com/zhangjinzhu>

## License

Released under the [MIT License](https://github.com/jinzhu/copier/blob/master/License).
