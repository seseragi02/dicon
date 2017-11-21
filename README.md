# dicon

DICONtainer Generator for go.

[![CircleCI](https://circleci.com/gh/akito0107/dicon.svg?style=svg)](https://circleci.com/gh/akito0107/dicon)

## Getting Started

### Prerequisites
- Go 1.9+
- make

### Installing
```
$ go get -u github.com/akito0107/dicon
```

### How to use
1. Write container interface and comment `+DICON` over it.
```.go
// +DICON
type Container interface {
    UserService() UserService
    UserReposiotry() UserRepository
}
```
2. Prepare dependencies. You must write constructor which meets below requirements:
- method name must be `New` + Interface name
- return type must be single Interface.
- dependencies which use this instance must be passed via the constructor.

```userservice.go
type UserService interface {
    Find(id int64) (*entity.User, error)
}

type userService struct {
    repo UserRepository
}

func NewUserService(repo UserRepotiroy) UserService {
    return &userService{
        repo: repo,
    }
}
```

```userrepository.go
type UserRepository interface {
    FindById(id int64) (*entity.User, error)
}

type userRepository struct {}

func NewUserRepository() UserRepository {
    return &userRepository{}
}
```
3. generate!
```
$ dicon generate --pkg sample
```

4. You can get the container implementation!
```dicon_gen.go
// Code generated by "dicon"; DO NOT EDIT.
package main

import (
	"log"
)

type dicontainer struct {
	store map[string]interface{}
}

func NewDIContainer() Container {
	return &dicontainer{
		store: map[string]interface{}{},
	}
}

func (d *dicontainer) UserRepository() UserRepository {
	if i, ok := d.store["UserRepository"]; ok {
		if instance, ok := i.(UserRepository); ok {
			return instance
		}
		log.Fatal("cached instance is polluted")
	}
	instance := repository.NewUserRepository()
	d.store["UserRepository"] = instance
	return instance
}
func (d *dicontainer) UserService() UserService {
	if i, ok := d.store["UserService"]; ok {
		if instance, ok := i.(UserService); ok {
			return instance
		}
		log.Fatal("cached instance is polluted")
	}
	dep0 := d.UserRepository()
	instance := service.NewUserService(dep0)
	d.store["UserService"] = instance
	return instance
}
```

5. Use it!
```.go
di := NewDIContainer()
u := di.UserService()
....
```

### Generate Mock
dicon's target interfaces are often mocked in unit tests. 
So, dicon also provides a tool for automated mock creation.

You just type
```
$ dicon generate-mock --pkg sample
```
then, you get mocks (by the default, under the `mock` package)

```go
// Code generated by "dicon"; DO NOT EDIT.

package mock

type UserRepositoryMock struct {
	FindByIdMock func(a0 int64) (*entity.User, error)
}

func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{}
}

func (mk *UserRepositoryMock) FindById(a0 int64) (*entity.User, error) {
	return mk.FindByIdMock(a0)
}

type UserServiceMock struct {
	FindMock   func(a1 int64) (*entity.User, error)
}

func NewUserServiceMock() *UserServiceMock {
	return &UserServiceMock{}
}

func (mk *UserServiceMock) Find(a0 int64) (*entity.User, error) {
	return mk.FindMock(a0)
}
```
Generated mocks have `XXXMock` func as a field (XXX is same as interface method name).
In testing, you can freely rewrite behaviors by assigning `func` to this field.
```go
func TestUserService_Find(t *testing.T) {
	mock := mock.NewUserRepositoryMock()
	mock.FindByIdMock = func(id int64) (*entity.User, error) {
		
		// mocking logic....
		
		return user, nil
	}
	
	service := NewUserService(UserRepository(mock)) // passing the mock
	
	if _, err := service.Find(id); err != nil {
		t.Error(err)
	}
}
```


## Options
- generate
```
$ dicon generate -h
NAME:
   dicon generate - generate dicon_gen file

USAGE:
   dicon generate [command options] [arguments...]

OPTIONS:
   --pkg value, -p value  target package(s).
   --out value, -o value  output file name (default: "dicon_gen")
   --dry-run
```
- generate mock
```
$ dicon generate-mock -h
NAME:
   dicon generate-mock - generate dicon_mock file

USAGE:
   dicon generate-mock [command options] [arguments...]

OPTIONS:
   --pkg value, -p value   target package(s).
   --out value, -o value   output file name (default: "dicon_mock")
   --dist value, -d value  output package name (default: "mock")
   --dry-run
```

## License
This project is licensed under the Apache License 2.0 License - see the [LICENSE](LICENSE) file for details
