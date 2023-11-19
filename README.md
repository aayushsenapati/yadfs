# yadfs
Sad attempt at recreating a distributed communist file system.


## Docker Config
```yml
version: '3'
services:
  nn-container:
    image: alpine:latest
    container_name: nn-container-devel
    networks:
      - my-network
    volumes:
      - type: bind
        source: "{source directory to mount}"
        target: /data
    tty: true

  dn1-container:
    image: alpine:latest
    container_name: dn1-container-devel
    networks:
      - my-network
    volumes:
      - type: bind
        source: "{source directory to mount}"
        target: /data
    tty: true

  dn2-container:
    image: alpine:latest
    container_name: dn2-container-devel
    networks:
      - my-network
    volumes:
      - type: bind
        source: "{source directory to mount}"
        target: /data
    tty: true

  dn3-container:
    image: alpine:latest
    container_name: dn3-container-devel
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
- for all container names:
     in terminal execute 'docker exec -it {put container name here} /bin/ash'
     in the alpine ash shell run 'apk add go'
