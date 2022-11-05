go build -o .\bins\client.exe .\client\
go build -o .\bins\server.exe .\server\
docker build -t tcp-client-server .
docker tag tcp-client-server princepereira/tcp-client-server
docker push princepereira/tcp-client-server