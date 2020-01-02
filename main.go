package main

import (
	"fmt"
	"jiakao/utils"
	"sync"
	"time"
)

var (
	//等待组
	wg sync.WaitGroup
)

func main() {
	for i := 0; i < 20; i++ {
		chNames <- utils.GetRandomName()
	}
	close(chNames)

	//巡考
	go Patrol()

	//考生并发考试
	for name := range chNames {
		wg.Add(1)
		go TakeExam(name)
	}
	wg.Wait()
	chStopExam <- 1
	// close(chStopExam)
	fmt.Println("考试完毕!")

	//录入考试成绩
	wg.Add(1)
	go func() {
		utils.WriteScore2Mysql(scoreMap)
		wg.Done()
	}()
	//故意给一个时间间隔，确保WriteScore2Mysql先抢到数据库的读写锁
	<-time.After(1 * time.Second)

	//考生查询成绩
	for _, name := range examers {
		wg.Add(1)
		go QueryScore(name)
	}

	<-time.After(1 * time.Second)
	fmt.Println("【再次查询】")
	for _, name := range examers {
		wg.Add(1)
		go QueryScore(name)
	}

	wg.Wait()
	fmt.Println("END")
}
