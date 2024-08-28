package dbutils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func CheckError(err error) {
	if err != nil {
		log.Println(err)
	}
}

var PsqlInfo string
var Db *sql.DB
var Err error
var F, e = os.OpenFile("/data/restsim.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
var Name, Table_name string
var Usertable_name, Sessiontable_name, Motable_name, Space_name, One_event, Event_name, Filter_table string
var Ref_db string

func Get_table() {
	One_event = "one_event"
	Filter_table = "filters"
	Event_name = "events"
	Motable_name = "modb"
	Space_name = "namespace"
	Table_name = "uridb"
	Usertable_name = "usertable"
	Sessiontable_name = "sessiontable"
	Ref_db = os.Getenv("REF_DB")
	Name = os.Getenv("DEPLOYMENT")
	if Ref_db == "" {
		Ref_db = "enmstub_2"
	}
	if Name != "" {
		Table_name += "_" + Name
		Usertable_name += "_" + Name
		Sessiontable_name += "_" + Name
		Motable_name += "_" + Name
		Space_name += "_" + Name
	}
}

func Db_uniq(Motable_name string, uri string, node string) {
	err, nodes := Db_select(Motable_name, "data", "uri=$1", uri)
	f := 0
	if err == nil {
		val1 := strings.Split(nodes, ";;;")
		for i := 0; i < len(val1); i++ {
			if val1[i] == node {
				f = 1
				break
			}
		}
		if f == 0 {
			Db_update(Motable_name, "data", []byte(nodes+";;;"+node), "uri=$2", uri)
		}
	} else {
		//fmt.Println("dn",uri,node)
		Db_insert(Motable_name, []string{"uri", "data"}, uri, node)
	}
}

func Db_insert(tableName string, columns []string, values ...interface{}) error {
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		columnList(columns),
		placeholderList(len(values)))
	ps, err := Db.Prepare(stmt)
	if err != nil {
		log.Println("error in insert prepare: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", columns, "values: ", values)
		return err
	}
	defer ps.Close()
	_, err = ps.Exec(values...)
	if err != nil {
		log.Println("error in insertion execution: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", columns, "values: ", values)
		return err
	}
	return nil
}
func columnList(columns []string) string {
	return fmt.Sprintf("%s",
		strings.Join(columns, ", "))
}
func placeholderList(n int) string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("%s", strings.Join(placeholders, ", "))
}

func Db_select(tableName string, column string, whereClause string, args ...interface{}) (error, string) {
	stmt := fmt.Sprintf("SELECT %s FROM %s",
		column, tableName)
	if whereClause != "" {
		stmt += fmt.Sprintf(" WHERE %s", whereClause)
	}
	ps, err := Db.Prepare(stmt)
	if err != nil {
		log.Println("error in select prepare: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", column, "whereclause: ", whereClause, "args: ", args)
		return err, ""
	}
	defer ps.Close()
	var result string
	err = ps.QueryRow(args...).Scan(&result)
	if err != nil {
		log.Println("error in select: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", column, "whereclause: ", whereClause, "args: ", args)
		return err, ""
	}
	return nil, result
}

func Db_select_multirows(tableName string, column string, whereClause string, args ...interface{}) (error, []string) {
	stmt := fmt.Sprintf("SELECT %s FROM %s",
		column, tableName)
	if whereClause != "" {
		stmt += fmt.Sprintf(" WHERE %s", whereClause)
	}
	ps, err := Db.Prepare(stmt)
	if err != nil {
		log.Println("error in select prepare: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", column, "whereclause: ", whereClause, "args: ", args)
		return err, []string{}
	}
	defer ps.Close()
	var res string
	var result []string
	rows, err2 := ps.Query(args...)
	for rows.Next() {
		e := rows.Scan(&res)
		fmt.Println(e)
		result = append(result, res)
	}
	defer rows.Close()

	if err2 != nil {
		log.Println("error in select: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", column, "whereclause: ", whereClause, "args: ", args)
		return err, []string{}
	}
	return nil, result

}

func Db_delete(tableName string, whereClause string, args ...interface{}) int64 {
	stmt := fmt.Sprintf("DELETE FROM %s", tableName)
	if whereClause != "" {
		stmt += fmt.Sprintf(" WHERE %s", whereClause)
	}
	ps, err := Db.Prepare(stmt)
	if err != nil {
		log.Println("error in delete prepare: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "whereclause: ", whereClause, "args: ", args)
		return 0
	}
	defer ps.Close()
	result, err := ps.Exec(args...)
	if err != nil {
		log.Println("error in deletion: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName)
		log.Println("whereclause: ", whereClause)
		log.Println("args: ", args)
		return 0
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("error in fetching rowsaffected: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName)
		log.Println("whereclause: ", whereClause)
		log.Println("args: ", args)
		return 0
	}
	return rowsAffected
}
func Db_update(tableName string, columnName string, value []byte, whereClause string, args interface{}) int64 {
	stmt := fmt.Sprintf("UPDATE %s SET %s = $1", tableName, columnName)
	if whereClause != "" {
		stmt += fmt.Sprintf(" WHERE %s", whereClause)
	}
	ps, err := Db.Prepare(stmt)
	if err != nil {
		log.Println("error in update prepare", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", columnName, "value: ", value, "whereclause: ", whereClause, "args: ", args)
		return 0
	}
	defer ps.Close()
	result, err := ps.Exec(string(value), args)
	if err != nil {
		log.Println("error in updation: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", columnName, "value: ", value, "whereclause: ", whereClause, "args: ", args)
		return 0
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("error in updated rowsaffected: ", err)
		log.Println(stmt)
		log.Println("Tablename: ", tableName, "columname: ", columnName, "value: ", value, "whereclause: ", whereClause, "args: ", args)
		return 0
	}
	return rowsAffected
}
func Drop_table(tablename string) error {
	_, err := Db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tablename))
	if err != nil {
		log.Println("error in drop table: ", err)
		log.Println("TableNamne: ", tablename)
		return err
	}
	return nil
}
func Delete_entries(tablename string) error {
	b := Db.QueryRow(fmt.Sprintf("SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_CATALOG = 'restsim' AND TABLE_NAME = '%s';", tablename))
	var boole string
	err := b.Scan(&boole)
	log.Println("error in deleting entries: ", err)
	log.Println("Tablename: ", tablename)
	if boole == "1" {
		_, err := Db.Exec(fmt.Sprintf("DELETE FROM %s;", tablename))
		if err != nil {
			return err
		}
		return nil
	} else {
		_, err := Db.Exec(fmt.Sprintf("create table %s (username varchar, cookie varchar, expirytime timestamp);", tablename))
		if err != nil {
			return err
		}
		return nil
	}
}

func Copy_table(sourceTable string, destinationTable string) error {
	_, err := Db.Exec(fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s;", destinationTable, sourceTable))
	if err != nil {
		log.Println("error in copying table: ", err)
		return err
	}
	return nil
}
func CreateTable(tableName string, columns []string, columnTypes []string, primaryKey string) error {
	if len(columns) != len(columnTypes) {
		log.Printf("number of columns and column types should be equal")
		return fmt.Errorf("number of columns and column types should be equal")
	}
	statement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)
	for i := 0; i < len(columns); i++ {
		statement += fmt.Sprintf("%s %s,", columns[i], columnTypes[i])
	}
	statement += fmt.Sprintf("PRIMARY KEY (%s))", primaryKey)
	_, err := Db.Exec(statement)
	if err != nil {
		log.Printf("failed to create table: %v", err, statement)
		return fmt.Errorf("failed to create table: %v", err, statement)
	}
	return nil
}
