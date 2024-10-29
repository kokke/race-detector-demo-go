all:
	@echo make all
	@go build ws

clean:
	@echo make clean
	@go clean
	@rm -f ws

test: all
	@echo make test
	@killall -q ws || true
	@go build -race ws
	@echo ""
	@echo testing thread-safe version
	@./ws -safety &
	@sleep 1
	@curl -X PUT -H "Content-Type: application/json" -d '{"key1": "value1", "key2": "value2"}' http://localhost:8080/test       > /dev/null 2>&1 
	@echo '{"k1": "v1", "k2": "v2"}' | ab -c 10 -n 1000000 -t 10 -p /dev/stdin -T application/json "http://localhost:8080/test" > /dev/null 2>&1  &
	@ab -c 10 -n 1000000 -t 10 "http://localhost:8080/test"                                                                     > /dev/null 2>&1 
	@killall -q ws ab || true
	@sleep 1
	@echo ""
	@echo testing thread-unsafe version
	@go build -race ws
	@./ws &
	@sleep 1
	@curl -X PUT -H "Content-Type: application/json" -d '{"key1": "value1", "key2": "value2"}' http://localhost:8080/test       > /dev/null 2>&1 
	@echo '{"k1": "v1", "k2": "v2"}' | ab -c 10 -n 1000000 -t 10 -p /dev/stdin -T application/json "http://localhost:8080/test" > /dev/null 2>&1  &
	@ab -c 10 -n 1000000 -t 10 "http://localhost:8080/test"                                                                     > /dev/null 2>&1 
	@killall -q ws ab || true
	@echo ""

