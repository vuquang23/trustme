# Trustme

## Setting

### Dependencies

```
$ go mod tidy
```


## Run

```
$ go run cmd/app/main.go
```


## APIs
### Get current block
```
curl --location 'http://localhost:8080/api/current-block'

```

### Subscribe an address
```
curl --location 'http://localhost:8080/api/subscribe' \
--header 'Content-Type: application/json' \
--data '{
    "address": "0x1f9090aaE28b8a3dCeaDf281B0F12828e676c326"
}'

```

### Get transactions
```
curl --location 'http://localhost:8080/api/txs?address=0x1f9090aaE28b8a3dCeaDf281B0F12828e676c326'
```
