# モバファク社内ISUCON

手元で立てる際は次を参考にしてください

```shell
# portal

% cd portal
% mysql -uroot YOUR_PORTAL_DATABASE < db/schema.sql
% carton install

## start portal web app
% carton exec -- plackup -Ilib portal.psgi

## start bench worker
% carton exec -- perl -Ilib worker.pl


# webapp
% cd webapp
% mysql -uroot geomemo < db/geomemo.sql
% carton install
% carton exec -- plackup -Ilib app.psgi
```

