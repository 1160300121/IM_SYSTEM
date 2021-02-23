package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP string
	Port int
	//消息管道
	Message chan Info
	//用户在线列表
	OnlineMap map[string]*User
	maplock sync.RWMutex
}

//创建一个server
func NewServer(IP string,port int) *Server {
	server:= &Server{
		IP:        IP,
		Port:      port,
		Message:   make(chan Info),
		OnlineMap: make(map[string]*User),
	}
	return server
}

//Server 监听 message channel 广播 有消息发给全部用户
func (this *Server)AcceptMessage()  {
	for  {
		meg:=<-this.Message
		n:=meg.Name
		this.maplock.RLock()
		for name,usr:=range this.OnlineMap{
			if name!=n{
				usr.C<-meg
			}
		}
		this.maplock.RUnlock()
	}
}

// 广播消息推送
func (this *Server) BoardCast(meg Info){
	this.Message<-meg
}

//服务器处理
func (this *Server) Handler(conn net.Conn){
	fmt.Println("connection is established successfully!")
	//将用户加入到OnlineMap中
	user:=NewUser(conn,this)
	//用户注册
	user.SendMessage("Please input your name:")
	for {
		buf:=make([]byte,1024)
		n, err := conn.Read(buf)
		if err!=nil&&err!=io.EOF{
			fmt.Println("Connection err:",err)
			return
		}
		if n==0{
			user.SendMessage("Wrong input,please input your name again:")
			continue
		}else{
			name:=string(buf[:n-1])
			_,ok:=user.server.OnlineMap[name]
			if ok{
				//用户已注册
				user.SendMessage("You is online now!:"+name+"\n")
				break
			}else{
				//用户未注册
				user.SendMessage("You haven't register ,Do you want the name you input to be your new name?(Y/n):")
				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if n==0||err!=nil{
						user.SendMessage("Wrong input,please input your name again:1")
						continue
					}else if op:=string(buf[:n-1]);op=="Y"{
						user.Register(name)
						user.SendMessage("You is online now!"+user.Name)
						break
					}else if op=="n"{
						user.SendMessage("OK!,please input your name again:")
						for  {
							buf := make([]byte, 1024)
							n, err := conn.Read(buf[:n-1])
							if n==0||err!=nil {
								user.SendMessage("Wrong input,please input your name again:2")
								continue
							}else{
								name:=string(buf[:n-1])
								user.Register(name)
								user.SendMessage("You is online now!:"+user.Name)
								break
							}
						}
						break
					}else{
						user.SendMessage("Wrong input,please input again(Y/n):")
						continue
					}
				}
				break
			}
		}
	}
	buf:=make([]byte,1024)
	isLive :=make(chan bool)
	//接受客户端的数据
	go func() {
		for {
			n,err:=conn.Read(buf)
			if n==0{
				user.OffLine()
				return
			}
			if err!=nil&&err!=io.EOF{
				fmt.Println("Connection err:",err)
				return
			}
		//提取用户的消息
		msg:=string(buf[:n-1])
		//将得到的消息进行处理
		user.DoMessage(msg)
		//用户活跃
		isLive<-true
		}
	}()
	//超时踢人功能
	for{
		select {
			case <-time.After(time.Second*6000): //已经超时
				user.SendMessage("您被踢出")
				close(user.C)
				conn.Close()
				return
			case <-isLive:
			//do nothing
		}
	}
}

//启动服务器
func (this *Server) Start(){
	//socket listen
	listener,err:=net.Listen("tcp" , fmt.Sprintf("%s:%d",this.IP,this.Port))
	if err!=nil{
		fmt.Println(err)
		return
	}
	defer listener.Close()
	//accept
	//启动监听message
	go this.AcceptMessage()

	for  {
		conn,err:=listener.Accept()
		if err!=nil{
			fmt.Println(err)
			continue
		}
		go this.Handler(conn)
	}
	//do handle

	//close the socket
}
