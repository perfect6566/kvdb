package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
	"kvdb.com/kvdbimp"

	"fmt"
	"strings"
)
var expiredtime=0
var cc=kvdbimp.Newcache()


func main(){
	r:=gin.Default()
	r.GET("/Get",Get)
	r.GET("/Set",Set)
	r.GET("/Count",Count)
	r.GET("/Watch",Watch)
	r.Run(":8888")
}


func Set(c *gin.Context){
	log.Println("expired ",c.Query("expired"))
	if c.Query("expired")==""{
		expiredtime=10
	}else {

		expiredtime, _ = strconv.Atoi(c.Query("expired"))
	}



	key:=c.Query("key")
	value:=c.Query("value")
	log.Println(key,value)

	cc.Set(key,value,time.Duration(expiredtime*1000*1000*1000))
	c.String(http.StatusOK,"Set ok",key,value,time.Duration(expiredtime*1000*1000*1000))
}


func Get(c *gin.Context){

	key:=c.Query("key")
	log.Println("Get",key)



	v,b:=cc.Get(key)

	log.Println("Get",key,v,b)
	rr:=kvdbimp.Result{}
	rr.Key=key
	rr.Value=v
	log.Println(rr)
	vv,err:=v.(string) //interface{}类型的使用strings.contain包需要先做断言成string
	if err{
		log.Println(err)
	}
	fmt.Printf("T%",vv)
	if b {
		log.Println("response is ",rr)
		//cc.SaveToFile("tmp/getdb")
		c.JSON(http.StatusOK,rr)
		}else if strings.Contains(vv,"Expired") {
		c.String(404, "Key Expired")
	}else {
		c.String(404, "Key Not Found")
	}

}

func Count(c *gin.Context){

	amount:=cc.Count()
	c.String(200, fmt.Sprintf("%v",amount))

}

func Watch(c *gin.Context){
	key:=c.Query("key")
	amount:=cc.Watch(key)
	c.String(200, fmt.Sprintf("%v",amount))

}