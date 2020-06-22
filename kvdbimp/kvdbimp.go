package kvdbimp

import (
	"time"
	"sync"
	"log"
	"kvdb.com/Configs"
	"os"
	"encoding/gob"
	"fmt"

)

type Item struct {
	Value interface{}
	Expiredtime int64
}

func (I Item)Expired() bool  {
	if I.Expiredtime==0{
		return false
	}
	return time.Now().UnixNano()>I.Expiredtime
}

type Result struct {
	Key string `json:"key"`
	Value interface{} `json:"value"`
}
type Request struct {
	Key string `json:"key"`
	Value interface{} `json:"value"`
	Expiredtime uint32 `json:"expired"`

}

type Cache struct {
	Defaultexpired int64
	Items map[string]Item
	chs map[string]func(key string)
	Gcinterval time.Duration
	mu sync.RWMutex
	Stopgc chan bool
}

func (c *Cache)Set(k string,v interface{},expiredtime time.Duration)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	e:=time.Now().Add(expiredtime).UnixNano()
	c.Items[k]=Item{Value:v,Expiredtime:e}

	for k,v:=range c.chs{
		log.Println("here ",k,v)

	}
	_,ok:=c.chs[k]
	if ok{
		c.chs[k](k)
	}

log.Println(k,v,"set")
}

func (c *Cache)Get(k string)(i interface{},result bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	item,found:=c.Items[k]
	if !found{
		return "Not Found",false
	}
	if item.Expired(){
		return "Expired",false
	}
	_,ok:=c.chs[k]
	if ok{
		c.chs[k](k)
	}
	return item.Value,true
}

func (c *Cache)Delete(k string)(i interface{},result bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	_,found:=c.Items[k]
	if !found{
		return "Key: "+k+" Not Exist",false
	}
	delete(c.Items,k)

	_,ok:=c.chs[k]
	if ok{
		c.chs[k](k)
	}

	return "Key: "+k+" Successful Deleted",true
}

func (c *Cache)Update(k string,v interface{},d time.Duration)(i interface{},result bool)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	_,found:=c.Items[k]
	if !found{
		return "Key: "+k+" Not Exist",false
	}
	e:=time.Now().Add(d).UnixNano()
	c.Items[k]=Item{Value:v,Expiredtime:e}
	_,ok:=c.chs[k]
	if ok{
		c.chs[k](k)
	}
	return "Key: "+k+" Updated",true
}

func (c *Cache)Count()int{
	return len(c.Items)
}
func (c *Cache)DeletedExpiredKey()  {

	for k,v:=range c.Items{
		log.Println(k,v)
		if v.Expiredtime<time.Now().UnixNano(){
			delete(c.Items,k)
			log.Println(k,"expired and deleted",len(c.Items))
		}
	}

}

func (c *Cache)Save(file string)error  {

	f,err:=os.Create(file)
	if err != nil{
		return err
	}
	defer f.Close()
	defer func() {
		if e:=recover();e!=nil{
			err = fmt.Errorf("Error registring item type with gob library!")
		}

	}()

	en:=gob.NewEncoder(f)

	c.mu.Lock()
	defer c.mu.Unlock()
	gob.Register(c.Items)
	en.Encode(c.Items)
return err

}




//从 io.Reader 中读取数据
func (c *Cache) Load(file string) error{
	f,err:=os.Open(file)
	if err!=nil{
		return err
	}
	dec := gob.NewDecoder(f)
	items := map[string]Item{}
	err = dec.Decode(&items)
	if err == nil{
		c.mu.Lock()
		defer c.mu.Unlock()
		for k,v := range items{
			//ov,found := c.items[k]
			//if !found || ov.Expired() {
			//	c.items[k] = v
			//}
			c.Items[k]=v
			fmt.Println(time.Now(),"Load from file abc",k,v)

		}
	}
	return err
}




//从 io.Reader 中读取数据
func (c *Cache) Watch(key string) string{
ch:=make(chan string,0)
var resp string
c.mu.Lock()
	c.chs[key]= func(k string) {
		ch<-key
	}
c.mu.Unlock()
	select {
	case ch1:=<- ch:
		resp=ch1+"Changed"
		delete(c.chs,ch1)
	case <-time.Tick(100*time.Second):
		resp="chaoshi"


}
return resp
}


func (c *Cache)gccleanloop(){
	t:=time.Tick(c.Gcinterval)
	for{
		select {
		case <-t:
			log.Println("GC work")

			c.DeletedExpiredKey()
			c.Save("binlog")

		case <-c.Stopgc:
			return

		}
	}
}


func Newcache()*Cache{

	cc:=&Cache{
		Defaultexpired:Configs.Defaultexpired,
		Items:make(map[string]Item),
		Gcinterval:Configs.Gcinterval,
		chs:make(map[string]func(k string)),
		Stopgc:            make(chan bool),
	}
	cc.Load("binlog")
go cc.gccleanloop()
	return cc
}
