#!/bin/bash

go get "github.com/go-sql-driver/mysql"
go get "github.com/jmoiron/sqlx"
go get "github.com/zenazn/goji"

go build -o app
