# go-crud-api

[![Go](https://github.com/junlapong/go-crud-api/workflows/Go/badge.svg?branch=master)](https://github.com/junlapong/go-crud-api/actions)

A Go port of the full [php-crud-api](https://github.com/mevdschee/php-crud-api) project (single file REST API)

NB: WORK IN PROGRESS - NOT FINISHED YET!

```
sqlite3 db/crud.sqlite

SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.
sqlite>
```

```sql
CREATE TABLE users (
    id integer primary key AUTOINCREMENT,
    username text not null UNIQUE,
    email text not null UNIQUE,
    firstname text not null,
    lastname text not null,
    created_date datetime not null default current_timestamp,
    updated_date datetime
);
```