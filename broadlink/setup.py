import broadlink

devices = broadlink.discover(timeout=5)

print(devices)