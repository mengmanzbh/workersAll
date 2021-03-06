package utils
import (
    "crypto/md5"
    "database/sql"
    "encoding/hex"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "strconv"
    "time"
)
const (
    DB_Driver = "root:my-secret-pw@tcp(3.81.214.206:3306)/data?charset=utf8"
)
func OpenDB() (success bool, db *sql.DB) {
    var isOpen bool
    db, err := sql.Open("mysql", DB_Driver)
    if err != nil {
        isOpen = false
    } else {
        isOpen = true
    }
    CheckErr(err)
    return isOpen, db
}
func insertToDB(db *sql.DB) {
    uid := GetNowtimeMD5()
    nowTimeStr := GetTime()
    stmt, err := db.Prepare("insert userinfo set username=?,departname=?,created=?,password=?,uid=?")
    CheckErr(err)
    res, err := stmt.Exec("wangbiao", "zhangqi", nowTimeStr, "123456", uid)
    CheckErr(err)
    id, err := res.LastInsertId()
    CheckErr(err)
    if err != nil {
        fmt.Println("插入数据失败")
    } else {
        fmt.Println("记录用户付款行为,插入数据成功：", id)
    }
}
func QueryFromDB(db *sql.DB) {
    rows, err := db.Query("SELECT * FROM userinfo")
    CheckErr(err)
    if err != nil {
        fmt.Println("error:", err)
    } else {
    }
    for rows.Next() {
        var uid string
        var username string
        var departmentname string
        var created string
        var password string
        var autid string
        CheckErr(err)
        err = rows.Scan(&uid, &username, &departmentname, &created, &password, &autid)
        fmt.Println(autid)
        fmt.Println(username)
        fmt.Println(departmentname)
        fmt.Println(created)
        fmt.Println(password)
        fmt.Println(uid)
    }
}
func UpdateDB(db *sql.DB, uid string) {
    stmt, err := db.Prepare("update userinfo set username=? where uid=?")
    CheckErr(err)
    res, err := stmt.Exec("zhangqi", uid)
    affect, err := res.RowsAffected()
    fmt.Println("更新数据：", affect)
    CheckErr(err)
}
func DeleteFromDB(db *sql.DB, autid int) {
    stmt, err := db.Prepare("delete from userinfo where autid=?")
    CheckErr(err)
    res, err := stmt.Exec(autid)
    affect, err := res.RowsAffected()
    fmt.Println("删除数据：", affect)
}
func CheckErr(err error) {
    if err != nil {
        panic(err)
        fmt.Println("err:", err)
    }
}

func GetTime() string {
    const shortForm = "2006-01-02 15:04:05"
    t := time.Now()
    temp := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
    str := temp.Format(shortForm)
    fmt.Println(t)
    return str
}

func GetMD5Hash(text string) string {
    haser := md5.New()
    haser.Write([]byte(text))
    return hex.EncodeToString(haser.Sum(nil))
}

func GetNowtimeMD5() string {
    t := time.Now()
    timestamp := strconv.FormatInt(t.UTC().UnixNano(), 10)
    return GetMD5Hash(timestamp)
}
func OpenAndInsertToDB() {
    opend, db := OpenDB()
    if opend {
        fmt.Println("open success")
        // DeleteFromDB(db, 10)
        //QueryFromDB(db)
        //DeleteFromDB(db, 1)
        //UpdateDB(db, 5)
        //UpdateUID(db, 5)
        //UpdateTime(db, 4)
        insertToDB(db)
    } else {
        fmt.Println("open faile:")
    }

}
// func main() {
//     opend, db := OpenDB()
//     if opend {
//         fmt.Println("open success")
//     } else {
//         fmt.Println("open faile:")
//     }
//     // DeleteFromDB(db, 10)
//     //QueryFromDB(db)
//     //DeleteFromDB(db, 1)
//     //UpdateDB(db, 5)
//     insertToDB(db)
//     //UpdateUID(db, 5)
//     //UpdateTime(db, 4)

// }