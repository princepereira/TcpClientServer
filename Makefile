client: go build -o .\bins\client.exe .\client\

server: go build -o .\bins\server.exe .\server\

docker: docker build -t tcp-client-server .