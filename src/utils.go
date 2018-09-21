package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"io"
	"bufio"
	"strings"

	"encoding/base64"
	"syscall"
	"sync"
	"time"
	"strconv"
)

type ConfigInfo struct{
	Interface string
	IpPort string
	ThreadNum int
	ImgAPath string
	ImgBPath string
	RequestNums int

	IntPut interface{}
	OutPut interface{}

}
var Configinfo ConfigInfo
var Ch chan int
var mtx sync.Mutex
var wg sync.WaitGroup
func ReadJson(path string) error{
	by,err := ioutil.ReadFile(path)
	if err!=nil{
		return err
	}
	if err = json.Unmarshal(by,&Configinfo);err !=nil{
		return err
	}
	return nil
}

func RequestHttp(interfacestr string,param map[string]interface{})(map[string]interface{}){
	var ret = make(map[string]interface{})

	b, err := json.Marshal(param)
	body := bytes.NewBuffer([]byte(b))

	if err != nil{
		fmt.Printf("%s\n",err)
	}

	urlstr := fmt.Sprintf("%+v%+v",Configinfo.IpPort,interfacestr)

	response,err:= http.Post(urlstr,"application/json",body)
	if err != nil{
		fmt.Printf("%s\n",err)
		ret["result"] = -1
		return ret
	}
	if response.Status == "200 OK"{
		var user map[string]interface{}
		bodyPost, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(bodyPost, &user)
		return user
	}
	ret["result"] = 1
	return ret
}


func ReadTxt2List(path string)[]string{
	var temp []string
	f,err := os.Open(path)
	if err!= nil{
		ErrInfo(path," list open err, ",err)
	}
	rd := bufio.NewReader(f)

	for{
		line,err := rd.ReadString('\n')

		if err != nil || io.EOF == err{
			break
		}
		line = strings.Replace(line,"\r","",-1)
		line = strings.Replace(line,"\n","",-1)
		temp = append(temp,line)
		DebugInfo(line)
	}

	return temp
}

func SingleGo(i int,imgpath string,result map[string]interface{},mapArr []string,fp *os.File){
	param := make(map[string]interface{})

	if img,err := ioutil.ReadFile(imgpath);err ==nil{
		mgbase64 := base64.StdEncoding.EncodeToString(img)

		for k,v := range Configinfo.IntPut.(map[string]interface{}){
			if k == "img"{
				param["img"] = mgbase64
			}else{
				param[k] = v
			}

		}

		result = RequestHttp(Configinfo.Interface,param)
		retempStr := fmt.Sprintf("%+v,%+v,",i,imgpath)

		for _,ko := range mapArr{
			for kRo,vro := range result{
				if ko == kRo{
					retempStr = fmt.Sprintf("%+v%+v,",retempStr,vro)
				}
			}
		}
		mtx.Lock()
		fp.WriteString(retempStr+"\n")
		DebugInfo(retempStr)
		mtx.Unlock()
	}else{
		mtx.Lock()
		fp.WriteString(fmt.Sprintf("%+v,%+v,null\n",i,imgpath))
		mtx.Unlock()
	}

	Ch <- 1
	wg.Done()
}

func DoubleGo(i int,imgA string,imgB string,result map[string]interface{},mapArr[]string,fp *os.File){
	param := make(map[string]interface{})

	if imgbyteA,err := ioutil.ReadFile(imgA);err ==nil{
		mgAbase64 := base64.StdEncoding.EncodeToString(imgbyteA)
		if imgbyteB,err := ioutil.ReadFile(imgB);err ==nil{
			mgBbase64 := base64.StdEncoding.EncodeToString(imgbyteB)
			for k,v := range Configinfo.IntPut.(map[string]interface{}){
				if k == "imgA"{
					param["imgA"] = mgAbase64
				}else if k == "imgB"{
					param["imgB"] = mgBbase64
				}else{
					param[k] = v
				}
			}

			result = RequestHttp(Configinfo.Interface,param)

			retempStr := fmt.Sprintf("%+v,%+v,%+v,",i,imgA,imgB)

			for _,ko := range mapArr{
				for kRo,vro := range result{
					if ko == kRo{
						retempStr = fmt.Sprintf("%+v%+v,",retempStr,vro)
					}
				}
			}
			mtx.Lock()
			fp.WriteString(retempStr+"\n")
			DebugInfo(retempStr)
			mtx.Unlock()

		}else{
			mtx.Lock()
			fp.WriteString(fmt.Sprintf("%+v,%+v,%+v,null\n",i,imgA,imgB))
			mtx.Unlock()
		}
	}else{
		mtx.Lock()
		fp.WriteString(fmt.Sprintf("%+v,%+v,%+v,null\n",i,imgA,imgB))
		mtx.Unlock()
	}

	Ch <- 1
	wg.Done()
}


