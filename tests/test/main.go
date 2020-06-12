package main

import (
	"encoding/json"
	"fmt"
)

type Job struct{
	Name string `json:"name"`
	Age string `json:"age"`
}

func All(data interface{}){
	bytes, _ := json.Marshal([]Job{{Name: "name", Age: "age"}})
	json.Unmarshal(bytes, &data)
	return
}

func main(){
	jobList := make([]*Job, 0)
	//json.Unmarshal(bytes, &jobList)
	//fmt.Println(*jobList[0])
	All(&jobList)
	fmt.Println(jobList)
}
