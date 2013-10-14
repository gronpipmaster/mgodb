Mgodb
=====

Orm for mongodb, wrapper from labix.org/v2/mgo
Example usage in http://robfig.github.io/revel/ framework see example/init.revel.go and example/model.revel.go, support hooks,

    func (self *SomeModelName) PreInsert(_ mgodb.DbExecutor) error {}
    func (self *SomeModelName) PostInsert(_ mgodb.DbExecutor) error {}
    func (self *SomeModelName) PreUpdate(_ mgodb.DbExecutor) error {}
    func (self *SomeModelName) PostUpdate(_ mgodb.DbExecutor) error {}
    func (self *SomeModelName) PreDelete(_ mgodb.DbExecutor) error {}
    func (self *SomeModelName) PostDelete(_ mgodb.DbExecutor) error {}

Example usage default CRUD:

    err = Dbm.Insert(&models.SomeModelName{Username: "Ale", Password: "+55 53 8116 9639"},
    &models.SomeModelName{Username: "Cla", Password: "+55 53 8402 8510"})
    if err != nil {
      fmt.Fatal(err)
    }

    var results []models.SomeModelName
    err = Dbm.Find("users", bson.M{"u": "Ale"}).All(&results)// All methods support from http://godoc.org/labix.org/v2/mgo#Collection.Find
    if err != nil {
      fmt.Fatal(err)
    }

    fmt.Println("Result:", results)

