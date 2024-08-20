run:
	go run cmd/main.go

push:
	git add .
	git commit -m "update"
	git push origin main

push-main:
	git add .
	git commit -m "update"
	git push origin main