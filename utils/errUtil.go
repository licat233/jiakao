package utils



import (
	"fmt"
	"os"
)

//HandlerError 处理err
func HandlerError(err error, when string) {
	if err != nil {
		fmt.Println(when, err)
		os.Exit(1)
	}
}
