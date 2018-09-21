package main

import (
	"reflect"
	"os"
)


func StrInvokeFunc(f interface{},param []interface{})(ret []reflect.Value){
	if f == ""{
		return ret
	}
	in := make([]reflect.Value,len(param))
	funname := reflect.ValueOf(f)

	for i,arg := range param{
		in[i] = reflect.ValueOf(arg)
	}
	ret = funname.Call(in)

	return ret
}


func main(){
	if err := ReadJson("./Config.json");err!=nil{
		ErrInfo(err)
		os.Exit(-1)
	}
	DebugInfo(Configinfo)

	if len(os.Args) <2{
		ErrInfo("*****[exe] [命令]*****")
		os.Exit(-1)
	}

	//可持续新增
	key_value := map[string]interface{}{"single":SingleList,"double":DoubleList,"performance":UnnumberedList}
	var ags []interface{}
	var cmd interface{}
	//var tmp interface{}
	Flag := false
	for i,_ := range os.Args{
		if i == 0 {
			continue
		}

		if Is_key(os.Args[i],key_value){
			//执行一次
			if Flag == true{
				StrInvokeFunc(cmd,ags)
				cmd = key_value[os.Args[i]]
				ags = []interface{}{}
				//Flag = false
			}else {
				cmd = key_value[os.Args[i]]
				Flag = true
			}

		}else{
			ags = append(ags,os.Args[i])
		}
	}
	StrInvokeFunc(cmd,ags)

}
