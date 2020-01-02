package main

import (
	"fmt"
	"jiakao/utils"
	"time"
)

var (
	//全局管道，用来放入名字
	chNames = make(chan string, 100)
	//全局切片，用来记录参与考试的考生名字
	examers = make([]string, 0)
	//全局管道，用来放车道,限制5个一组
	chLanes = make(chan int, 5)
	//全局管道，用来防止违规者的名字
	chFouls = make(chan string, 100)
	//全局管道，用来监测考试结束状态
	chStopExam = make(chan int, 5)
	//全局管道，用来监测考试开始状态
	chStartExam = make(chan int, 5)
	//考试成绩
	scoreMap = make(map[string]int)
)

//Patrol 巡考官
func Patrol() {
	//400 微秒执行一次
	ticker := time.NewTicker(400 * time.Millisecond)
	// LOOP:
	for {
		select {
		case <-chStopExam:
			// break LOOP
			return
		case <-chStartExam:
			// fmt.Println("考官正在巡考...")
			select {
			case name := <-chFouls:
				fmt.Println(name, "考试违纪！！！！")
			default:
				fmt.Println("考场秩序良好")
			}
		default:
		}
		<-ticker.C
	}
}

//TakeExam 参加考试
func TakeExam(name string) {
	//每次考试占用1个车道
	chLanes <- 1
	//开始考试
	chStartExam <- 1
	fmt.Println(name, "正在考试...")
	//将参加考试的考生姓名记录到examers切片里
	examers = append(examers, name)
	//考试持续2秒
	<-time.After(2 * time.Second)

	//生成考试成绩
	score := utils.GetRandomInt(0, 100)
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

//QueryScore 二级缓存查询成绩
func QueryScore(name string) {
	score, err := utils.QueryScoreFromrRedis(name)
	if err != nil {
		fmt.Println("未能从Redis中查到数据")
		score, _ := utils.QueryScoreFromMysql(name)
		fmt.Println("Mysql成绩：", name, ":", score)
		//将数据写入Redis
		utils.WriteScore2Redis(name, score)
	} else {
		fmt.Println("Redis成绩：", name, ":", score)
	}
	wg.Done()
}