func SingleList(a...string)(ret string ){

	if Configinfo.ImgAPath == ""{
		return "Configinfo.ImgAPath == nil"
	}

	listA := ReadTxt2List(Configinfo.ImgAPath)

	var result map[string]interface{}

	fp ,_ := os.OpenFile("SingleList.csv",syscall.O_CREAT |syscall.O_WRONLY,666)

	strtmp:="序列,图片路径,"
	var mapArr []string
	for kk,_ := range Configinfo.OutPut.(map[string]interface{}){
		strtmp = fmt.Sprintf("%+v%+v,",strtmp,kk)
		mapArr = append(mapArr,kk)
	}
	fp.WriteString(strtmp+"\n")
	Ch = make(chan int,Configinfo.ThreadNum)
	for i,imgpath := range listA{
		wg.Add(1)
		go SingleGo(i,imgpath,result,mapArr,fp)

		if i%Configinfo.ThreadNum==Configinfo.ThreadNum-1 {
			for j:=0 ;j<Configinfo.ThreadNum;j++{
				<-Ch
			}
		}
	}
	wg.Wait()

	close(Ch)
	return "Single Run End"
}

func DoubleList(a...string)(ret string ){
	if Configinfo.ImgAPath == "" || Configinfo.ImgBPath == ""{
		ErrInfo("Configinfo.ImgAPath == nil || Configinfo.ImgBPath == nil")
		return "Configinfo.ImgAPath == nil || Configinfo.ImgBPath == nil"
	}

	listA := ReadTxt2List(Configinfo.ImgAPath)
	listB := ReadTxt2List(Configinfo.ImgBPath)

	if len(listA) != len(listB){
		ErrInfo("len(listA) != len(listB)")
		os.Exit(1)
	}

	var result map[string]interface{}

	fp ,_ := os.OpenFile("DoubleList.csv",syscall.O_CREAT |syscall.O_WRONLY,666)

	strtmp:="序列,图片路径1,图片路径2,"
	var mapArr []string
	for kk,_ := range Configinfo.OutPut.(map[string]interface{}){
		strtmp = fmt.Sprintf("%+v%+v,",strtmp,kk)
		mapArr = append(mapArr,kk)
	}
	fp.WriteString(strtmp+"\n")
	Ch = make(chan int,Configinfo.ThreadNum)
	for i,_ := range listA{
		wg.Add(1)
		go DoubleGo(i,listA[i],listB[i],result,mapArr,fp)

		if i%Configinfo.ThreadNum==Configinfo.ThreadNum-1 {
			for j:=0 ;j<Configinfo.ThreadNum;j++{
				<-Ch
			}
		}
	}
	wg.Wait()
	close(Ch)
	return "Double Run End!"
}


func MulRequestHttp(i int){
	var result map[string]interface{}
	for n := 1 ;n <= Configinfo.RequestNums;n++{
		start := time.Now()
		result = RequestHttp(Configinfo.Interface,Configinfo.IntPut.(map[string]interface{}))
		end := time.Now()
		DebugInfo(i,"Thread --> ",n,"Request Result :",result["result"],", cost: ",end.Sub(start))
		if result["result"] != float64(0){
			ResponseTimeErrArr = append(ResponseTimeErrArr,end.Sub(start))
		}
		ResponseTimeArr = append(ResponseTimeArr,end.Sub(start))
	}
	wg.Done()
}

func SortOrder(arr []time.Duration)[]time.Duration{
	for i := 0 ;i <len(arr)-1;i++{
		for j := i ;j <len(arr);j++{
			if arr[i] > arr[j]{
				arr[i],arr[j] = arr[j],arr[i]
			}
		}
	}
	return arr
}

