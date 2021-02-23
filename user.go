package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C chan Info
	passwd string
	conn net.Conn
	//属于服务器
	server *Server
}
type Info struct {
	Name string
	Meg string
}

func (info Info)toString() string {
	return info.Name+info.Meg
}
//创建一个新用户
func NewUser(conn net.Conn,server *Server) *User{
	addr:=conn.RemoteAddr().String()
	user:= &User{
		Name: addr,
		Addr: addr,
		C:    make(chan Info),
		conn: conn,
		server: server,
	}
	//不断监听user
	go user.ListenMessage()
	return user
}

//用户注册
func (this *User)Register(name string){
	this.server.maplock.Lock()
	this.Name=name
	this.server.OnlineMap[name]=this
	this.server.maplock.Unlock()
	info:=Info{
		Name: this.Name,
		Meg:  "上线",
	}
	this.server.BoardCast(info)
}

//离线
func (this *User)OffLine()  {
	this.server.maplock.Lock()
	delete(this.server.OnlineMap,this.Name)
	this.server.maplock.Unlock()
	info:=Info{
		Name: this.Name,
		Meg:  "离线",
	}
	this.server.BoardCast(info)
}

//发送信息
func (this *User)SendMessage(meg string){
	this.conn.Write([]byte(meg))
}

//处理消息
func (this *User)DoMessage(meg string){
	if meg=="who"{
		//查询当前在线用户
		this.server.maplock.Lock()
		for _,usr:=range this.server.OnlineMap{
			onlinemeg:=fmt.Sprintf("%s 在线",usr.Name)
			this.SendMessage(onlinemeg)
		}
		this.server.maplock.Unlock()
	}else if len(meg)>7&&meg[:7]=="rename|"{
		//更改姓名
		name:=strings.Split(meg,"|")[1]
		_,ok:=this.server.OnlineMap[name]
		if ok{
			this.SendMessage("当前用户名已使用")
		}else{
			this.server.maplock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[name]=this
			this.server.maplock.Unlock()
			this.Name=name
			this.SendMessage("姓名已更改:"+this.Name)
		}
	}else if len(meg)>4&&meg[:3]=="to|"{
		//获取谈话对象
		toName:=strings.Split(meg,"|")[1]
		content:=strings.Split(meg,"|")[2]
		if toName==""{
			this.SendMessage("姓名格式不正确")
			return
		}
		toUser,ok:=this.server.OnlineMap[toName]
		if !ok{
			this.SendMessage("该用户不存在")
			return
		}else{
			toUser.SendMessage(content+"from"+this.Name)
		}
	}else{
		info:=Info{
			Name: this.Name,
			Meg: meg,
		}
		this.server.BoardCast(info)
	}
}

//监听user channel 接受从服务器的信息
func (this *User)ListenMessage(){
	for{
		meg:=<-this.C
		this.conn.Write([]byte(meg.toString()+"\n"))
	}
}
