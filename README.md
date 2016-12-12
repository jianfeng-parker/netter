# netter
Golang 实现TCP通讯的简易封装

### TO DO LIST
>* Server中实现用户自定义的Handler: 1. 单个Handler；2. Handler Chain 
>* 支持Session关闭时回调功能
>* 增加Server端增加Session的管理功能
>* Session对象增加Receive Loop功能
>* 提供默认Protocol实现

### 示例
#### Server:

```java

  import(
      "netter"
  )

  // 实现 core中定义的Handler接口
  type HandlerA struct{
      // ...
  }
  
  func (h *handlerA) Handler(){

  }

  type ProtocolA struct{
      // ...
  }

  protocol := NewProtocolA() // 也可使用框架提供的默认实现

  handler := NewHandlerA()

  server, err := netter.Listen("tcp", "0.0.0.0",handler,protocol,1000)
  server.Start()

```
#### Client:

``` java

  import(
      "netter"
  )

  protocol := 使用默认框架的默认实现/自定义实现
  session, err := netter.Dial("tcp","127.0.0.1",protocol,1000)
  // ...

```
