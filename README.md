# PMM Grafana backup tool

This util is responsible to back up Grafana dashboards as a JSON files and restore from the JSON files to Grafana Dashboards.

## Backup dashboards as a code

To fetch dashboards from the Grafana and save it to the local directory run:
> GRAFANA_URL=<your url> GRAFANA_TOKEN=<your token> make run_backup

* GRAFANA_URL should contain url to your Grafana
* GRAFANA_TOKEN - token of Grafana with admin scope


By default it would be backed up to the `./_OUTPUT_` directory. 
There they will be stored as JSON files.

You can run it in the docker container. For this run:
> GRAFANA_URL=<your url> GRAFANA_TOKEN=<your token> make docker_run_backup

## Restore dashboards from the code to Grafana

To restore run:
> GRAFANA_URL=<your url> GRAFANA_TOKEN=<your token> make run_restore

It expects that Grafana dashboards are stored in the `./_OUTPUT_` directory as JSON files in the project.

You can run it in the docker container. For this run:
> GRAFANA_URL=<your url> GRAFANA_TOKEN=<your token> make docker_run_backup

## Issues
* It backup/restores only dashboards. It's needed to backup folders/library panels/etc.
* Need to override backup directory to which JSON files will be stored.
