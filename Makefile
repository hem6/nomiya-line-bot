.PHONY: all

serve:
	GOPATH=./gopath dev_appserver.py --enable_host_checking=false gae/app.yaml

dep:
	dep ensure

deploy:
	GOPATH=./gopath gcloud app deploy gae/app.yaml

browse:
	gcloud app browse
