package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	//全局管道，用来放入名字
	chNames = make(chan string, 100)
	//全局切片，用来记录参与考试的考生名字
	examers = make([]string, 0)
	//全局管道，用来放车道
	chLanes = make(chan int, 5)
	//全局管道，用来防止违规者的名字
	chFouls = make(chan string, 100)
	//考试成绩
	scoreMap = make(map[string]int)
	//等待组
	wg sync.WaitGroup
	//随机数的互斥锁（确保随机数函数不能被并发访问）
	randomMutex sync.Mutex
	//数据库的读写锁
	dbMutex sync.RWMutex
	//姓氏
	familyNames = []string{"赵", "钱", "孙", "李", "周", "吴", "郑", "王", "冯", "陈", "楚", "卫", "蒋", "沈", "韩", "杨", "张", "欧阳", "东门", "西门", "上官", "诸葛", "司徒", "夏侯", "司空"}
	//辈分
	middleNamesMap = map[string][]string{}
	//名字
	lastNames = []string{"春", "夏", "秋", "冬", "风", "霜", "雨", "雪", "木", "禾", "米", "竹", "山", "石", "田", "土", "福", "禄", "寿", "喜", "文", "武", "才", "华"}
)

//ExamScore 考试成绩
type ExamScore struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Score int    `db:"score"`
}

func init() {
	middleNamesMap["赵"] = []string{"大", "国", "益", "之", "仕", "世", "秉", "忠", "德", "全", "立", "志", "承", "先", "泽", "诗", "书", "继", "祖", "传", "代", "远", "永", "佑", "启", "家", "邦", "振", "万", "年"}
	middleNamesMap["欧阳"] = []string{"元", "梦", "应", "祖", "子", "添", "永", "秀", "文", "才", "思", "颜", "承", "德", "正", "道", "积", "享", "荣", "华", "洪", "范", "征", "恩", "锡", "彝", "伦", "叙", "典", "常"}
	for _, x := range familyNames {
		if x != "欧阳" {
			middleNamesMap[x] = middleNamesMap["赵"]
		} else {
			middleNamesMap[x] = middleNamesMap["欧阳"]
		}
	}
}

func main() {
	for i := 0; i < 20; i++ {
		chNames <- GetRandomName()
	}
	close(chNames)

	//巡考
	go Patrol()

	//考生并发考试
	for name := range chNames {
		wg.Add(1)
		go QueryScore(name)
	}

	wg.Wait()
	fmt.Println("考试完毕!")

	//录入考试成绩
	wg.Add(1)
	go WriteScore2Mysql(scoreMap)
	//故意给一个时间间隔，确保WriteScore2Mysql先抢到数据库的读写锁
	time.After(1 * time.Second)

	//考生查询成绩
	for _, name := range examers {
		wg.Add(1)
		go QueryScore(name)
	}

	for _, name := range examers {
		wg.Add(1)
		go QueryScore(name)
	}

	wg.Wait()
	fmt.Println("END")
}

//Patrol 巡考官
func Patrol() {
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		// fmt.Println("考官正在巡考...")
		select {
		case name := <-chFouls:
			fmt.Println(name, "考试违纪！！！！")
		default:
			fmt.Println("考场秩序良好")
		}
		<-ticker.C
	}
}

//TakeExam 参加考试
func TakeExam(name string) {
	chLanes <- 1
	fmt.Println(name, "正在考试...")
	//将参加考试的考生姓名记录到examers切片里
	examers = append(examers, name)
	//考试持续5秒
	<-time.After(2 * time.Second)

	//生成考试成绩
	score := GetRandomInt(0, 100)
	// fmt.Println(score)
	if score < 10 {
		score = 0
		chFouls <- name
		// fmt.Println(name, "考试违纪！！", score)
	}
	scoreMap[name] = score
	<-chLanes
	wg.Done()
}

//GetRandomName 获取随机名字
func GetRandomName() (name string) {
	familyName := familyNames[GetRandomInt(0, len(familyNames)-1)]
	middleName := middleNamesMap[familyName][GetRandomInt(0, len(middleNamesMap[familyName])-1)]
	lastName := lastNames[GetRandomInt(0, len(lastNames)-1)]
	return familyName + middleName + lastName
}

//GetRandomInt 获取随机整数
func GetRandomInt(start, end int) int {
	randomMutex.Lock()
	<-time.After(1 * time.Nanosecond)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := start + r.Intn(end-start+1)
	randomMutex.Unlock()
	return n
}

//HandlerError 处理err
func HandlerError(err error, when string) {
	if err != nil {
		fmt.Println(when, err)
		os.Exit(1)
	}
}

//QueryScore 二级缓存查询成绩
func QueryScore(name string) {
	score, err := QueryScoreFromrRedis(name)
	if err != nil {
		score, _ := QueryScoreFromMysql(name)
		fmt.Println(name, ":", score)
		//将数据写入Redis
		WriteScore2Redis(name, score)
	} else {
		fmt.Println(name, ":", score)
	}
	wg.Done()
}

//QueryScoreFromMysql 从MYSQL数据表查询成绩,这里只有读操作
func QueryScoreFromMysql(name string) (score int, err error) {
	fmt.Println("QueryScoreFromMysql")
	//读锁，当有写锁时，无法加载读锁，当只有读锁或者没有锁时，可以加载读锁，读锁可以加载多个，所以适用于“读多写少”的场景。
	dbMutex.RLock()
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")
	HandlerError(err, `sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")`)

	//创建一个临时切片来存储查询的信息
	examScores := make([]ExamScore, 0)

	err = db.Select(&examScores, "select * from score where name=?;", name)
	if err != nil {
		fmt.Println(err, `db.Select(&examScores, "select * from score where name=?;", name)`)
		return
	}
	fmt.Println(examScores)

	dbMutex.RUnlock()
	return examScores[0].Score, nil
}

//QueryScoreFromrRedis 从Redis查询成绩
func QueryScoreFromrRedis(name string) (score int, e error) {
	fmt.Println("QueryScoreFromrRedis")
	conn, err := redis.Dial("tcp", "localhost:6379")
	HandlerError(err, `redis.Dial("tcp","localhost:6379")`)
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
	fmt.Println(name, ":", score)
	return score, nil
}

//WriteScore2Mysql 向MySQL数据库写入成绩
func WriteScore2Mysql(scoreMap map[string]int) {
	//锁定为写模式，写入期间不允许读访问
	dbMutex.Lock()
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")
	HandlerError(err, `sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/driving_exam")`)
	for name, score := range scoreMap {
		_, err := db.Exec("insert into score(name,score) values(?,?);", name, score)
		HandlerError(err, `db.Exec("insert into score(name,score) values(?,?);", name, score)`)
		fmt.Println("插入成功！")
	}
	fmt.Println("成绩录入完毕")
	//解锁数据库，开放查询
	dbMutex.Unlock()
	wg.Done()
}

//WriteScore2Redis 向Redis写入成绩
func WriteScore2Redis(name string, score int) error {
	conn, err := redis.Dial("tcp", "localhost:6379")
	HandlerError(err, `redis.Dial("tcp","localhost:6379")`)
	_, err = conn.Do("set", name, score)
	fmt.Println("Redis写入成功")
	return err
}
