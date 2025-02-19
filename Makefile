docker-build:
	docker build  -t azcon1 .

docker-run:
	docker run --env-file .env -it azcon1
