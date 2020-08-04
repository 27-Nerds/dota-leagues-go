## Development

Create new db:

```bash
arangosh
```

in arango shell:

```bash
> db._createDatabase("dota_leagues", {}, [{username: "login", passwd: "pass", active: true}])
```

run project

```bash
go run dota_league
```

------

## Deployment

```bash
rsync  --cvs-exclude -av ./dota_league deploy@46.101.217.107:work
```

database dump:

```bash
arangodump --server.database dota_leagues

```

### start as a service
```bash 
sudo cp dota-league.service /lib/systemd/system/dota-league.service

sudo systemctl start dota-league
sudo systemctl enable dota-league
```
get service logs:

```bash
journalctl -u dota-league.service
```