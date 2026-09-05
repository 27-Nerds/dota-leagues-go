## Development

start ArangoDB with docker (no manual install needed):

```bash
docker compose up -d
```

create config file `config.json` (you can edit it to set arangodb credentials), or just copy an example

```bash
cp config.json.example config.json
```

auth is disabled on the test instance, so any credentials work — the example values (`login`/`pass`) are fine as-is. the database itself is created automatically on first run.

run project

```bash
go run .
```

navigate to http://localhost:1323

stop the db when you are done: `docker compose down`

------

## Deployment

copy files to the server:

```bash
rsync  --cvs-exclude -av ./dota_league deploy@46.101.217.107:work
```

build executable

```bash
go build dota_league
```

run

```
./dota_league
```

or start as a service

```bash 
sudo cp dota-league.service /lib/systemd/system/dota-league.service

sudo systemctl start dota-league
sudo systemctl enable dota-league
```

----

### help tasks

get service logs:

```bash
journalctl -u dota-league.service
```

arango database dump:

```bash
arangodump --server.database dota_leagues

```