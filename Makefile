
DEST=pi4

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build

install:
	ssh $(DEST) /usr/sbin/service amnon stop
	rsync -av covid19 static templates $(DEST):/home/pi/
	ssh $(DEST) /usr/sbin/service amnon start
