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

For other dockerfiles, copy the corresponding docker images to root directory, then
```
cp -Force .\dockerfiles\Dockerfile.WindowsServer2022 Dockerfile
docker build -t tcp-client-server:WindowsServer2022 .

or

cp -Force .\dockerfiles\Dockerfile.Windows2022 Dockerfile
docker build -t tcp-client-server:Windows2022 .
```

#### Tag Docker: 
```
docker tag tcp-client-server princepereira/tcp-client-server
```
For other dockerfiles
```
docker tag tcp-client-server:Windows2022 princepereira/tcp-client-server:Windows2022
docker tag tcp-client-server:WindowsServer2022 princepereira/tcp-client-server:WindowsServer2022
```

#### Push Docker: 
```
docker push princepereira/tcp-client-server
```
For other dockerfiles
```
docker push princepereira/tcp-client-server:Windows2022
docker push princepereira/tcp-client-server:WindowsServer2022
```

#### Pull Docker: 
```
docker pull princepereira/tcp-client-server
```
For other dockerfiles
```
docker pull princepereira/tcp-client-server:Windows2022
docker pull princepereira/tcp-client-server:WindowsServer2022
```

#### Run Server: 
```
.\server.exe -p 4444
```

#### Server Help: 
```
PS> .\server.exe -h

#==============================#

Format : .\server.exe -p <Port>

Eg : .\server.exe -p 4444

Parameters (Optional, Mandatory*):

   -p   : (*) Port number of the server
   -pr  :     Proto used. Options: TCP/UDP/All. Default: TCP
   -pw  :     Timeout for prestop action in seconds.

#==============================#
```

#### Run Client: 
```
PS> .\client.exe -i 127.0.0.1 -p 4444 -c 10 -r 10 -d 50

```

#### Client Help: 
```

PS> .\client.exe -h

#==============================#

Format : .\client.exe -i <IP> -p <Port> -c <Number of Connections> -r <Number of Requests/Connection> -d <Delay (in ms) between each request>

Eg : .\client.exe -i 127.0.0.1 -p 4444 -c 1 -r 10000 -d 1

Parameters (Optional, Mandatory*):

   -i   : (*) IP Address of the server
   -p   : (*) Port number of the server
   -c   : (*) Number of clients/threads/connections
   -r   : (*) Number of requests per connection
   -d   : (*) Delay/Sleep/Time between each request for a single connection (in milliseconds)
   -it  :     Number of iterations. Default: 1
   -pr  :     Proto used. Options: TCP/UDP. Default: TCP
   -dka :     Disable KeepAlive. Options: True/False. Default: False
   -tka :     KeepAlive Time in milliseconds. Default: 15 seconds

#==============================#

```


# Windows Tcp Client Server Deployment

This will enable client server test connection tool which can be run in Windows AKS.


#### 1. Create TCP Server Deployment

Create namespace demo
```
kubectl create namespace demo
```
```
kubectl create -f yamls\server-deployment.yaml
```

#### 2. Create TCP Server Service

```
kubectl create -f yamls\server-svc.yaml
```

#### 3. Create TCP Client POD

```
kubectl run -it tcp-client -n demo --image=princepereira/tcp-client-server --command -- cmd
```

If the above session is ended, resume using below command:
```
kubectl attach tcp-client -c tcp-client -i -t -n demo
```

#### 4. Establish cient server connection

Get the service IP
```
kubectl get services -n demo
```

Connect client to server
```
client -i <Service IP> -p <Internal Port> -c <No: of Connections> -r <Reqs per Conn> -d <Delay in ms for each req per conn>
```
```
Eg: client >> client -i 127.0.0.1 -p 4444 -c 10 -r 10 -d 50
```
Or you can use the domain name also to make the connection
```
Domain name: <service-name>.<namespace>.svc.cluster.local
```
```
Eg: tcp-server-.demo.svc.cluster.local
client >> client -i tcp-server.demo.svc.cluster.local -p 4444 -c 10 -r 10 -d 50
```