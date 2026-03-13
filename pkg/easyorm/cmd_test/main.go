package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"rest_demo/pkg/easyorm"
)

func main() {
	engine, _ := easyorm.NewEngine("mysql", "edward:edward@tcp(192.168.199.131:3306)/edward?charset=utf8mb4&parseTime=True&loc=Local")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
