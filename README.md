# Gdbc 一个 `go`版本的`mysql`连接驱动

# 一.基本使用
```golang
import (
    "database/sql"
    _ "gdbc"
    "gdbc/client"
)

func main() {
    client.Debug = true //开启调试，可以查看sql语句
    //1.连接
	db, _ := sql.Open("mysql", "root:123456@127.0.0.1:3306/test")
	db.SetMaxIdleConns(1)
	db.SetMaxIdleConns(2)
    db.Ping()

    //2.查询,请求参数可以是任何类型
    //var ai int64 = -20
	//var ai string = "20"
	//var ai []byte = nil
	var ai time.Time = time.Date(2018, 06, 3, 0, 0, 0, 0, time.Local)
	//var ai float64 = -10.2
	//var ai uint = 20
	row, err := db.Query("select * from student where age > ?", ai)
	if err != nil {
		panic(err)
    }
    defer row.Close()
	var id, age int
	var name string
	var create_time, update_time time.Time
	for row.Next() {
		row.Scan(&id, &name, &age, &create_time, &update_time)
		fmt.Printf("id:%d\t", id)
		fmt.Printf("name:%s\t", name)
		fmt.Printf("age:%d\n", age)
		fmt.Println(create_time.String())
		fmt.Println(update_time.String())
    }
    
    //3.使用Prepare
    stmt, _ := db.Prepare("select * from student where id=?")
    row, _ = stmt.Query(16138)
	for row.Next() {
		row.Scan(&id, &name, &age, &create_time, &update_time)
		fmt.Printf("stmt-id:%d\t", id)
		fmt.Printf("stmt-name:%s\t", name)
		fmt.Printf("stmt-age:%d\n", age)
		fmt.Printf(create_time.String())
		fmt.Printf(update_time.String())
    }
    // 4. insert
    //insert
	stmt, e := db.Prepare("insert into student(`name`,age) values (?,?)")
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	defer stmt.Close()
	result, e := stmt.Exec("test", "22")
	if e != nil {
		fmt.Println(e)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()
    fmt.Println(rowsAffected, lastInsertId)
    //5.update
	stmt, _ = db.Prepare("update student set name=? where id = ?")
	stmt.Exec("test1", 23192)
	fmt.Println(result.RowsAffected())
    fmt.Println(result.LastInsertId())
    //6.delete
	stmt, _ = db.Prepare("delete from student where id=?")
	stmt.Exec(23191)
	fmt.Println(result.RowsAffected())
	fmt.Println(result.LastInsertId())
	//7.使用预编译 select
	stmt, e = db.Prepare("select * from student where id=?")
	if e != nil {
		fmt.Println(e)
		return
	}
	rows, _ := stmt.Query(23192)
	defer rows.Close()
	for rows.Next() {
        row.Scan(&id, &name, &age, &create_time, &update_time)
		fmt.Printf("stmt-id:%d\t", id)
		fmt.Printf("stmt-name:%s\t", name)
		fmt.Printf("stmt-age:%d\n", age)
		fmt.Printf(create_time.String())
		fmt.Printf(update_time.String())
	}
	//8.使用事务，使用事务时，各种操作都要用tx对象来使用，才能实现事务，默认调用事务回滚，这样可以释放连接,也可以手动释放连接
	//使用前如果事务自动提交是开启的需要先进行关闭。SET AUTOCOMMIT = 0
	tx, err := db.Begin()
	if err != nil {
		//fmt.Println(err)
		panic(err)
	}
	defer tx.Rollback()
	_, err = tx.Exec("update student set name=? where id = ?", "test", 16138)
	if err != nil {
		tx.Rollback()
	}
	_, err = tx.Exec("delete from student where id=?", 16138)
	//
	commit := tx.Commit()
	if commit != nil {
		tx.Rollback()
	}
	//或者不捕获错误，使用上面的自动回滚方法
	tx.Commit()
}
```

# 二、实现特性
- 完全和mysql的协议对应，从结构上和流程上看代码都会更加清楚。
- 和上层`database/sql/driver`契合很好
- 功能单一，只作为一个`driver`使用
- 