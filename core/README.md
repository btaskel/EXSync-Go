# 概述：
## 通讯：

**端口设置：**

* 客户端：
* 指令发送端口：`[random]`
* 数据传输端口：`[random]`


* 服务端：
* 数据传输端口：`[server_port]`
* 指令接收端口：`[server_port + 1]`

**大致通讯模型：**

    主机A                               主机B

    client(主动控制方—连接对方server)      client (主动控制方—连接对方server)
    server(被动控制方-监听所有client)      server (被动控制方-监听所有client)

    Server: CommandSocket持续接收来自Client的指令。
            DataSocket持续接收来自Client的数据并转交给TimeChannel，同时可以主动发送到Client-DataSocket数据。
    Client: CommandSocket主动向Server-CommandSocket发起指令。
            DataSocket主动向Server-DataSocket发送数据，同时可以接收来自Server-DataSocket的数据并转交给TimeChannel。

    主动&被动操作: 只有A创建Client并与B的Server连接才能与B主动发起操作，B亦然。
    数据接收隔离: 每一个Client与Server连接后隔离建立TimeChannel。
    服务连接建立: Server有且仅有一个实例存在, Client无限制。

**核心权限分级：**

1. guest(主要用于验证是否为用户)：0
2. user(主要用于安全范围内的文件操作)：10
3. admin(使用shell远程控制主机等全部权限)：20

**客户端发送至服务端指令套接字：**

指令以 随机8字节的Mark 开头，并且连接字典来传递参数。
在后续的异步发送文件中，每一条指令的发送会自动加上 8个字节的mark 标识。

    更改指令格式为：
    [8bytesMark]{
                    "command": "data"/"comm", # 命令类型(数据类型/命令类型)
                    "type": "file",         # 操作类型(具体操作方式)
                    "method": "get",        # 操作方法(get/post 获取与提交)
                    "data": {               # 参数数据(具体参数)
                        "a": 1
                        ....
                        "stat": "errorText" # 远程返回的错误信息
                    }
                }

**服务端答复至客户端指令套接字：**

更改使用TimeChannel实现数据的总体接收与解密。
每一个Client与Server连接后隔离建立TimeChannel。

## 同步：

**使用相对路径进行文件同步：**

每个主机的同步空间绝对路径没有必要保持一致。

    1.当主机A发送文件同步请求，会首先在索引文件中寻找当前文件信息。
    2.主机B接收文件时会根据同步空间名称来推断对方发送的文件路径。
        
    A. SpaceName | ./a/b/c.txt。
    B. 查询SpaceName的路径信息，与 ./a/b/c.txt 相连接得到文件绝对路径。

**双方文件同步判断**

作用于双方的同一同步空间, 现有文件A与文件B, 且文件名相同。

    假设:
        文件A的
        
**同步池**

在与某一台主机的同步空间进行同步时，为了尽可能地进行负载均衡, exsync会使用文件池动态调整文件拉取的目标主机。
    