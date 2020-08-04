## Development

install arangodb. follow this [guide](https://www.arangodb.com/docs/stable/getting-started-installation.html)

Create new db:

```bash
arangosh
```

in arango shell:

```bash
> db._createDatabase("dota_leagues", {}, [{username: "login", passwd: "pass", active: true}])
```

create config file `config.json` (you can edit it to set arangodb credentials), or just copy an example

```bash
cp config.json.example config.json
```

run project

```bash
go run dota_league
```

navigate to http://localhost:1323

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