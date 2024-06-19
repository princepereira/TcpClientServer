client: go build -o .\bins\client.exe .\client\

server: go build -o .\bins\server.exe .\server\

nodeexporter: go build -o .\bins\nodeexporter.exe .\nodeexporter\

docker: docker build -t tcp-client-server .

tag : docker tag tcp-client-server princepereira/tcp-client-server

push : docker push princepereira/tcp-client-server

pull : docker pull princepereira/tcp-client-server
