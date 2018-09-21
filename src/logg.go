package main

import (
	"runtime"
	"time"
	"fmt"
)

func DebugInfo(a ... interface{}){
	timestr := time.Now().Format("2006-01-02 15:04:05")
	funname,file,line,ok :=runtime.Caller(1)
	if ok{
		funname = funname
		fmt.Println(fmt.Sprintf("[%+v] [%+v %+v %+v] %+v","Info",timestr,file ,line,fmt.Sprint(a...)))
	}
}


func ErrInfo(a ... interface{}){
	timestr := time.Now().Format("2006-01-02 15:04:05")
	funname,file,line,ok :=runtime.Caller(1)
	if ok{
		funname = funname
		fmt.Println(fmt.Sprintf("[%+v] [%+v %+v %+v] %+v","Error",timestr,file ,line,fmt.Sprint(a...)))
	}
}
