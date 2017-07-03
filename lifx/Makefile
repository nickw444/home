lifx:
	cd hklifx && go build -o ../lifx

clean:
	rm lifx lifx-arm

upload:
	cd hklifx && GOOS=linux GOARCH=arm GOARM=5 go build -o ../lifx-arm
	scp lifx-arm sprinkle@sprinkle:/sprinkle/bin/lifx
