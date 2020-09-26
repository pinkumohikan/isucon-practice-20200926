.PHONY: gogo

gogo: stop-services build truncate-logs start-services

build:
	echo "TODO: build"

stop-services:
	ssh isucon-app2 sudo systemctl stop isucoin

start-services:
	ssh isucon-app2 sudo systemctl start isucoin

truncate-logs:
	echo "TODO: truncate-logs"
	#sudo truncate --size 0 /var/log/nginx/access.log

bench:
	
