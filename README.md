# yadfs
Sad attempt at recreating a distributed file system.


## Docker Config
```yml
version: '3'
services:
  server-container:
    image: alpine:latest
    container_name: server-container-devel
    networks:
      - my-network
    volumes:
      - type: bind
        source: "{source directory to mount}"
        target: /data
    tty: true

  client-container:
    image: alpine:latest
    container_name: client-container-devel
    networks:
      - my-network
    volumes:
      - type: bind
        source: "{source directory to mount}"
        target: /data
    tty: true
networks:
  my-network:
```
- Save this as docker-compose.yml
- 'run docker compose up'
- in terminal execute 'docker exec -it /bin/ash'
- in the alpine ash shell run 'apk add go'
