.PHONY: all

serve:
	dev_appserver.py gae/app.yaml

dep:
	dep ensure

deploy:
	GOPATH=./gopath gcloud app deploy gae/app.yaml

browse:
	gcloud app browse
