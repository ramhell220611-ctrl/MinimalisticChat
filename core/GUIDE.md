# Full Guide 🏴‍☠️

## First way
```bash
# Copy gh files
git clone https://github.com/FlowRamAlltimes/MinimalisticChat
# Go to core dir then create certs
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=MyLocalCA"
sudo nano server.conf
# Enter your config
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extfile server.conf -extensions v3_ext
```
### After that, just make binary & run it
```bash
./server
```

## Second way
# Go to releases by this [link](https://github.com/FlowRamAlltimes/MinimalisticChat/releases)
```bash
# Download last version & run it
./CrystalCoreTlsEncrypted-v1.8
```
