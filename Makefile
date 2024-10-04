run:
	@swag init -g cmd/main.go && go run cmd/main.go

push:
	git add .
	git commit -m "update"
	git push origin main

push-main:
	git add .
	git commit -m "update"
	git push origin main

deploy:
	@./scripts/deploy.sh