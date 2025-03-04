build:
	docker build -t registry.cn-guangzhou.aliyuncs.com/scutrobot/rm-search:latest .

push:
	docker push registry.cn-guangzhou.aliyuncs.com/scutrobot/rm-search:latest
