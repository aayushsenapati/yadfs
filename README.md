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
      - type: volume
        source: mount
        target: /data
    environment:
      - CONTAINER_NAME=nn-container-devel
    tty: true

  dn1-container:
    image: alpine:latest
    container_name: dn1-container-devel
    networks:
      - my-network
    volumes:
      - type: volume
        source: mount
        target: /data
    environment:
      - CONTAINER_NAME=dn1-container-devel
      - NN_CONTAINER_NAME=nn-container-devel
    tty: true

  dn2-container:
    image: alpine:latest
    container_name: dn2-container-devel
    networks:
      - my-network
    volumes:
      - type: volume
        source: mount
        target: /data
    environment:
      - CONTAINER_NAME=dn2-container-devel
      - NN_CONTAINER_NAME=nn-container-devel
      
    tty: true

  dn3-container:
    image: alpine:latest
    container_name: dn3-container-devel
    networks:
      - my-network
    volumes:
      - type: volume
        source: mount
        target: /data
    environment:
      - CONTAINER_NAME=dn3-container-devel
      - NN_CONTAINER_NAME=nn-container-devel
    tty: true

  client-container:
    image: alpine:latest
    container_name: client-container-devel
    networks:
      - my-network
    volumes:
      - type: volume
        source: mount
        target: /data
    environment:
      - CONTAINER_NAME=client-container-devel
      - NN_CONTAINER_NAME=nn-container-devel
    tty: true

volumes:
  mount:
    external: true
    name: yadfs
  
networks:
  my-network:
```
- Save this as docker-compose.yml
- 'run docker compose up'
- for all container names:
     in terminal execute 'docker exec -it {put container name here} /bin/ash'
     in the alpine ash shell run 'apk add go'
