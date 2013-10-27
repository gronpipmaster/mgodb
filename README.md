Mgodb
=====

Orm Active Record http://en.wikipedia.org/wiki/ActiveRecord for mongodb, wrapper from http://labix.org/v2/mgo
Example usage in framework http://robfig.github.io/revel/ see example/init.revel.go and example/model.revel.go, support hooks concept from https://github.com/coopernurse/gorp

### CRUD ###

Example usage default CRUD:
```go
var err error
user := models.NewUser()
user.Username = "Bar"
user.Password = "ssdf"
err = user.Save() //Insert object
if err != nil {
    panic(err)
}
fmt.Println(user)
user.Username = "Foo"
err = user.Save() //Update object
if err != nil {
    panic(err)
}
fmt.Println(user)
loadUser, err := models.FindUser(user.GetId()) //Get object
if err != nil {
    panic(err)
}
fmt.Println(loadUser)
loadUser.Username = "NewFoo"
err = loadUser.Save() //Update object
if err != nil {
    panic(err)
}
fmt.Println(loadUser)
err = loadUser.Delete() //Delete object
if err != nil {
    panic(err)
}
```
### More examples ###

Example search one by field "Username" = "Foo":
```go
user, err := models.FindUserBy(&models.User{Username: "Foo"})
if err != nil {
  panic(err)
}
fmt.Println(user.GetId())//Get object
user.Username = "Bar"
err = user.Save() //Update object
if err != nil {
  panic(err)
}
fmt.Println(user.GetId())
```
Example get count by field "Username" = "Foo":
```go
count, err := models.CountUsers(&models.User{Username: "Foo"})
if err != nil {
  panic(err)
}
fmt.Println(count)
```
Example get all by field "Username" = "Bar":
```go
query := mgodb.Query{QueryDoc: &models.User{Username: "Bar"}, Limit: 0, Skip: 0}
users, err := models.FindUsers(query) //Get users
if err != nil {
  panic(err)
}
for _, item := range users {
  item.Username = "Foo"
  if err = item.Save(); err != nil { //Update users
    panic(err)
  }
}
```


### Hooks ###

Use hooks to before/after saving/delete to the db.
```go
//Full list of hooks that you can implement:
func (self *SomeModelName) BeforeInsert() error {}
func (self *SomeModelName) AfterInsert() error {}

func (self *SomeModelName) BeforeUpdate() error {}
func (self *SomeModelName) AfterUpdate() error {}

func (self *SomeModelName) BeforeDelete() error {}
func (self *SomeModelName) AfterDelete() error {}
```