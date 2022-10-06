# Some set of commands


#### Make Client: 
```
go build -o .\bins\client.exe .\client\
```

#### Make Server: 
```
go build -o .\bins\server.exe .\server\
```

#### Make Docker: 
```
docker build -t tcp-client-server .
```

#### Tag Docker: 
```
docker tag tcp-client-server princepereira/tcp-client-server
```

#### Push Docker: 
```
docker push princepereira/tcp-client-server
```

#### Pull Docker: 
```
docker pull princepereira/tcp-client-server
```

#### Run Server: 
```
.\server.exe -p 4444
```

#### Run Client: 
```
.\client.exe -i 127.0.0.1 -p 4444 -c 10 -r 10 -d 50

```


# Windows Tcp Client Server Deployment

This will enable client server test connection tool which can be run in Windows AKS.


#### 1. Create TCP Server Deployment

Create namespace demo
```
>> kubectl create namespace demo
```
```
>> kubectl create -f yamls\server-deployment.yaml
```

#### 2. Create TCP Server Service

```
>> kubectl create -f yamls\server-svc.yaml
```

#### 3. Create TCP Client POD

```
>> kubectl run -it tcp-client -n demo --image=princepereira/tcp-client-server --command -- cmd
```

If the above session is ended, resume using below command:
```
>> kubectl attach tcp-client -c tcp-client -i -t -n demo
```

#### 4. Establish cient server connection

Get the service IP
```
>> kubectl get services -n demo
```

Connect client to server
```
client >> client -i <Service IP> -p <Internal Port> -c <No: of Connections> -r <Reqs per Conn> -d <Delay in ms for each req per conn>

Eg: client >> client -i 127.0.0.1 -p 4444 -c 10 -r 10 -d 50
```