func AveTime(arr []time.Duration)time.Duration{
	var ave time.Duration
	for i := 0 ;i <len(arr);i++{
		ave += arr[i]
	}
	ave = ave/time.Duration(len(arr))
	return ave
}

var ResponseTimeArr []time.Duration
var ResponseTimeErrArr []time.Duration

func Time2Float32(total time.Duration,cost time.Duration)float64{
	single := cost/total

	//ms
	flag := strings.Contains(single.String(),"ms")
	if flag{
		new := strings.Replace(single.String(),"ms","",-1)
		msvalue,_ := strconv.ParseFloat(new,32)
		return float64(1000)/msvalue
	}


	//m,s或s
	new := strings.Replace(single.String(),"m","-",-1)
	new = strings.Replace(single.String(),"s","",-1)
	newArr := strings.Split(new,"-")
	if len(newArr) == 2{
		m,_ := strconv.ParseFloat(newArr[0],32)
		s,_ := strconv.ParseFloat(newArr[1],32)
		return float64(1)/(m*60.+s)
	}else if len(newArr) == 1{
		s,_ := strconv.ParseFloat(newArr[0],32)
		return float64(1)/s
	}

	return 0.0
}

func UnnumberedList(a ...string)(ret string){

	if Configinfo.RequestNums < 1 {
		ErrInfo("RequestNums < 1")
		return "RequestNums < 1"
	}
	var runArr []int

	if len(a) < 1{
		runArr = append(runArr,Configinfo.ThreadNum)
	}else{
		for _,tempp := range a{
			num ,err := strconv.ParseFloat(tempp,32)
			if err !=nil{
				ErrInfo("input ThreadNum err")
				continue
			}
			runArr = append(runArr,int(num))
		}
	}

	for _,numm := range runArr{
		temp := strings.Split(Configinfo.Interface,"/")
		fp,_:= os.OpenFile(fmt.Sprintf("TPS_%+v.csv",temp[len(temp)-1]),syscall.O_CREAT |syscall.O_WRONLY|syscall.O_APPEND,666)
		fp.WriteString(fmt.Sprintf("###==%+v==####==%+v ThreadNum==###\n",Configinfo.Interface,numm))
		fp.WriteString("Samples,AveTime,90%Line,95%Line,99%Line,Err%,TPS\n")

		ResponseTimeArr = []time.Duration{}
		ResponseTimeErrArr = []time.Duration{}

		start := time.Now()
		for i:= int(0) ;i < numm;i++{
			wg.Add(1)
			go MulRequestHttp(i)

		}
		wg.Wait()
		end := time.Now()
		cost := end.Sub(start)

		ResponseTimeArr = SortOrder(ResponseTimeArr)
		DebugInfo(len(ResponseTimeArr),ResponseTimeArr)

		//输出结果
		//var Percentage = []float32{0.9,0.95,0.99}
		total := float32(len(ResponseTimeArr))
		var Index = []int{int(0.9*total),int(0.95*total),int(0.99*total)}

		fp.WriteString(fmt.Sprintf("%+v,%+v,%+v,%+v,%+v,%+v,%.2f\n",total,AveTime(ResponseTimeArr),
			ResponseTimeArr[Index[0]],ResponseTimeArr[Index[1]]	,ResponseTimeArr[Index[2]],
			float32(100*len(ResponseTimeErrArr))/total,
			Time2Float32(time.Duration(total),time.Duration(cost))))

		fmt.Println("Samples,AveTime,90%Line,95%Line,99%Line,Err%,TPS")
		fmt.Println(fmt.Sprintf("%+v,%+v,%+v,%+v,%+v,%+v,%.2f",total,AveTime(ResponseTimeArr),
			ResponseTimeArr[Index[0]],ResponseTimeArr[Index[1]]	,ResponseTimeArr[Index[2]],
			float32(100*len(ResponseTimeErrArr))/total,
			Time2Float32(time.Duration(total),time.Duration(cost))))

		fp.WriteString("\n\n")
	}

	return " Unnumbered Run End"
}


func Is_key(is_key string, keymap map[string]interface{})(ret bool){
	ret = false
	for k,_ := range keymap{
		if k == is_key{
			ret = true
			break
		}
	}

	return ret
}