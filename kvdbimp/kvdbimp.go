package kvdbimp

import (
	"time"
	"sync"
	"log"
	"kvdb.com/Configs"
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


type Cache struct {
	Defaultexpired int64
	Items map[string]Item
	Gcinterval time.Duration
	mu sync.RWMutex
	Stopgc chan bool
}

func (c *Cache)Set(k string,v interface{},expiredtime time.Duration)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	e:=time.Now().Add(expiredtime).UnixNano()
	c.Items[k]=Item{Value:v,Expiredtime:e}

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
func (c *Cache)gccleanloop(){
	t:=time.Tick(c.Gcinterval)
	for{
		select {
		case <-t:
			log.Println("GC work")
			c.DeletedExpiredKey()
		case <-c.Stopgc:
			return

		}
	}
}


func Newcache()*Cache{

	cc:=&Cache{
		Defaultexpired:Configs.Defaultexpired,
		Items:make(map[string]Item,10),
		Gcinterval:Configs.Gcinterval,
		Stopgc:            make(chan bool),
	}
go cc.gccleanloop()
	return cc
}
