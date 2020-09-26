.PHONY: gogo

gogo: stop-services build truncate-logs start-services

build:
	echo "TODO: build"

distribution-app:
	scp webapp/go/isucoin isucon-app2:/home/ubuntu/isucon8-final/webapp/go/

stop-services:
	ssh isucon-app2 sudo systemctl stop isucoin.go
	sudo systemctl stop nginx

start-services:
	ssh isucon-app2 sudo systemctl start isucoin.go
	sudo systemctl start nginx

truncate-logs:
	echo "TODO: truncate-logs"
	sudo truncate --size 0 /var/log/nginx/access.log

bench:
	ssh isucon-bench sh start.sh


kataribe:
	cd ../ && sudo cat /var/log/nginx/access.log | ./kataribe

log: 
	ssh isucon-app4 sudo mv /var/log/mysql/slow.log ./
	ssh isucon-app4 sudo chmod 777 ./slow.log
	scp isucon-app4:~/slow.log ./
