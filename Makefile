.PHONY: gogo

gogo: stop-services build distribution-app truncate-logs start-services

build:
	make -C webapp/go clean
	make -C webapp/go build

distribution-app:
	scp webapp/go/isucoin isucon-app2:/home/ubuntu/isucon8-final/webapp/go/

stop-services:
	sudo systemctl stop varnish
	sudo systemctl stop nginx
	ssh isucon-app2 sudo systemctl stop isucoin.go
	ssh isucon-app4 sudo systemctl stop mysql

start-services:
	ssh isucon-app4 sudo systemctl start mysql
	ssh isucon-app2 sudo systemctl start isucoin.go
	sudo systemctl start nginx
	sudo systemctl start varnish

truncate-logs:
	sudo truncate --size 0 /var/log/nginx/access.log
	ssh isucon-app4 sudo truncate --size 0 /var/log/mysql/slow.log

bench:
	ssh isucon-bench sh start.sh


kataribe:
	cd ../ && sudo cat /var/log/nginx/access.log | ./kataribe

log: 
	ssh isucon-app4 sudo cp /var/log/mysql/slow.log ./
	ssh isucon-app4 sudo chmod 777 ./slow.log
	scp isucon-app4:~/slow.log ../log/
	ssh isucon-app4 rm ./slow.log
