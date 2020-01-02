package utils

import (
	"errors"
	"fmt"
	"sync"

	"github.com/garyburd/redigo/redis"
	//mysql驱动包
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	//数据库的读写锁
	dbMutex sync.RWMutex
)

//ExamScore 考试成绩
type ExamScore struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Score int    `db:"score"`
}

//QueryScoreFromMysql 从MYSQL数据表查询成绩,这里只有读操作
func QueryScoreFromMysql(name string) (score int, err error) {
	// fmt.Println("QueryScoreFromMysql")
	//读锁，当有写锁时，无法加载读锁，当只有读锁或者没有锁时，可以加载读锁，读锁可以加载多个，所以适用于“读多写少”的场景。
	dbMutex.RLock()
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")
	HandlerError(err, `sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")`)
	defer db.Close()
	//创建一个临时切片来存储查询的信息
	examScores := make([]ExamScore, 0)

	e := db.Select(&examScores, "select * from score where name=?;", name)
	if e != nil {
		fmt.Println(e, `db.Select(&examScores, "select * from score where name=?;", name)`)
		return
	}
	// fmt.Println(examScores)

	dbMutex.RUnlock()
	return examScores[0].Score, nil
}

//QueryScoreFromrRedis 从Redis查询成绩
func QueryScoreFromrRedis(name string) (score int, e error) {
	// fmt.Println("QueryScoreFromrRedis")
	conn, err := redis.Dial("tcp", "localhost:6379")
	HandlerError(err, `redis.Dial("tcp","localhost:6379")`)
	defer conn.Close()
	reply, e := conn.Do("get", name)
	if reply != nil {
		score, e = redis.Int(reply, e)
	} else {
		return 0, errors.New("未能从Redis查到数据")
	}

	if e != nil {
		fmt.Println(err, `conn.Do("get", name)或者redis.Int(reply, err)`)
		return 0, e
	}
	// fmt.Println(name, ":", score)
	return score, nil
}

//WriteScore2Mysql 向MySQL数据库写入成绩
func WriteScore2Mysql(scoreMap map[string]int) {
	//锁定为写模式，写入期间不允许读访问
	dbMutex.Lock()
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")
	HandlerError(err, `sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")`)
	defer db.Close()
	for name, score := range scoreMap {
		_, err := db.Exec("insert into score(name,score) values(?,?);", name, score)
		HandlerError(err, `db.Exec("insert into score(name,score) values(?,?);", name, score)`)
		// fmt.Println("Msql录入成功！")
	}
	fmt.Println("成绩录入完毕")
	//解锁数据库，开放查询
	dbMutex.Unlock()

}

//WriteScore2Redis 向Redis写入成绩
func WriteScore2Redis(name string, score int) error {
	conn, err := redis.Dial("tcp", "localhost:6379")
	HandlerError(err, `redis.Dial("tcp","localhost:6379")`)
	defer conn.Close()
	_, err = conn.Do("set", name, score)
	// fmt.Println("Redis写入成功")
	return err
}
