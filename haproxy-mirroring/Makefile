.PHONY: haproxy
haproxy:
	BACKEND_SOCKET_ADDRESS=/tmp/backend.sock SPOA_SOCKET_ADDRESS=/tmp/spoa.sock haproxy -W -db -f haproxy.cfg

.PHONY: backend
backend:
	SOCKET_ADDRESS=/tmp/backend.sock go run backend/main.go

.PHONY: spoa
spoa:
	SOCKET_ADDRESS=/tmp/spoa.sock DEST_ADDRESS=localhost:3000 go run spoa/main.go

.PHONY: dest
dest:
	ADDRESS=localhost:3000 go run dest/main.go
