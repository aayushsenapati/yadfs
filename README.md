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
- run command: docker volume create yadfs
- clone this repository into a folder
- move contents of the above folder into this directory  `\\wsl.localhost\docker-desktop-data\data\docker\volumes` (paste this in file explorer to find it)
- Save the above code as docker-compose.ymlanywhere in your computer.(preferably one folder for your yml files)
- run command: `docker compose up` in the location of the above folder
-
  1. for all container names:
     - in terminal execute 'docker exec -it {put container name here} /bin/ash'
     - in the alpine ash shell run 'apk add go'
  2. or in vscode download the docker extension -> do **ctrl/cmd +shift+p** and type **docker attach shell**
      - attach the shells for the different containers in vs code
      - in each shell run apk add go
- in any one of the above shells run command: `chmod +x datanode.sh` and run `./datanode.sh dn1`, `./datanode.sh dn2`, `./datanode.sh dn3`
- in the respective shells cd to nn,dn1,dn2,dn3 and run command `go run .`
